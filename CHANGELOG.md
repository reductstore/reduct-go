# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
* Set up project and CI/CD GitHub actions, [PR-9](https://github.com/reductstore/reduct-go/pull/9)
* Implement Bucket API [PR-10](https://github.com/reductstore/reduct-go/pull/10)
* Implement Query API with streaming support for large records
  * Add support for querying records with time intervals using `Start` and `Stop` parameters
  * Add support for querying records with label conditions using MongoDB-like syntax
  * Add support for continuous polling with configurable intervals
  * Add efficient streaming support for large records (>10MB)
  * Add HEAD request support for metadata-only queries
* Improve batch operations
  * Add support for batch writing with proper error handling
  * Add support for batch updates with label modifications
  * Add support for batch removals
* Improve error handling and type safety
  * Use proper error types and error wrapping
  * Add proper type conversions for timestamps and sizes
  * Use UTC timestamps consistently
* Add comprehensive test coverage for all new features

