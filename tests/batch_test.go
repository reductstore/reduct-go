package tests

import (
	"context"
	"encoding/json"
	"fmt"
	reductgo "reduct-go"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBatchReading(t *testing.T) {
	exists, err := mainTestBucket.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)

	// First, let's write some test data using batch write
	batch := mainTestBucket.BeginWriteBatch("batch-test-entry")

	// Add a few records with different timestamps
	data1 := map[string]any{"key": "value1"}
	data2 := map[string]any{"key": "value2"}

	jsonData1, _ := json.Marshal(data1) //nolint:errcheck //not needed
	jsonData2, _ := json.Marshal(data2) //nolint:errcheck //not needed

	ts1 := time.Now().UTC().UnixMicro()
	ts2 := ts1 + 1000

	batch.Add(ts1, jsonData1, "application/json", map[string]any{"label1": "value1"})
	batch.Add(ts2, jsonData2, "application/json", map[string]any{"label2": "value2"})

	// Write the batch
	err = batch.Write(context.Background())
	assert.NoError(t, err)

	// Now test reading the batch
	t.Run("read batch records", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		reader, err := mainTestBucket.BeginRead(ctx, "batch-test-entry", nil, false)
		assert.NoError(t, err)
		assert.NotNil(t, reader)

		// Read and verify the data
		data, err := reader.Read()
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify record metadata
		assert.Equal(t, "application/json", reader.ContentType())
		assert.NotZero(t, reader.Time())
		assert.NotZero(t, reader.Size())
	})

	t.Run("read batch records with head request", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		reader, err := mainTestBucket.BeginRead(ctx, "batch-test-entry", nil, true)
		assert.NoError(t, err)
		assert.NotNil(t, reader)

		// For head request, we should still get metadata but no content
		assert.Equal(t, "application/json", reader.ContentType())
		assert.NotZero(t, reader.Time())
		assert.NotZero(t, reader.Size())
	})
}

func TestFetchAndParseBatchedRecords(t *testing.T) {
	exists, err := mainTestBucket.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)
	// First, let's write some test data using batch write
	batch := mainTestBucket.BeginWriteBatch("batch-test-entry")

	// Add a few records with different timestamps
	data1 := map[string]any{"key": "value1"}
	data2 := map[string]any{"key": "value2"}

	jsonData1, _ := json.Marshal(data1) //nolint:errcheck //not needed
	jsonData2, _ := json.Marshal(data2) //nolint:errcheck //not needed

	ts1 := time.Now().UTC().UnixMicro()
	ts2 := ts1 + 1000

	batch.Add(ts1, jsonData1, "application/json", map[string]any{"label1": "value1"})
	batch.Add(ts2, jsonData2, "application/json", map[string]any{"label2": "value2"})
	err = batch.Write(context.Background())
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// get the id of the last record
	queryResult, err := mainTestBucket.ExecuteQuery(ctx, "batch-test-entry", nil)
	assert.NoError(t, err)
	id := queryResult.ID
	fetchResult, err := mainTestBucket.FetchAndParseBatchedRecords(
		ctx,
		"batch-test-entry",
		id,
		true,
		5*time.Second,
		false,
	)
	assert.NoError(t, err)

	for record := range fetchResult.Records() {
		// process each record
		fmt.Println(record)
	}
}

func TestBatchWrite(t *testing.T) {
	ctx := context.Background()
	entry := "test-batch-write"

	// Create a batch
	batch := mainTestBucket.BeginWriteBatch(entry)
	assert.NotNil(t, batch)

	// Add records
	now := time.Now().UTC().UnixMicro()
	batch.Add(now, []byte("data1"), "text/plain", nil)
	batch.Add(now+1, []byte("data2"), "text/plain", map[string]any{"label": "value"})
	batch.Add(now+2, []byte("data3"), "text/plain", nil)

	// Write batch
	err := batch.Write(ctx)
	assert.NoError(t, err)

	// Verify records
	queryResult, err := mainTestBucket.Query(ctx, entry, nil)
	assert.NoError(t, err)

	count := 0
	for record := range queryResult.Records() {
		count++
		data, err := record.Read()
		assert.NoError(t, err)
		assert.Contains(t, string(data), "data")
		if record.Time() == now+1 {
			assert.Equal(t, "value", record.Labels()["label"])
		}
	}
	assert.Equal(t, 3, count)
}

func TestBatchUpdate(t *testing.T) {
	ctx := context.Background()
	entry := "test-batch-update"

	// First write some records
	batch := mainTestBucket.BeginWriteBatch(entry)
	now := time.Now().UTC().UnixMicro()
	batch.Add(now, []byte("data1"), "text/plain", nil)
	batch.Add(now+1, []byte("data2"), "text/plain", nil)
	err := batch.Write(ctx)
	assert.NoError(t, err)

	// Update labels
	updateBatch := mainTestBucket.BeginUpdateBatch(entry)
	updateBatch.AddOnlyLabels(now, map[string]any{"updated": "true"})
	err = updateBatch.Write(ctx)
	assert.NoError(t, err)

	// Verify updates
	queryResult, err := mainTestBucket.Query(ctx, entry, &reductgo.QueryOptions{
		Start: &now,
		Stop:  &now,
	})
	assert.NoError(t, err)

	for record := range queryResult.Records() {
		assert.Equal(t, "true", record.Labels()["updated"])
		break
	}
}

func TestBatchRemove(t *testing.T) {
	ctx := context.Background()
	entry := "test-batch-remove"

	// First write some records
	batch := mainTestBucket.BeginWriteBatch(entry)
	now := time.Now().UTC().UnixMicro()
	batch.Add(now, []byte("data1"), "text/plain", nil)
	batch.Add(now+1, []byte("data2"), "text/plain", nil)
	err := batch.Write(ctx)
	assert.NoError(t, err)

	// Remove records
	removeBatch := mainTestBucket.BeginRemoveBatch(entry)
	err = removeBatch.Write(ctx)
	assert.NoError(t, err)

	// Verify removal
	queryResult, err := mainTestBucket.Query(ctx, entry, &reductgo.QueryOptions{
		Start: &now,
		Stop:  &now,
	})
	assert.NoError(t, err)

	count := 0
	for range queryResult.Records() {
		count++
	}
	assert.Equal(t, 0, count)

}

func TestBatchErrors(t *testing.T) {
	ctx := context.Background()
	entry := "test-batch-errors"

	// Test empty batch
	batch := mainTestBucket.BeginWriteBatch(entry)
	err := batch.Write(ctx)
	assert.Error(t, err)

	// Test batch with invalid timestamps
	batch = mainTestBucket.BeginWriteBatch(entry)
	tm := -1
	batch.Add(int64(tm), []byte("invalid"), "text/plain", nil)
	err = batch.Write(ctx)
	assert.Error(t, err)
}
