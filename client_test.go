package reductgo

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

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

func TestGetBucket_NotFound(t *testing.T) {
	_, err := client.GetBucket(context.Background(), "not-exist-bucket")
	assert.Error(t, err)

	var apiErr model.APIError
	errors.As(err, &apiErr)
	assert.Equal(t, 404, apiErr.Status)
	assert.Equal(t, "bucket 'not-exist-bucket' not found", apiErr.Message)
}

func TestGetBucketInfo(t *testing.T) {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucket, err := client.CreateOrGetBucket(context.Background(), "test-bucket", &settings)
	assert.NoError(t, err)
	info, err := bucket.GetInfo(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-bucket", info.Name)
	assert.Equal(t, int64(0), info.Size)
}
func TestGetBucketEntries(t *testing.T) {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucket, err := client.CreateOrGetBucket(context.Background(), "test-bucket", &settings)
	assert.NoError(t, err)
	entries, err := bucket.GetEntries(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(entries))
	// write some entries
	writer := bucket.BeginWrite(context.Background(), "test-entry", nil)
	err = writer.Write([]byte("test-data"))
	assert.NoError(t, err)
	entries, err = bucket.GetEntries(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	// delete bucket
	err = client.RemoveBucket(context.Background(), "test-bucket")
	assert.NoError(t, err)
}

func TestGetBucketFullInfo(t *testing.T) {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucket, err := client.CreateOrGetBucket(context.Background(), "test-bucket", &settings)
	assert.NoError(t, err)
	info, err := bucket.GetFullInfo(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-bucket", info.Info.Name)
	assert.Equal(t, model.QuotaTypeFifo, info.Settings.QuotaType)
	assert.Equal(t, int64(1024*1024*1024), info.Settings.QuotaSize)
	assert.Equal(t, 0, len(info.Entries))
	// write some entries
	writer := bucket.BeginWrite(context.Background(), "test-entry", nil)
	err = writer.Write([]byte("test-data"))
	assert.NoError(t, err)
	info, err = bucket.GetFullInfo(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(info.Entries))
	// delete bucket
	err = client.RemoveBucket(context.Background(), "test-bucket")
	assert.NoError(t, err)
}

func TestBucketRemoveEntry(t *testing.T) {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucket, err := client.CreateOrGetBucket(context.Background(), "test-bucket", &settings)
	assert.NoError(t, err)
	writer := bucket.BeginWrite(context.Background(), "test-entry", nil)
	err = writer.Write([]byte("test-data"))
	assert.NoError(t, err)
	err = bucket.RemoveEntry(context.Background(), "test-entry")
	assert.NoError(t, err)
	entries, err := bucket.GetEntries(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(entries))
	// delete bucket
	err = client.RemoveBucket(context.Background(), "test-bucket")
	assert.NoError(t, err)
}

func TestBucketRenameEntry(t *testing.T) {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucket, err := client.CreateOrGetBucket(context.Background(), "test-bucket", &settings)
	assert.NoError(t, err)
	writer := bucket.BeginWrite(context.Background(), "test-entry", nil)
	err = writer.Write([]byte("test-data"))
	assert.NoError(t, err)
	err = bucket.RenameEntry(context.Background(), "test-entry", "test-entry-new")
	assert.NoError(t, err)
	entries, err := bucket.GetEntries(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "test-entry-new", entries[0].Name)
	// delete bucket
	err = client.RemoveBucket(context.Background(), "test-bucket")
	assert.NoError(t, err)
}

func TestBucketRemoveRecord(t *testing.T) {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).WithMaxBlockSize(1024).Build()
	bucket, err := client.CreateOrGetBucket(context.Background(), "test-bucket", &settings)
	assert.NoError(t, err)
	// write a record
	now := time.Now().UTC().UnixMicro()
	writer := bucket.BeginWrite(context.Background(), "test-entry", &WriteOptions{Timestamp: now})
	err = writer.Write([]byte("test-data"))
	assert.NoError(t, err)
	// remove the record
	err = bucket.RemoveRecord(context.Background(), "test-entry", now)
	assert.NoError(t, err)
	// check if the record is removed
	record, err := bucket.BeginRead(context.Background(), "test-entry", &now)
	assert.Error(t, err, "Could not read record after removal")
	data, err := record.Read()
	assert.Error(t, err, "Expected error when reading removed record")
	assert.Equal(t, "", string(data))
	// check if the entry is not removed
	entries, err := bucket.GetEntries(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entries))

	// delete bucket
	err = client.RemoveBucket(context.Background(), "test-bucket")
	assert.NoError(t, err)
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

func TestGetInfo_VersionCheck(t *testing.T) {
	info, err := client.GetInfo(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, info)

	// Parse the server version
	serverVersion, err := model.ParseVersion(info.Version)
	assert.NoError(t, err)
	minVersion, err := model.ParseVersion("v1.5.0")
	assert.NoError(t, err)
	// If the server version is older than min version by 3 minor versions,
	// the warning will be logged but the function will still succeed
	if serverVersion.IsOlderThan(minVersion, 3) {
		t.Logf("WARNING: Server version %s is at least 3 minor versions older than minimum supported version %s",
			serverVersion.String(), minVersion.String())
	}
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

func teardownToken(tokenName string) {
	_ = client.RemoveToken(context.Background(), tokenName) //nolint:errcheck // ignore error.
}
