package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifecycleInfoUnmarshal(t *testing.T) {
	t.Run("with last_run", func(t *testing.T) {
		payload := []byte(`{"name":"test-lifecycle","is_provisioned":true,"is_running":false,"type":"compress","mode":"enabled","last_run":"2026-06-15T06:53:37Z"}`)

		var info LifecycleInfo
		require.NoError(t, json.Unmarshal(payload, &info))

		expected := time.Date(2026, 6, 15, 6, 53, 37, 0, time.UTC)
		assert.Equal(t, "test-lifecycle", info.Name)
		assert.Equal(t, LifecycleTypeCompress, info.LifecycleType)
		assert.Equal(t, LifecycleModeEnabled, info.Mode)
		require.NotNil(t, info.LastRun)
		assert.True(t, info.LastRun.Equal(expected))
	})

	t.Run("without last_run", func(t *testing.T) {
		payload := []byte(`{"name":"test-lifecycle","is_provisioned":true,"is_running":false,"type":"delete","mode":"enabled"}`)

		var info LifecycleInfo
		require.NoError(t, json.Unmarshal(payload, &info))

		assert.Equal(t, LifecycleTypeDelete, info.LifecycleType)
		assert.Nil(t, info.LastRun)
	})
}
