package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucketInfoStatusSerialization(t *testing.T) {
	t.Run("serialize BucketInfo with READY status", func(t *testing.T) {
		info := BucketInfo{
			Name:       "test-bucket",
			EntryCount: 10,
			Size:       1024,
			Status:     StatusReady,
		}

		data, err := json.Marshal(info)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"status":"READY"`)
	})

	t.Run("serialize BucketInfo with DELETING status", func(t *testing.T) {
		info := BucketInfo{
			Name:       "test-bucket",
			EntryCount: 10,
			Size:       1024,
			Status:     StatusDeleting,
		}

		data, err := json.Marshal(info)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"status":"DELETING"`)
	})

	t.Run("serialize BucketInfo without status (empty)", func(t *testing.T) {
		info := BucketInfo{
			Name:       "test-bucket",
			EntryCount: 10,
			Size:       1024,
		}

		data, err := json.Marshal(info)
		assert.NoError(t, err)
		// omitempty should exclude the status field when empty
		assert.NotContains(t, string(data), `"status"`)
	})

	t.Run("deserialize BucketInfo with status", func(t *testing.T) {
		jsonData := `{"name":"test-bucket","entry_count":10,"size":1024,"status":"DELETING"}`
		var info BucketInfo
		err := json.Unmarshal([]byte(jsonData), &info)
		assert.NoError(t, err)
		assert.Equal(t, "test-bucket", info.Name)
		assert.Equal(t, StatusDeleting, info.Status)
	})

	t.Run("deserialize BucketInfo without status", func(t *testing.T) {
		jsonData := `{"name":"test-bucket","entry_count":10,"size":1024}`
		var info BucketInfo
		err := json.Unmarshal([]byte(jsonData), &info)
		assert.NoError(t, err)
		assert.Equal(t, "test-bucket", info.Name)
		assert.Equal(t, Status(""), info.Status)
	})
}

func TestEntryInfoStatusSerialization(t *testing.T) {
	t.Run("serialize EntryInfo with READY status", func(t *testing.T) {
		info := EntryInfo{
			Name:        "test-entry",
			RecordCount: 100,
			Size:        2048,
			Status:      StatusReady,
		}

		data, err := json.Marshal(info)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"status":"READY"`)
	})

	t.Run("serialize EntryInfo with DELETING status", func(t *testing.T) {
		info := EntryInfo{
			Name:        "test-entry",
			RecordCount: 100,
			Size:        2048,
			Status:      StatusDeleting,
		}

		data, err := json.Marshal(info)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"status":"DELETING"`)
	})

	t.Run("serialize EntryInfo without status (empty)", func(t *testing.T) {
		info := EntryInfo{
			Name:        "test-entry",
			RecordCount: 100,
			Size:        2048,
		}

		data, err := json.Marshal(info)
		assert.NoError(t, err)
		// omitempty should exclude the status field when empty
		assert.NotContains(t, string(data), `"status"`)
	})

	t.Run("deserialize EntryInfo with status", func(t *testing.T) {
		jsonData := `{"name":"test-entry","block_count":5,"record_count":100,"size":2048,"status":"READY"}`
		var info EntryInfo
		err := json.Unmarshal([]byte(jsonData), &info)
		assert.NoError(t, err)
		assert.Equal(t, "test-entry", info.Name)
		assert.Equal(t, StatusReady, info.Status)
	})

	t.Run("deserialize EntryInfo without status", func(t *testing.T) {
		jsonData := `{"name":"test-entry","block_count":5,"record_count":100,"size":2048}`
		var info EntryInfo
		err := json.Unmarshal([]byte(jsonData), &info)
		assert.NoError(t, err)
		assert.Equal(t, "test-entry", info.Name)
		assert.Equal(t, Status(""), info.Status)
	})
}

func TestStatusConstants(t *testing.T) {
	t.Run("status constants are defined correctly", func(t *testing.T) {
		assert.Equal(t, Status("READY"), StatusReady)
		assert.Equal(t, Status("DELETING"), StatusDeleting)
	})
}
