package model

// DiagnosticsItem represents a collection of diagnostic metrics.
type DiagnosticsItem struct {
	// Number of successful operations
	Ok int64 `json:"ok"`
	// Number of failed operations
	Errored int64 `json:"errored"`
	// Errors mapped by error code
	Errors map[int64]*DiagnosticsError `json:"errors"`
}

// Diagnostics represents replication statistics.
type Diagnostics struct {
	// Hourly diagnostics
	Hourly *DiagnosticsItem `json:"hourly"`
}

// DiagnosticsError represents error information in diagnostics.
type DiagnosticsError struct {
	// Number of errors
	Count int64 `json:"count"`
	// Last error message
	LastMessage string `json:"last_message"`
}
