package reductgo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
)

const (
	recordBatchHeaderPrefix  = "x-reduct-"
	recordBatchErrorPrefix   = "x-reduct-error-"
	recordBatchEntriesHeader = "x-reduct-entries"
	recordBatchStartTSHeader = "x-reduct-start-ts"
	recordBatchLabelsHeader  = "x-reduct-labels"
)

// RecordBatchErrorMap represents errors per entry and timestamp.
type RecordBatchErrorMap map[string]ErrorMap

type recordBatchKey struct {
	entry string
	ts    int64
}

type recordBatchRecord struct {
	entry       string
	timestamp   int64
	data        []byte
	contentType string
	labels      LabelMap
}

type recordBatchMeta struct {
	contentType string
	labels      map[string]string
}

// RecordBatch is a batch of records across multiple entries (Batch Protocol v2).
type RecordBatch struct {
	bucketName string
	httpClient httpclient.HTTPClient
	batchType  BatchType
	records    map[recordBatchKey]*recordBatchRecord
	totalSize  int64
	lastAccess time.Time
	mu         sync.Mutex
}

// newRecordBatch creates a new record batch.
func newRecordBatch(bucket string, client httpclient.HTTPClient, batchType BatchType) *RecordBatch {
	return &RecordBatch{
		bucketName: bucket,
		httpClient: client,
		batchType:  batchType,
		records:    make(map[recordBatchKey]*recordBatchRecord),
		totalSize:  0,
		lastAccess: time.Time{},
		mu:         sync.Mutex{},
	}
}

// Add adds a record to the batch with entry name.
func (b *RecordBatch) Add(entry string, ts int64, data []byte, contentType string, labels LabelMap) {
	if entry == "" {
		return
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if labels == nil {
		labels = LabelMap{}
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.totalSize += int64(len(data))
	b.lastAccess = time.Now().UTC()
	key := recordBatchKey{entry: entry, ts: ts}
	b.records[key] = &recordBatchRecord{
		entry:       entry,
		timestamp:   ts,
		data:        data,
		contentType: contentType,
		labels:      labels,
	}
}

// Send sends the batch to the server using Batch Protocol v2.
func (b *RecordBatch) Send(ctx context.Context) (RecordBatchErrorMap, error) {
	b.mu.Lock()
	items := make([]*recordBatchRecord, 0, len(b.records))
	for _, record := range b.records {
		items = append(items, record)
	}
	b.mu.Unlock()

	switch b.batchType {
	case BatchWrite:
		headers, body, entries, startTS, contentLength, err := buildRecordBatchWriteRequest(items)
		if err != nil {
			return nil, err
		}

		path := fmt.Sprintf("/io/%s/write", b.bucketName)
		req, err := b.httpClient.NewRequestWithContext(ctx, http.MethodPost, path, body)
		if err != nil {
			return nil, err
		}
		req.Header = headers
		req.Header.Set("Content-Type", "application/octet-stream")
		req.ContentLength = contentLength

		resp, err := b.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		return parseRecordBatchErrors(resp.Header, entries, startTS)
	case BatchUpdate:
		headers, entries, startTS, err := buildRecordBatchUpdateRequest(items)
		if err != nil {
			return nil, err
		}

		path := fmt.Sprintf("/io/%s/update", b.bucketName)
		req, err := b.httpClient.NewRequestWithContext(ctx, http.MethodPatch, path, nil)
		if err != nil {
			return nil, err
		}
		req.Header = headers
		req.ContentLength = 0

		resp, err := b.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		return parseRecordBatchErrors(resp.Header, entries, startTS)
	default:
		return nil, fmt.Errorf("invalid batch type")
	}
}

// Size returns the total size of the batch.
func (b *RecordBatch) Size() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.totalSize
}

// LastAccessTime returns the last access time of the batch.
func (b *RecordBatch) LastAccessTime() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.lastAccess
}

// RecordCount returns the number of records in the batch.
func (b *RecordBatch) RecordCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.records)
}

