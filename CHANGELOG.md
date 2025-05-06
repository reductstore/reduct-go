# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
* Set up project and CI/CD GitHub actions, [PR-9](https://github.com/reductstore/reduct-go/pull/9)
## Added Bucket api [PR-10](https://github.com/reductstore/reduct-go/pull/10)
* Github Actions workflow
* Handle X-Reduct-error on http client
* Added client methods 
    - CreateBucket 
    - CreateOrGetBucket
    - GetBucket
    - CheckBucketExists
    - RemoveBucket
* Added bucket methods 
    - CheckExists
    - GetInfo
    - GetSettings
    - SetSettings
    - Rename
    - Remove
