package reductgo

import (
	"context"
	"net/http"
	"testing"

	"github.com/reductstore/reduct-go/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrGetBucket_Success(t *testing.T) {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucket, err := client.CreateOrGetBucket(context.Background(), mainTestBucket.Name, &settings)
	assert.NoError(t, err)
	assert.Equal(t, bucket.Name, mainTestBucket.Name)
}

func teardownToken(tokenName string) {
	_ = client.RemoveToken(context.Background(), tokenName) //nolint:errcheck // ignore error.
}

func TestTokenAPI(t *testing.T) {
	tokenName := "test-token"
	teardownToken(tokenName)
	t.Run("Create Token", func(t *testing.T) {
		token, err := client.CreateToken(context.Background(), tokenName, model.TokenPermissions{
			FullAccess: true,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
	t.Run("Get Token", func(t *testing.T) {
		token, err := client.GetToken(context.Background(), tokenName)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("Get Current Token", func(t *testing.T) {
		token, err := client.GetCurrentToken(context.Background())
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
	t.Run("Remove Token", func(t *testing.T) {
		err := client.RemoveToken(context.Background(), tokenName)
		assert.NoError(t, err)
	})
	// check if the token is removed
	t.Run("Check if Token is Removed", func(t *testing.T) {
		_, err := client.GetToken(context.Background(), tokenName)
		assert.Error(t, err)
	})
}
func TestGetInfo(t *testing.T) {
	info, err := client.GetInfo(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestGetBuckets(t *testing.T) {
	buckets, err := client.GetBuckets(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, buckets)
}

func TestHealth(t *testing.T) {
	isLive, err := client.IsLive(context.Background())
	assert.NoError(t, err)
	assert.True(t, isLive)
}

// Creating a new bucket should succeed.
func TestCreateBucket_Success(t *testing.T) {
	ctx := context.Background()

	newBucketName := getRandomBucketName()
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	info, err := client.CreateBucket(ctx, newBucketName, &settings)
	assert.NoError(t, err)
	assert.Equal(t, newBucketName, info.Name)

	// remove the created bucket
	err = client.RemoveBucket(ctx, info.Name)
	assert.NoError(t, err)
}
func TestReplicationAPI(t *testing.T) {
	ctx := context.Background()
	sourceBucketName := getRandomBucketName()
	destinationBucketName := getRandomBucketName()
	task := model.ReplicationSettings{
		SrcBucket: sourceBucketName,
		DstBucket: destinationBucketName,
		DstHost:   "http://localhost:8383",
	}
	// create the source bucket
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	_, _ = client.CreateBucket(ctx, sourceBucketName, &settings) //nolint:errcheck // ignore error
	// create the destination bucket
	settings = model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	_, _ = client.CreateBucket(ctx, destinationBucketName, &settings) //nolint:errcheck // ignore error

	t.Run("CreateReplicationTask", func(t *testing.T) {
		err := client.CreateReplicationTask(ctx, "test-replication", task)
		assert.NoError(t, err)
	})
	t.Run("GetReplicationTask", func(t *testing.T) {
		task, err := client.GetReplicationTask(ctx, "test-replication")
		assert.NoError(t, err)
		assert.Equal(t, task.Info.Name, "test-replication")
	})
	t.Run("UpdateReplicationTask", func(t *testing.T) {
		err := client.UpdateReplicationTask(ctx, "test-replication", task)
		assert.NoError(t, err)
	})

	t.Run("GetReplicationTasks", func(t *testing.T) {
		tasks, err := client.GetReplicationTasks(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})

	t.Run("RemoveReplicationTask", func(t *testing.T) {
		err := client.RemoveReplicationTask(ctx, "test-replication")
		assert.NoError(t, err)
	})
	// check its removed
	t.Run("GetReplicationTask", func(t *testing.T) {
		_, err := client.GetReplicationTask(ctx, "test-replication")
		assert.Error(t, err)
	})

}

// Creating an existing bucket should fail.
func TestCreateBucket_Fail(t *testing.T) {
	ctx := context.Background()
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucketName := getRandomBucketName()
	_, err := client.CreateBucket(ctx, bucketName, &settings)

	assert.NoError(t, err)
	// trying to create existing bucket again
	_, err = client.CreateBucket(ctx, bucketName, &settings)
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
func TestReductStoreHealth(t *testing.T) {
	healthURL := "http://127.0.0.1:8383/api/v1/alive"

	req, err := http.NewRequest(http.MethodHead, healthURL, http.NoBody)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, resp.StatusCode, http.StatusOK)
}
