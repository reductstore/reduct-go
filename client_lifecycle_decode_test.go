package reductgo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/reductstore/reduct-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLifecycleDecodesLifecycleInfoFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/lifecycles/test-lifecycle", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Reduct-API", "v1.3")
		_, err := w.Write([]byte(`{"info":{"name":"test-lifecycle","is_provisioned":true,"is_running":false,"type":"compress","mode":"enabled","last_run":"2026-06-15T06:53:37Z"},"settings":{"bucket":"bucket-1","older_than":"1h","mode":"enabled"}}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := ReductClient{HTTPClient: httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: time.Second})}

	lifecycle, err := client.GetLifecycle(context.Background(), "test-lifecycle")
	require.NoError(t, err)
	require.NotNil(t, lifecycle.Info)
	require.NotNil(t, lifecycle.Settings)

	expected := time.Date(2026, 6, 15, 6, 53, 37, 0, time.UTC)
	assert.Equal(t, "test-lifecycle", lifecycle.Info.Name)
	assert.Equal(t, model.LifecycleTypeCompress, lifecycle.Info.LifecycleType)
	assert.Equal(t, model.LifecycleModeEnabled, lifecycle.Info.Mode)
	require.NotNil(t, lifecycle.Info.LastRun)
	assert.True(t, lifecycle.Info.LastRun.Equal(expected))
	assert.Equal(t, "bucket-1", lifecycle.Settings.Bucket)
}

func TestGetLifecyclesDecodesLifecycleInfoFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/lifecycles", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Reduct-API", "v1.3")
		_, err := w.Write([]byte(`{"lifecycles":[{"name":"test-lifecycle","is_provisioned":true,"is_running":false,"type":"delete","mode":"dry_run"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := ReductClient{HTTPClient: httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: time.Second})}

	lifecycles, err := client.GetLifecycles(context.Background())
	require.NoError(t, err)
	require.Len(t, lifecycles, 1)

	assert.Equal(t, "test-lifecycle", lifecycles[0].Name)
	assert.Equal(t, model.LifecycleTypeDelete, lifecycles[0].LifecycleType)
	assert.Equal(t, model.LifecycleModeDryRun, lifecycles[0].Mode)
	assert.Nil(t, lifecycles[0].LastRun)
}