// Clear clears the batch.
func (b *RecordBatch) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.records = make(map[recordBatchKey]*recordBatchRecord)
	b.totalSize = 0
	b.lastAccess = time.Time{}
}

type indexedRecord struct {
	entryIndex int
	timestamp  int64
	record     *recordBatchRecord
}

func buildRecordBatchWriteRequest(records []*recordBatchRecord) (http.Header, io.Reader, []string, int64, int64, error) {
	headers := http.Header{}
	if len(records) == 0 {
		headers.Set(recordBatchEntriesHeader, "")
		headers.Set(recordBatchStartTSHeader, "0")
		headers.Set("Content-Type", "application/octet-stream")
		headers.Set("Content-Length", "0")
		return headers, bytes.NewReader(nil), []string{}, 0, 0, nil
	}

	items := append([]*recordBatchRecord(nil), records...)
	slices.SortFunc(items, func(a, b *recordBatchRecord) int {
		if a.entry != b.entry {
			if a.entry < b.entry {
				return -1
			}
			return 1
		}
		switch {
		case a.timestamp < b.timestamp:
			return -1
		case a.timestamp > b.timestamp:
			return 1
		default:
			return 0
		}
	})

	entries := make([]string, 0)
	entryIndexLookup := map[string]int{}
	indexed := make([]indexedRecord, 0, len(items))

	startTS := items[0].timestamp
	for _, record := range items {
		if record.timestamp < startTS {
			startTS = record.timestamp
		}
	}

	for _, record := range items {
		entryIndex, ok := entryIndexLookup[record.entry]
		if !ok {
			entryIndex = len(entries)
			entries = append(entries, record.entry)
			entryIndexLookup[record.entry] = entryIndex
		}
		indexed = append(indexed, indexedRecord{
			entryIndex: entryIndex,
			timestamp:  record.timestamp,
			record:     record,
		})
	}

	headers.Set(recordBatchEntriesHeader, encodeHeaderList(entries))
	headers.Set(recordBatchStartTSHeader, strconv.FormatInt(startTS, 10))
	headers.Set("Content-Type", "application/octet-stream")

	labelIndex := map[string]int{}
	labelNames := make([]string, 0)
	lastMeta := map[int]recordBatchMeta{}

	var contentLength int64
	chunks := make([]io.Reader, 0, len(indexed))

	for _, item := range indexed {
		record := item.record
		contentLength += int64(len(record.data))
		chunks = append(chunks, bytes.NewReader(record.data))

		delta := record.timestamp - startTS
		contentType := record.contentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		currentLabels := normalizeLabels(record.labels)
		prev := lastMeta[item.entryIndex]
		var prevPtr *recordBatchMeta
		if prev.labels != nil || prev.contentType != "" {
			prevPtr = &prev
		}
		labelDelta := buildLabelDelta(currentLabels, prevPtr, labelIndex, &labelNames)
		hasLabels := labelDelta != ""

		contentTypePart := contentType
		if prevPtr != nil && prevPtr.contentType == contentType {
			contentTypePart = ""
		}

		parts := []string{strconv.Itoa(len(record.data))}
		if contentTypePart != "" || hasLabels {
			parts = append(parts, contentTypePart)
		}
		if hasLabels {
			parts = append(parts, labelDelta)
		}

		headerName := fmt.Sprintf("%s%d-%d", recordBatchHeaderPrefix, item.entryIndex, delta)
		headers.Set(headerName, strings.Join(parts, ","))

		lastMeta[item.entryIndex] = recordBatchMeta{
			contentType: contentType,
			labels:      currentLabels,
		}
	}

	if len(labelNames) > 0 {
		headers.Set(recordBatchLabelsHeader, encodeHeaderList(labelNames))
	}

	headers.Set("Content-Length", strconv.FormatInt(contentLength, 10))

	body := io.MultiReader(chunks...)
	return headers, body, entries, startTS, contentLength, nil
}

