package reductgo

import (
	"context"
	"net/http"
	"reduct-go/bucket"
	"reduct-go/httpclient"
	"reduct-go/model"
	"time"
)

const (
	DefaultBaseURL = "http://localhost:8383"
	APIVersion     = "v1"
	DefaultTimeout = 60 * time.Second
)

// add optional options
type ReductClientOption func(*ReductClient)

func WithURL(url string) ReductClientOption {
	//TODO: validate and sanitize url
	if url == "" {
		url = DefaultBaseURL
	}
	return func(c *ReductClient) {
		c.url = url
	}
}

func WithTimeout(timeout time.Duration) ReductClientOption {
	return func(c *ReductClient) {
		c.timeout = timeout
	}
}

type ReductClient struct {
	url      string
	timeout  time.Duration
	ApiToken string
	// this is a custom http client
	httpClient httpclient.HTTPClient
}

func NewReductClient(apiToken string, options ...ReductClientOption) *ReductClient {
	client := &ReductClient{
		url:      DefaultBaseURL,
		timeout:  DefaultTimeout,
		ApiToken: apiToken,
	}
	client.httpClient = httpclient.NewHTTPClient(DefaultTimeout, client.setClientHeaders)

	for _, fn := range options {
		fn(client)
	}
	return client
}

func (c *ReductClient) Bucket(name string) bucket.Client {
	return bucket.NewBucketClient(name, c.ApiToken, c.url, APIVersion, c.httpClient)
}

func (c *ReductClient) setClientHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.ApiToken)
	req.Header.Set("Content-Type", "application/json")
}

func (c *ReductClient) CreateBucket(ctx context.Context, req model.CreateBucketRequest) (model.BucketInfo, error) {
	// NOTE: this is a temporary implementation, might be better to use the bucket client
	// url := fmt.Sprintf("%s/api/%s/b/%s", c.url, APIVersion, req.BucketName)
	// response := &model.CreateBucketResponse{}
	// err := c.httpClient.Do(ctx, url, "POST", req, response)
	// if err != nil {
	// 	return nil, err
	// }

	// TODO: validate request body
	res, err := c.Bucket(req.BucketName).Create(ctx, model.BucketSetting{
		MaxBlockSize:    int64(req.MaxBlockSize),
		MaxBlockRecords: int64(req.MaxBlockRecords),
		QuotaType:       req.QuotaType,
		QuotaSize:       int64(req.QuotaSize),
	})
	return res, err
}

func (c *ReductClient) DeleteBucket(ctx context.Context, name string) error {
	return c.Bucket(name).Delete(ctx)
}

func (c *ReductClient) GetBucketInfo(ctx context.Context, name string) (model.BucketInfo, error) {
	return c.Bucket(name).GetInfo(ctx)
}

func (c *ReductClient) GetBucketSettings(ctx context.Context, name string) (model.BucketSetting, error) {
	return c.Bucket(name).GetSettings(ctx)
}

func (c *ReductClient) SetBucketSettings(ctx context.Context, name string, settings model.BucketSetting) error {
	return c.Bucket(name).SetSettings(ctx, settings)
}

func (c *ReductClient) RenameBucket(ctx context.Context, name string, newName string) error {
	return c.Bucket(name).Rename(ctx, newName)
}
