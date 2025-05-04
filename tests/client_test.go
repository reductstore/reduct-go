package tests

import (
	"context"
	"fmt"
	"math/rand"
	reductgo "reduct-go"
	"reduct-go/model"
	"testing"
)

var apiToken = "4a420b22-2912-11f0-a74c-00155d4c972d"

func getRandomBucketName() string {
	return fmt.Sprintf("test-bucket-%d", rand.Intn(1000000))
}
func TestCreateBucket(t *testing.T) {
	client := reductgo.NewReductClient(apiToken)
	bucketName := getRandomBucketName()
	createBucketRequest := model.CreateBucketRequest{
		BucketName:      bucketName,
		MaxBlockSize:    1024,
		MaxBlockRecords: 1000,
		QuotaType:       model.QuotaTypeFifo,
		QuotaSize:       1024 * 1024 * 1024,
	}

	createBucketResponse, err := client.CreateBucket(context.Background(), createBucketRequest)
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	fmt.Printf("Bucket created: %v", createBucketResponse)
	// delete the bucket
	err = client.DeleteBucket(context.Background(), bucketName)
	if err != nil {
		t.Fatalf("Failed to delete bucket: %v", err)
	}
}
