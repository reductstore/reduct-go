package reductgo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/model"
	"github.com/stretchr/testify/assert"
)

func TestRenameBucket(t *testing.T) {
	// check if the bucket exists
	err := mainTestBucket.Rename(context.Background(), getRandomBucketName())
	assert.NoError(t, err)
}

func TestUpdateSettings(t *testing.T) {
	// update settings
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(2048).
		WithMaxBlockRecords(2000).WithMaxBlockSize(3000).Build()
	err := mainTestBucket.SetSettings(context.Background(), settings)
	assert.NoError(t, err)
	// get the bucket settings and check if the settings are updated
	settings, err = mainTestBucket.GetSettings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(3000), settings.MaxBlockSize)
	assert.Equal(t, int64(2000), settings.MaxBlockRecords)
	assert.Equal(t, model.QuotaTypeNone, settings.QuotaType)
	assert.Equal(t, int64(2048), settings.QuotaSize)
}

func TestBucketExists(t *testing.T) {
	exists, err := mainTestBucket.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestEntryRecordWriterAndReader(t *testing.T) {
	exists, err := mainTestBucket.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)

	// create a new entry record writer
	writer := mainTestBucket.BeginWrite(context.Background(), "entry-1", nil)
	data := map[string]any{
		"key1": "value1",
		"key2": float64(2),
		"key3": map[string]any{
			"key3.1": "value3.1",
			"key3.2": 3.2,
		},
	}
	dataByte, err := json.Marshal(data)
	assert.NoError(t, err)
	err = writer.Write(dataByte)
	assert.NoError(t, err)
	// we should be able to read the written data
	reader, err := mainTestBucket.BeginRead(context.Background(), "entry-1", nil)
	assert.NoError(t, err)
	// read the data
	resp, err := reader.Read()
	assert.NoError(t, err)
	// check if the data is the same
	var readDataMap map[string]any
	err = json.Unmarshal(resp, &readDataMap)
	assert.NoError(t, err)
	// check if the data is the same
	assert.Equal(t, data, readDataMap)
}

func TestEntryRecordStreamWriterAndChunkedReader(t *testing.T) {
	exists, err := mainTestBucket.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Begin writing to entry using stream
	writer := mainTestBucket.BeginWrite(context.Background(), "entry-stream-chunked", nil)

	chunks := []byte(`{"part": "one","more": 123,"nested": {"inner": "value"}}`)

	err = writer.Write(chunks)
	assert.NoError(t, err)

	// Begin reading with streaming reader
	reader, err := mainTestBucket.BeginRead(context.Background(), "entry-stream-chunked", nil)
	assert.NoError(t, err)
	// Stream read in chunks (e.g., 16 bytes at a time)
	streamReader := reader.Stream()
	buf := make([]byte, 16)
	var result []byte

	for {
		n, err := streamReader.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
	}

	expectedJSON := `{"part": "one","more": 123,"nested": {"inner": "value"}}`
	assert.JSONEq(t, expectedJSON, string(result))
}

func TestUpdateRecordLabels(t *testing.T) {
	ctx := context.Background()
	entry := "test-update-labels"

	// First write a record with initial labels
	now := time.Now().UTC().UnixMicro()
	writer := mainTestBucket.BeginWrite(ctx, entry, &WriteOptions{
		Timestamp: now,
		Labels: LabelMap{
			"initial": "value",
		},
		Size: int64(9),
	})
	reader := bytes.NewReader([]byte("test data"))

	err := writer.Write(reader)
	assert.NoError(t, err)

	// Update the labels
	newLabels := LabelMap{
		"initial": "",          // This should remove the label
		"updated": "new-value", // This should add a new label
	}
	err = mainTestBucket.Update(ctx, entry, now, newLabels)
	assert.NoError(t, err)

	// Verify the labels were updated correctly
	record, err := mainTestBucket.BeginMetadataRead(ctx, entry, &now)
	assert.NoError(t, err)
	labels := record.Labels()
	assert.NotContains(t, labels, "initial", "initial label should be removed")
	assert.Equal(t, "new-value", labels["updated"], "updated label should be set")
}

func TestRemoveBucket(t *testing.T) {
	// check if the bucket exists
	err := mainTestBucket.Remove(context.Background())
	assert.NoError(t, err)
}
