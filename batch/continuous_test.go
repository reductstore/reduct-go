package batch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const emptyStartPayload = "hello"

// emptyThenDataServer answers the first read with 204 (no records yet) and
// every later read with a single final record.
func emptyThenDataServer(t *testing.T, v2 bool) *httptest.Server {
	t.Helper()

	var calls atomic.Int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.20")
		if calls.Add(1) == 1 {
			w.Header().Set("x-reduct-error", "No content")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if v2 {
			w.Header().Set(entriesHeader, "entry")
			w.Header().Set(startTSHeader, "0")
			w.Header().Set(headerPrefix+"0-1000", strconv.Itoa(len(emptyStartPayload))+",text/plain")
			w.Header().Set(lastHeader, "true")
		} else {
			w.Header().Set("x-reduct-time-1000", strconv.Itoa(len(emptyStartPayload))+",text/plain")
			w.Header().Set("x-reduct-last", "true")
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(emptyStartPayload)); err != nil {
			panic(err)
		}
	}))
}

// collectWithinDeadline drains a query, failing rather than hanging if the
// records channel never produces anything.
func collectWithinDeadline(t *testing.T, records <-chan *Record) int {
	t.Helper()

	done := make(chan int, 1)
	go func() {
		count := 0
		for rec := range records {
			count++
			if rec.Last {
				break
			}
		}
		done <- count
	}()

	select {
	case count := <-done:
		return count
	case <-time.After(10 * time.Second):
		t.Fatal("continuous query that started with no data never delivered a record")
		return 0
	}
}

// A continuous query whose first read returns 204 must keep polling. The
// readers used to hand the first batch back as a nil channel in that case, and
// ranging over a nil channel blocks forever.
func TestFetchAndParseContinuesWhenFirstBatchIsEmpty(t *testing.T) {
	server := emptyThenDataServer(t, false)
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: 10 * time.Second})

	records, errCh, err := FetchAndParse(context.Background(), client, "bucket", "entry", 1, true, 10*time.Millisecond, false)
	require.NoError(t, err)

	assert.Equal(t, 1, collectWithinDeadline(t, records))
	assert.NoError(t, drainErr(errCh))
}

func TestFetchAndParseV2ContinuesWhenFirstBatchIsEmpty(t *testing.T) {
	server := emptyThenDataServer(t, true)
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: 10 * time.Second})

	records, errCh, err := FetchAndParseV2(context.Background(), client, "bucket", 1, true, 10*time.Millisecond, false)
	require.NoError(t, err)

	assert.Equal(t, 1, collectWithinDeadline(t, records))
	assert.NoError(t, drainErr(errCh))
}

func drainErr(errCh <-chan error) error {
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// A non-continuous query that finds no records must finish immediately.
func TestFetchAndParseStopsWhenFirstBatchIsEmptyAndNotContinuous(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.20")
		w.Header().Set("x-reduct-error", "No content")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: 10 * time.Second})

	records, errCh, err := FetchAndParse(context.Background(), client, "bucket", "entry", 1, false, time.Second, false)
	require.NoError(t, err)

	assert.Equal(t, 0, collectWithinDeadline(t, records))
	assert.NoError(t, drainErr(errCh))
}
