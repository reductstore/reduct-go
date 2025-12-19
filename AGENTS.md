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
Starting with ReductStore v1.18, bucket and entry deletions are performed asynchronously in the background. Resources expose a `Status` field that can be `READY` or `DELETING`:

- **Status Field**: Added to `BucketInfo` and `EntryInfo` models as an optional field (`omitempty` JSON tag for backward compatibility)
- **Status Values**: 
  - `model.StatusReady` - Resource is ready for normal operations
  - `model.StatusDeleting` - Resource is being deleted in the background; operations will return HTTP 409 (Conflict) until deletion completes
- **Testing**: Status field is tested in existing integration tests (`client_test.go` for buckets, entries checked in `GetBucketEntries` and `GetBucketFullInfo`)
- **Backward Compatibility**: The status field uses `omitempty`, so it won't be serialized if empty, maintaining compatibility with older ReductStore versions
