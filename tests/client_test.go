package tests

import (
	"context"
	"os"
	reductgo "reduct-go"
	"reduct-go/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

var serverUrl = "http://localhost:8383"

// Creating a new bucket should succeed
func TestCreateBucket_Success(t *testing.T) {
	ctx := context.Background()
	var apiToken = os.Getenv("RS_API_TOKEN")
	client := reductgo.NewClient(serverUrl, reductgo.ClientOptions{
		ApiToken: apiToken,
	})
	var newBucketName = getRandomBucketName()
	info, err := client.CreateBucket(ctx, newBucketName, model.BucketSetting{
		MaxBlockSize:    1024,
		MaxBlockRecords: 1000,
		QuotaType:       model.QuotaTypeFifo,
		QuotaSize:       1024,
	})
	assert.NoError(t, err)
	assert.Equal(t, newBucketName, info.Name)

	// remove the created bucket
	err = client.RemoveBucket(ctx, info.Name)
	assert.NoError(t, err)
}

// Creating an existing bucket should fail
func TestCreateBucket_Fail(t *testing.T) {
	ctx := context.Background()
	var apiToken = os.Getenv("RS_API_TOKEN")
	client := reductgo.NewClient(serverUrl, reductgo.ClientOptions{
		ApiToken: apiToken,
	})
	bucketName := getRandomBucketName()
	_, err := client.CreateBucket(ctx, bucketName, model.BucketSetting{
		MaxBlockSize:    1024,
		MaxBlockRecords: 1000,
		QuotaType:       model.QuotaTypeFifo,
		QuotaSize:       1024,
	})

	assert.NoError(t, err)
	// trying to create existing bucket again
	_, err = client.CreateBucket(ctx, bucketName, model.BucketSetting{
		MaxBlockSize:    1024,
		MaxBlockRecords: 1000,
		QuotaType:       model.QuotaTypeFifo,
		QuotaSize:       1024,
	})
	assert.Error(t, err)

	// remove bucket
	err = client.RemoveBucket(ctx, bucketName)
	assert.NoError(t, err)
}

func TestBucketExistsFail(t *testing.T) {
	var apiToken = os.Getenv("RS_API_TOKEN")
	ctx := context.Background()
	client := reductgo.NewClient(serverUrl, reductgo.ClientOptions{
		ApiToken: apiToken,
	})
	exists, err := client.CheckBucketExists(ctx, "new-not-exist")
	assert.Error(t, err)
	assert.False(t, exists)
}
