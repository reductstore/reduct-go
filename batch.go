package reductgo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
)

type BatchType int

const (
	BatchWrite BatchType = iota
	BatchUpdate
	BatchRemove
)

type LabelMap map[string]any

type Record struct {
	Data        []byte
	ContentType string
	Labels      LabelMap
}

type Batch struct {
	bucketName string
	entryName  string
	httpClient httpclient.HTTPClient
	batchType  BatchType
	records    map[int64]*Record
	totalSize  int64
	lastAccess time.Time
	mu         sync.Mutex
}

type BatchOptions struct{}

// NewBatch creates a new batch.
func NewBatch(bucket, entry string, client httpclient.HTTPClient, batchType BatchType) *Batch {
	return &Batch{
		bucketName: bucket,
		entryName:  entry,
		httpClient: client,
		batchType:  batchType,
		records:    make(map[int64]*Record),
		totalSize:  0,
		lastAccess: time.Now().UTC(),
		mu:         sync.Mutex{},
	}
}

// Add adds a record to the batch.
func (b *Batch) Add(ts int64, data []byte, contentType string, labels LabelMap) {
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
	b.records[ts] = &Record{Data: data, ContentType: contentType, Labels: labels}
}

// AddOnlyLabels adds an empty record with only labels.
func (b *Batch) AddOnlyLabels(ts int64, labels LabelMap) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.records[ts] = &Record{Data: []byte{}, ContentType: "", Labels: labels}
}

// AddOnlyTimestamp adds an empty record with only a timestamp.
func (b *Batch) AddOnlyTimestamp(ts int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.records[ts] = &Record{Data: []byte{}, ContentType: "", Labels: LabelMap{}}
}

// Write writes the batch to the server.
func (b *Batch) Write(ctx context.Context) error {
	b.mu.Lock()
	headers := http.Header{}
	var chunks bytes.Buffer
	var contentLength int64
	timestamps := make([]int64, 0, len(b.records))
	for ts := range b.records {
		timestamps = append(timestamps, ts)
	}
	slices.Sort(timestamps)
	for _, ts := range timestamps {
		rec := b.records[ts]
		contentLength += int64(len(rec.Data))
		header := fmt.Sprintf("x-reduct-time-%d", ts)

		headerValue := "0,"
		if b.batchType == BatchWrite {
			headerValue = fmt.Sprintf("%d,%s", len(rec.Data), rec.ContentType)
		}
		for k, v := range rec.Labels {
			valStr, ok := v.(string)
			if ok && strings.Contains(valStr, ",") {
				headerValue += fmt.Sprintf(",%s=%q", k, valStr)
			} else {
				headerValue += fmt.Sprintf(",%s=%s", k, valStr)
			}
		}
		headers.Set(header, headerValue)
		chunks.Write(rec.Data)
	}

	b.mu.Unlock()

	var req *http.Request
	var err error
	path := fmt.Sprintf("/b/%s/%s/batch", b.bucketName, b.entryName)

	switch b.batchType {
	case BatchWrite:
		req, err = b.httpClient.NewRequestWithContext(ctx, http.MethodPost, path, &chunks)
		req.Header = headers
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Length", strconv.Itoa(chunks.Len()))
	case BatchUpdate:
		req, err = b.httpClient.NewRequestWithContext(ctx, http.MethodPatch, path, nil)
		req.Header = headers
	case BatchRemove:
		req, err = b.httpClient.NewRequestWithContext(ctx, http.MethodDelete, path, nil)
		req.Header = headers
	default:
		return errors.New("invalid batch type")
	}

	if err != nil {
		return err
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	errs := make([]string, 0)
	for key, val := range resp.Header {
		if strings.HasPrefix(strings.ToLower(key), "x-reduct-error-") {
			tsStr := strings.TrimPrefix(key, "x-reduct-error-")
			ts, err := strconv.ParseInt(tsStr, 10, 64)
			if err == nil {
				parts := strings.SplitN(val[0], ",", 2)
				if len(parts) == 2 {
					code, _ := strconv.Atoi(parts[0]) //nolint:errcheck //not needed
					errs = append(errs, fmt.Sprintf("error code %d: %s for timestamp %d", code, parts[1], ts))
				}
			}
		}
	}
	if len(errs) > 0 {
		return model.APIError{
			Status:   http.StatusBadRequest,
			Message:  "some records failed to write",
			Original: errors.New(strings.Join(errs, "\n")),
		}
	}
	return nil
}

// Size returns the total size of the batch.
func (b *Batch) Size() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.totalSize
}

// LastAccessTime returns the last access time of the batch.
func (b *Batch) LastAccessTime() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.lastAccess
}

// RecordCount returns the number of records in the batch.
func (b *Batch) RecordCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.records)
}

// Clear removes all records from the batch and resets its state.
func (b *Batch) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.records = make(map[int64]*Record)
	b.totalSize = 0
	b.lastAccess = time.Time{}
}
