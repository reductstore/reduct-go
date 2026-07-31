package batch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
)

const (
	headerPrefix      = "x-reduct-"
	errorHeaderPrefix = "x-reduct-error-"
	entriesHeader     = "x-reduct-entries"
	startTSHeader     = "x-reduct-start-ts"
	labelsHeader      = "x-reduct-labels"
	lastHeader        = "x-reduct-last"
)

type recordHeader struct {
	entryIndex int
	delta      int64
	rawValue   string
}

type parsedHeader struct {
	contentLength int64
	contentType   string
	labels        map[string]any
}

// FetchAndParseV2 reads records for a query ID using Batch Protocol v2.
// The first batch is fetched synchronously so that hard errors (e.g. invalid SQL)
// are returned immediately as a normal error. Any error occurring in subsequent
// batches is sent to the returned error channel, which is closed when streaming ends.
func FetchAndParseV2(ctx context.Context, client httpclient.HTTPClient, bucketName string, id int64, continueQuery bool, pollInterval time.Duration, head bool) (<-chan *Record, <-chan error, error) { //nolint:gocritic // directional channels cannot be named returns
	firstBatch, err := readBatchedRecordsV2(ctx, client, bucketName, id, head)
	if err != nil {
		var apiErr model.APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusNoContent {
			if !continueQuery {
				ch := make(chan *Record)
				close(ch)
				errCh := make(chan error)
				close(errCh)
				return ch, errCh, nil
			}
			firstBatch = nil
		} else {
			return nil, nil, err
		}
	}

	records := make(chan *Record, 100)
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		defer close(records)

		for _, rec := range firstBatch {
			select {
			case <-ctx.Done():
				return
			case records <- rec:
				if rec.Last {
					return
				}
			}
		}

		for {
			batch, err := readBatchedRecordsV2(ctx, client, bucketName, id, head)
			if err != nil {
				var apiErr model.APIError
				if errors.As(err, &apiErr) && apiErr.Status == http.StatusNoContent {
					if continueQuery {
						select {
						case <-ctx.Done():
							return
						case <-time.After(pollInterval):
							continue
						}
					}
					return
				}
				errCh <- err
				return
			}

			if len(batch) == 0 {
				if continueQuery {
					select {
					case <-ctx.Done():
						return
					case <-time.After(pollInterval):
						continue
					}
				}
				return
			}

			for _, rec := range batch {
				select {
				case <-ctx.Done():
					return
				case records <- rec:
					if rec.Last {
						return
					}
				}
			}
		}
	}()

	return records, errCh, nil
}

