package tests

import (
	"context"
	"testing"

	"reduct-go/model"

	"github.com/stretchr/testify/assert"
)

var serverURL = "http://localhost:8383"

func TestCreateOrGetBucket_Success(t *testing.T) {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucket, err := client.CreateOrGetBucket(context.Background(), mainTestBucket.Name, settings)
	assert.NoError(t, err)
	assert.Equal(t, bucket.Name, mainTestBucket.Name)
}

// Creating a new bucket should succeed.
func TestCreateBucket_Success(t *testing.T) {
	ctx := context.Background()

	newBucketName := getRandomBucketName()
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	info, err := client.CreateBucket(ctx, newBucketName, settings)
	assert.NoError(t, err)
	assert.Equal(t, newBucketName, info.Name)

	// remove the created bucket
	err = client.RemoveBucket(ctx, info.Name)
	assert.NoError(t, err)
}

// Creating an existing bucket should fail.
func TestCreateBucket_Fail(t *testing.T) {
	ctx := context.Background()
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucketName := getRandomBucketName()
	_, err := client.CreateBucket(ctx, bucketName, settings)

	assert.NoError(t, err)
	// trying to create existing bucket again
	_, err = client.CreateBucket(ctx, bucketName, settings)
	assert.Error(t, err)

	// remove bucket
	err = client.RemoveBucket(ctx, bucketName)
	assert.NoError(t, err)
}

func TestBucketExistsFail(t *testing.T) {
	ctx := context.Background()

	exists, err := client.CheckBucketExists(ctx, "new-not-exist")
	assert.Error(t, err)
	assert.False(t, exists)
}
