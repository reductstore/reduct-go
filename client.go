// Package reductgo provides functionality for managing Reduct object storage, including
// client operations, bucket management
package reductgo

import (
	"context"
	"fmt"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
)

var defaultClientTimeout = 60 * time.Second

// this is a client for a ReductStore instance.
type Client interface {
	// Create a new bucket
	CreateBucket(ctx context.Context, name string, settings model.BucketSetting) (Bucket, error)
	// Create a new bucket if it doesn't exist and return it
	CreateOrGetBucket(ctx context.Context, name string, settings model.BucketSetting) (Bucket, error)
	// Get a bucket
	GetBucket(ctx context.Context, name string) (Bucket, error)
	// Check if a bucket exists
	CheckBucketExists(ctx context.Context, name string) (bool, error)
	// Remove a bucket
	RemoveBucket(ctx context.Context, name string) error
}

type ClientOptions struct {
	APIToken  string
	Timeout   time.Duration
	VerifySSL bool
}
type ReductClient struct {
	url      string
	timeout  time.Duration
	APIToken string
	// this is a custom http client
	HTTPClient httpclient.HTTPClient
}

func NewClient(url string, options ClientOptions) Client {
	if options.Timeout.Seconds() == 0 {
		options.Timeout = defaultClientTimeout
	}
	client := &ReductClient{
		url:      url,
		timeout:  options.Timeout,
		APIToken: options.APIToken,
	}
	client.HTTPClient = httpclient.NewHTTPClient(httpclient.Option{
		APIToken: options.APIToken,
		Timeout:  options.Timeout,
		BaseURL:  url,
	})

	return client
}

func (c *ReductClient) GetBucket(ctx context.Context, name string) (Bucket, error) {
	err := c.HTTPClient.Get(ctx, fmt.Sprintf(`/b/%s`, name), nil)
	if err != nil {
		return Bucket{}, model.APIError{Message: err.Error(), Original: err}
	}
	return NewBucket(name, c.HTTPClient), nil
}

func (c *ReductClient) CreateBucket(ctx context.Context, name string, settings model.BucketSetting) (Bucket, error) {
	err := c.HTTPClient.Post(ctx, fmt.Sprintf("/b/%s", name), settings, nil)
	if err != nil {
		return Bucket{}, err
	}

	return NewBucket(name, c.HTTPClient), err
}

func (c *ReductClient) CreateOrGetBucket(ctx context.Context, name string, settings model.BucketSetting) (Bucket, error) {
	err := c.HTTPClient.Post(ctx, fmt.Sprintf("/b/%s", name), settings, nil)
	if err != nil {
		if apiErr, ok := err.(*model.APIError); ok { //nolint:errorlint //error.As does not give access to status check
			if apiErr.Status == 409 {
				return c.GetBucket(ctx, name)
			}
		} else {
			return Bucket{}, err
		}
	}

	return NewBucket(name, c.HTTPClient), err
}

func (c *ReductClient) CheckBucketExists(ctx context.Context, name string) (bool, error) {
	err := c.HTTPClient.Head(ctx, fmt.Sprintf(`/b/%s`, name))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *ReductClient) RemoveBucket(ctx context.Context, name string) error {
	return c.HTTPClient.Delete(ctx, fmt.Sprintf(`/b/%s`, name))
}
