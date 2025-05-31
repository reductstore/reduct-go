package model

// TokenPermissions represents the permissions associated with a token.
type TokenPermissions struct {
	// FullAccess indicates if the token has complete access to create, remove and update settings of buckets,
	// manage tokens and read and write data
	FullAccess bool `json:"full_access"`
	// Read is a list of buckets allowed to read
	Read []string `json:"read,omitempty"`
	// Write is a list of buckets allowed to write
	Write []string `json:"write,omitempty"`
}

// Token represents information about an access token.
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
