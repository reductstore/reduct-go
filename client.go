package reductgo

import (
	"context"
	"fmt"
	"net/http"
	"reduct-go/model"
	"time"
)

const (
	DefaultBaseURL = "http://localhost:8383"
	APIVersion     = "v1"
	DefaultTimeout = 60 * time.Second
)

// add optional options
type ClientOption func(*Client)

func WithURL(url string) ClientOption {
	//TODO: validate and sanitize url
	if url == "" {
		url = DefaultBaseURL
	}
	return func(c *Client) {
		c.url = url
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func (c *Client) WithVerifySSL() {
	c.verifySSL = true
}

type Client struct {
	url          string
	timeout      time.Duration
	verifySSL    bool
	ApiToken     string
	ExtraHeaders map[string]string
}

func NewClient(apiToken string, options ...ClientOption) *Client {
	client := &Client{
		url:          DefaultBaseURL,
		timeout:      DefaultTimeout,
		ApiToken:     apiToken,
		ExtraHeaders: make(map[string]string),
	}
	for _, fn := range options {
		fn(client)
	}
	return client
}
func (c *Client) CreateToken(ctx context.Context, req model.CreateTokenRequest) (*model.CreateTokenResponse, error) {
	url := fmt.Sprintf("%s/api/%s/tokens/%s", c.url, APIVersion, req.Name)
	response := &model.CreateTokenResponse{}
	err := DoRequest(ctx, url, "POST", func(req *http.Request) error {
		req.Header.Set("Content-Type", "application/json")
		return nil
	}, req, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) CreateBucket(ctx context.Context, req model.CreateBucketRequest) (*model.CreateBucketResponse, error) {
	// TODO: validate request
	url := fmt.Sprintf("%s/api/%s/b/%s", c.url, APIVersion, req.BucketName)
	err := DoRequest(ctx, url, "POST", func(req *http.Request) error {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.ApiToken)
		return nil
	}, req, nil)
	if err != nil {
		return nil, err
	}
	// TODO: if it gives result, parse it
	return &model.CreateBucketResponse{}, nil
}
