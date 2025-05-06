package reductgo

import (
	"context"
	"fmt"
	"reduct-go/httpclient"
	"reduct-go/model"
)

type Bucket struct {
	httpClient httpclient.HTTPClient
	name       string
}

func NewBucket(name string, httpClient httpclient.HTTPClient) Bucket {
	return Bucket{
		httpClient: httpClient,
		name:       name,
	}
}

func (b *Bucket) CheckExists(ctx context.Context) (bool, error) {
	err := b.httpClient.Head(ctx, fmt.Sprintf("/b/%s", b.name))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *Bucket) Create(ctx context.Context, settings model.BucketSetting) (model.BucketInfo, error) {
	resp := &model.BucketInfo{}
	err := b.httpClient.Post(ctx, fmt.Sprintf("/b/%s", b.name), settings, resp)
	if err != nil {
		return model.BucketInfo{}, err
	}
	return *resp, nil
}

func (b *Bucket) GetInfo(ctx context.Context) (model.BucketInfo, error) {
	resp := &model.FullBucketDetail{}
	err := b.httpClient.Get(ctx, fmt.Sprintf("/b/%s", b.name), resp)
	if err != nil {
		return model.BucketInfo{}, err
	}
	return resp.Info, nil
}

func (b *Bucket) GetSettings(ctx context.Context) (model.BucketSetting, error) {
	resp := &model.FullBucketDetail{}
	err := b.httpClient.Get(ctx, fmt.Sprintf("/b/%s", b.name), resp)
	if err != nil {
		return model.BucketSetting{}, err
	}
	return resp.Settings, nil
}

func (b *Bucket) SetSettings(ctx context.Context, settings model.BucketSetting) error {
	err := b.httpClient.Put(ctx, fmt.Sprintf("/b/%s", b.name), settings, nil)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bucket) Rename(ctx context.Context, newName string) error {
	err := b.httpClient.Put(ctx, fmt.Sprintf("/b/%s/rename", b.name), map[string]string{"new_name": newName}, nil)
	if err != nil {
		return err
	}
	b.name = newName
	return nil
}
func (b *Bucket) Delete(ctx context.Context) error {
	err := b.httpClient.Delete(ctx, fmt.Sprintf("/b/%s", b.name))
	if err != nil {
		return err
	}
	return nil
}
