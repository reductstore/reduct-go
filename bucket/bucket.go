package bucket

import (
	"context"
	"fmt"
	"reduct-go/httpclient"
	"reduct-go/model"
)

type Client interface {
	// Create a new bucket
	Create(ctx context.Context, settings model.BucketSetting) (model.BucketInfo, error)
	// rename bucket
	Rename(ctx context.Context, newName string) error
	// Check if a bucket exists
	CheckExists(ctx context.Context) (bool, error)
	// Get bucket info
	GetInfo(ctx context.Context) (model.BucketInfo, error)
	// Get bucket settings
	GetSettings(ctx context.Context) (model.BucketSetting, error)
	// Set bucket settings
	SetSettings(ctx context.Context, settings model.BucketSetting) error
	// Remove a bucket with all its entries and stored data
	Delete(ctx context.Context) error
}

type bucketClient struct {
	httpClient httpclient.HTTPClient
	apiVersion string
	apiToken   string
	url        string
	name       string
}

func NewBucketClient(name string, apiToken string, url string, apiVersion string, httpClient httpclient.HTTPClient) Client {
	return &bucketClient{
		httpClient: httpClient,
		name:       name,
		apiToken:   apiToken,
		url:        url,
		apiVersion: apiVersion,
	}
}

func (b *bucketClient) CheckExists(ctx context.Context) (bool, error) {
	err := b.httpClient.Do(ctx, fmt.Sprintf("%s/api/%s/b/%s", b.url, b.apiVersion, b.name), "HEAD", nil, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *bucketClient) Create(ctx context.Context, settings model.BucketSetting) (model.BucketInfo, error) {
	resp := &model.BucketInfo{}
	err := b.httpClient.Do(ctx, fmt.Sprintf("%s/api/%s/b/%s", b.url, b.apiVersion, b.name), "POST", settings, resp)
	if err != nil {
		return model.BucketInfo{}, err
	}
	return *resp, nil
}

func (b *bucketClient) GetInfo(ctx context.Context) (model.BucketInfo, error) {
	resp := &model.FullBucketDetail{}
	err := b.httpClient.Do(ctx, fmt.Sprintf("%s/api/%s/b/%s", b.url, b.apiVersion, b.name), "GET", nil, resp)
	if err != nil {
		return model.BucketInfo{}, err
	}
	return resp.Info, nil
}

func (b *bucketClient) GetSettings(ctx context.Context) (model.BucketSetting, error) {
	resp := &model.FullBucketDetail{}
	err := b.httpClient.Do(ctx, fmt.Sprintf("%s/api/%s/b/%s", b.url, b.apiVersion, b.name), "GET", nil, resp)
	if err != nil {
		return model.BucketSetting{}, err
	}
	return resp.Settings, nil
}

func (b *bucketClient) SetSettings(ctx context.Context, settings model.BucketSetting) error {
	err := b.httpClient.Do(ctx, fmt.Sprintf("%s/api/%s/b/%s", b.url, b.apiVersion, b.name), "PUT", settings, nil)
	if err != nil {
		return err
	}
	return nil
}

func (b *bucketClient) Rename(ctx context.Context, newName string) error {
	err := b.httpClient.Do(ctx, fmt.Sprintf("%s/api/%s/b/%s/rename", b.url, b.apiVersion, b.name), "PUT", map[string]string{"new_name": newName}, nil)
	if err != nil {
		return err
	}
	b.name = newName
	return nil
}
func (b *bucketClient) Delete(ctx context.Context) error {
	err := b.httpClient.Do(ctx, fmt.Sprintf("%s/api/%s/b/%s", b.url, b.apiVersion, b.name), "DELETE", nil, nil)
	if err != nil {
		return err
	}
	return nil
}
