package tests

import (
	"context"
	reductgo "reduct-go"
	"reduct-go/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucketClient_CheckExists(t *testing.T) {
	bucketName := getRandomBucketName()
	client := reductgo.NewReductClient(apiToken).Bucket(bucketName)
	// create the bucket first
	_, err := client.Create(context.Background(), model.BucketSetting{
		MaxBlockSize:    1024,
		MaxBlockRecords: 1000,
		QuotaType:       model.QuotaTypeFifo,
		QuotaSize:       1024,
	})
	assert.NoError(t, err)
	// check if the bucket exists
	exists, err := client.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)
	// rename the bucket
	err = client.Rename(context.Background(), "new-bucket-name")
	assert.NoError(t, err)
	// get the bucket info and check if the name is updated
	info, err := client.GetInfo(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "new-bucket-name", info.Name)
	// update settings
	err = client.SetSettings(context.Background(), model.BucketSetting{
		MaxBlockSize:    2048,
		MaxBlockRecords: 2000,
		QuotaType:       model.QuotaTypeNone,
		QuotaSize:       2048,
	})
	assert.NoError(t, err)
	// get the bucket settings and check if the settings are updated
	settings, err := client.GetSettings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(2048), settings.MaxBlockSize)
	assert.Equal(t, int64(2000), settings.MaxBlockRecords)
	assert.Equal(t, model.QuotaTypeNone, settings.QuotaType)
	assert.Equal(t, int64(2048), settings.QuotaSize)
	// delete the bucket
	err = client.Delete(context.Background())
	assert.NoError(t, err)
}
