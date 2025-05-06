package reductgo

import (
	"context"
	"fmt"
	"reduct-go/httpclient"
	"reduct-go/model"
	"time"
)

var defaultClientTimeout = 60 * time.Second

// this is a client for a ReductStore instance
type Client interface {
	// Create a new bucket
	CreateBucket(ctx context.Context, name string, settings model.BucketSetting) (Bucket, error)
	// Create a new bucket if it doesn't exist and return it
	CreateOrGetBucket(ctx context.Context, name string, settings model.BucketSetting) (Bucket, error)
	// Get a bucket
	GetBucket(ctx context.Context, name string) (Bucket, error)
	// Check if a bucket exists
	CheckExists(ctx context.Context, name string) (bool, error)
}

type ClientOptions struct {
	ApiToken  string
	Timeout   time.Duration
	VerifySSL bool
}
type ReductClient struct {
	url      string
	timeout  time.Duration
	ApiToken string
	// this is a custom http client
	httpClient httpclient.HTTPClient
}

func NewClient(url string, options ClientOptions) *ReductClient {
	if options.Timeout.Seconds() == 0 {
		options.Timeout = defaultClientTimeout
	}
	client := &ReductClient{
		url:      url,
		timeout:  options.Timeout,
		ApiToken: options.ApiToken,
	}
	client.httpClient = httpclient.NewHTTPClient(httpclient.HttpClientOption{
		ApiToken: options.ApiToken,
		Timeout:  options.Timeout,
		BaseUrl:  url,
	})

	return client
}

func (c *ReductClient) GetBucket(ctx context.Context, name string) (Bucket, error) {

	err := c.httpClient.Get(ctx, fmt.Sprintf(`/b/%s`, name), nil)
	if err != nil {
		return Bucket{}, model.APIError{Message: err.Error(), Original: err}
	}
	return NewBucket(name, c.httpClient), nil
}

func (c *ReductClient) CreateBucket(ctx context.Context, name string, settings model.BucketSetting) (Bucket, error) {

	err := c.httpClient.Post(ctx, fmt.Sprintf("/b/%s", name), settings, nil)
	if err != nil {
		return Bucket{}, err
	}

	return NewBucket(name, c.httpClient), err
}
func (c *ReductClient) CreateOrGetBucket(ctx context.Context, name string, settings model.BucketSetting) (Bucket, error) {

	err := c.httpClient.Post(ctx, fmt.Sprintf("/b/%s", name), settings, nil)
	if err != nil {
		// try get it
		return c.GetBucket(ctx, name)
	}

	return NewBucket(name, c.httpClient), err
}

func (c *ReductClient) CheckExists(ctx context.Context, name string) (bool, error) {

	err := c.httpClient.Head(ctx, fmt.Sprintf(`/b/%s`, name))
	if err != nil {
		return false, err
	}
	return true, nil
}
