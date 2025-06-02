package reductgo

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	ctx := context.Background()
	entry := "test-query"

	// Write test data
	batch := mainTestBucket.BeginWriteBatch(ctx, entry)
	now := time.Now().UTC().UnixMicro()
	batch.Add(now+1, []byte("data0"), "application/json", map[string]any{"type": "test0"})
	batch.Add(now+2, []byte("data1"), "application/json", map[string]any{"type": "test1"})
	batch.Add(now+3, []byte("data2"), "application/json", map[string]any{"type": "test2"})
	err := batch.Write(ctx)
	assert.NoError(t, err)

	t.Run("Query All Records", func(t *testing.T) {
		queryResult, err := mainTestBucket.Query(ctx, entry, nil)
		assert.NoError(t, err)
		count := 0
		for record := range queryResult.Records() {
			assert.NoError(t, err)

			count++
			data, err := record.Read()
			fmt.Printf("time:%d  data:%s labels:%v\n", record.Time(), string(data), record.Labels())
			assert.NoError(t, err)
			if record.IsLast() {
				break
			}

		}
		assert.Equal(t, 3, count)
	})

	t.Run("Query with Time Range", func(t *testing.T) {
		start := now
		end := now + 3
		queryOptions := NewQueryOptionsBuilder().
			WithStart(start).
			WithStop(end).
			Build()
		queryResult, err := mainTestBucket.Query(ctx, entry, &queryOptions)
		assert.NoError(t, err)

		count := 0
		for record := range queryResult.Records() {
			count++

			if record.IsLast() {
				break
			}
		}
		assert.Equal(t, 2, count)
	})

	t.Run("Query with Invalid Time Range", func(t *testing.T) {
		start := now + 4
		end := now
		queryOptions := NewQueryOptionsBuilder().
			WithStart(start).
			WithStop(end).
			Build()
		queryResult, err := mainTestBucket.Query(ctx, entry, &queryOptions)
		assert.Error(t, err)
		for record := range queryResult.Records() {
			data, err := record.Read()
			assert.NoError(t, err)
			assert.Empty(t, data)
			if record.IsLast() {
				break
			}
		}

	})

	t.Run("Query with Labels", func(t *testing.T) {
		options := &QueryOptions{
			When: map[string]any{"&type": map[string]any{"$eq": "test1"}},
		}
		queryResult, err := mainTestBucket.Query(ctx, entry, options)
		assert.NoError(t, err)

		count := 0
		for record := range queryResult.Records() {
			count++
			assert.Equal(t, "test1", record.Labels()["type"])
			if record.IsLast() {
				break
			}
		}
		assert.Equal(t, 1, count)
	})

	t.Run("Query Head Only", func(t *testing.T) {
		options := &QueryOptions{
			Head: true,
		}
		queryResult, err := mainTestBucket.Query(ctx, entry, options)
		assert.NoError(t, err)

		for record := range queryResult.Records() {
			assert.NoError(t, err)
			// Head request should return empty data but valid metadata
			data, err := record.Read()
			assert.NoError(t, err)
			assert.Empty(t, data)
			assert.NotZero(t, record.Time())
			assert.NotEmpty(t, record.Labels())
			if record.IsLast() {
				break
			}
		}
	})
	t.Run("Query with last record big data", func(t *testing.T) {
		queryResult, err := mainTestBucket.Query(ctx, entry, nil)
		assert.NoError(t, err)
		for record := range queryResult.Records() {

			if record.IsLast() {
				// users can read the stream how they want
				stream := record.Stream()
				data, err := io.ReadAll(stream)
				fmt.Printf("time:%d  last record data:%s labels:%v\n", record.Time(), string(data), record.Labels())
				assert.NoError(t, err)
				break
			}
			data, err := record.Read()
			assert.NoError(t, err)
			fmt.Printf("time:%d  data:%s labels:%v\n", record.Time(), string(data), record.Labels())
		}
	})

	t.Run("Query with large last record (>10MB)", func(t *testing.T) {
		// Create a new entry for large data test
		entryLarge := "test-query-large"

		// Write some small records first
		batch := mainTestBucket.BeginWriteBatch(ctx, entryLarge)
		now := time.Now().UTC().UnixMicro()
		batch.Add(now+1, []byte("small1"), "text/plain", nil)
		batch.Add(now+2, []byte("small2"), "text/plain", nil)

		// Create a large record (12MB)
		largeSize := 12 * 1024 * 1024 // 12MB
		largeData := make([]byte, largeSize)
		for i := 0; i < largeSize; i++ {
			largeData[i] = byte(i % 256) // Fill with repeating pattern
		}

		// Add the large record as the last one
		batch.Add(now+3, largeData, "application/octet-stream", map[string]any{"size": "large"})

		// Write the batch
		err := batch.Write(ctx)
		assert.NoError(t, err)

		// Query and verify the records
		queryResult, err := mainTestBucket.Query(ctx, entryLarge, nil)
		assert.NoError(t, err)

		recordCount := 0
		var lastRecordSize int
		for record := range queryResult.Records() {
			recordCount++

			if record.IsLastInBatch() {
				// Verify the large record
				assert.Equal(t, int64(largeSize), record.Size())
				assert.Equal(t, "large", record.Labels()["size"])

				// Read using stream to handle large data efficiently
				stream := record.Stream()
				data, err := io.ReadAll(stream)
				assert.NoError(t, err)
				lastRecordSize = len(data)

				// Verify the content (check first and last few bytes)
				for i := 0; i < 1024; i++ { // Check first 1KB
					assert.Equal(t, byte(i%256), data[i])
				}
				for i := len(data) - 1024; i < len(data); i++ { // Check last 1KB
					assert.Equal(t, byte(i%256), data[i])
				}
			} else {
				// Verify small records
				data, err := record.Read()
				assert.NoError(t, err)
				assert.Contains(t, string(data), "small")
			}
		}

		assert.Equal(t, 3, recordCount)
		assert.Equal(t, largeSize, lastRecordSize)
	})
}

