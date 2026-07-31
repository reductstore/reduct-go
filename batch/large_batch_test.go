package batch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// largeBatchRecordCount exceeds the 100-slot buffered channel the batch readers
// used to hand records through. The reader filled that channel while the caller
// was still blocked waiting for the reader to finish, so any batch with more
// than 100 records deadlocked.
const largeBatchRecordCount = 150

const recordPayloadSize = 8

// fetchWithinDeadline runs fetch on a goroutine so a regression shows up as a
// test failure instead of hanging the suite until the global timeout.
func fetchWithinDeadline(t *testing.T, fetch func() ([]*Record, error)) []*Record {
	t.Helper()

	type result struct {
		records []*Record
		err     error
	}

	done := make(chan result, 1)
	go func() {
		records, err := fetch()
		done <- result{records: records, err: err}
	}()

	select {
	case got := <-done:
		require.NoError(t, got.err)
		return got.records
	case <-time.After(10 * time.Second):
		t.Fatal("reading a batch larger than the record channel deadlocked")
		return nil
	}
}

func TestReadBatchedRecordsHandlesBatchLargerThanChannelBuffer(t *testing.T) {
	payload := make([]byte, recordPayloadSize*largeBatchRecordCount)
	for i := range payload {
		payload[i] = byte('a' + i%26)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.20")
		for i := 0; i < largeBatchRecordCount; i++ {
			w.Header().Set(fmt.Sprintf("x-reduct-time-%d", i*1000), strconv.Itoa(recordPayloadSize)+",text/plain")
		}
		w.Header().Set("x-reduct-last", "true")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(payload); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: 10 * time.Second})

	records := fetchWithinDeadline(t, func() ([]*Record, error) {
		return readBatchedRecords(context.Background(), client, "bucket", "entry", 1, false)
	})

	require.Len(t, records, largeBatchRecordCount)
	assertRecordsReadable(t, records, payload)
	assert.True(t, records[largeBatchRecordCount-1].Last, "final record should end the query")
}

func TestReadBatchedRecordsV2HandlesBatchLargerThanChannelBuffer(t *testing.T) {
	payload := make([]byte, recordPayloadSize*largeBatchRecordCount)
	for i := range payload {
		payload[i] = byte('a' + i%26)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.20")
		w.Header().Set(entriesHeader, "entry")
		w.Header().Set(startTSHeader, "0")
		for i := 0; i < largeBatchRecordCount; i++ {
			value := strconv.Itoa(recordPayloadSize)
			if i == 0 {
				// The first record of an entry carries the content type.
				value += ",text/plain"
			}
			w.Header().Set(fmt.Sprintf("%s0-%d", headerPrefix, i*1000), value)
		}
		w.Header().Set(lastHeader, "true")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(payload); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: 10 * time.Second})

	records := fetchWithinDeadline(t, func() ([]*Record, error) {
		return readBatchedRecordsV2(context.Background(), client, "bucket", 1, false)
	})

	require.Len(t, records, largeBatchRecordCount)
	assertRecordsReadable(t, records, payload)
	assert.True(t, records[largeBatchRecordCount-1].Last, "final record should end the query")
}

// assertRecordsReadable checks every record reports and yields its own slice of
// the batch payload, which is what the shared batch buffer must preserve.
func assertRecordsReadable(t *testing.T, records []*Record, payload []byte) {
	t.Helper()

	for i, rec := range records {
		assert.Equal(t, int64(recordPayloadSize), rec.Size, "record %d size", i)

		data, err := io.ReadAll(rec.Body)
		require.NoError(t, err, "record %d body", i)

		want := payload[i*recordPayloadSize : (i+1)*recordPayloadSize]
		assert.Equal(t, want, data, "record %d payload", i)
	}
}
