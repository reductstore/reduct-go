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
