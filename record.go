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

func (w *WritableRecord) WithSize(size int64) *WritableRecord {
	w.options.Size = size
	return w
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
// calling this method on a last record is not recommended, use Stream() instead.
func (r *ReadableRecord) Read() ([]byte, error) {
	// read from stream
	data, err := io.ReadAll(r.stream)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *ReadableRecord) ReadAsString() (string, error) {
	data, err := io.ReadAll(r.stream)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *ReadableRecord) Stream() io.Reader {
	return r.stream
}

func (r *ReadableRecord) IsLast() bool {
	return r.last
}

func (r *ReadableRecord) Size() int64 {
	return r.size
}

func (r *ReadableRecord) Labels() LabelMap {
	return r.labels
}

func (r *ReadableRecord) ContentType() string {
	return r.contentType
}

func (r *ReadableRecord) Time() int64 {
	return r.time
}
