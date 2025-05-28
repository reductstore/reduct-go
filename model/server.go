package model

// ServerDefaults represents default settings for the server
type ServerDefaults struct {
	Bucket BucketSetting `json:"bucket"`
}

// LicenseInfo represents license information for the server
type LicenseInfo struct {
	// Licensee name
	Licensee string `json:"licensee"`
	// Invoice number
	Invoice string `json:"invoice"`
	// Expiry date as unix timestamp in milliseconds
	ExpiryDate int64 `json:"expiry_date"`
	// Plan name
	Plan string `json:"plan"`
	// Number of devices
	DeviceNumber int64 `json:"device_number"`
	// Disk quota
	DiskQuota int64 `json:"disk_quota"`
	// Fingerprint
	Fingerprint string `json:"fingerprint"`
}

// ServerInfo represents information about the storage server
type ServerInfo struct {
	// Version storage server
	Version string `json:"version"`
	// Number of buckets
	BucketCount int64 `json:"bucket_count"`
	// Stored data in bytes
	Usage int64 `json:"usage"`
	// Server uptime in seconds
	Uptime int64 `json:"uptime"`
	// Unix timestamp of the oldest record in microseconds
	OldestRecord int64 `json:"oldest_record"`
	// Unix timestamp of the latest record in microseconds
	LatestRecord int64 `json:"latest_record"`
	// License information
	License *LicenseInfo `json:"license,omitempty"`
	// Default settings
	Defaults ServerDefaults `json:"defaults"`
}
