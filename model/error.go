package model

import "fmt"

type APIError struct {
	Status   int    `json:"status"`
	Message  string `json:"message"`
	Original any    `json:"original"`
}

func NewAPIError(status int, message string, original any) *APIError {
	return &APIError{
		Status:   status,
		Message:  message,
		Original: original,
	}
}

func (e *APIError) Error() string {
	return fmt.Sprintf("APIError: %d %s", e.Status, e.Message)
}
