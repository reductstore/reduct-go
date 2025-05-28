package model

import "time"

// TokenPermissions represents the permissions associated with a token
type TokenPermissions struct {
	// FullAccess indicates if the token has complete access to create, remove and update settings of buckets,
	// manage tokens and read and write data
	FullAccess bool `json:"full_access"`
	// Read is a list of buckets allowed to read
	Read []string `json:"read,omitempty"`
	// Write is a list of buckets allowed to write
	Write []string `json:"write,omitempty"`
}

// ParseTokenPermissions creates a TokenPermissions struct from a map
func ParseTokenPermissions(data map[string]interface{}) *TokenPermissions {
	var read []string
	var write []string

	if readData, ok := data["read"].([]interface{}); ok {
		read = make([]string, len(readData))
		for i, v := range readData {
			if str, ok := v.(string); ok {
				read[i] = str
			}
		}
	}

	if writeData, ok := data["write"].([]interface{}); ok {
		write = make([]string, len(writeData))
		for i, v := range writeData {
			if str, ok := v.(string); ok {
				write[i] = str
			}
		}
	}

	return &TokenPermissions{
		FullAccess: data["full_access"].(bool),
		Read:       read,
		Write:      write,
	}
}

// Serialize converts TokenPermissions to a map suitable for JSON serialization
func (tp *TokenPermissions) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"full_access": tp.FullAccess,
		"read":        tp.Read,
		"write":       tp.Write,
	}
}

// Token represents information about an access token
type Token struct {
	// Name of the token
	Name string `json:"name"`
	// CreatedAt is the creation time of the token as unix timestamp in milliseconds
	CreatedAt string `json:"created_at"`
	// IsProvisioned indicates if the token is provisioned and can't be removed or changed
	IsProvisioned bool `json:"is_provisioned,omitempty"`
	// Permissions of the token
	Permissions *TokenPermissions `json:"permissions,omitempty"`
}

// ParseToken creates a Token struct from a map
func ParseToken(data map[string]interface{}) (*Token, error) {
	createdAtStr, ok := data["created_at"].(string)
	if !ok {
		return nil, NewAPIError("invalid created_at format", 400, nil)
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, NewAPIError("invalid created_at format", 400, err)
	}

	var permissions *TokenPermissions
	if permData, ok := data["permissions"].(map[string]interface{}); ok {
		permissions = ParseTokenPermissions(permData)
	}

	isProvisioned := false
	if provData, ok := data["is_provisioned"].(bool); ok {
		isProvisioned = provData
	}

	return &Token{
		Name:          data["name"].(string),
		CreatedAt:     createdAt.Format(time.RFC3339),
		IsProvisioned: isProvisioned,
		Permissions:   permissions,
	}, nil
}
