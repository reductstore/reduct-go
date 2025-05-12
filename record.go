package reductgo

import (
	"bytes"
	"io"
)

type readableRecord struct {
	time        uint64
	size        uint64
	last        bool
	head        bool
	stream      io.Reader
	labels      map[string]any
	contentType string
	buf         *bytes.Buffer
}

func NewReadableRecord(time uint64, size uint64, last bool, head bool, stream io.Reader, labels map[string]any, contentType string) *readableRecord {
	var buf *bytes.Buffer
	if head {
		buf = bytes.NewBuffer([]byte{})
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return &readableRecord{
		time:        time,
		size:        size,
		last:        last,
		head:        head,
		stream:      stream,
		labels:      labels,
		contentType: contentType,
		buf:         buf,
	}
}

func (r *readableRecord) Read() ([]byte, error) {
	if r.buf != nil {
		return r.buf.Bytes(), nil
	}
	// read from stream
	buf := make([]byte, r.size)
	_, err := io.ReadFull(r.stream, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
