package model

// Represents bucket settings
type BucketSetting struct {
	// max block content_length in bytes
	MaxBlockSize int64 `json:"max_block_size"`
	// max number of records in a block
	MaxBlockRecords int64 `json:"max_block_records"`
	// quota type, [NONE, FIFO, HARD]
	QuotaType QuotaType `json:"quota_type"`
	// quota content_length in bytes
	QuotaSize int64 `json:"quota_size"`
}

type QuotaType string

const (
	QuotaTypeNone QuotaType = "NONE"
	QuotaTypeFifo QuotaType = "FIFO"
	QuotaTypeHard QuotaType = "HARD"
)

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
}

// Information about the bucket in JSON format
type FullBucketDetail struct {
	// bucket settings
	Settings BucketSetting `json:"settings"`
	// bucket info
	Info BucketInfo `json:"info"`
	// entries in the bucket
	Entries []EntryInfo `json:"entries"`
}
