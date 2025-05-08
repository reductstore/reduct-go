package reductgo

import (
	"context"
	"fmt"

	"reduct-go/httpclient"
	"reduct-go/model"
)

type Bucket struct {
	HTTPClient httpclient.HTTPClient
	Name       string
}

func NewBucket(name string, httpClient httpclient.HTTPClient) Bucket {
	return Bucket{
		HTTPClient: httpClient,
		Name:       name,
	}
}

func (b *Bucket) CheckExists(ctx context.Context) (bool, error) {
	err := b.HTTPClient.Head(ctx, fmt.Sprintf("/b/%s", b.Name))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *Bucket) GetInfo(ctx context.Context) (model.BucketInfo, error) {
	resp := &model.FullBucketDetail{}
	err := b.HTTPClient.Get(ctx, fmt.Sprintf("/b/%s", b.Name), resp)
	if err != nil {
		return model.BucketInfo{}, err
	}
	return resp.Info, nil
}

func (b *Bucket) GetSettings(ctx context.Context) (model.BucketSetting, error) {
	resp := &model.FullBucketDetail{}
	err := b.HTTPClient.Get(ctx, fmt.Sprintf("/b/%s", b.Name), resp)
	if err != nil {
		return model.BucketSetting{}, err
	}
	return resp.Settings, nil
}

func (b *Bucket) SetSettings(ctx context.Context, settings model.BucketSetting) error {
	err := b.HTTPClient.Put(ctx, fmt.Sprintf("/b/%s", b.Name), settings, nil)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bucket) Rename(ctx context.Context, newName string) error {
	err := b.HTTPClient.Put(ctx, fmt.Sprintf("/b/%s/rename", b.Name), map[string]string{"new_name": newName}, nil)
	if err != nil {
		return err
	}
	b.Name = newName
	return nil
}

func (b *Bucket) Remove(ctx context.Context) error {
	err := b.HTTPClient.Delete(ctx, fmt.Sprintf("/b/%s", b.Name))
	if err != nil {
		return err
	}
	return nil
}
