package reductgo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"reduct-go/httpclient"
	"reduct-go/model"
)

type Bucket struct {
	HTTPClient httpclient.HTTPClient
	Name       string
}

func NewBucket(name string, httpClient httpclient.HTTPClient) Bucket {
	return Bucket{
		HTTPClient: httpClient,
		Name:       name,
	}
}

func (b *Bucket) CheckExists(ctx context.Context) (bool, error) {
	err := b.HTTPClient.Head(ctx, fmt.Sprintf("/b/%s", b.Name))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *Bucket) GetInfo(ctx context.Context) (model.BucketInfo, error) {
	resp := &model.FullBucketDetail{}
	err := b.HTTPClient.Get(ctx, fmt.Sprintf("/b/%s", b.Name), resp)
	if err != nil {
		return model.BucketInfo{}, err
	}
	return resp.Info, nil
}

func (b *Bucket) GetSettings(ctx context.Context) (model.BucketSetting, error) {
	resp := &model.FullBucketDetail{}
	err := b.HTTPClient.Get(ctx, fmt.Sprintf("/b/%s", b.Name), resp)
	if err != nil {
		return model.BucketSetting{}, err
	}
	return resp.Settings, nil
}

func (b *Bucket) SetSettings(ctx context.Context, settings model.BucketSetting) error {
	err := b.HTTPClient.Put(ctx, fmt.Sprintf("/b/%s", b.Name), settings, nil)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bucket) Rename(ctx context.Context, newName string) error {
	err := b.HTTPClient.Put(ctx, fmt.Sprintf("/b/%s/rename", b.Name), map[string]string{"new_name": newName}, nil)
	if err != nil {
		return err
	}
	b.Name = newName
	return nil
}

func (b *Bucket) Remove(ctx context.Context) error {
	err := b.HTTPClient.Delete(ctx, fmt.Sprintf("/b/%s", b.Name))
	if err != nil {
		return err
	}
	return nil
}

// BeginRead starts reading a record from the given entry at the specified timestamp.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - entry: Name of the entry to read from.
//   - ts: Optional A UNIX timestamp in microseconds. If it is empty, the latest record is returned.
//   - id: Optional A query ID to read the next record in the query. If it is set, the parameter ts is ignored.
//   - head: If true, performs a HEAD request to fetch metadata only.
//
// It returns a readableRecord or an error if the read fails.
//
// Use readableRecord.Read() to read the content of the reader
func (b *Bucket) BeginRead(ctx context.Context, entry string, ts, id *string, head bool) (*readableRecord, error) {
	return b.readRecord(ctx, entry, ts, id, head)
}

// readRecord prepares an entry record reader from the reductstore server
func (b *Bucket) readRecord(ctx context.Context, entry string, ts, id *string, head bool) (*readableRecord, error) {
	query := url.Values{}
	if ts != nil {
		query.Set("ts", *ts)
	}
	if id != nil {
		query.Set("q", *id)
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

	resp, err := b.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	errorMessage := resp.Header.Get("x-reduct-error")
	if resp.StatusCode == 204 {
		if errorMessage == "" {
			errorMessage = "No content"
		}
		return nil, model.APIError{Status: 204, Message: errorMessage}
	}
	timeStr := resp.Header.Get("x-reduct-time")
	sizeStr := resp.Header.Get("content-length")
	last := resp.Header.Get("x-reduct-last") == "1"

	labels := make(map[string]any)
	for key, values := range resp.Header {
		if strings.HasPrefix(key, "x-reduct-label-") {
			labels[strings.TrimPrefix(key, "x-reduct-label-")] = values[0]
		}
	}

	timeVal, _ := strconv.ParseUint(timeStr, 10, 64)
	sizeVal, _ := strconv.ParseUint(sizeStr, 10, 64)
	record := NewReadableRecord(timeVal, sizeVal, last, head, resp.Body, labels, resp.Header.Get("Content-Type"))
	return record, nil

}

func (b *Bucket) BeginWrite(entry string, options *WriteOptions) *writableRecord {
	var localOptions = WriteOptions{Timestamp: 0}
	if options != nil {
		localOptions = *options
	}
	if localOptions.Timestamp == 0 {
		// NOTE: time.Now() would give time on the callers server/machine timezone
		localOptions.Timestamp = uint64(time.Now().UnixMicro())
	}
	return NewWritableRecord(b.Name, entry, b.HTTPClient, localOptions)
}
