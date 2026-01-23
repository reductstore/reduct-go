package batch

import (
	"bytes"
	"context"
	"errors"
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
func FetchAndParse(ctx context.Context, client httpclient.HTTPClient, bucketName, entry string, id int64, continueQuery bool, pollInterval time.Duration, head bool) (<-chan *Record, error) {
	records := make(chan *Record, 100)

	go func() {
		defer close(records)

		for {
			record, err := readBatchedRecords(ctx, client, bucketName, entry, id, head)
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
				return
			}

			if record == nil {
				return
			}

			for rec := range record {
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

	return records, nil
}

// CSVRowResult represents the parsed result of a CSV row.
type CSVRowResult struct {
	Size        int64  `json:"size"`
	ContentType string `json:"content_type,omitempty"`
	Labels      map[string]any
}

// ParseCSVRow parses a CSV row with support for escaped values.
func ParseCSVRow(row string) CSVRowResult {
	items := make([]string, 0)
	escaped := ""
	current := ""

	// Split the row into parts, handling escaped quotes
	for _, char := range row {
		if char == ',' && escaped == "" {
			if current != "" {
				items = append(items, current)
				current = ""
			}
			continue
		}

		if char == '"' {
			if escaped == "" {
				escaped = current
				current = ""
			} else {
				items = append(items, escaped+current)
				escaped = ""
				current = ""
			}
			continue
		}

		current += string(char)
	}

	// Add the last item if exists
	if current != "" {
		if escaped != "" {
			items = append(items, escaped+current)
		} else {
			items = append(items, current)
		}
	}

	// Parse the results
	result := CSVRowResult{
		Labels: make(map[string]any),
	}

	// Parse size
	if len(items) > 0 {
		size, err := strconv.ParseInt(items[0], 10, 64)
		if err == nil {
			result.Size = size
		}
	}

	// Parse content type
	if len(items) > 1 {
		result.ContentType = items[1]
	}

	// Parse labels
	for _, item := range items[2:] {
		if !strings.Contains(item, "=") {
			continue
		}

		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			result.Labels[parts[0]] = parts[1]
		}
	}

	return result
}

func readBatchedRecords(ctx context.Context, client httpclient.HTTPClient, bucketName, entry string, id int64, head bool) (chan *Record, error) {
	path := fmt.Sprintf("/b/%s/%s/batch?q=%d", bucketName, entry, id)
	records := make(chan *Record, 100)
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

	resp, err := client.Do(req) //nolint:bodyclose //intentionally needed for streaming
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNoContent {
		errorMessage := resp.Header.Get("x-reduct-error")
		if errorMessage == "" {
			errorMessage = "No content"
		}
		return nil, model.APIError{Status: http.StatusNoContent, Message: errorMessage}
	}

	timeHeaders := make([]int64, 0)
	for header := range resp.Header {
		header = strings.ToLower(header)
		if strings.HasPrefix(header, "x-reduct-time-") {
			tsStr := strings.TrimPrefix(header, "x-reduct-time-")
			ts, err := strconv.ParseInt(tsStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid timestamp %s: %w", tsStr, err)
			}
			timeHeaders = append(timeHeaders, ts)
		}
	}
	slices.Sort(timeHeaders)
	if len(timeHeaders) == 0 {
		return nil, fmt.Errorf("no records found")
	}
	total := len(timeHeaders)
	var leftover []byte
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer close(records)
		defer wg.Done()
		for i, ts := range timeHeaders {
			select {
			case <-ctx.Done():
				err = fmt.Errorf("context canceled")
				return
			default:
				value := resp.Header.Get(fmt.Sprintf("x-reduct-time-%d", ts))
				if value == "" {
					err = fmt.Errorf("no record found for timestamp %d", ts)
					return
				}

				parsed := ParseCSVRow(value)
				isLastInBatch := i == total-1
				isLastInQuery := resp.Header.Get("x-reduct-last") == "true" && isLastInBatch
				var buffer = make([]byte, parsed.Size)
				var body io.ReadCloser
				switch {
				case head:
					body = io.NopCloser(bytes.NewReader([]byte{}))
				case isLastInBatch:
					if leftover != nil {
						body = io.NopCloser(io.MultiReader(bytes.NewReader(leftover), resp.Body))
						leftover = nil
					} else {
						body = resp.Body
					}
				default:
					var n int
					if leftover != nil {
						n = copy(buffer, leftover)
						leftover = leftover[n:]
					}

					remaining := parsed.Size - int64(n)
					if remaining > 0 {
						_, err = io.ReadFull(resp.Body, buffer[n:parsed.Size])
						if err != nil {
							return
						}
					}

					body = io.NopCloser(bytes.NewReader(buffer[:parsed.Size]))
				}

				contentType := parsed.ContentType
				if contentType == "" {
					contentType = "application/octet-stream"
				}

				record := &Record{
					Entry:       entry,
					Time:        ts,
					Size:        parsed.Size,
					Last:        isLastInQuery,
					LastInBatch: isLastInBatch,
					Body:        body,
					Labels:      parsed.Labels,
					ContentType: contentType,
				}
				select {
				case <-ctx.Done():
					err = fmt.Errorf("context canceled")
					return
				default:
					records <- record
				}

				if isLastInQuery {
					return
				}
			}
		}
	}()
	wg.Wait()
	return records, err
}
