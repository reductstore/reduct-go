# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- Support replication modes (`enabled`, `paused`, `disabled`), [PR-44](https://github.com/reductstore/reduct-go/pull/44)
- Support for non-blocking deletions: Add `Status` field to `BucketInfo` and `EntryInfo` models to track deletion state (READY or DELETING), [PR-46](https://github.com/reductstore/reduct-go/pull/46)
- Support for multiple entry API, [PR-48](https://github.com/reductstore/reduct-go/pull/48)

### Fixed

- Fix crash when multi-entry query receives empty batch, [PR-50](https://github.com/reductstore/reduct-go/pull/50)

## 1.17.1 - 2025-11-17

## Added

- Add `baseUrl` to `Bucket.CreateQueryLink`, [PR-43](https://github.com/reductstore/reduct-go/pull/43)

## 1.17.0 - 2025-10-21

### Added

- Implement Bucket.CreateQueryLink method, [PR-39](https://github.com/reductstore/reduct-go/pull/39)

## 1.16.1  - 2025-08-26

### Fixed

- Fix error handling in Client.GetBucket method, [PR-38](https://github.com/reductstore/reduct-go/pull/38)

## 1.16.0 - 2025-07-31

Align with ReductStore v1.16 API.

## 1.15.2 - 2025-07-28

### Fixed

* Fix serialization of `head` flag in `QueryOptions`, [PR-37](https://github.com/reductstore/reduct-go/pull/37)

## 1.15.1 - 2025-06-23

### Fixed

* Use embeding instead of init for version file loading [PR-34](https://github.com/reductstore/reduct-go/pull/34)

## 1.15.0 - 2025-06-11

### Breaking Changes

* Remove deprecated `Include` and `Exclude` methods from `ReplicationSettings`, [PR-33](https://github.com/reductstore/reduct-go/pull/33)

## 1.15.0-beta.8 - 2025-06-05

### Added

* Missing methods  `RemoveEntry`, `RemoveRecord`, `RenamEntry`, [PR-32](https://github.com/reductstore/reduct-go/pull/32)

## 1.15.0-beta.7 - 2025-06-04

### Breaking Changes

* Remove unnecessary WritableRecord.WithSize method, [PR-29](https://github.com/reductstore/reduct-go/pull/29)

### Added

* `Bucket.GetEntries()` and `Bucket.GetFullInfo()` methods, [PR-30](https://github.com/reductstore/reduct-go/pull/30)

## 1.15.0-beta.5 - 2025-06-03

### Fixed

* HTTP error handling and down casting, [PR-28](https://github.com/reductstore/reduct-go/pull/28)

## 1.15.0-beta.4 - 2025-06-03

### Fixed

* Use response status code by default, [PR-27](https://github.com/reductstore/reduct-go/pull/27)

## 1.15.0-beta.3 - 2025-06-03

### Fixed

* Avoid double HTTP error wrapping [PR-26](https://github.com/reductstore/reduct-go/pull/26)

## 1.15.0-beta.2 - 2025-06-03

### Breaking Changes

* Fix record error parsing in Batch.Write() and return record errors as result, [PR-25](https://github.com/reductstore/reduct-go/pull/25)

## 1.15.0-beta.1 - 2025-06-02

### Added

* Set up project and CI/CD GitHub actions, [PR-9](https://github.com/reductstore/reduct-go/pull/9)
* Implement Bucket API, [PR-10](https://github.com/reductstore/reduct-go/pull/10)
* Implement Entry API, [PR-12](https://github.com/reductstore/reduct-go/pull/12)
* Implement Server API, [PR-14](https://github.com/reductstore/reduct-go/pull/14)
* Implement Token API, [PR-15](https://github.com/reductstore/reduct-go/pull/15)
* Implement Replication API, [PR-16](https://github.com/reductstore/reduct-go/pull/16)
* Improvement for Bucket API, [PR-18](https://github.com/reductstore/reduct-go/pull/18)
