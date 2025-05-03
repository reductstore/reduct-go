package model

// ReplicationInfo represents information about a replication
type ReplicationInfo struct {
	Name           string `json:"name"`
	IsActive       bool   `json:"is_active"`
	IsProvisioned  bool   `json:"is_provisioned"`
	PendingRecords int64  `json:"pending_records"`
}

// FullReplicationInfo represents complete replication information including settings and diagnostics
type FullReplicationInfo struct {
	Info        ReplicationInfo     `json:"info"`
	Settings    ReplicationSettings `json:"settings"`
	Diagnostics Diagnostics         `json:"diagnostics"`
}

// FullReplicationInfoResponse represents the raw response format for replication info
type FullReplicationInfoResponse struct {
	Info        ReplicationInfo             `json:"info"`
	Settings    OriginalReplicationSettings `json:"settings"`
	Diagnostics Diagnostics                 `json:"diagnostics"`
}

// OriginalReplicationSettings represents the raw format of replication settings
type OriginalReplicationSettings struct {
	SrcBucket string            `json:"src_bucket"`
	DstBucket string            `json:"dst_bucket"`
	DstHost   string            `json:"dst_host"`
	DstToken  string            `json:"dst_token"`
	Entries   []string          `json:"entries"`
	Include   map[string]string `json:"include,omitempty"`
	Exclude   map[string]string `json:"exclude,omitempty"`
	EachS     *int              `json:"each_s,omitempty"`
	EachN     *int64            `json:"each_n,omitempty"`
	When      interface{}       `json:"when,omitempty"`
}

// ReplicationSettings represents the normalized format of replication settings
type ReplicationSettings struct {
	SrcBucket string            `json:"srcBucket"`
	DstBucket string            `json:"dstBucket"`
	DstHost   string            `json:"dstHost"`
	DstToken  string            `json:"dstToken"`
	Entries   []string          `json:"entries"`
	Include   map[string]string `json:"include,omitempty"`
	Exclude   map[string]string `json:"exclude,omitempty"`
	EachS     *int              `json:"eachS,omitempty"`
	EachN     *int64            `json:"eachN,omitempty"`
	When      interface{}       `json:"when,omitempty"`
}
