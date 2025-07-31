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

## Getting Started

To get started with the ReductStore Client SDK for Golang, you'll need to have ReductStore installed and running on your machine. 
You can find instructions for installing ReductStore [here](https://www.reduct.store/docs/getting-started#docker).

Once you have ReductStore up and running, you can install the ReductStore Client SDK for Golang using go get:

```bash
go get github.com/reductstore/reduct-go
```

Then, you can use the following example code to start interacting with your ReductStore database from your Go application:

```go
package main

import (
	"context"
	reduct "github.com/reductstore/reduct-go"
	model "github.com/reductstore/reduct-go/model"
	"time"
)

func main() {
	ctx := context.Background()
	// 1. Create a ReductStore client
	client := reduct.NewClient("http://localhost:8383", reduct.ClientOptions{
		APIToken: "my-token",
	})

	// 2. Get or create a bucket with 1Gb quota
	settings := model.NewBucketSettingBuilder().
		WithQuotaType(model.QuotaTypeFifo).
		WithQuotaSize(1_000_000_000).
		Build()

	bucket, err := client.CreateOrGetBucket(ctx, "my-bucket", &settings)
	if err != nil {
		panic(err)
	}

	// 3. Write some data with timestamps in the 'entry-1' entry
	ts := time.Now().UnixMicro()
	writer := bucket.BeginWrite(ctx, "entry-1",
		&reduct.WriteOptions{Timestamp: ts, Labels: map[string]any{"score": 10}})
	err = writer.Write("<Blob data>")
	if err != nil {
		panic(err)
	}

	writer = bucket.BeginWrite(ctx, "entry-1",
		&reduct.WriteOptions{Timestamp: ts + 1, Labels: map[string]any{"score": 20}})
	err = writer.Write("<Blob data 2>")
	if err != nil {
		panic(err)
	}

	// 4. Query the data by time range and condition
	queryOptions := reduct.NewQueryOptionsBuilder().
		WithStart(ts).
		WithStop(ts + 2).
		WithWhen(map[string]any{"&score": map[string]any{"$gt": 15}}).
		Build()

	query, err := bucket.Query(ctx, "entry-1", &queryOptions)
	if err != nil {
		panic(err)
	}

	for rec := range query.Records() {
		data, err := rec.Read()
		if err != nil {
			panic(err)
		}
		timestamp := rec.Time()
		labels := rec.Labels()
		println("Record at time:", timestamp)
		println("Labels:", labels)
		println("Data:", string(data))
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

