# ReductStore Client SDK for Golang

[![Go Reference](https://pkg.go.dev/badge/github.com/reductstore/reduct-go.svg)](https://pkg.go.dev/github.com/reductstore/reduct-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/reductstore/reduct-go)](https://goreportcard.com/report/github.com/reductstore/reduct-go)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/reductstore/reduct-go/ci.yml?branch=main)](https://github.com/reductstore/reduct-go/actions)

The ReductStore Client SDK for Golang is an HTTP client wrapper for interacting with a [ReductStore](https://www.reduct.store) instance from a Golang application. It provides a set of APIs for accessing and manipulating data stored in ReductStore.

## Requirements

- Go 1.24 or later

## Features

- HTTP client wrapper with context support
- Clean API design following Go idioms
- Support for [ReductStore HTTP API v1.16](https://www.reduct.store/docs/http-api)
- Token-based authentication for secure access to the database
- Labeling for read-write operations and querying
- Batch operations for efficient data processing

## Version Compatibility

This SDK follows a version compatibility policy to ensure smooth operation with ReductStore servers:

- Current SDK Version: 1.8.0
- Minimum Supported Server Version: 1.5.0

The SDK will issue a warning if it detects that the server version is 3 or more minor versions older than the minimum supported version. In such cases, you should either:
- Upgrade your ReductStore server to a newer version, or
- Downgrade the SDK to a version compatible with your server

This policy helps maintain compatibility while allowing for API evolution. Legacy code for deprecated API features is removed after 3 minor version releases.

## Getting Started

To get started with the ReductStore Client SDK for Golang, you'll need to have ReductStore installed and running on your machine. You can find instructions for installing ReductStore [here](https://www.reduct.store/docs/getting-started#docker).

Once you have ReductStore up and running, you can install the ReductStore Client SDK for Golang using go get:

```bash
go get github.com/reductstore/reduct-go
```

Then, you can use the following example code to start interacting with your ReductStore database from your Go application:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/reductstore/reduct-go"
    "github.com/reductstore/reduct-go/model"
)

func main() {
    // 1. Create a ReductStore client
    client := reductgo.NewClient("http://127.0.0.1:8383", reductgo.ClientOptions{
        APIToken: "my-token",
    })

    ctx := context.Background()

    // 2. Get or create a bucket with 1Gb quota
    settings := model.NewBucketSettingBuilder().
        WithQuotaType(model.QuotaTypeFifo).
        WithQuotaSize(1e9).
        Build()

    bucket, err := client.CreateOrGetBucket(ctx, "my-bucket", &settings)
    if err != nil {
        panic(err)
    }

    // 3. Write some data with timestamps and labels to the 'sensor-1' entry
    ts := func(dateStr string) int64 {
        t, _ := time.Parse(time.RFC3339, dateStr)
        return t.UnixMicro()
    }

    record := bucket.BeginWrite(ctx, "sensor-1", &reductgo.WriteOptions{
        Timestamp: ts("2021-01-01T11:00:00Z"),
        Labels: reductgo.LabelMap{
            "score": 10,
        },
    })
    if err := record.Write([]byte("<Blob data>")); err != nil {
        panic(err)
    }

    record = bucket.BeginWrite(ctx, "sensor-1", &reductgo.WriteOptions{
        Timestamp: ts("2021-01-01T11:00:01Z"),
        Labels: reductgo.LabelMap{
            "score": 20,
        },
    })
    if err := record.Write([]byte("<Blob data>")); err != nil {
        panic(err)
    }

    // 4. Query the data by time range and condition
    query := reductgo.NewQueryOptionsBuilder().
        WithStart(ts("2021-01-01T11:00:00Z")).
        WithStop(ts("2021-01-01T11:00:02Z")).
        WithWhen(map[string]interface{}{
            "&score": map[string]interface{}{
                "$gt": 10,
            },
        }).
        Build()

    result, err := bucket.Query(ctx, "sensor-1", query)
    if err != nil {
        panic(err)
    }

    for record := range result.Records() {
        fmt.Printf("Record timestamp: %d\n", record.Time())
        fmt.Printf("Record size: %d\n", record.Size())
        data, _ := record.ReadAsString()
        fmt.Println(data)
    }
}
```

For more examples, see the [Guides](https://www.reduct.store/docs/guides) section in the ReductStore documentation.

## Supported ReductStore Versions and Backward Compatibility

The library is backward compatible with the previous versions. However, some methods have been deprecated and will be removed in future releases. Please refer to the [Changelog](CHANGELOG.md) for more details.

The SDK supports the following ReductStore API versions:
- v1.16
- v1.15
- v1.14

It can work with newer and older versions, but it is not guaranteed that all features will work as expected because the API may change and some features may be deprecated or the SDK may not support them yet.

