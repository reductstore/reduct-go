package model

import (
	"fmt"
	"time"
)

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

// ParseLicenseInfo parses the original license info into a LicenseInfo struct
func ParseLicenseInfo(data map[string]interface{}) (*LicenseInfo, error) {
	expiryStr, ok := data["expiry_date"].(string)
	if !ok {
		return nil, NewAPIError("invalid expiry_date format", 400, nil)
	}

	expiryTime, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		return nil, NewAPIError("invalid expiry_date format", 400, err)
	}

	deviceNum, _ := data["device_number"].(float64)
	diskQuota, _ := data["disk_quota"].(float64)

	return &LicenseInfo{
		Licensee:     data["licensee"].(string),
		Invoice:      data["invoice"].(string),
		ExpiryDate:   expiryTime.UnixMilli(),
		Plan:         data["plan"].(string),
		DeviceNumber: int64(deviceNum),
		DiskQuota:    int64(diskQuota),
		Fingerprint:  data["fingerprint"].(string),
	}, nil
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

// ParseServerInfo parses the original server info into a ServerInfo struct
func ParseServerInfo(data map[string]interface{}) (*ServerInfo, error) {
	bucketCount, _ := data["bucket_count"].(string)
	uptime, _ := data["uptime"].(string)
	usage, _ := data["usage"].(string)
	oldestRecord, _ := data["oldest_record"].(string)
	latestRecord, _ := data["latest_record"].(string)

	var license *LicenseInfo
	if licenseData, ok := data["license"].(map[string]interface{}); ok {
		var err error
		license, err = ParseLicenseInfo(licenseData)
		if err != nil {
			return nil, err
		}
	}

	defaults := ServerDefaults{}
	if defaultsData, ok := data["defaults"].(map[string]interface{}); ok {
		if bucketData, ok := defaultsData["bucket"].(map[string]interface{}); ok {
			quotaSize, _ := bucketData["quota_size"].(float64)
			maxBlockSize, _ := bucketData["max_block_size"].(float64)
			maxBlockRecords, _ := bucketData["max_block_records"].(float64)
			quotaType, _ := bucketData["quota_type"].(string)

			defaults.Bucket = BucketSetting{
				QuotaSize:       int64(quotaSize),
				MaxBlockSize:    int64(maxBlockSize),
				MaxBlockRecords: int64(maxBlockRecords),
				QuotaType:       QuotaType(quotaType),
			}
		}
	}

	return &ServerInfo{
		Version:      data["version"].(string),
		BucketCount:  parseStringToInt64(bucketCount),
		Uptime:       parseStringToInt64(uptime),
		Usage:        parseStringToInt64(usage),
		OldestRecord: parseStringToInt64(oldestRecord),
		LatestRecord: parseStringToInt64(latestRecord),
		License:      license,
		Defaults:     defaults,
	}, nil
}

// Helper function to parse string to int64
func parseStringToInt64(s string) int64 {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0
	}
	return result
}
