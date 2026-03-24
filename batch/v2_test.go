package batch

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
)

func TestParseHeaderList(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		want        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "empty header returns nil",
			header:  "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "whitespace only header returns nil",
			header:  "   ",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "single entry",
			header:  "entry1",
			want:    []string{"entry1"},
			wantErr: false,
		},
		{
			name:    "multiple entries",
			header:  "entry1,entry2,entry3",
			want:    []string{"entry1", "entry2", "entry3"},
			wantErr: false,
		},
		{
			name:    "entries with spaces",
			header:  " entry1 , entry2 , entry3 ",
			want:    []string{"entry1", "entry2", "entry3"},
			wantErr: false,
		},
		{
			name:    "entries with percent encoding",
			header:  "entry%201,entry%202",
			want:    []string{"entry 1", "entry 2"},
			wantErr: false,
		},
		{
			name:        "empty entry in list",
			header:      "entry1,,entry3",
			want:        nil,
			wantErr:     true,
			errContains: "invalid entries/labels header",
		},
		{
			name:        "invalid percent encoding",
			header:      "entry%ZZ",
			want:        nil,
			wantErr:     true,
			errContains: "invalid URL escape",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeaderList(tt.header)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseHeaderList() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseHeaderList() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("parseHeaderList() unexpected error = %v", err)
				return
			}
			if !equalStringSlices(got, tt.want) {
				t.Errorf("parseHeaderList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFetchAndParseV2_ContinueQueryOnEmptyBatch(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/io/bucket/read" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

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
		if _, err := w.Write([]byte("hello")); err != nil {
			t.Fatalf("failed to write response body: %v", err)
		}
	}))
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	records, err := FetchAndParseV2(ctx, client, "bucket", 1, true, 10*time.Millisecond, false)
	if err != nil {
		t.Fatalf("FetchAndParseV2() error = %v", err)
	}

	select {
	case rec, ok := <-records:
		if !ok {
			t.Fatal("records channel closed before receiving data")
		}
		if rec == nil {
			t.Fatal("record is nil")
		}
		if rec.Entry != "entry" {
			t.Fatalf("record entry = %q, want %q", rec.Entry, "entry")
		}
		if !rec.Last {
			t.Fatal("expected record.Last = true")
		}
		body, readErr := io.ReadAll(rec.Body)
		if readErr != nil {
			t.Fatalf("failed to read record body: %v", readErr)
		}
		_ = rec.Body.Close()
		if string(body) != "hello" {
			t.Fatalf("record body = %q, want %q", string(body), "hello")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for record after empty batch")
	}

	if atomic.LoadInt32(&requests) < 2 {
		t.Fatalf("expected at least 2 read requests, got %d", requests)
	}
}
