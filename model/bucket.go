package model

type BucketSetting struct {
	MaxBlockSize    int32       `json:"max_block_size"`
	MaxBlockRecords int32       `json:"max_block_records"`
	QuotaType       []QuotaType `json:"quota_type"`
	QuotaSize       int32       `json:"quota_size"`
}

type QuotaType string

const (
	QuotaTypeNone QuotaType = "NONE"
	QuotaTypeFifo QuotaType = "FIFO"
	QuotaTypeHard QuotaType = "HARD"
)

type BucketInfo struct {
	Name          string `json:"name"`
	EntryCount    int32  `json:"entry_count"`
	Size          int32  `json:"size"`
	OldestRecord  int64  `json:"oldest_record"`
	LatestRecord  int64  `json:"latest_record"`
	IsProvisioned bool   `json:"is_provisioned"`
}

type CreateBucketRequest struct {
	BucketName      string    `json:"bucket_name"`
	MaxBlockSize    int32     `json:"max_block_size"`
	MaxBlockRecords int32     `json:"max_block_records"`
	QuotaType       QuotaType `json:"quota_type"`
	QuotaSize       int32     `json:"quota_size"`
}

type CreateBucketResponse struct {
}
type GetBucketResponse struct {
	Settings BucketSetting `json:"settings"`
	Info     BucketInfo    `json:"info"`
	Entries  []EntryInfo   `json:"entries"`
}
