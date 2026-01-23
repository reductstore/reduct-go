package reductgo

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/model"

	"github.com/stretchr/testify/assert"
)

func TestBatchReading(t *testing.T) {
	exists, err := mainTestBucket.CheckExists(context.Background())
	assert.NoError(t, err)
	assert.True(t, exists)

	// First, let's write some test data using batch write
	batch := mainTestBucket.BeginWriteBatch(context.Background(), "batch-test-entry")

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
	errMap, err := batch.Write(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, errMap, "All records should be written successfully")

	// Now test reading the batch
	t.Run("read batch records", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		reader, err := mainTestBucket.BeginRead(ctx, "batch-test-entry", nil)
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

		reader, err := mainTestBucket.BeginMetadataRead(ctx, "batch-test-entry", nil)
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
	batch := mainTestBucket.BeginWriteBatch(context.Background(), "batch-test-entry")

	// Add a few records with different timestamps
	data1 := map[string]any{"key": "value1"}
	data2 := map[string]any{"key": "value2"}

	jsonData1, _ := json.Marshal(data1) //nolint:errcheck //not needed
	jsonData2, _ := json.Marshal(data2) //nolint:errcheck //not needed

	ts1 := time.Now().UTC().UnixMicro()
	ts2 := ts1 + 1000

	batch.Add(ts1, jsonData1, "application/json", map[string]any{"label1": "value1"})
	batch.Add(ts2, jsonData2, "application/json", map[string]any{"label2": "value2"})
	errMap, err := batch.Write(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, errMap, "All records should be written successfully")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// get the id of the last record
	queryResult, err := mainTestBucket.executeQuery(ctx, "batch-test-entry", nil)
	assert.NoError(t, err)
	id := queryResult.ID
	fetchResult, err := mainTestBucket.fetchAndParseBatchedRecords(
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
	batch := mainTestBucket.BeginWriteBatch(ctx, entry)
	assert.NotNil(t, batch)

	// Add records
	now := time.Now().UTC().UnixMicro()
	batch.Add(now, []byte("data1"), "text/plain", nil)
	batch.Add(now+1, []byte("data2"), "text/plain", map[string]any{"label": "value"})
	batch.Add(now+2, []byte("data3"), "text/plain", nil)

	// Write batch
	errMap, err := batch.Write(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errMap, "All records should be written successfully")

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

func TestRecordBatchWrite(t *testing.T) {
	ctx := context.Background()
	skipVersingLower(ctx, t, "1.18.0")

	batch := mainTestBucket.BeginWriteRecordBatch(ctx)
	assert.NotNil(t, batch)

	ts1 := time.Now().UTC().UnixMicro()
	ts2 := ts1 + 1000

	batch.Add("record-batch-entry-1", ts1, []byte("alpha"), "text/plain", map[string]any{"label": "a"})
	batch.Add("record-batch-entry-2", ts2, []byte("beta"), "", nil)

	errs, err := batch.Send(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errs)

	recordsEntry1, err := mainTestBucket.Query(ctx, "record-batch-entry-1", nil)
	assert.NoError(t, err)
	record1 := <-recordsEntry1.Records()
	assert.NotNil(t, record1)
	assert.Equal(t, ts1, record1.Time())
	assert.Equal(t, int64(5), record1.Size())
	assert.Equal(t, "text/plain", record1.ContentType())
	assert.Equal(t, "a", record1.Labels()["label"])
	content1, err := record1.Read()
	assert.NoError(t, err)
	assert.Equal(t, "alpha", string(content1))

	recordsEntry2, err := mainTestBucket.Query(ctx, "record-batch-entry-2", nil)
	assert.NoError(t, err)
	record2 := <-recordsEntry2.Records()
	assert.NotNil(t, record2)
	assert.Equal(t, ts2, record2.Time())
	assert.Equal(t, int64(4), record2.Size())
	assert.Equal(t, "application/octet-stream", record2.ContentType())
	assert.Empty(t, record2.Labels())
	content2, err := record2.Read()
	assert.NoError(t, err)
	assert.Equal(t, "beta", string(content2))
}

func TestRecordBatchWriteErrors(t *testing.T) {
	ctx := context.Background()
	skipVersingLower(ctx, t, "1.18.0")

	batch := mainTestBucket.BeginWriteRecordBatch(ctx)
	assert.NotNil(t, batch)

	entry := "record-batch-error-entry"
	ts := time.Now().UTC().UnixMicro()

	batch.Add(entry, ts, []byte("first"), "", nil)
	errs, err := batch.Send(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errs)

	batch.Clear()
	batch.Add(entry, ts, []byte("dup"), "", nil)
	errs, err = batch.Send(ctx)
	assert.NoError(t, err)

	entryErrors := errs[entry]
	if assert.NotNil(t, entryErrors) {
		assert.Equal(t, model.APIError{
			Status:  409,
			Message: fmt.Sprintf("A record with timestamp %d already exists", ts),
		}, entryErrors[ts])
	}
}

func TestRecordBatchUpdate(t *testing.T) {
	ctx := context.Background()
	skipVersingLower(ctx, t, "1.18.0")

	entry1 := "record-batch-update-entry-1"
	entry2 := "record-batch-update-entry-2"

	writeBatch := mainTestBucket.BeginWriteRecordBatch(ctx)
	assert.NotNil(t, writeBatch)

	ts1 := time.Now().UTC().UnixMicro()
	ts2 := ts1 + 1000

	writeBatch.Add(entry1, ts1, []byte("alpha"), "text/plain", map[string]any{"keep": "one", "remove": "gone"})
	writeBatch.Add(entry2, ts2, []byte("beta"), "text/plain", map[string]any{"keep": "two"})
	errs, err := writeBatch.Send(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errs)

	updateBatch := mainTestBucket.BeginUpdateRecordBatch(ctx)
	assert.NotNil(t, updateBatch)
	updateBatch.AddOnlyLabels(entry1, ts1, map[string]any{"keep": "one-updated", "remove": ""})
	updateBatch.AddOnlyLabels(entry2, ts2, map[string]any{"new": "added"})
	errs, err = updateBatch.Send(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errs)

	recordsEntry1, err := mainTestBucket.Query(ctx, entry1, nil)
	assert.NoError(t, err)
	record1 := <-recordsEntry1.Records()
	assert.NotNil(t, record1)
	assert.Equal(t, ts1, record1.Time())
	assert.Equal(t, int64(5), record1.Size())
	assert.Equal(t, "text/plain", record1.ContentType())
	content1, err := record1.Read()
	assert.NoError(t, err)
	assert.Equal(t, "alpha", string(content1))
	assert.Equal(t, "one-updated", record1.Labels()["keep"])
	_, ok := record1.Labels()["remove"]
	assert.False(t, ok)

	recordsEntry2, err := mainTestBucket.Query(ctx, entry2, nil)
	assert.NoError(t, err)
	record2 := <-recordsEntry2.Records()
	assert.NotNil(t, record2)
	assert.Equal(t, ts2, record2.Time())
	assert.Equal(t, int64(4), record2.Size())
	assert.Equal(t, "text/plain", record2.ContentType())
	content2, err := record2.Read()
	assert.NoError(t, err)
	assert.Equal(t, "beta", string(content2))
	assert.Equal(t, "two", record2.Labels()["keep"])
	assert.Equal(t, "added", record2.Labels()["new"])
}

func TestRecordBatchUpdateErrors(t *testing.T) {
	ctx := context.Background()
	skipVersingLower(ctx, t, "1.18.0")

	entry := "record-batch-update-error-entry"
	ts := time.Now().UTC().UnixMicro()

	writeBatch := mainTestBucket.BeginWriteRecordBatch(ctx)
	assert.NotNil(t, writeBatch)
	writeBatch.Add(entry, ts, []byte("alpha"), "text/plain", nil)
	errs, err := writeBatch.Send(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errs)

	updateBatch := mainTestBucket.BeginUpdateRecordBatch(ctx)
	assert.NotNil(t, updateBatch)
	updateBatch.AddOnlyLabels(entry, ts, map[string]any{"ok": "true"})
	updateBatch.AddOnlyLabels("missing-entry", ts, map[string]any{"bad": "true"})
	errs, err = updateBatch.Send(ctx)
	assert.NoError(t, err)

	entryErrors := errs["missing-entry"]
	if assert.NotNil(t, entryErrors) {
		apiErr, ok := entryErrors[ts]
		assert.True(t, ok)
		assert.Equal(t, 404, apiErr.Status)
	}
}

func TestRecordBatchRemove(t *testing.T) {
	ctx := context.Background()
	skipVersingLower(ctx, t, "1.18.0")

	entry1 := "record-batch-remove-entry-1"
	entry2 := "record-batch-remove-entry-2"

	writeBatch := mainTestBucket.BeginWriteRecordBatch(ctx)
	assert.NotNil(t, writeBatch)

	ts1 := time.Now().UTC().UnixMicro()
	ts2 := ts1 + 1000

	writeBatch.Add(entry1, ts1, []byte("alpha"), "text/plain", nil)
	writeBatch.Add(entry2, ts2, []byte("beta"), "text/plain", nil)
	errs, err := writeBatch.Send(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errs)

	removeBatch := mainTestBucket.BeginRemoveRecordBatch(ctx)
	assert.NotNil(t, removeBatch)
	removeBatch.AddOnlyTimestamp(entry1, ts1)
	removeBatch.AddOnlyTimestamp(entry2, ts2)
	errs, err = removeBatch.Send(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errs)

	_, err = mainTestBucket.BeginRead(ctx, entry1, &ts1)
	assert.Error(t, err)
	_, err = mainTestBucket.BeginRead(ctx, entry2, &ts2)
	assert.Error(t, err)
}

func TestRecordBatchRemoveErrors(t *testing.T) {
	ctx := context.Background()
	skipVersingLower(ctx, t, "1.18.0")

	entry := "record-batch-remove-error-entry"
	ts := time.Now().UTC().UnixMicro()

	writeBatch := mainTestBucket.BeginWriteRecordBatch(ctx)
	assert.NotNil(t, writeBatch)
	writeBatch.Add(entry, ts, []byte("alpha"), "text/plain", nil)
	errs, err := writeBatch.Send(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errs)

	removeBatch := mainTestBucket.BeginRemoveRecordBatch(ctx)
	assert.NotNil(t, removeBatch)
	removeBatch.AddOnlyTimestamp(entry, ts+1)
	removeBatch.AddOnlyTimestamp("missing-entry", ts)
	errs, err = removeBatch.Send(ctx)
	assert.NoError(t, err)

	missingEntryErrors := errs["missing-entry"]
	if assert.NotNil(t, missingEntryErrors) {
		apiErr, ok := missingEntryErrors[ts]
		assert.True(t, ok)
		assert.Equal(t, 404, apiErr.Status)
	}

	entryErrors := errs[entry]
	if assert.NotNil(t, entryErrors) {
		apiErr, ok := entryErrors[ts+1]
		assert.True(t, ok)
		assert.Equal(t, 404, apiErr.Status)
	}
}

func TestBatchUpdate(t *testing.T) {
	ctx := context.Background()
	entry := "test-batch-update"

	// First write some records
	batch := mainTestBucket.BeginWriteBatch(ctx, entry)
	now := time.Now().UTC().UnixMicro()
	batch.Add(now, []byte("data1"), "text/plain", nil)
	batch.Add(now+1, []byte("data2"), "text/plain", nil)
	errMap, err := batch.Write(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errMap, "All records should be written successfully")

	// Update labels
	updateBatch := mainTestBucket.BeginUpdateBatch(ctx, entry)
	updateBatch.AddOnlyLabels(now, map[string]any{"updated": "true"})
	errMap, err = updateBatch.Write(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errMap, "All records should be written successfully")

	queryOptions := NewQueryOptionsBuilder().
		WithStart(now).
		WithStop(now).
		Build()
	// Verify updates
	queryResult, err := mainTestBucket.Query(ctx, entry, &queryOptions)
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
	batch := mainTestBucket.BeginWriteBatch(ctx, entry)
	now := time.Now().UTC().UnixMicro()
	batch.Add(now, []byte("data1"), "text/plain", nil)
	batch.Add(now+1, []byte("data2"), "text/plain", nil)
	errMap, err := batch.Write(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errMap, "All records should be written successfully")

	// Remove records
	removeBatch := mainTestBucket.BeginRemoveBatch(ctx, entry)
	errMap, err = removeBatch.Write(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errMap, "All records should be removed successfully")

	// Verify removal
	queryOptions := NewQueryOptionsBuilder().
		WithStart(now).
		WithStop(now).
		Build()
	queryResult, err := mainTestBucket.Query(ctx, entry, &queryOptions)
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

	// Test batch with invalid timestamps
	batch := mainTestBucket.BeginWriteBatch(ctx, entry)
	tm := int64(1)
	batch.Add(tm, []byte("new"), "text/plain", nil)
	errMap, err := batch.Write(ctx)
	assert.NoError(t, err)
	assert.Empty(t, errMap, "All records should be written successfully")

	batch.Clear()
	batch.Add(tm, []byte("exists"), "text/plain", nil)
	errMap, err = batch.Write(ctx)

	assert.NoError(t, err)
	assert.Equal(t, errMap[tm], model.APIError{
		Status:  409,
		Message: fmt.Sprintf("A record with timestamp %d already exists", tm),
	})
}
