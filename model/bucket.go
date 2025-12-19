// Package model defines the core data structures (models) used throughout the application.
// components. Each model typically includes fields, validation rules, and methods for
// interacting with the underlying data.
package model

import "encoding/json"

// BucketSetting represents bucket settings.
type BucketSetting struct {
	// max block content_length in bytes
	MaxBlockSize int64 `json:"max_block_size,omitempty"`
	// max number of records in a block
	MaxBlockRecords int64 `json:"max_block_records,omitempty"`
	// quota type, [NONE, FIFO, HARD]
	QuotaType QuotaType `json:"quota_type,omitempty"`
	// quota content_length in bytes
	QuotaSize int64 `json:"quota_size,omitempty"`
}

// MarshalJSON is the custom marshaller for serializing BucketSetting object.
func (b BucketSetting) MarshalJSON() ([]byte, error) {
	tmp := make(map[string]any)

	if b.MaxBlockSize != 0 {
		tmp["max_block_size"] = b.MaxBlockSize
	}
	if b.MaxBlockRecords != 0 {
		tmp["max_block_records"] = b.MaxBlockRecords
	}
	if b.QuotaSize != 0 {
		tmp["quota_size"] = b.QuotaSize
	}
	if b.QuotaType != "" { // assuming QuotaNone is your zero/default
		tmp["quota_type"] = b.QuotaType
	}
	return json.Marshal(tmp)
}

type BucketSettingBuilder struct {
	bucket BucketSetting
}

// NewBucketSettingBuilder creates a new instance of BucketSettingBuilder with can be used to build a BucketSetting object.
func NewBucketSettingBuilder() *BucketSettingBuilder {
	return &BucketSettingBuilder{}
}

// WithMaxBlockSize sets MaxBlockSize value of the builder.
func (b *BucketSettingBuilder) WithMaxBlockSize(size int64) *BucketSettingBuilder {
	b.bucket.MaxBlockSize = size
	return b
}

// WithMaxBlockRecords sets MaxBlockRecords value of the builder.
func (b *BucketSettingBuilder) WithMaxBlockRecords(records int64) *BucketSettingBuilder {
	b.bucket.MaxBlockRecords = records
	return b
}

// WithQuotaSize sets QuotaSize value of the builder.
func (b *BucketSettingBuilder) WithQuotaSize(size int64) *BucketSettingBuilder {
	b.bucket.QuotaSize = size
	return b
}

// WithQuotaType sets the QuotaType value of the builder.
// options are ["NONE","FIFO", "HARD"].
func (b *BucketSettingBuilder) WithQuotaType(qt QuotaType) *BucketSettingBuilder {
	b.bucket.QuotaType = qt
	return b
}

// Build Builds BucketSetting model.
// Uses ["NONE"] QotaType if not set.
func (b *BucketSettingBuilder) Build() BucketSetting {
	if b.bucket.QuotaType == "" {
		b.bucket.QuotaType = QuotaTypeNone
	}
	return b.bucket
}

type QuotaType string

const (
	QuotaTypeNone QuotaType = "NONE"
	QuotaTypeFifo QuotaType = "FIFO"
	QuotaTypeHard QuotaType = "HARD"
)

// Status represents the current status of a bucket or entry.
type Status string

const (
	// StatusReady indicates the bucket or entry is ready for operations.
	StatusReady Status = "READY"
	// StatusDeleting indicates the bucket or entry is being deleted in the background.
	// Operations on resources with this status will return HTTP 409.
	StatusDeleting Status = "DELETING"
)

// BucketInfo Represents information about a bucket.
type BucketInfo struct {
	// bucket name
	Name string `json:"name"`
	// number of entries in the bucket
	EntryCount int64 `json:"entry_count"`
	// disk usage in bytes
	Size int64 `json:"size"`
	// unix timestamp of oldest record in microseconds
	OldestRecord uint64 `json:"oldest_record"`
	// unix timestamp of latest record in microseconds
	LatestRecord uint64 `json:"latest_record"`
	// true if the bucket is provisioned
	IsProvisioned bool `json:"is_provisioned"`
	// status of the bucket (READY or DELETING)
	Status Status `json:"status,omitempty"`
}

// FullBucketDetail Information about the bucket in JSON format.
type FullBucketDetail struct {
	// bucket settings
	Settings BucketSetting `json:"settings"`
	// bucket info
	Info BucketInfo `json:"info"`
	// entries in the bucket
	Entries []EntryInfo `json:"entries"`
}
