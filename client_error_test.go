package reductgo

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/reductstore/reduct-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errNotImplemented = errors.New("not implemented")

type stubHTTPClient struct {
	getErr  error
	headErr error
}

func (s stubHTTPClient) Post(context.Context, string, any, any) error { return nil }

func (s stubHTTPClient) Put(context.Context, string, any, any) error { return nil }

func (s stubHTTPClient) Patch(context.Context, string, any, any) error { return nil }

func (s stubHTTPClient) Get(_ context.Context, _ string, _ any) error { return s.getErr }

func (s stubHTTPClient) Head(_ context.Context, _ string) error { return s.headErr }

func (s stubHTTPClient) Delete(context.Context, string) error { return nil }

func (s stubHTTPClient) Do(*http.Request) (*http.Response, error) { return nil, errNotImplemented }

func (s stubHTTPClient) NewRequest(string, string, io.Reader) (*http.Request, error) {
	return nil, errNotImplemented
}

func (s stubHTTPClient) NewRequestWithContext(context.Context, string, string, io.Reader) (*http.Request, error) {
	return nil, errNotImplemented
}

func TestIsLivePropagatesOriginalError(t *testing.T) {
	expectedErr := &model.APIError{Status: -3, Message: "network error"}
	client := ReductClient{HTTPClient: stubHTTPClient{headErr: expectedErr}}

	_, err := client.IsLive(context.Background())

	require.Same(t, expectedErr, err)
}

func TestGetInfoPropagatesOriginalError(t *testing.T) {
	expectedErr := &model.APIError{Status: -3, Message: "network error"}
	client := ReductClient{HTTPClient: stubHTTPClient{getErr: expectedErr}}

	_, err := client.GetInfo(context.Background())

	require.Same(t, expectedErr, err)
}

type queryLinkStubHTTPClient struct {
	postCalled bool
	postPath   string
	postBody   any
}

func (s *queryLinkStubHTTPClient) Post(_ context.Context, path string, requestBody, responseData any) error {
	s.postCalled = true
	s.postPath = path
	s.postBody = requestBody

	if responseData != nil {
		payload, err := json.Marshal(map[string]string{"link": "http://localhost/fake-link"})
		if err != nil {
			return err
		}
		return json.Unmarshal(payload, responseData)
	}

	return nil
}

func (s *queryLinkStubHTTPClient) Put(context.Context, string, any, any) error {
	return errNotImplemented
}
func (s *queryLinkStubHTTPClient) Patch(context.Context, string, any, any) error {
	return errNotImplemented
}
func (s *queryLinkStubHTTPClient) Get(context.Context, string, any) error { return errNotImplemented }
func (s *queryLinkStubHTTPClient) Head(context.Context, string) error     { return errNotImplemented }
func (s *queryLinkStubHTTPClient) Delete(context.Context, string) error   { return errNotImplemented }
func (s *queryLinkStubHTTPClient) Do(*http.Request) (*http.Response, error) {
	return nil, errNotImplemented
}
func (s *queryLinkStubHTTPClient) NewRequest(string, string, io.Reader) (*http.Request, error) {
	return nil, errNotImplemented
}
func (s *queryLinkStubHTTPClient) NewRequestWithContext(context.Context, string, string, io.Reader) (*http.Request, error) {
	return nil, errNotImplemented
}

func TestCreateQueryLinkUsesRecordIdentityArguments(t *testing.T) {
	httpClient := &queryLinkStubHTTPClient{}
	bucket := Bucket{
		HTTPClient: httpClient,
		Name:       "test-bucket",
	}

	recordTimestamp := int64(123456789)
	options := NewQueryLinkOptionsBuilder().
		WithRecordEntry("source-entry").
		WithRecordTimestamp(recordTimestamp).
		WithRecordIndex(2).
		Build()

	_, err := bucket.CreateQueryLink(context.Background(), "entry-1", options)
	require.NoError(t, err)
	require.True(t, httpClient.postCalled)
	assert.Equal(t, "/links/entry-1_2", httpClient.postPath)

	rawPayload, err := json.Marshal(httpClient.postBody)
	require.NoError(t, err)

	payload := map[string]any{}
	err = json.Unmarshal(rawPayload, &payload)
	require.NoError(t, err)

	assert.Equal(t, "test-bucket", payload["bucket"])
	assert.Equal(t, "entry-1", payload["entry"])
	assert.Equal(t, "source-entry", payload["record_entry"])
	assert.Equal(t, float64(recordTimestamp), payload["record_timestamp"])
	assert.Equal(t, float64(2), payload["index"])
}

func TestCreateQueryLinkRejectsNegativeRecordIndex(t *testing.T) {
	httpClient := &queryLinkStubHTTPClient{}
	bucket := Bucket{
		HTTPClient: httpClient,
		Name:       "test-bucket",
	}

	options := NewQueryLinkOptionsBuilder().WithRecordIndex(-1).Build()
	_, err := bucket.CreateQueryLink(context.Background(), "entry-1", options)

	require.EqualError(t, err, "record index must be a non-negative integer")
	assert.False(t, httpClient.postCalled)
}