func buildRecordBatchUpdateRequest(records []*recordBatchRecord) (http.Header, []string, int64, error) {
	headers := http.Header{}
	if len(records) == 0 {
		headers.Set(recordBatchEntriesHeader, "")
		headers.Set(recordBatchStartTSHeader, "0")
		headers.Set("Content-Length", "0")
		return headers, []string{}, 0, nil
	}

	items := append([]*recordBatchRecord(nil), records...)
	slices.SortFunc(items, func(a, b *recordBatchRecord) int {
		if a.entry != b.entry {
			if a.entry < b.entry {
				return -1
			}
			return 1
		}
		switch {
		case a.timestamp < b.timestamp:
			return -1
		case a.timestamp > b.timestamp:
			return 1
		default:
			return 0
		}
	})

	entries := make([]string, 0)
	entryIndexLookup := map[string]int{}

	startTS := items[0].timestamp
	for _, record := range items {
		if record.timestamp < startTS {
			startTS = record.timestamp
		}
	}

	labelIndex := map[string]int{}
	labelNames := make([]string, 0)

	for _, record := range items {
		entryIndex, ok := entryIndexLookup[record.entry]
		if !ok {
			entryIndex = len(entries)
			entries = append(entries, record.entry)
			entryIndexLookup[record.entry] = entryIndex
		}

		delta := record.timestamp - startTS
		labelDelta := buildUpdateLabelDelta(record.labels, labelIndex, &labelNames)

		headerName := fmt.Sprintf("%s%d-%d", recordBatchHeaderPrefix, entryIndex, delta)
		if labelDelta == "" {
			headers.Set(headerName, "0")
		} else {
			headers.Set(headerName, "0,,"+labelDelta)
		}
	}

	headers.Set(recordBatchEntriesHeader, encodeHeaderList(entries))
	headers.Set(recordBatchStartTSHeader, strconv.FormatInt(startTS, 10))
	headers.Set("Content-Length", "0")

	if len(labelNames) > 0 {
		headers.Set(recordBatchLabelsHeader, encodeHeaderList(labelNames))
	}

	return headers, entries, startTS, nil
}

func parseRecordBatchErrors(headers http.Header, entries []string, startTS int64) (RecordBatchErrorMap, error) {
	errs := RecordBatchErrorMap{}
	for key, values := range headers {
		name := strings.ToLower(key)
		if !strings.HasPrefix(name, recordBatchErrorPrefix) {
			continue
		}

		// Header format: x-reduct-error-<entryIndex>-<delta>=<status>,<message>
		suffix := name[len(recordBatchErrorPrefix):]
		lastDash := strings.LastIndex(suffix, "-")
		if lastDash == -1 {
			return nil, fmt.Errorf("invalid error header '%s'", key)
		}

		entryRaw := suffix[:lastDash]
		deltaRaw := suffix[lastDash+1:]

		// Parse entry index and timestamp delta.
		entryIndex, err := strconv.Atoi(entryRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid error header '%s'", key)
		}
		delta, err := strconv.ParseInt(deltaRaw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid error header '%s'", key)
		}
		if entryIndex < 0 || entryIndex >= len(entries) {
			return nil, fmt.Errorf("invalid error header '%s'", key)
		}
		if len(values) == 0 {
			continue
		}

		// Split status and message from the header value.
		parts := strings.SplitN(values[0], ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid error header '%s'", key)
		}
		code, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid error header '%s'", key)
		}

		// Reconstruct full timestamp and map it to the entry name.
		ts := startTS + delta
		entryName := entries[entryIndex]
		entryErrors := errs[entryName]
		if entryErrors == nil {
			entryErrors = ErrorMap{}
			errs[entryName] = entryErrors
		}

		// Store the parsed error for this entry/timestamp.
		entryErrors[ts] = model.APIError{
			Status:  code,
			Message: parts[1],
		}
	}
	return errs, nil
}

