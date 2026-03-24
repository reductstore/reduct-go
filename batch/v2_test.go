package batch

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeaderList(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		want        []string
		errContains string
	}{
		{
			name:   "empty header returns nil",
			header: "",
			want:   nil,
		},
		{
			name:   "whitespace only header returns nil",
			header: "   ",
			want:   nil,
		},
		{
			name:   "single entry",
			header: "entry1",
			want:   []string{"entry1"},
		},
		{
			name:   "multiple entries",
			header: "entry1,entry2,entry3",
			want:   []string{"entry1", "entry2", "entry3"},
		},
		{
			name:   "entries with spaces",
			header: " entry1 , entry2 , entry3 ",
			want:   []string{"entry1", "entry2", "entry3"},
		},
		{
			name:   "entries with percent encoding",
			header: "entry%201,entry%202",
			want:   []string{"entry 1", "entry 2"},
		},
		{
			name:        "empty entry in list",
			header:      "entry1,,entry3",
			errContains: "invalid entries/labels header",
		},
		{
			name:        "invalid percent encoding",
			header:      "entry%ZZ",
			errContains: "invalid URL escape",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeaderList(tt.header)

			if tt.errContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFetchAndParseV2_ContinueQueryOnEmptyBatch(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/io/bucket/read", r.URL.Path)

		w.Header().Set("X-Reduct-API", "v1.3")
		w.Header().Set("x-reduct-entries", "entry")
		w.Header().Set("x-reduct-start-ts", "100")

		if atomic.AddInt32(&requests, 1) == 1 {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("x-reduct-0-0", "5,text/plain")
		w.Header().Set("x-reduct-last", "true")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("hello"))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	records, err := FetchAndParseV2(ctx, client, "bucket", 1, true, 10*time.Millisecond, false)
	require.NoError(t, err)

	select {
	case rec, ok := <-records:
		require.True(t, ok, "records channel closed before receiving data")
		require.NotNil(t, rec)

		assert.Equal(t, "entry", rec.Entry)
		assert.True(t, rec.Last)

		body, readErr := io.ReadAll(rec.Body)
		require.NoError(t, readErr)
		require.NoError(t, rec.Body.Close())
		assert.Equal(t, "hello", string(body))
	case <-ctx.Done():
		t.Fatal("timed out waiting for record after empty batch")
	}

	assert.GreaterOrEqual(t, atomic.LoadInt32(&requests), int32(2), "expected at least two read requests")
}
