package tests

import (
	"context"
	"reduct-go/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenameBucket(t *testing.T) {
	// check if the bucket exists
	err := mainTestBucket.Rename(context.Background(), "new-bucket-name")
	assert.NoError(t, err)
}

func TestUpdateSettings(t *testing.T) {
	// update settings
	err := mainTestBucket.SetSettings(context.Background(), model.BucketSetting{
		MaxBlockSize:    2048,
		MaxBlockRecords: 2000,
		QuotaType:       model.QuotaTypeNone,
		QuotaSize:       2048,
	})
	assert.NoError(t, err)
	// get the bucket settings and check if the settings are updated
	settings, err := mainTestBucket.GetSettings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(2048), settings.MaxBlockSize)
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
