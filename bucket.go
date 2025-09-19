package reductgo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
)

type Bucket struct {
	HTTPClient httpclient.HTTPClient
	Name       string
}

func newBucket(name string, httpClient httpclient.HTTPClient) Bucket {
	return Bucket{
		HTTPClient: httpClient,
		Name:       name,
	}
}

// CheckExists checks if the bucket exists on the server.
func (b *Bucket) CheckExists(ctx context.Context) (bool, error) {
	err := b.HTTPClient.Head(ctx, fmt.Sprintf("/b/%s", b.Name))
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetInfo retrieves the basic information about the bucket, such as its name, size, and quota.
func (b *Bucket) GetInfo(ctx context.Context) (model.BucketInfo, error) {
	resp := model.FullBucketDetail{}
	err := b.HTTPClient.Get(ctx, fmt.Sprintf("/b/%s", b.Name), &resp)
	if err != nil {
		return model.BucketInfo{}, err
	}
	return resp.Info, nil
}

// GetEntries retrieves the list of entries and their information in the bucket.
func (b *Bucket) GetEntries(ctx context.Context) ([]model.EntryInfo, error) {
	resp := &model.FullBucketDetail{}
	err := b.HTTPClient.Get(ctx, fmt.Sprintf("/b/%s", b.Name), resp)
	if err != nil {
		return nil, err
	}
	return resp.Entries, nil
}

// GetFullInfo retrieves the full details of the bucket, including its settings and entries.
func (b *Bucket) GetFullInfo(ctx context.Context) (model.FullBucketDetail, error) {
	resp := model.FullBucketDetail{}
	err := b.HTTPClient.Get(ctx, fmt.Sprintf("/b/%s", b.Name), &resp)
	if err != nil {
		return model.FullBucketDetail{}, err
	}
	return resp, nil
}

// GetSettings retrieves the settings of the bucket.
func (b *Bucket) GetSettings(ctx context.Context) (model.BucketSetting, error) {
	resp := &model.FullBucketDetail{}
	err := b.HTTPClient.Get(ctx, fmt.Sprintf("/b/%s", b.Name), resp)
	if err != nil {
		return model.BucketSetting{}, err
	}
	return resp.Settings, nil
}

// SetSettings updates the settings of the bucket.
func (b *Bucket) SetSettings(ctx context.Context, settings model.BucketSetting) error {
	return b.HTTPClient.Put(ctx, fmt.Sprintf("/b/%s", b.Name), settings, nil)
}

// Rename changes the name of the bucket.
func (b *Bucket) Rename(ctx context.Context, newName string) error {
	err := b.HTTPClient.Put(ctx, fmt.Sprintf("/b/%s/rename", b.Name), map[string]string{"new_name": newName}, nil)
	if err != nil {
		return err
	}
	b.Name = newName
	return nil
}

// Remove deletes the bucket from the server.
func (b *Bucket) Remove(ctx context.Context) error {
	return b.HTTPClient.Delete(ctx, fmt.Sprintf("/b/%s", b.Name))
}

// RemoveRecord removes a record from the bucket.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - entry: Name of the entry to remove the record from.
//   - ts: Timestamp of the record to remove in microseconds.
func (b *Bucket) RemoveRecord(ctx context.Context, entry string, ts int64) error {
	return b.HTTPClient.Delete(ctx, fmt.Sprintf("/b/%s/%s?ts=%d", b.Name, entry, ts))
}

// RemoveEntry removes an entry from the bucket and all its records.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - entry: Name of the entry to remove.
func (b *Bucket) RemoveEntry(ctx context.Context, entry string) error {
	return b.HTTPClient.Delete(ctx, fmt.Sprintf("/b/%s/%s", b.Name, entry))
}

// RenameEntry renames an entry.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - entry: Name of the entry to rename.
//   - newName: New name of the entry.
func (b *Bucket) RenameEntry(ctx context.Context, entry, newName string) error {
	return b.HTTPClient.Put(ctx, fmt.Sprintf("/b/%s/%s/rename", b.Name, entry), map[string]string{"new_name": newName}, nil)
}

// BeginRead starts reading a record from the given entry at the specified timestamp.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - entry: Name of the entry to read from.
//   - ts: Optional A UNIX timestamp in microseconds. If it is empty, the latest record is returned.
//
// It returns a readableRecord or an error if the read fails.
//
// Use readableRecord.Read() to read the content of the reader.
func (b *Bucket) BeginRead(ctx context.Context, entry string, ts *int64) (*ReadableRecord, error) {
	if ts == nil {
		// If no timestamp is provided, read the latest record
		return b.readRecord(ctx, entry, nil, false)
	}
	strTs := strconv.FormatInt(*ts, 10)
	return b.readRecord(ctx, entry, &strTs, false)
}

// BeginMetadataRead starts reading only the metadata of a record from the given entry at the specified timestamp.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - entry: Name of the entry to read from.
//   - ts: Optional A UNIX timestamp in microseconds. If it is empty, the latest record is returned.
//
// It returns a readableRecord or an error if the read fails.
//
// Use readableRecord.Read() to read the content of the reader.
func (b *Bucket) BeginMetadataRead(ctx context.Context, entry string, ts *int64) (*ReadableRecord, error) {
	// If no timestamp is provided, read the latest record
	if ts == nil {
		return b.readRecord(ctx, entry, nil, true)
	}
	strTs := strconv.FormatInt(*ts, 10)
	return b.readRecord(ctx, entry, &strTs, true)
}

// readRecord prepares an entry record reader from the reductstore server.
func (b *Bucket) readRecord(ctx context.Context, entry string, ts *string, head bool) (*ReadableRecord, error) {
	query := url.Values{}
	if ts != nil {
		query.Set("ts", *ts)
	}

	path := fmt.Sprintf("/b/%s/%s?%s", b.Name, entry, query.Encode())

	var req *http.Request
	var err error
	if head {
		req, err = b.HTTPClient.NewRequestWithContext(ctx, http.MethodHead, path, nil)
	} else {
		req, err = b.HTTPClient.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	}
	if err != nil {
		return nil, err
	}

	if !head {
		req.Header.Set("Accept", "application/octet-stream")
	}

	resp, err := b.HTTPClient.Do(req) //nolint:bodyclose //intentionally needed for streaming
	if err != nil {
		return nil, err
	}
	errorMessage := resp.Header.Get("x-reduct-error")
	if resp.StatusCode == http.StatusNoContent {
		if errorMessage == "" {
			errorMessage = "No content"
		}
		return nil, model.APIError{Status: http.StatusNoContent, Message: errorMessage}
	}
	// check there is data in the response
	if resp.ContentLength == 0 || resp.Body == nil {
		return nil, model.APIError{Status: http.StatusNoContent, Message: "No content"}
	}

	timeStr := resp.Header.Get("x-reduct-time")
	sizeStr := resp.Header.Get("content-length")
	last := resp.Header.Get("x-reduct-last") == "1"

	labels := make(map[string]any)
	for key, values := range resp.Header {
		key = strings.ToLower(key)
		if strings.HasPrefix(key, "x-reduct-label-") {
			labels[strings.TrimPrefix(key, "x-reduct-label-")] = values[0]
		}
	}

	timeVal, _ := strconv.ParseInt(timeStr, 10, 64) //nolint:errcheck //not needed
	sizeVal, _ := strconv.ParseInt(sizeStr, 10, 64) //nolint:errcheck //not needed
	record := NewReadableRecord(timeVal, sizeVal, last, resp.Body, labels, resp.Header.Get("Content-Type"))
	return record, nil

}

// BeginWrite starts a record writer for an entry.
//
// Parameters:
//   - entry the name of the entry to write the record to.
//   - options:
//   - TimeStamp: timestamp in microseconds, it is set to current time if not provided
//   - ContentType: "text/plain"
//   - Labels: record label kev:value pairs  {label1: "value1", label2: "value2"}.
func (b *Bucket) BeginWrite(_ context.Context, entry string, options *WriteOptions) *WritableRecord {
	var localOptions = WriteOptions{Timestamp: 0}
	if options != nil {
		localOptions = *options
	}
	if localOptions.Timestamp == 0 {
		localOptions.Timestamp = time.Now().UTC().UnixMicro()
	}
	if localOptions.ContentType == "" {
		localOptions.ContentType = "application/octet-stream"
	}
	return NewWritableRecord(b.Name, entry, b.HTTPClient, localOptions)
}

func (b *Bucket) BeginWriteBatch(_ context.Context, entry string) *Batch {
	return newBatch(b.Name, entry, b.HTTPClient, BatchWrite)
}

func (b *Bucket) BeginUpdateBatch(_ context.Context, entry string) *Batch {
	return newBatch(b.Name, entry, b.HTTPClient, BatchUpdate)
}

func (b *Bucket) BeginRemoveBatch(_ context.Context, entry string) *Batch {
	return newBatch(b.Name, entry, b.HTTPClient, BatchRemove)
}

// QueryType represents the type of query to run.
type QueryType string

const (
	QueryTypeQuery  QueryType = "QUERY"
	QueryTypeRemove QueryType = "REMOVE"
)

// QueryOptions represents a query to run on an entry.
type QueryOptions struct {
	QueryType    QueryType     `json:"query_type"`
	Start        int64         `json:"start,omitempty"`
	Stop         int64         `json:"stop,omitempty"`
	When         any           `json:"when,omitempty"`
	Ext          any           `json:"ext,omitempty"`
	Strict       bool          `json:"strict,omitempty"`
	Continuous   bool          `json:"continuous,omitempty"`
	Head         bool          `json:"head,omitempty"`
	PollInterval time.Duration `json:"-"`
}
type QueryOptionsBuilder struct {
	query QueryOptions
}

func NewQueryOptionsBuilder() *QueryOptionsBuilder {
	return &QueryOptionsBuilder{}
}

// WithStart sets the start timestamp (in microseconds) for the query.
// Returns the QueryOptionsBuilder to allow method chaining.
func (q *QueryOptionsBuilder) WithStart(start int64) *QueryOptionsBuilder {
	q.query.Start = start
	return q
}

// WithStop sets the stop timestamp (in microseconds) for the query.
// Returns the QueryOptionsBuilder to allow method chaining.
func (q *QueryOptionsBuilder) WithStop(stop int64) *QueryOptionsBuilder {
	q.query.Stop = stop
	return q
}

// WithWhen sets the when condition for the query.
// Example: map[string]any{"&label": map[string]any{"$eq": "test"}}
// Returns the QueryOptionsBuilder to allow method chaining.
func (q *QueryOptionsBuilder) WithWhen(when any) *QueryOptionsBuilder {
	q.query.When = when
	return q
}

// WithExt sets the ext field for the query to pass additional parameters to extensions
// Returns the QueryOptionsBuilder to allow method chaining.
func (q *QueryOptionsBuilder) WithExt(ext any) *QueryOptionsBuilder {
	q.query.Ext = ext
	return q
}

// WithStrict sets the strict mode for the query.
// Returns the QueryOptionsBuilder to allow method chaining.
func (q *QueryOptionsBuilder) WithStrict(strict bool) *QueryOptionsBuilder {
	q.query.Strict = strict
	return q
}

// WithContinuous WithQueryType makes the query continuous.
// If set, the query doesn't finish if no records are found and waits for new records to be added.
// Returns the QueryOptionsBuilder to allow method chaining.
func (q *QueryOptionsBuilder) WithContinuous(continuous bool) *QueryOptionsBuilder {
	q.query.Continuous = continuous
	return q
}

// WithHead if set to true, only metadata is fetched without the content.
// Returns the QueryOptionsBuilder to allow method chaining.
func (q *QueryOptionsBuilder) WithHead(head bool) *QueryOptionsBuilder {
	q.query.Head = head
	return q
}

// WithPollInterval sets the interval for polling the query if it is continuous.
// Returns the QueryOptionsBuilder to allow method chaining.
func (q *QueryOptionsBuilder) WithPollInterval(pollInterval time.Duration) *QueryOptionsBuilder {
	q.query.PollInterval = pollInterval
	return q
}

// Build builds the QueryOptions from the builder.
func (q *QueryOptionsBuilder) Build() QueryOptions {
	return q.query
}

// QueryResponse represents the response from a query operation.
type QueryResponse struct {
	ID             int64 `json:"id,omitempty"`
	RemovedRecords int64 `json:"removed_records,omitempty"`
}

type QueryResult struct {
	records <-chan *ReadableRecord
}

func (q *QueryResult) Records() <-chan *ReadableRecord {
	if q.records == nil {
		ch := make(chan *ReadableRecord)
		close(ch)
		return ch
	}
	return q.records
}

// Query queries records for a time interval and returns them through a channel
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - entry: Name of the entry to query
//   - start: Optional start point of the time period in microseconds
//   - end: Optional end point of the time period in microseconds
//   - options: Optional query options for filtering and controlling the query behavior
//
// Example:
//
//	records, err := bucket.Query(ctx, "entry-1", start, end, nil)
//	if err != nil {
//	    return err
//	}
//	for record := range records {
//	    fmt.Printf("Time: %d, Size: %d\n", record.Time(), record.Size())
//	    fmt.Printf("Labels: %v\n", record.Labels())
//	    content, err := record.Read()
//	    if err != nil {
//	        return err
//	    }
//	    // Process content...
//	}
func (b *Bucket) Query(ctx context.Context, entry string, options *QueryOptions) (*QueryResult, error) {
	if options == nil {
		options = &QueryOptions{
			QueryType:    QueryTypeQuery,
			Head:         false,
			PollInterval: time.Second,
		} // default options
	}
	if options.PollInterval == 0 {
		options.PollInterval = time.Second // default poll interval
	}
	if options.QueryType == "" {
		options.QueryType = QueryTypeQuery
	}
	resp, err := b.executeQuery(ctx, entry, options)
	if err != nil {
		return &QueryResult{}, err
	}

	return b.fetchAndParseBatchedRecords(ctx, entry, resp.ID, options.Continuous, options.PollInterval, options.Head)
}

// RemoveQuery removes records by query.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - entry: Name of the entry
//   - start: Optional start point of the time period in microseconds. If nil, starts from the first record
//   - end: Optional end point of the time period in microseconds. If nil, ends at the last record
//   - options: Optional query options. Only When and Ext fields are used, other options are ignored
//
// Note: remove is exclusive of the end point. [start, end)
// Returns the number of records removed.
func (b *Bucket) RemoveQuery(ctx context.Context, entry string, options *QueryOptions) (int64, error) {
	if options == nil {
		options = &QueryOptions{}
	}
	options.QueryType = QueryTypeRemove

	resp, err := b.executeQuery(ctx, entry, options)
	if err != nil {
		return 0, err
	}

	return resp.RemovedRecords, nil

}

// executeQuery runs a query on an entry, it returns the query ID or an error.
func (b *Bucket) executeQuery(ctx context.Context, entry string, option *QueryOptions) (QueryResponse, error) {
	path := fmt.Sprintf("/b/%s/%s/q", b.Name, entry)
	if option == nil {
		option = &QueryOptions{
			QueryType: QueryTypeQuery,
		}
	}
	resp := QueryResponse{}
	err := b.HTTPClient.Post(ctx, path, option, &resp)
	if err != nil {
		return QueryResponse{}, err
	}

	return resp, nil
}

// fetchAndParseBatchedRecords fetches and parses batched records with optional polling.
//
// It takes a context, entry name, query ID, whether to continue polling on 204 status,
// polling interval duration, and whether this is a HEAD request.
// It returns a QueryResult containing the records channel or an error.
// If continueQuery is true and a 204 status is received, it will wait pollInterval
// duration and retry the request once before returning.
func (b *Bucket) fetchAndParseBatchedRecords(ctx context.Context, entry string, id int64, continueQuery bool, pollInterval time.Duration, head bool) (*QueryResult, error) {
	// Create a buffered channel for records
	records := make(chan *ReadableRecord, 100)

	// Start a goroutine to continuously fetch records
	go func() {
		defer close(records)

		for {
			record, err := b.readBatchedRecords(ctx, entry, id, head)
			if err != nil {
				var apiErr model.APIError
				if errors.As(err, &apiErr) && apiErr.Status == 204 {
					if continueQuery {
						select {
						case <-ctx.Done():
							return
						case <-time.After(pollInterval):
							continue // Try again after polling interval
						}
					}
					return
				}
				return
			}

			if record == nil {
				return
			}

			// Send records to channel
			for rec := range record {
				select {
				case <-ctx.Done():
				case records <- rec:
					if rec.IsLast() {
						return
					}
				}
			}

		}
	}()

	return &QueryResult{
		records: records,
	}, nil
}

// readBatchedRecords prepares an entry record reader from the reductstore server.
func (b *Bucket) readBatchedRecords(ctx context.Context, entry string, id int64, head bool) (chan *ReadableRecord, error) {
	path := fmt.Sprintf("/b/%s/%s/batch?q=%d", b.Name, entry, id)
	// Create buffered channels
	records := make(chan *ReadableRecord, 100)
	var req *http.Request
	var err error
	if head {
		req, err = b.HTTPClient.NewRequestWithContext(ctx, http.MethodHead, path, nil)
	} else {
		req, err = b.HTTPClient.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err == nil {
			req.Header.Set("Accept", "application/octet-stream")
		}
	}
	if err != nil {
		return nil, err
	}

	resp, err := b.HTTPClient.Do(req) //nolint:bodyclose //intentionally needed for streaming
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

	// Find all timestamp headers first
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

				// Parse the CSV row
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
				record := NewReadableRecord(ts, parsed.Size, isLastInQuery, body, parsed.Labels, parsed.ContentType)
				record.lastInBatch = isLastInBatch
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

// CSVRowResult represents the parsed result of a CSV row.
type CSVRowResult struct {
	Size        int64    `json:"size"`
	ContentType string   `json:"content_type,omitempty"`
	Labels      LabelMap `json:"labels"`
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
		Labels: make(LabelMap),
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

// Update updates the labels of an existing record.
// If a label has an empty string value, it will be removed.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - entry: Name of the entry
//   - ts: Timestamp of record in microseconds
//   - labels: Labels to update
func (b *Bucket) Update(ctx context.Context, entry string, ts int64, labels LabelMap) error {
	headers := make(map[string]string)

	for key, value := range labels {
		headers[fmt.Sprintf("x-reduct-label-%s", key)] = fmt.Sprint(value)
	}

	path := fmt.Sprintf("/b/%s/%s?ts=%d", b.Name, entry, ts)
	req, err := b.HTTPClient.NewRequestWithContext(ctx, http.MethodPatch, path, nil)
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := b.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return model.APIError{
			Status:  resp.StatusCode,
			Message: resp.Header.Get("x-reduct-error"),
		}
	}

	return nil
}

type QueryLinkOptions struct {
	Bucket       string       `json:"bucket"`
	Entry        string       `json:"entry"`
	QueryOptions QueryOptions `json:"query"`
	RecordIndex  int          `json:"index"`
	ExpireAt     int64        `json:"expire_at"`
	fileName     string
}

type QueryLinkOptionsBuilder struct {
	options QueryLinkOptions
}

// NewQueryLinkOptionsBuilder creates a new QueryLinkOptionsBuilder with default values.
// Default values:
//   - recordIndex: 0 (first record)
//   - expireAt: 24 hours from now
func NewQueryLinkOptionsBuilder() *QueryLinkOptionsBuilder {
	return &QueryLinkOptionsBuilder{
		options: QueryLinkOptions{
			RecordIndex: 0,
			ExpireAt:    time.Now().Add(24 * time.Hour).Unix(),
			QueryOptions: QueryOptions{
				QueryType: QueryTypeQuery,
			},
		},
	}
}

func (q *QueryLinkOptionsBuilder) WithQueryOptions(queryOptions QueryOptions) *QueryLinkOptionsBuilder {
	q.options.QueryOptions = queryOptions
	return q
}

func (q *QueryLinkOptionsBuilder) WithRecordIndex(recordIndex int) *QueryLinkOptionsBuilder {
	q.options.RecordIndex = recordIndex
	return q
}

func (q *QueryLinkOptionsBuilder) WithExpireAt(expireAt int64) *QueryLinkOptionsBuilder {
	q.options.ExpireAt = expireAt
	return q
}

func (q *QueryLinkOptionsBuilder) WithFileName(fileName string) *QueryLinkOptionsBuilder {
	q.options.fileName = fileName
	return q
}

func (q *QueryLinkOptionsBuilder) Build() QueryLinkOptions {
	return q.options
}

// CreateQueryLink creates a temporary link to access a specific record from a query.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - entry: Name of the entry to query
//   - options: Options for the query link, including query parameters, record index, expiration time, and file name
func (b *Bucket) CreateQueryLink(ctx context.Context, entry string, options QueryLinkOptions) (string, error) {
	options.Bucket = b.Name
	options.Entry = entry
	if options.fileName == "" {
		options.fileName = fmt.Sprintf("%s_%d", entry, options.RecordIndex)
	}

	response := struct {
		Link string `json:"link"`
	}{}

	err := b.HTTPClient.Post(ctx, fmt.Sprintf("/links/%s", options.fileName), options, &response)
	if err != nil {
		return "", err
	}

	return response.Link, nil
}
