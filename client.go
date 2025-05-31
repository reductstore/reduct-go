// Package reductgo provides functionality for managing Reduct object storage, including
// client operations, bucket management
package reductgo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
)

type tokenInfo struct {
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

var defaultClientTimeout = 60 * time.Second

// this is a client for a ReductStore instance.
type Client interface {
	// Get Info
	GetInfo(ctx context.Context) (model.ServerInfo, error)
	// Check if the storage engine is working
	IsLive(ctx context.Context) (bool, error)
	// Get a list of the buckets with their stats
	GetBuckets(ctx context.Context) ([]model.BucketInfo, error)
	// Create a new bucket
	CreateBucket(ctx context.Context, name string, settings *model.BucketSetting) (Bucket, error)
	// Create a new bucket if it doesn't exist and return it
	CreateOrGetBucket(ctx context.Context, name string, settings *model.BucketSetting) (Bucket, error)
	// Get a bucket
	GetBucket(ctx context.Context, name string) (Bucket, error)
	// Check if a bucket exists
	CheckBucketExists(ctx context.Context, name string) (bool, error)
	// Remove a bucket
	RemoveBucket(ctx context.Context, name string) error
	// Get a list of Tokens
	GetTokens(ctx context.Context) ([]model.Token, error)
	// Show Information about a Token
	GetToken(ctx context.Context, name string) (model.Token, error)
	// Create a New Token
	CreateToken(ctx context.Context, name string, permissions model.TokenPermissions) (string, error)
	// Remove a Token
	RemoveToken(ctx context.Context, name string) error
	// Get Full Information about Current API Token
	GetCurrentToken(ctx context.Context) (model.Token, error)
	// Get a list of Replication Tasks
	GetReplicationTasks(ctx context.Context) ([]model.ReplicationInfo, error)
	// Get a Replication Task
	GetReplicationTask(ctx context.Context, name string) (model.FullReplicationInfo, error)
	// Create a Replication Task
	CreateReplicationTask(ctx context.Context, name string, task model.ReplicationSettings) error
	// Update a Replication Task
	UpdateReplicationTask(ctx context.Context, name string, task model.ReplicationSettings) error
	// Remove a Replication Task
	RemoveReplicationTask(ctx context.Context, name string) error
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

// NewClient creates a new ReductClient.
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

// GetInfo returns information about the server.
func (c *ReductClient) GetInfo(ctx context.Context) (model.ServerInfo, error) {
	var info model.ServerInfo
	err := c.HTTPClient.Get(ctx, "/info", &info)
	if err != nil {
		return model.ServerInfo{}, model.APIError{Message: err.Error(), Original: err}
	}
	return info, nil
}

// IsLive checks if the server is live.
func (c *ReductClient) IsLive(ctx context.Context) (bool, error) {
	err := c.HTTPClient.Head(ctx, "/alive")
	if err != nil {
		return false, model.APIError{Message: err.Error(), Original: err}
	}
	return true, nil
}

// GetBuckets returns a list of buckets with their stats.
func (c *ReductClient) GetBuckets(ctx context.Context) ([]model.BucketInfo, error) {
	var buckets map[string][]model.BucketInfo
	err := c.HTTPClient.Get(ctx, "/list", &buckets)
	if err != nil {
		return nil, model.APIError{Message: err.Error(), Original: err}
	}
	return buckets["buckets"], nil
}

// GetBucket returns a bucket.
func (c *ReductClient) GetBucket(ctx context.Context, name string) (Bucket, error) {
	err := c.HTTPClient.Get(ctx, fmt.Sprintf(`/b/%s`, name), nil)
	if err != nil {
		return Bucket{}, model.APIError{Message: err.Error(), Original: err}
	}
	return NewBucket(name, c.HTTPClient), nil
}

func (c *ReductClient) CreateBucket(ctx context.Context, name string, settings *model.BucketSetting) (Bucket, error) {
	err := c.HTTPClient.Post(ctx, fmt.Sprintf("/b/%s", name), settings, nil)
	if err != nil {
		return Bucket{}, err
	}

	return NewBucket(name, c.HTTPClient), err
}

func (c *ReductClient) CreateOrGetBucket(ctx context.Context, name string, settings *model.BucketSetting) (Bucket, error) {
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

// CheckBucketExists checks if a bucket exists.
func (c *ReductClient) CheckBucketExists(ctx context.Context, name string) (bool, error) {
	err := c.HTTPClient.Head(ctx, fmt.Sprintf(`/b/%s`, name))
	if err != nil {
		return false, err
	}
	return true, nil
}

// RemoveBucket removes a bucket.
func (c *ReductClient) RemoveBucket(ctx context.Context, name string) error {
	return c.HTTPClient.Delete(ctx, fmt.Sprintf(`/b/%s`, name))
}

// GetTokens returns a list of tokens.
func (c *ReductClient) GetTokens(ctx context.Context) ([]model.Token, error) {
	var tokens map[string][]model.Token
	err := c.HTTPClient.Get(ctx, "/tokens", &tokens)
	if err != nil {
		return nil, model.APIError{Message: err.Error(), Original: err}
	}
	return tokens["tokens"], nil
}

// GetToken returns information about a token.
func (c *ReductClient) GetToken(ctx context.Context, name string) (model.Token, error) {
	var token model.Token
	err := c.HTTPClient.Get(ctx, fmt.Sprintf("/tokens/%s", name), &token)
	if err != nil {
		return model.Token{}, model.APIError{Message: err.Error(), Original: err}
	}
	return token, nil
}

// CreateToken creates a new token.
func (c *ReductClient) CreateToken(ctx context.Context, name string, permissions model.TokenPermissions) (string, error) {

	var token tokenInfo
	err := c.HTTPClient.Post(ctx, fmt.Sprintf("/tokens/%s", name), permissions, &token)
	if err != nil {
		return "", model.APIError{Message: err.Error(), Original: err}
	}
	return token.Value, nil
}

// RemoveToken removes a token.
func (c *ReductClient) RemoveToken(ctx context.Context, name string) error {
	err := c.HTTPClient.Delete(ctx, fmt.Sprintf("/tokens/%s", name))
	if err != nil {
		apiErr := &model.APIError{}
		if errors.As(err, &apiErr) {
			if apiErr.Status == 404 {
				return nil
			}
		}
		return err
	}
	return nil
}

// GetCurrentToken returns the current token.
func (c *ReductClient) GetCurrentToken(ctx context.Context) (model.Token, error) {
	var token model.Token
	err := c.HTTPClient.Get(ctx, "/me", &token)
	if err != nil {
		return model.Token{}, model.APIError{Message: err.Error(), Original: err}
	}
	return token, nil
}

// GetReplicationTasks returns a list of replication tasks.
func (c *ReductClient) GetReplicationTasks(ctx context.Context) ([]model.ReplicationInfo, error) {
	var tasks map[string][]model.ReplicationInfo
	err := c.HTTPClient.Get(ctx, "/replications", &tasks)
	if err != nil {
		return nil, err
	}
	return tasks["replications"], nil
}

// GetReplicationTask returns a replication task.
func (c *ReductClient) GetReplicationTask(ctx context.Context, name string) (model.FullReplicationInfo, error) {
	var task model.FullReplicationInfo
	err := c.HTTPClient.Get(ctx, fmt.Sprintf("/replications/%s", name), &task)
	if err != nil {
		return model.FullReplicationInfo{}, err
	}
	return task, nil
}

// CreateReplicationTask creates a new replication task.
func (c *ReductClient) CreateReplicationTask(ctx context.Context, name string, task model.ReplicationSettings) error {
	// validate the task
	if task.SrcBucket == "" {
		return fmt.Errorf("src_bucket is required")
	}
	if task.DstBucket == "" {
		return fmt.Errorf("dst_bucket is required")
	}
	if task.DstHost == "" {
		return fmt.Errorf("dst_host is required")
	}
	if name == "" {
		return fmt.Errorf("name is required")
	}
	var fullTask model.FullReplicationInfo
	err := c.HTTPClient.Post(ctx, fmt.Sprintf("/replications/%s", name), task, &fullTask)
	if err != nil {
		return err
	}
	return nil
}

// UpdateReplicationTask updates an existing replication task.
func (c *ReductClient) UpdateReplicationTask(ctx context.Context, name string, task model.ReplicationSettings) error {
	var fullTask model.FullReplicationInfo
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if task.SrcBucket == "" {
		return fmt.Errorf("src_bucket is required")
	}
	if task.DstBucket == "" {
		return fmt.Errorf("dst_bucket is required")
	}
	if task.DstHost == "" {
		return fmt.Errorf("dst_host is required")
	}
	err := c.HTTPClient.Put(ctx, fmt.Sprintf("/replications/%s", name), task, &fullTask)
	if err != nil {
		return err
	}
	return nil
}

// RemoveReplicationTask removes a replication task.
func (c *ReductClient) RemoveReplicationTask(ctx context.Context, name string) error {
	return c.HTTPClient.Delete(ctx, fmt.Sprintf("/replications/%s", name))
}
