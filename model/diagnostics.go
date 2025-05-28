package model

import "strconv"

// DiagnosticsItem represents a collection of diagnostic metrics
type DiagnosticsItem struct {
	// Number of successful operations
	Ok int64 `json:"ok"`
	// Number of failed operations
	Errored int64 `json:"errored"`
	// Errors mapped by error code
	Errors map[int64]*DiagnosticsError `json:"errors"`
}

// ParseDiagnosticsItem creates a DiagnosticsItem struct from a map
func ParseDiagnosticsItem(data map[string]interface{}) *DiagnosticsItem {
	okCount := int64(0)
	if okVal, exists := data["ok"].(float64); exists {
		okCount = int64(okVal)
	}

	errored := int64(0)
	if erroredVal, ok := data["errored"].(float64); ok {
		errored = int64(erroredVal)
	}

	errors := make(map[int64]*DiagnosticsError)
	if errorsData, ok := data["errors"].(map[string]interface{}); ok {
		for k, v := range errorsData {
			if errorData, ok := v.(map[string]interface{}); ok {
				if code, err := strconv.ParseInt(k, 10, 64); err == nil {
					errors[code] = ParseDiagnosticsError(errorData)
				}
			}
		}
	}

	return &DiagnosticsItem{
		Ok:      okCount,
		Errored: errored,
		Errors:  errors,
	}
}

// Diagnostics represents replication statistics
type Diagnostics struct {
	// Hourly diagnostics
	Hourly *DiagnosticsItem `json:"hourly"`
}

// ParseDiagnostics creates a Diagnostics struct from a map
func ParseDiagnostics(data map[string]interface{}) *Diagnostics {
	var hourly *DiagnosticsItem
	if hourlyData, ok := data["hourly"].(map[string]interface{}); ok {
		hourly = ParseDiagnosticsItem(hourlyData)
	}

	return &Diagnostics{
		Hourly: hourly,
	}
}

// DiagnosticsError represents error information in diagnostics
type DiagnosticsError struct {
	// Number of errors
	Count int64 `json:"count"`
	// Last error message
	LastMessage string `json:"last_message"`
}

// ParseDiagnosticsError creates a DiagnosticsError struct from a map
func ParseDiagnosticsError(data map[string]interface{}) *DiagnosticsError {
	count := int64(0)
	if countVal, ok := data["count"].(float64); ok {
		count = int64(countVal)
	}

	lastMessage := ""
	if msg, ok := data["last_message"].(string); ok {
		lastMessage = msg
	}

	return &DiagnosticsError{
		Count:       count,
		LastMessage: lastMessage,
	}
}
