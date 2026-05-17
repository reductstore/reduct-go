package batch

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchAndParse_FirstBatchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.3")
		w.Header().Set("X-Reduct-Error", `Invalid SQL: SQL error: ParserError("bad query")`)
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: time.Second})

	_, _, err := FetchAndParse(context.Background(), client, "bucket", "entry", 1, false, time.Second, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid SQL")
}

func TestFetchAndParse_SubsequentBatchError(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.3")

		if atomic.AddInt32(&requests, 1) == 1 {
			// First batch: one record, not last
			ts := int64(100)
			w.Header().Set(fmt.Sprintf("x-reduct-time-%d", ts), "5,text/plain")
			w.Header().Set("x-reduct-last", "false")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("hello"))
			require.NoError(t, err)
		} else {
			// Second batch: server error
			w.Header().Set("X-Reduct-Error", "mid-stream failure")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: time.Second})

	records, errCh, err := FetchAndParse(context.Background(), client, "bucket", "entry", 1, false, time.Second, false)
	require.NoError(t, err)

	for rec := range records {
		_ = rec
	}

	err = <-errCh
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mid-stream failure")
}
