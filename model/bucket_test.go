package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusSerialization(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "READY status",
			status:   StatusReady,
			expected: `"status":"READY"`,
		},
		{
			name:     "DELETING status",
			status:   StatusDeleting,
			expected: `"status":"DELETING"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := BucketInfo{Name: "test", Status: tt.status}
			data, err := json.Marshal(info)
			assert.NoError(t, err)
			assert.Contains(t, string(data), tt.expected)
		})
	}

	t.Run("omitempty excludes empty status", func(t *testing.T) {
		info := BucketInfo{Name: "test"}
		data, err := json.Marshal(info)
		assert.NoError(t, err)
		assert.NotContains(t, string(data), `"status"`)
	})
}

