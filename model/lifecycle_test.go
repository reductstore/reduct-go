package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifecycleSettingsProcessingIntervalJSON(t *testing.T) {
	t.Run("MarshalWithProcessingInterval", func(t *testing.T) {
		settings := LifecycleSettings{
			LifecycleType:      LifecycleTypeCompress,
			Bucket:             "measurements",
			OlderThan:          "30d",
			Interval:           "1h",
			ProcessingInterval: "12h",
			Mode:               LifecycleModeEnabled,
		}

		raw, err := json.Marshal(settings)
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(raw, &payload))
		assert.Equal(t, "12h", payload["processing_interval"])
	})

	t.Run("MarshalOmitsEmptyProcessingInterval", func(t *testing.T) {
		settings := LifecycleSettings{
			LifecycleType: LifecycleTypeDelete,
			Bucket:        "measurements",
			OlderThan:     "30d",
		}

		raw, err := json.Marshal(settings)
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(raw, &payload))
		assert.NotContains(t, payload, "processing_interval")
	})

	t.Run("UnmarshalFullLifecycleInfoWithProcessingInterval", func(t *testing.T) {
		raw := []byte(`{
			"info": {
				"name": "compact-fast",
				"type": "compress",
				"mode": "enabled",
				"is_provisioned": false,
				"is_running": false
			},
			"settings": {
				"type": "compress",
				"bucket": "measurements",
				"older_than": "30d",
				"interval": "1h",
				"processing_interval": "12h",
				"mode": "enabled"
			}
		}`)

		var lifecycle FullLifecycleInfo
		require.NoError(t, json.Unmarshal(raw, &lifecycle))
		require.NotNil(t, lifecycle.Settings)
		assert.Equal(t, "12h", lifecycle.Settings.ProcessingInterval)
	})
}
