package model

// DiagnosticsError represents an error in diagnostics
type DiagnosticsError struct {
	Count       int    `json:"count"`
	LastMessage string `json:"last_message"`
}

// DiagnosticsItem represents a diagnostics item with success/failure counts
type DiagnosticsItem struct {
	Ok      int64                    `json:"ok"`
	Errored int64                    `json:"errored"`
	Errors  map[int]DiagnosticsError `json:"errors"`
}

// Diagnostics contains diagnostic information
type Diagnostics struct {
	Hourly DiagnosticsItem `json:"hourly"`
}
