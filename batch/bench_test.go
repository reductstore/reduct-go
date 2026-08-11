package batch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/httpclient"
)

// benchConfig describes a synthetic query response served by the fake server.
type benchConfig struct {
	recordSize int
	perBatch   int
	batches    int
	labels     int
}

// newBenchServer serves cfg.batches batches of cfg.perBatch records using Batch
// Protocol v1. It keeps the server side as cheap as possible so that the
// measurement reflects SDK-side parsing and allocation cost rather than server
// or storage behaviour. The returned reset function rewinds the response
// sequence so each benchmark iteration replays the same query.
func newBenchServer(cfg benchConfig) (server *httptest.Server, reset func()) {
	payload := make([]byte, cfg.recordSize*cfg.perBatch)
	for i := range payload {
		payload[i] = byte('a' + i%26)
	}

	var labelSuffix strings.Builder
	for i := 0; i < cfg.labels; i++ {
		fmt.Fprintf(&labelSuffix, ",label%d=value%d", i, i)
	}
	csvRow := fmt.Sprintf("%d,text/plain%s", cfg.recordSize, labelSuffix.String())
	contentLength := strconv.Itoa(len(payload))

	var served atomic.Int64
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.20")
		index := served.Add(1) - 1
		if index >= int64(cfg.batches) {
			w.Header().Set("X-Reduct-Error", "No content")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		base := index * int64(cfg.perBatch) * 1000
		for i := 0; i < cfg.perBatch; i++ {
			w.Header().Set(fmt.Sprintf("x-reduct-time-%d", base+int64(i)*1000), csvRow)
		}
		w.Header().Set("x-reduct-last", strconv.FormatBool(index == int64(cfg.batches)-1))
		w.Header().Set("Content-Length", contentLength)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(payload); err != nil {
			panic(err)
		}
	}))

	return server, func() { served.Store(0) }
}

// drain consumes every record exactly the way a well-behaved caller would: one
// reusable buffer, read to EOF, no per-record allocation on the caller side.
func drain(b *testing.B, records <-chan *Record, buf []byte) int64 {
	b.Helper()

	var total int64
	for rec := range records {
		for {
			n, err := rec.Body.Read(buf)
			total += int64(n)
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatalf("read record: %v", err)
			}
		}
		if rec.Last {
			break
		}
	}
	return total
}

func runQueryBenchmark(b *testing.B, cfg benchConfig) {
	b.Helper()

	server, reset := newBenchServer(cfg)
	defer server.Close()

	client := httpclient.NewHTTPClient(httpclient.Option{BaseURL: server.URL, Timeout: 30 * time.Second})
	ctx := context.Background()
	buf := make([]byte, 64*1024)
	wantBytes := int64(cfg.recordSize) * int64(cfg.perBatch) * int64(cfg.batches)
	iterations := 0

	b.SetBytes(wantBytes)
	b.ReportAllocs()

	for b.Loop() {
		reset()
		records, errCh, err := FetchAndParse(ctx, client, "bucket", "entry", 1, false, time.Second, false)
		if err != nil {
			b.Fatalf("fetch: %v", err)
		}
		if got := drain(b, records, buf); got != wantBytes {
			b.Fatalf("read %d bytes, want %d", got, wantBytes)
		}
		for range errCh { //nolint:revive // drain
		}
		iterations++
	}

	b.StopTimer()
	// Normalise to per-record cost, which is what the SDK controls.
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(iterations*cfg.perBatch*cfg.batches), "ns/record")
}

// Record sizes mirror the cross-SDK benchmark suite, where the Go SDK loses the
// most ground at small records: https://github.com/reductstore/sdk-benchmarks
func BenchmarkQueryV1_1KiB(b *testing.B) {
	runQueryBenchmark(b, benchConfig{recordSize: 1024, perBatch: 80, batches: 25, labels: 0})
}

func BenchmarkQueryV1_8KiB(b *testing.B) {
	runQueryBenchmark(b, benchConfig{recordSize: 8192, perBatch: 80, batches: 25, labels: 0})
}

func BenchmarkQueryV1_64KiB(b *testing.B) {
	runQueryBenchmark(b, benchConfig{recordSize: 65536, perBatch: 80, batches: 5, labels: 0})
}

func BenchmarkQueryV1_1MiB(b *testing.B) {
	runQueryBenchmark(b, benchConfig{recordSize: 1024 * 1024, perBatch: 8, batches: 5, labels: 0})
}

// With labels the per-record CSV header parser is on the hot path.
func BenchmarkQueryV1_1KiB_7Labels(b *testing.B) {
	runQueryBenchmark(b, benchConfig{recordSize: 1024, perBatch: 80, batches: 25, labels: 7})
}

func BenchmarkParseCSVRow(b *testing.B) {
	row := "1024,text/plain,label0=value0,label1=value1,label2=value2,label3=value3,label4=value4,label5=value5,label6=value6"
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		result := ParseCSVRow(row)
		if result.Size != 1024 {
			b.Fatal("unexpected size")
		}
	}
}
