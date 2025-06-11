package model

// ReplicationSettings represents the settings for replication.
type ReplicationSettings struct {
	// Source bucket. Must exist.
	SrcBucket string `json:"src_bucket"`
	// Destination bucket. Must exist.
	DstBucket string `json:"dst_bucket"`
	// Destination host. Must exist.
	DstHost string `json:"dst_host"`
	// Destination token. Must have write access to the destination bucket.
	DstToken string `json:"dst_token,omitempty"`
	// List of entries to replicate. If empty, all entries are replicated. Wildcards are supported.
	Entries []string `json:"entries,omitempty"`
	// Replicate a record every S seconds
	EachS int64 `json:"each_s,omitempty"`
	// Replicate every Nth record
	EachN int64 `json:"each_n,omitempty"`
	// Conditional query
	When any `json:"when,omitempty"`
}

// ReplicationInfo represents basic information about a replication.
type ReplicationInfo struct {
	// Name of the replication
	Name string `json:"name"`
	// Whether the remote instance is available and replication is active
	IsActive bool `json:"is_active"`
	// Whether the replication is provisioned
	IsProvisioned bool `json:"is_provisioned"`
	// Number of records pending replication
	PendingRecords int64 `json:"pending_records"`
}

// FullReplicationInfo represents complete information about a replication.
type FullReplicationInfo struct {
	// Basic replication information
	Info *ReplicationInfo `json:"info"`
	// Replication settings
	Settings *ReplicationSettings `json:"settings"`
	// Replication statistics
	Diagnostics *Diagnostics `json:"diagnostics"`
}
