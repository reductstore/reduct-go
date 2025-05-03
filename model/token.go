package model

// TokenPermissions represents the permissions associated with a token
type TokenPermissions struct {
	// FullAccess indicates if the token has complete access to all operations
	FullAccess bool `json:"full_access"`

	// Read contains list of buckets with read access
	Read []string `json:"read,omitempty"`

	// Write contains list of buckets with write access
	Write []string `json:"write,omitempty"`
}

type CreateTokenRequest struct {
	Name       string   `json:"name"`
	FullAccess bool     `json:"full_access,omitempty"`
	Read       []string `json:"read,omitempty"`
	Write      []string `json:"write,omitempty"`
}

type CreateTokenResponse struct {
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

// Token represents an access token and its associated metadata
type Token struct {
	// Name is the identifier for the token
	Name string `json:"name"`

	// CreatedAt is the timestamp when the token was created
	CreatedAt string `json:"created_at"`

	// IsProvisioned indicates if the token is provisioned and cannot be modified
	IsProvisioned bool `json:"is_provisioned,omitempty"`

	// Permissions defines the access rights for this token
	Permissions *TokenPermissions `json:"permissions,omitempty"`
}