// readBatchedRecordsV2 fetches one batch of records for a query using Batch
// Protocol v2. Like the v1 reader it buffers every record but the last in a
// single allocation and streams the last one straight off the response body.
func readBatchedRecordsV2(ctx context.Context, client httpclient.HTTPClient, bucketName string, id int64, head bool) ([]*Record, error) {
	path := fmt.Sprintf("/io/%s/read", bucketName)

	var req *http.Request
	var err error
	if head {
		req, err = client.NewRequestWithContext(ctx, http.MethodHead, path, nil)
	} else {
		req, err = client.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err == nil {
			req.Header.Set("Accept", "application/octet-stream")
		}
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-reduct-query-id", strconv.FormatInt(id, 10))

	// The response body is handed to the last record of the batch, which the
	// caller drains; every other exit path closes it explicitly.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNoContent {
		errorMessage := resp.Header.Get("x-reduct-error")
		if errorMessage == "" {
			errorMessage = "No content"
		}
		closeBody(resp)
		return nil, model.APIError{Status: http.StatusNoContent, Message: errorMessage}
	}

	entriesHeaderValue := resp.Header.Get(entriesHeader)
	if entriesHeaderValue == "" {
		closeBody(resp)
		return nil, fmt.Errorf("%s header is required", entriesHeader)
	}
	startHeaderValue := resp.Header.Get(startTSHeader)
	if startHeaderValue == "" {
		closeBody(resp)
		return nil, fmt.Errorf("%s header is required", startTSHeader)
	}

	entries, err := parseHeaderList(entriesHeaderValue)
	if err != nil {
		closeBody(resp)
		return nil, err
	}
	startTS, err := strconv.ParseInt(startHeaderValue, 10, 64)
	if err != nil {
		closeBody(resp)
		return nil, fmt.Errorf("invalid %s header: %w", startTSHeader, err)
	}

	var labelNames []string
	if labelsHeaderValue := resp.Header.Get(labelsHeader); labelsHeaderValue != "" {
		labelNames, err = parseHeaderList(labelsHeaderValue)
		if err != nil {
			closeBody(resp)
			return nil, err
		}
	}

	recordHeaders := parseRecordHeaders(resp.Header)
	// An empty batch is valid.
	if len(recordHeaders) == 0 {
		closeBody(resp)
		return nil, nil
	}

	if err = ctx.Err(); err != nil {
		closeBody(resp)
		return nil, err
	}

	total := len(recordHeaders)
	lastHeaderValue := strings.EqualFold(resp.Header.Get(lastHeader), "true")

	// Parse every record header first so the buffered payload size is known
	// before the body is touched.
	parsed := make([]parsedHeader, total)
	lastHeaderPerEntry := map[int]parsedHeader{}
	var bufferedSize int64
	for i, header := range recordHeaders {
		if header.entryIndex < 0 || header.entryIndex >= len(entries) {
			closeBody(resp)
			return nil, fmt.Errorf("invalid header '%s%d-%d': entry index out of range", headerPrefix, header.entryIndex, header.delta)
		}

		prev, ok := lastHeaderPerEntry[header.entryIndex]
		var prevPtr *parsedHeader
		if ok {
			prevPtr = &prev
		}
		entryHeader, parseErr := parseRecordHeader(header.rawValue, prevPtr, labelNames)
		if parseErr != nil {
			closeBody(resp)
			return nil, parseErr
		}
		lastHeaderPerEntry[header.entryIndex] = entryHeader

		parsed[i] = entryHeader
		if i < total-1 && entryHeader.contentLength > 0 {
			bufferedSize += entryHeader.contentLength
		}
	}

	// Every record but the last is read in one pass into one buffer.
	var buffered []byte
	if !head && bufferedSize > 0 {
		buffered = make([]byte, bufferedSize)
		if _, err = io.ReadFull(resp.Body, buffered); err != nil {
			closeBody(resp)
			return nil, err
		}
	}

	records := make([]*Record, total)
	backing := make([]Record, total)
	readers := make([]sliceReader, total)

	var offset int64
	for i, header := range recordHeaders {
		isLastInBatch := i == total-1

		var body io.ReadCloser
		switch {
		case head:
			body = emptyBody{}
		case isLastInBatch:
			body = resp.Body
		default:
			readers[i].data = buffered[offset : offset+parsed[i].contentLength]
			offset += parsed[i].contentLength
			body = &readers[i]
		}

		backing[i] = Record{
			Entry:       entries[header.entryIndex],
			Time:        startTS + header.delta,
			Size:        parsed[i].contentLength,
			Last:        lastHeaderValue && isLastInBatch,
			LastInBatch: isLastInBatch,
			Body:        body,
			Labels:      parsed[i].labels,
			ContentType: parsed[i].contentType,
		}
		records[i] = &backing[i]
	}

	if head {
		// Nothing streams from a HEAD response, so release the connection now.
		closeBody(resp)
	}

	return records, nil
}

func parseRecordHeaders(headers http.Header) []recordHeader {
	parsed := make([]recordHeader, 0, len(headers))
	for key, values := range headers {
		// Compare case-insensitively against the canonical form rather than
		// lower-casing every header name, which would allocate per record.
		if len(key) <= len(headerPrefix) || !strings.EqualFold(key[:len(headerPrefix)], headerPrefix) {
			continue
		}
		if strings.EqualFold(key, entriesHeader) ||
			strings.EqualFold(key, startTSHeader) ||
			strings.EqualFold(key, labelsHeader) ||
			strings.EqualFold(key, lastHeader) ||
			(len(key) >= len(errorHeaderPrefix) && strings.EqualFold(key[:len(errorHeaderPrefix)], errorHeaderPrefix)) {
			continue
		}

		suffix := key[len(headerPrefix):]
		lastDash := strings.LastIndex(suffix, "-")
		if lastDash == -1 {
			continue
		}
		entryRaw := suffix[:lastDash]
		deltaRaw := suffix[lastDash+1:]
		if entryRaw == "" || deltaRaw == "" {
			continue
		}
		entryIndex, err := strconv.Atoi(entryRaw)
		if err != nil {
			continue
		}
		delta, err := strconv.ParseInt(deltaRaw, 10, 64)
		if err != nil {
			continue
		}
		rawValue := ""
		if len(values) > 0 {
			rawValue = values[0]
		}
		parsed = append(parsed, recordHeader{
			entryIndex: entryIndex,
			delta:      delta,
			rawValue:   rawValue,
		})
	}

	slices.SortFunc(parsed, func(a, b recordHeader) int {
		if a.entryIndex != b.entryIndex {
			return a.entryIndex - b.entryIndex
		}
		switch {
		case a.delta < b.delta:
			return -1
		case a.delta > b.delta:
			return 1
		default:
			return 0
		}
	})

	return parsed
}

