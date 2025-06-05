package reductgo

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
)

type WriteOptions struct {
	Timestamp   int64
	ContentType string
	Labels      LabelMap
	Size        int64
}

type WritableRecord struct {
	bucketName string
	entryName  string
	httpClient httpclient.HTTPClient
	options    WriteOptions
}

func NewWritableRecord(bucketName string,
	entryName string,
	httpClient httpclient.HTTPClient,
	options WriteOptions,
) *WritableRecord {
	return &WritableRecord{
		bucketName: bucketName,
		entryName:  entryName,
		httpClient: httpClient,
		options:    options,
	}
}

// Write writes the record to the bucket.
//
// data can be a string, []byte, or io.Reader.
// size is the size of the data to write.
// if size is not provided, it will be calculated from the data.
func (w *WritableRecord) Write(data any) error {
	if w.options.Timestamp == 0 {
		return fmt.Errorf("timestamp must be set")
	}
	if data == nil {
		return fmt.Errorf("no data to write")
	}

	var reader io.Reader
	var contentLength int64
	var err error

	switch v := data.(type) {
	case string:
		reader = bytes.NewBufferString(v)
		contentLength = int64(len(v))
	case []byte:
		reader = bytes.NewReader(v)
		contentLength = int64(len(v))
	case io.Reader:
		reader = v
		if w.options.Size != 0 {
			contentLength = w.options.Size
		}
	default:
		return fmt.Errorf("unsupported data type")
	}

	url := fmt.Sprintf("/b/%s/%s?ts=%d", w.bucketName, w.entryName, w.options.Timestamp)

	req, err := w.httpClient.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", w.options.ContentType)
	req.Header.Set("Content-Length", strconv.FormatInt(contentLength, 10))

	// Custom label headers
	for k, v := range w.options.Labels {
		req.Header.Set(fmt.Sprintf("x-reduct-label-%s", k), fmt.Sprint(v))
	}

	resp, err := w.httpClient.Do(req)
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

type ReadableRecord struct {
	time        int64
	size        int64
	last        bool
	lastInBatch bool
	stream      io.Reader
	labels      LabelMap
	contentType string
}

func NewReadableRecord(time int64,
	size int64,
	last bool,
	stream io.Reader,
	labels LabelMap,
	contentType string,
) *ReadableRecord {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return &ReadableRecord{
		time:        time,
		size:        size,
		last:        last,
		stream:      stream,
		labels:      labels,
		contentType: contentType,
	}
}

// Read reads the record from the stream.
//
// note: calling read on last record will return no error, but may return empty data.
//
// calling this method on a last record is not recommended, use Stream().Read() instead.
func (r *ReadableRecord) Read() ([]byte, error) {
	if r == nil {
		return nil, model.APIError{
			Status:  400,
			Message: "record is nil, nothing to read",
		}
	}
	if r.stream == nil {
		return nil, model.APIError{
			Status:  400,
			Message: "stream is nil, nothing to read",
		}
	}
	// read from stream
	data, err := io.ReadAll(r.stream)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ReadAsString reads the record from the stream and returns it as a string.
//
// use this to read the record at once.
func (r *ReadableRecord) ReadAsString() (string, error) {
	data, err := io.ReadAll(r.stream)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Stream returns the stream of the record.
//
// use this to read the record in a stream.
func (r *ReadableRecord) Stream() io.Reader {
	return r.stream
}

// IsLast is true if this is the last record in the query.
//
// This is not the same as IsLastInBatch(), which is true if this is the last record in the batch.
func (r *ReadableRecord) IsLast() bool {
	return r.last
}

// IsLastInBatch is true if this is the last record in the batch.
//
// This is not the same as IsLast(), which is true if this is the last record in the query.
// use this to check if the record is the last in the batch which has to be processed in a stream.
func (r *ReadableRecord) IsLastInBatch() bool {
	return r.lastInBatch
}

// Size returns the size of the record.
func (r *ReadableRecord) Size() int64 {
	return r.size
}

// Labels returns the labels of the record.
func (r *ReadableRecord) Labels() LabelMap {
	return r.labels
}

// ContentType returns the content type of the record.
func (r *ReadableRecord) ContentType() string {
	return r.contentType
}

// Time returns the timestamp of the record.
func (r *ReadableRecord) Time() int64 {
	return r.time
}
