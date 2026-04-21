package reductgo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/reductstore/reduct-go/model"
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
