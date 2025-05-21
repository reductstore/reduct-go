package reductgo

import (
	"bytes"
	"fmt"
	"io"
	"reduct-go/httpclient"
	"reduct-go/model"
	"strconv"
)

type WriteOptions struct {
	Timestamp   uint64
	ContentType string
	Labels      map[string]any
}

type writableRecord struct {
	bucketName string
	entryName  string
	httpClient httpclient.HTTPClient
	options    WriteOptions
}

func NewWritableRecord(bucketName string, entryName string, httpClient httpclient.HTTPClient, options WriteOptions) *writableRecord {
	return &writableRecord{
		bucketName: bucketName,
		entryName:  entryName,
		httpClient: httpClient,
		options:    options,
	}
}

func (w *writableRecord) Write(data any, size int64) error {
	if w.options.Timestamp == 0 {
		return fmt.Errorf("timestamp must be set")
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
		if size <= 0 {
			return fmt.Errorf("stream data requires a valid size")
		}
		contentLength = size
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

type readableRecord struct {
	time        uint64
	size        uint64
	last        bool
	stream      io.Reader
	labels      map[string]any
	contentType string
}

func NewReadableRecord(time uint64,
	size uint64,
	last bool,
	head bool,
	stream io.Reader,
	labels map[string]any,
	contentType string,
) *readableRecord {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return &readableRecord{
		time:        time,
		size:        size,
		last:        last,
		stream:      stream,
		labels:      labels,
		contentType: contentType,
	}
}

func (r *readableRecord) Read() ([]byte, error) {
	// read from stream
	buffer := bytes.NewBuffer([]byte{})
	_, err := io.ReadFull(r.stream, buffer.Bytes())
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (r *readableRecord) ReadAsString() (string, error) {
	buffer := bytes.NewBuffer([]byte{})
	_, err := io.ReadFull(r.stream, buffer.Bytes())
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (r *readableRecord) Stream() io.Reader {
	return r.stream
}

func (r *readableRecord) IsLast() bool {
	return r.last
}

func (r *readableRecord) Size() uint64 {
	return r.size
}

func (r *readableRecord) Labels() map[string]any {
	return r.labels
}

func (r *readableRecord) ContentType() string {
	return r.contentType
}

func (r *readableRecord) Time() uint64 {
	return r.time
}
