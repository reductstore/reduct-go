package tests

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	reductgo "reduct-go"
	"reduct-go/httpclient"
	"reduct-go/model"

	"github.com/joho/godotenv"
)

var (
	mainTestBucket = reductgo.Bucket{}
	client         = reductgo.ReductClient{}
)

func getRandomBucketName() string {
	return fmt.Sprintf("remove-test-bucket-%d", rand.Int()) //nolint:gosec //uses math
}

func getNewTestClient() *reductgo.ReductClient {
	apiToken := os.Getenv("RS_API_TOKEN")

	return &reductgo.ReductClient{
		APIToken: apiToken,
		HTTPClient: httpclient.NewHTTPClient(httpclient.Option{
			BaseURL:  serverURL,
			APIToken: apiToken,
		}),
	}
}

func setup() {
	settings := model.NewBucketSettingBuilder().
		WithQuotaSize(1024 * 1024 * 1024).
		WithQuotaType(model.QuotaTypeFifo).
		WithMaxBlockRecords(1000).
		WithMaxBlockSize(1024).Build()

	client = *getNewTestClient()

	_, err := client.CreateBucket(context.Background(), mainTestBucket.Name, settings)
	if err != nil {
		log.Fatal(err)
	}
}

func tearDown() {
	_ = client.RemoveBucket(context.Background(), mainTestBucket.Name) //nolint:errcheck //no need

}

func TestMain(m *testing.M) {
	fmt.Println("Setting up test environment...")

	_ = godotenv.Load("../.env") //nolint:errcheck //Loads env from .env into os.Environ

	mainTestBucket.Name = getRandomBucketName()
	mainTestBucket.HTTPClient = getNewTestClient().HTTPClient
	setup()
	// Run tests
	code := m.Run()

	fmt.Println("Tearing down test environment...")
	// delete the created bucket with all its entries
	tearDown()
	os.Exit(code)
}
