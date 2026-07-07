package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplicationSettingsDstPrefixJSON(t *testing.T) {
	t.Run("MarshalWithDstPrefix", func(t *testing.T) {
		settings := ReplicationSettings{
			SrcBucket: "edge",
			DstBucket: "fleet",
			DstHost:   "https://central.example",
			DstPrefix: "robot-1",
		}

		data, err := json.Marshal(settings)
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(data, &payload))
		assert.Equal(t, "robot-1", payload["dst_prefix"])
	})

	t.Run("MarshalWithoutDstPrefix", func(t *testing.T) {
		settings := ReplicationSettings{
			SrcBucket: "edge",
			DstBucket: "fleet",
			DstHost:   "https://central.example",
		}

		data, err := json.Marshal(settings)
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(data, &payload))
		assert.NotContains(t, payload, "dst_prefix")
	})

	t.Run("UnmarshalWithDstPrefix", func(t *testing.T) {
		var settings ReplicationSettings
		err := json.Unmarshal([]byte(`{
			"src_bucket": "edge",
			"dst_bucket": "fleet",
			"dst_host": "https://central.example",
			"dst_prefix": "robot-1"
		}`), &settings)
		require.NoError(t, err)

		assert.Equal(t, "robot-1", settings.DstPrefix)
	})

	t.Run("UnmarshalWithoutDstPrefix", func(t *testing.T) {
		var settings ReplicationSettings
		err := json.Unmarshal([]byte(`{
			"src_bucket": "edge",
			"dst_bucket": "fleet",
			"dst_host": "https://central.example"
		}`), &settings)
		require.NoError(t, err)

		assert.Empty(t, settings.DstPrefix)
	})
}
