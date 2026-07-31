package batch

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
)

type Record struct {
	Entry       string
	Time        int64
	Size        int64
	Last        bool
	LastInBatch bool
	Body        io.ReadCloser
	Labels      map[string]any
	ContentType string
}

// FetchAndParse reads records for a query ID using Batch Protocol v1.
// The first batch is fetched synchronously so that hard errors are returned
// immediately as a normal error. Any error in subsequent batches is sent to
// the returned error channel, which is closed when streaming ends.
func FetchAndParse(ctx context.Context, client httpclient.HTTPClient, bucketName, entry string, id int64, continueQuery bool, pollInterval time.Duration, head bool) (<-chan *Record, <-chan error, error) { //nolint:gocritic // directional channels cannot be named returns
	firstBatch, err := readBatchedRecords(ctx, client, bucketName, entry, id, head)
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
			batch, err := readBatchedRecords(ctx, client, bucketName, entry, id, head)
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

// CSVRowResult represents the parsed result of a CSV row.
type CSVRowResult struct {
	Size        int64  `json:"size"`
	ContentType string `json:"content_type,omitempty"`
	Labels      map[string]any
}

// ParseCSVRow parses a CSV row with support for escaped values.
//
// Fields are separated by commas, empty fields are skipped, and a comma inside
// a quoted section is treated as data. The first field is the content length,
// the second the content type, and the rest are `name=value` labels.
func ParseCSVRow(row string) CSVRowResult {
	result := CSVRowResult{
		Labels: make(map[string]any),
	}

	// The overwhelmingly common case has no quoted values, and every field is
	// then a substring of row, so it needs no copying at all.
	if strings.IndexByte(row, '"') < 0 {
		parseCSVRowPlain(row, &result)
	} else {
		parseCSVRowQuoted(row, &result)
	}

	return result
}

// setCSVField assigns the index-th non-empty field of a record header row.
func setCSVField(result *CSVRowResult, index int, field string) {
	switch index {
	case 0:
		if size, err := strconv.ParseInt(field, 10, 64); err == nil {
			result.Size = size
		}
	case 1:
		result.ContentType = field
	default:
		if eq := strings.IndexByte(field, '='); eq >= 0 {
			result.Labels[field[:eq]] = field[eq+1:]
		}
	}
}

func parseCSVRowPlain(row string, result *CSVRowResult) {
	index := 0
	for row != "" {
		var field string
		if comma := strings.IndexByte(row, ','); comma >= 0 {
			field, row = row[:comma], row[comma+1:]
		} else {
			field, row = row, ""
		}
		if field == "" {
			continue
		}
		setCSVField(result, index, field)
		index++
	}
}

func parseCSVRowQuoted(row string, result *CSVRowResult) {
	var current strings.Builder
	escaped := ""
	index := 0

	for _, char := range row {
		switch {
		case char == ',' && escaped == "":
			if current.Len() > 0 {
				setCSVField(result, index, current.String())
				index++
				current.Reset()
			}
		case char == '"':
			if escaped == "" {
				escaped = current.String()
				current.Reset()
			} else {
				setCSVField(result, index, escaped+current.String())
				index++
				escaped = ""
				current.Reset()
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		if escaped != "" {
			setCSVField(result, index, escaped+current.String())
		} else {
			setCSVField(result, index, current.String())
		}
	}
}

const timeHeaderPrefix = "x-reduct-time-"

// timeHeader is one `x-reduct-time-<ts>: <csv row>` response header.
type timeHeader struct {
	ts    int64
	value string
}

// collectTimeHeaders extracts and orders the record headers of a batch in a
// single pass over the response headers.
func collectTimeHeaders(headers http.Header) ([]timeHeader, error) {
	collected := make([]timeHeader, 0, len(headers))
	for name, values := range headers {
		if len(name) <= len(timeHeaderPrefix) || !strings.EqualFold(name[:len(timeHeaderPrefix)], timeHeaderPrefix) {
			continue
		}

		rawTS := name[len(timeHeaderPrefix):]
		ts, err := strconv.ParseInt(rawTS, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp %s: %w", rawTS, err)
		}
		if len(values) == 0 || values[0] == "" {
			return nil, fmt.Errorf("no record found for timestamp %d", ts)
		}

		collected = append(collected, timeHeader{ts: ts, value: values[0]})
	}

	slices.SortFunc(collected, func(a, b timeHeader) int {
		return cmp.Compare(a.ts, b.ts)
	})

	return collected, nil
}

// readBatchedRecords fetches one batch of records for a query.
//
// Every record but the last is buffered so that callers may consume records in
// any order; the last record streams straight off the response body. All the
// buffered payloads of a batch share a single allocation, and the Record and
// reader values are carved out of one backing array each, so per-record
// allocation is limited to the label map.
func readBatchedRecords(ctx context.Context, client httpclient.HTTPClient, bucketName, entry string, id int64, head bool) ([]*Record, error) {
	path := fmt.Sprintf("/b/%s/%s/batch?q=%d", bucketName, entry, id)
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

	timeHeaders, err := collectTimeHeaders(resp.Header)
	if err != nil {
		closeBody(resp)
		return nil, err
	}
	if len(timeHeaders) == 0 {
		closeBody(resp)
		return nil, fmt.Errorf("no records found")
	}

	total := len(timeHeaders)
	lastInQuery := strings.EqualFold(resp.Header.Get("x-reduct-last"), "true")

	// Parse every header up front so the buffered payload size is known before
	// touching the body.
	parsed := make([]CSVRowResult, total)
	var bufferedSize int64
	for i := range timeHeaders {
		parsed[i] = ParseCSVRow(timeHeaders[i].value)
		if i < total-1 && parsed[i].Size > 0 {
			bufferedSize += parsed[i].Size
		}
	}

	if err = ctx.Err(); err != nil {
		closeBody(resp)
		return nil, err
	}

	// Every record but the last is read in one pass into one buffer, which
	// costs a single allocation and a single read for the whole batch.
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
	for i := range timeHeaders {
		isLastInBatch := i == total-1

		var body io.ReadCloser
		switch {
		case head:
			body = emptyBody{}
		case isLastInBatch:
			body = resp.Body
		default:
			readers[i].data = buffered[offset : offset+parsed[i].Size]
			offset += parsed[i].Size
			body = &readers[i]
		}

		contentType := parsed[i].ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		backing[i] = Record{
			Entry:       entry,
			Time:        timeHeaders[i].ts,
			Size:        parsed[i].Size,
			Last:        lastInQuery && isLastInBatch,
			LastInBatch: isLastInBatch,
			Body:        body,
			Labels:      parsed[i].Labels,
			ContentType: contentType,
		}
		records[i] = &backing[i]
	}

	if head {
		// Nothing streams from a HEAD response, so release the connection now.
		closeBody(resp)
	}

	return records, nil
}

func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

// sliceReader is a zero-allocation body for a record already held in memory.
type sliceReader struct {
	data []byte
}

func (r *sliceReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func (r *sliceReader) WriteTo(w io.Writer) (int64, error) {
	if len(r.data) == 0 {
		return 0, nil
	}
	n, err := w.Write(r.data)
	r.data = r.data[n:]
	return int64(n), err
}

func (r *sliceReader) Close() error { return nil }

// emptyBody is the body of a metadata-only record.
type emptyBody struct{}

func (emptyBody) Read([]byte) (int, error) { return 0, io.EOF }
func (emptyBody) Close() error             { return nil }
