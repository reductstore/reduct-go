package model

type BucketSettings struct {
	QuotaSize       int64  `json:"quota_size"`
	MaxBlockSize    int64  `json:"max_block_size"`
	QuotaType       string `json:"quota_type"`
	MaxBlockRecords int64  `json:"max_block_records"`
}

type LicenseInfo struct {
	Licensee    string `json:"licensee"`
	Invoice     string `json:"invoice"`
	ExpiryDate  string `json:"expiry_date"`
	Plan        string `json:"plan"`
	DeviceCount int    `json:"device_number"`
	DiskQuota   int64  `json:"disk_quota"`
	Fingerprint string `json:"fingerprint"`
}

type ServerDefaults struct {
	Bucket BucketSettings `json:"bucket"`
}

type ServerInfo struct {
	Version      string         `json:"version"`
	BucketCount  int64          `json:"bucket_count"`
	Usage        int64          `json:"usage"`
	Uptime       int64          `json:"uptime"`
	OldestRecord int64          `json:"oldest_record"`
	LatestRecord int64          `json:"latest_record"`
	License      *LicenseInfo   `json:"license,omitempty"`
	Defaults     ServerDefaults `json:"defaults"`
}