func parseRecordHeader(raw string, previous *parsedHeader, labelNames []string) (parsedHeader, error) {
	commaIndex := strings.Index(raw, ",")
	contentLengthRaw := raw
	if commaIndex != -1 {
		contentLengthRaw = raw[:commaIndex]
	}
	contentLength, err := strconv.ParseInt(strings.TrimSpace(contentLengthRaw), 10, 64)
	if err != nil {
		return parsedHeader{}, fmt.Errorf("invalid content length: %w", err)
	}

	if commaIndex == -1 {
		if previous == nil {
			return parsedHeader{}, fmt.Errorf("content-type and labels must be provided for the first record of an entry")
		}
		labelsCopy := map[string]any{}
		for k, v := range previous.labels {
			labelsCopy[k] = v
		}
		return parsedHeader{
			contentLength: contentLength,
			contentType:   previous.contentType,
			labels:        labelsCopy,
		}, nil
	}

	rest := raw[commaIndex+1:]
	nextComma := strings.Index(rest, ",")
	contentTypeRaw := rest
	var labelsRaw *string
	if nextComma != -1 {
		contentTypeRaw = rest[:nextComma]
		labelsValue := rest[nextComma+1:]
		labelsRaw = &labelsValue
	}

	contentType := strings.TrimSpace(contentTypeRaw)
	if contentType == "" {
		if previous != nil && previous.contentType != "" {
			contentType = previous.contentType
		} else {
			contentType = "application/octet-stream"
		}
	}

	var labels map[string]any
	if labelsRaw == nil {
		labels = map[string]any{}
		if previous != nil {
			for k, v := range previous.labels {
				labels[k] = v
			}
		}
	} else {
		base := map[string]any{}
		if previous != nil {
			for k, v := range previous.labels {
				base[k] = v
			}
		}
		parsedLabels, parseErr := applyLabelDelta(*labelsRaw, base, labelNames)
		if parseErr != nil {
			return parsedHeader{}, parseErr
		}
		labels = parsedLabels
	}

	return parsedHeader{
		contentLength: contentLength,
		contentType:   contentType,
		labels:        labels,
	}, nil
}

type labelDeltaOp struct {
	key   string
	value *string
}

func applyLabelDelta(rawLabels string, base map[string]any, labelNames []string) (map[string]any, error) {
	labels := map[string]any{}
	for k, v := range base {
		labels[k] = v
	}
	ops, err := parseLabelDeltaOps(rawLabels, labelNames)
	if err != nil {
		return nil, err
	}
	for _, op := range ops {
		if op.value == nil {
			delete(labels, op.key)
		} else {
			labels[op.key] = *op.value
		}
	}
	return labels, nil
}

func parseLabelDeltaOps(rawLabels string, labelNames []string) ([]labelDeltaOp, error) {
	ops := make([]labelDeltaOp, 0)
	rest := strings.TrimSpace(rawLabels)
	if rest == "" {
		return ops, nil
	}

	for rest != "" {
		eqIndex := strings.Index(rest, "=")
		if eqIndex == -1 {
			return nil, fmt.Errorf("invalid batched header")
		}

		rawKey := strings.TrimSpace(rest[:eqIndex])
		key, err := resolveLabelName(rawKey, labelNames)
		if err != nil {
			return nil, err
		}

		valuePart := rest[eqIndex+1:]
		var value string
		var nextRest string

		if strings.HasPrefix(valuePart, "\"") {
			valuePart = valuePart[1:]
			endQuote := strings.Index(valuePart, "\"")
			if endQuote == -1 {
				return nil, fmt.Errorf("invalid batched header")
			}
			value = strings.TrimSpace(valuePart[:endQuote])
			nextRest = strings.TrimSpace(valuePart[endQuote+1:])
			if strings.HasPrefix(nextRest, ",") {
				nextRest = strings.TrimSpace(nextRest[1:])
			}
		} else {
			nextComma := strings.Index(valuePart, ",")
			if nextComma == -1 {
				value = strings.TrimSpace(valuePart)
				nextRest = ""
			} else {
				value = strings.TrimSpace(valuePart[:nextComma])
				nextRest = strings.TrimSpace(valuePart[nextComma+1:])
			}
		}

		if value == "" {
			ops = append(ops, labelDeltaOp{key: key, value: nil})
		} else {
			val := value
			ops = append(ops, labelDeltaOp{key: key, value: &val})
		}

		if nextRest == "" {
			return ops, nil
		}
		rest = nextRest
	}

	return ops, nil
}

func resolveLabelName(raw string, labelNames []string) (string, error) {
	if len(labelNames) > 0 {
		if _, err := strconv.Atoi(raw); err == nil {
			idx, _ := strconv.Atoi(raw) //nolint:errcheck // already checked
			if idx < 0 || idx >= len(labelNames) {
				return "", fmt.Errorf("label index '%s' is out of range", raw)
			}
			name := labelNames[idx]
			if name == "" {
				return "", fmt.Errorf("label index '%s' is out of range", raw)
			}
			return name, nil
		}
	}

	if strings.HasPrefix(raw, "@") {
		return "", fmt.Errorf("label names must not start with '@': reserved for computed labels")
	}

	return raw, nil
}

func parseHeaderList(header string) ([]string, error) {
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		// Empty batch has no entries
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	out := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("invalid entries/labels header")
		}
		decoded, err := url.PathUnescape(item)
		if err != nil {
			return nil, err
		}
		out = append(out, decoded)
	}
	return out, nil
}
