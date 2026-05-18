package reductgo

import (
	"context"
	"time"

	"github.com/reductstore/reduct-go/batch"
)

func (b *Bucket) fetchAndParseBatchedRecords(ctx context.Context, entry string, id int64, continueQuery bool, pollInterval time.Duration, head bool) (*QueryResult, error) {
	records, errCh, err := batch.FetchAndParse(ctx, b.HTTPClient, b.Name, entry, id, continueQuery, pollInterval, head)
	if err != nil {
		return &QueryResult{}, err
	}

	return wrapBatchRecords(ctx, records, errCh), nil
}

func (b *Bucket) fetchAndParseBatchedRecordsV2(ctx context.Context, id int64, continueQuery bool, pollInterval time.Duration, head bool) (*QueryResult, error) {
	records, errCh, err := batch.FetchAndParseV2(ctx, b.HTTPClient, b.Name, id, continueQuery, pollInterval, head)
	if err != nil {
		return &QueryResult{}, err
	}

	return wrapBatchRecords(ctx, records, errCh), nil
}

func wrapBatchRecords(ctx context.Context, records <-chan *batch.Record, errCh <-chan error) *QueryResult {
	out := make(chan *ReadableRecord, 100)
	outErrCh := make(chan error, 1)

	go func() {
		defer close(outErrCh)
		defer close(out)
		for rec := range records {
			if rec == nil {
				continue
			}

			select {
			case <-ctx.Done():
				return
			default:
			}

			var labels LabelMap
			if rec.Labels != nil {
				labels = LabelMap(rec.Labels)
			}

			record := NewReadableRecord(rec.Entry, rec.Time, rec.Size, rec.Last, rec.Body, labels, rec.ContentType)
			record.SetLastInBatch(rec.LastInBatch)

			select {
			case <-ctx.Done():
				return
			case out <- record:
				if record.IsLast() {
					return
				}
			}
		}

		if err, ok := <-errCh; ok && err != nil {
			outErrCh <- err
		}
	}()

	return &QueryResult{records: out, errCh: outErrCh}
}
