package tests

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	reductgo "reduct-go"
	"reduct-go/model"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var serverUrl = "http://localhost:8383"

func getRandomBucketName() string {
	return fmt.Sprintf("test-bucket-%d", rand.Intn(1000000))
}
func init() {
	_ = godotenv.Load("../.env") // Loads env from .env into os.Environ
}
func TestCreateBucket(t *testing.T) {
	var apiToken = os.Getenv("RS_API_TOKEN")

	client := reductgo.NewClient(serverUrl, reductgo.ClientOptions{
		ApiToken: apiToken,
	})
	bucketName := getRandomBucketName()
	settings := model.BucketSetting{
		MaxBlockSize:    1024,
		MaxBlockRecords: 1000,
		QuotaType:       model.QuotaTypeFifo,
		QuotaSize:       1024 * 1024 * 1024,
	}

	createBucketResponse, err := client.CreateBucket(context.Background(), bucketName, settings)
	assert.NoError(t, err)

	bucket, err := client.GetBucket(context.Background(), bucketName)
	assert.NoError(t, err)
	fmt.Printf("Bucket created: %v", createBucketResponse)
	// delete the bucket
	err = bucket.Delete(context.Background())
	assert.NoError(t, err)
}