func TestRemoveQuery(t *testing.T) {
	ctx := context.Background()
	entry := "test-remove-query-entry"
	// Write test data
	batch := mainTestBucket.BeginWriteBatch(ctx, entry)
	now := time.Now().UTC().UnixMicro()
	batch.Add(now, []byte("data1"), "text/plain", map[string]any{"type": "test1"})
	batch.Add(now+1, []byte("data2"), "text/plain", map[string]any{"type": "test2"})
	batch.Add(now+1000, []byte("data3"), "text/plain", map[string]any{"type": "test3"})
	err := batch.Write(ctx)
	assert.NoError(t, err)

	t.Run("Remove by Time Range", func(t *testing.T) {
		start := now
		end := now + 2
		removed, err := mainTestBucket.RemoveQuery(ctx, entry, &QueryOptions{
			Start: start,
			Stop:  end,
		})
		assert.NoError(t, err)
		fmt.Printf("removed:%d\n", removed)
		assert.Equal(t, int64(2), removed)

		// Verify remaining record
		queryResult, err := mainTestBucket.Query(ctx, entry, nil)
		assert.NoError(t, err)

		count := 0
		for record := range queryResult.Records() {
			count++
			data, err := record.Read()
			assert.NoError(t, err)
			assert.Equal(t, "data3", string(data))
			if record.IsLast() {
				break
			}
		}
		assert.Equal(t, 1, count)
	})

	t.Run("Remove by Labels", func(t *testing.T) {
		// Write more test data
		entry := "test-remove-by-label"
		batch := mainTestBucket.BeginWriteBatch(ctx, entry)
		now := time.Now().UTC().UnixMicro()
		batch.Add(now, []byte("data3"), "text/plain", map[string]any{"type": "testExisting"})
		batch.Add(now+1, []byte("data4"), "text/plain", map[string]any{"type": "testRemoved"})
		batch.Add(now+2, []byte("data5"), "text/plain", map[string]any{"type": "testRemoved"})
		err := batch.Write(ctx)
		assert.NoError(t, err)

		options := &QueryOptions{
			When: map[string]any{"&type": map[string]any{"$eq": "testRemoved"}},
		}
		removed, err := mainTestBucket.RemoveQuery(ctx, entry, options)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), removed)

		// Verify remaining records
		queryResult, err := mainTestBucket.Query(ctx, entry, nil)
		assert.NoError(t, err)

		count := 0
		for record := range queryResult.Records() {
			count++
			assert.NotEqual(t, "testRemoved", record.Labels()["type"])
		}
		assert.Equal(t, 1, count)
	})
}
