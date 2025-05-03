package reductgo

import (
	"context"
	"fmt"
	"reduct-go/model"
	"testing"
)

var apiToken = "test-token"

func TestCreateToken(t *testing.T) {
	client := NewClient(apiToken)
	createTokenRequest := model.CreateTokenRequest{
		Name: "test-token",
	}
	createTokenResponse, err := client.CreateToken(context.Background(), createTokenRequest)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	fmt.Printf("Token created: %v", createTokenResponse)
	apiToken = createTokenResponse.Value
}

func TestCreateBucket(t *testing.T) {
	client := NewClient(apiToken)
	bucketName := "test-bucket"
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
}
