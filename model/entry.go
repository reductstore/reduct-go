package model

type EntryInfo struct {
	Name         string `json:"name"`          // Name of the entry
	BlockCount   int64  `json:"block_count"`   // Number of blocks
	RecordCount  int64  `json:"record_count"`  // Number of records
	Size         int64  `json:"size"`          // Size of stored data in the bucket in bytes
	OldestRecord int64  `json:"oldest_record"` // Unix timestamp of the oldest record in microseconds
	LatestRecord int64  `json:"latest_record"` // Unix timestamp of the latest record in microseconds
}
