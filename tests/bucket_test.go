package tests

import (
	"context"
	"encoding/json"
	"testing"

	"reduct-go/model"

	"github.com/stretchr/testify/assert"
)

func TestRenameBucket(t *testing.T) {
	// check if the bucket exists
	err := mainTestBucket.Rename(context.Background(), "new-bucket-name")
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

func TestRemoveBucket(t *testing.T) {
	// check if the bucket exists
	err := mainTestBucket.Remove(context.Background())
	assert.NoError(t, err)
}

func TestEntryRecordWritterAndReader(t *testing.T) {
	exists, err := mainTestBucket.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)

	// create a new entry record writter
	writter := mainTestBucket.BeginWrite("entry-1", nil)
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
	err = writter.Write(dataByte, int64(len(dataByte)))
	assert.NoError(t, err)
	// we should be able to read the written data
	reader, err := mainTestBucket.BeginRead(context.Background(), "entry-1", nil, nil, false)
	assert.NoError(t, err)
	// read the data
	readData, err := reader.Read()
	assert.NoError(t, err)
	// check if the data is the same
	var readDataMap map[string]any
	err = json.Unmarshal(readData, &readDataMap)
	assert.NoError(t, err)
	// check if the data is the same
	assert.Equal(t, data, readDataMap)
}
