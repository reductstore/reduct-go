package tests

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	reductgo "reduct-go"
	"reduct-go/httpclient"
	"reduct-go/model"
	"testing"

	"github.com/joho/godotenv"
)

var mainTestBucket = reductgo.Bucket{}

func getRandomBucketName() string {
	return fmt.Sprintf("test-bucket-%d", rand.Intn(1000000))
}
func getNewTestClient() *reductgo.ReductClient {
	var apiToken = os.Getenv("RS_API_TOKEN")

	return &reductgo.ReductClient{
		ApiToken: apiToken,
		HttpClient: httpclient.NewHTTPClient(httpclient.HttpClientOption{
			BaseUrl:  serverUrl,
			ApiToken: apiToken,
		}),
	}
}
func setup() {
	settings := model.BucketSetting{
		MaxBlockSize:    1024,
		MaxBlockRecords: 1000,
		QuotaType:       model.QuotaTypeFifo,
		QuotaSize:       1024 * 1024 * 1024,
	}

	_, err := getNewTestClient().CreateBucket(context.Background(), mainTestBucket.Name, settings)
	if err != nil {
		log.Fatal(err)
	}
}

func tearDown() {
	_ = getNewTestClient().RemoveBucket(context.Background(), mainTestBucket.Name)
}

func TestMain(m *testing.M) {
	fmt.Println("Setting up test environment...")

	_ = godotenv.Load("../.env") // Loads env from .env into os.Environ
	mainTestBucket.Name = getRandomBucketName()
	mainTestBucket.HttpClient = getNewTestClient().HttpClient
	setup()
	// Run tests
	code := m.Run()

	fmt.Println("Tearing down test environment...")
	// delete the created bucket with all its entries
	tearDown()
	os.Exit(code)
}