func normalizeLabels(labels LabelMap) map[string]string {
	out := map[string]string{}
	for k, v := range labels {
		out[k] = fmt.Sprint(v)
	}
	return out
}

type labelDeltaOp struct {
	idx   int
	value *string
}

func buildLabelDelta(labels map[string]string, previous *recordBatchMeta, labelIndex map[string]int, labelNames *[]string) string {
	ops := make([]labelDeltaOp, 0)
	ensureLabel := func(name string) int {
		if idx, ok := labelIndex[name]; ok {
			return idx
		}
		idx := len(*labelNames)
		labelIndex[name] = idx
		*labelNames = append(*labelNames, name)
		return idx
	}

	if previous == nil {
		keys := make([]string, 0, len(labels))
		for k := range labels {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, key := range keys {
			idx := ensureLabel(key)
			val := formatLabelValue(labels[key])
			ops = append(ops, labelDeltaOp{idx: idx, value: &val})
		}
	} else {
		keys := map[string]struct{}{}
		for k := range previous.labels {
			keys[k] = struct{}{}
		}
		for k := range labels {
			keys[k] = struct{}{}
		}
		sorted := make([]string, 0, len(keys))
		for k := range keys {
			sorted = append(sorted, k)
		}
		slices.Sort(sorted)
		for _, key := range sorted {
			prevVal, okPrev := previous.labels[key]
			currVal, okCurr := labels[key]
			if okPrev && okCurr && prevVal == currVal {
				continue
			}
			idx := ensureLabel(key)
			if !okCurr {
				ops = append(ops, labelDeltaOp{idx: idx, value: nil})
				continue
			}
			val := formatLabelValue(currVal)
			ops = append(ops, labelDeltaOp{idx: idx, value: &val})
		}
	}

	slices.SortFunc(ops, func(a, b labelDeltaOp) int {
		return a.idx - b.idx
	})

	parts := make([]string, 0, len(ops))
	for _, op := range ops {
		if op.value == nil {
			parts = append(parts, fmt.Sprintf("%d=", op.idx))
		} else {
			parts = append(parts, fmt.Sprintf("%d=%s", op.idx, *op.value))
		}
	}
	return strings.Join(parts, ",")
}

func buildUpdateLabelDelta(labels LabelMap, labelIndex map[string]int, labelNames *[]string) string {
	if len(labels) == 0 {
		return ""
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	ensureLabel := func(name string) int {
		if idx, ok := labelIndex[name]; ok {
			return idx
		}
		idx := len(*labelNames)
		labelIndex[name] = idx
		*labelNames = append(*labelNames, name)
		return idx
	}

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		idx := ensureLabel(key)
		val := labels[key]
		if val == nil {
			parts = append(parts, fmt.Sprintf("%d=", idx))
			continue
		}

		valStr := fmt.Sprint(val)
		if valStr == "" {
			parts = append(parts, fmt.Sprintf("%d=", idx))
			continue
		}
		parts = append(parts, fmt.Sprintf("%d=%s", idx, formatLabelValue(valStr)))
	}

	return strings.Join(parts, ",")
}

func formatLabelValue(value string) string {
	if strings.Contains(value, ",") {
		return fmt.Sprintf("\"%s\"", value)
	}
	return value
}

func encodeHeaderList(values []string) string {
	encoded := make([]string, 0, len(values))
	for _, value := range values {
		encoded = append(encoded, encodeHeaderComponent(value))
	}
	return strings.Join(encoded, ",")
}

func encodeHeaderComponent(value string) string {
	var builder strings.Builder
	for _, b := range []byte(value) {
		if isTchar(b) {
			builder.WriteByte(b)
		} else {
			builder.WriteString(fmt.Sprintf("%%%02X", b))
		}
	}
	return builder.String()
}

func isTchar(b byte) bool {
	switch {
	case b >= '0' && b <= '9':
		return true
	case b >= 'A' && b <= 'Z':
		return true
	case b >= 'a' && b <= 'z':
		return true
	}
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}
