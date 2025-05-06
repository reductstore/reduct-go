package tests

import (
	"context"
	"os"
	reductgo "reduct-go"
	"reduct-go/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucketClient_CheckExists(t *testing.T) {
	bucketName := getRandomBucketName()
	var apiToken = os.Getenv("RS_API_TOKEN")
	ctx := context.Background()
	client := reductgo.NewClient(serverUrl, reductgo.ClientOptions{
		ApiToken: apiToken,
	})
	bucket, err := client.CreateOrGetBucket(ctx, bucketName, model.BucketSetting{
		MaxBlockSize:    1024,
		MaxBlockRecords: 1000,
		QuotaType:       model.QuotaTypeFifo,
		QuotaSize:       1024,
	})
	assert.NoError(t, err)
	// check if the bucket exists
	exists, err := bucket.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)
	// rename the bucket
	err = bucket.Rename(context.Background(), "new-bucket-name")
	assert.NoError(t, err)
	// get the bucket info and check if the name is updated
	info, err := bucket.GetInfo(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "new-bucket-name", info.Name)
	// update settings
	err = bucket.SetSettings(context.Background(), model.BucketSetting{
		MaxBlockSize:    2048,
		MaxBlockRecords: 2000,
		QuotaType:       model.QuotaTypeNone,
		QuotaSize:       2048,
	})
	assert.NoError(t, err)
	// get the bucket settings and check if the settings are updated
	settings, err := bucket.GetSettings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(2048), settings.MaxBlockSize)
	assert.Equal(t, int64(2000), settings.MaxBlockRecords)
	assert.Equal(t, model.QuotaTypeNone, settings.QuotaType)
	assert.Equal(t, int64(2048), settings.QuotaSize)
	// delete the bucket
	err = bucket.Delete(context.Background())
	assert.NoError(t, err)
}
