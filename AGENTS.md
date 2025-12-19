# Repository Guidelines

## Project Structure & Module Organization
- Core SDK lives in the root `reductgo` package (`client.go`, `bucket.go`, `record.go`, `batch.go`); this is where bucket, record, and batch operations are defined.
- HTTP adapters sit in `httpclient/`; API models, version helpers, and builders live in `model/` (e.g., `version.go`, `replication.go`); keep generated or API-shaped structs there.
- Tests live alongside code as `*_test.go`, with integration setup/teardown in `main_test.go`; lint artifacts are written to `logs/`.
- Key configs: `go.mod` (Go 1.24+), `.golangci.yml` (lint rules and formatters).

## Build, Test, and Development Commands
- `go test ./...` — runs the suite; integration tests hit `http://localhost:8383`. Ensure a ReductStore instance is running and `RS_API_TOKEN` is set (or provided via `.env` at repo root). Example: `RS_API_TOKEN=token go test ./...`.
- `golangci-lint run ./...` — uses the bundled config; also writes JUnit output to `logs/issues.xml` for CI.
- `go vet ./...` — optional static analysis to catch obvious issues during iteration.

## Coding Style & Naming Conventions
- Format with `gofmt`/`goimports` (enforced by golangci-lint); tabs and default Go widths.
- Follow Go idioms: context first, error last, short locals, no unchecked errors; keep API-facing names aligned with ReductStore terms (bucket, entry, record).
- Exported identifiers need doc comments; JSON tags use snake_case per the lint config.

## Testing Guidelines
- Tests rely on `stretchr/testify`; prefer table-driven cases for new behaviors.
- Integration tests create and clean buckets dynamically; reuse helpers in `main_test.go` instead of hard-coding names or tokens.
- Use `-run` filters for focused debugging when full integration runs are slow; document any test that needs special fixtures or environment.

## Commit & Pull Request Guidelines
- Commits are short and imperative (e.g., `Add base_url to query link`); releases follow `release vX.Y.Z` and often include PR numbers.
- PRs should describe behavior changes, link issues, and note compatibility with supported ReductStore API versions (v1.15–v1.17). Attach results for `go test ./...` and `golangci-lint run`.
- Update README/CHANGELOG when modifying public APIs or support matrix, and call out breaking changes early in the PR description.

## Non-Blocking Deletions (v1.18+)

Starting with ReductStore v1.18, bucket and entry deletions are performed asynchronously in the background. The operation returns immediately while the actual cleanup happens in the background.

During deletion, the bucket or entry status will be set to `DELETING`. While in this state:
- Read, write, or delete operations will return HTTP 409 (Conflict)
- You can check the status using the `Status` field in `BucketInfo` or `EntryInfo`

Example of checking bucket status:

```go
ctx := context.Background()

// Delete a bucket
err := client.RemoveBucket(ctx, "my-bucket")
if err != nil {
    panic(err)
}

// Check bucket status in list
buckets, err := client.GetBuckets(ctx)
if err != nil {
    panic(err)
}

for _, bucket := range buckets {
    if bucket.Status == model.StatusDeleting {
        fmt.Printf("Bucket %s is being deleted\n", bucket.Name)
    }
}
```

The status field can have the following values:
- `model.StatusReady` - The resource is ready for operations
- `model.StatusDeleting` - The resource is being deleted in the background
