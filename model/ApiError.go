package model

// APIError represents an HTTP error with optional status, message, and original error info.
type APIError struct {
	Status   int    `json:"status,omitempty"`   // HTTP status of the error (nil if communication issue)
	Message  string `json:"message,omitempty"`  // Parsed message from the storage engine
	Original any    `json:"original,omitempty"` // Original error (can be of any type)
}

// NewAPIError creates a new instance of APIError with given message, status, and original error.
func NewAPIError(message string, status int, original any) *APIError {
	return &APIError{
		Status:   status,
		Message:  message,
		Original: original,
	}
}

func (e APIError) Error() string {
	return e.Message
}
