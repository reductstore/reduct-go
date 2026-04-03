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

// TokenCreateOptions represents optional fields for token creation API v2.
type TokenCreateOptions struct {
	// Permissions of the token.
	Permissions TokenPermissions `json:"permissions"`
	// ExpiresAt sets absolute expiration time in RFC3339 format.
	ExpiresAt *string `json:"expires_at,omitempty"`
	// TTL sets inactivity timeout in seconds. Must be > 0 when provided.
	TTL *uint64 `json:"ttl,omitempty"`
	// IPAllowlist restricts token usage to exact IPs/CIDRs.
	IPAllowlist []string `json:"ip_allowlist,omitempty"`
}

// TokenCreateResponse is returned by create/rotate token endpoints.
type TokenCreateResponse struct {
	// Value is a token secret value.
	Value string `json:"value"`
	// CreatedAt is the creation time in RFC3339 format.
	CreatedAt string `json:"created_at"`
}

// Token represents information about an access token.
type Token struct {
	// Name of the token
	Name string `json:"name"`
	// CreatedAt is the creation time in RFC3339 format.
	CreatedAt string `json:"created_at"`
	// Value is empty for read endpoints and non-empty for create/rotate responses.
	Value string `json:"value,omitempty"`
	// IsProvisioned indicates if the token is provisioned and can't be removed or changed
	IsProvisioned bool `json:"is_provisioned,omitempty"`
	// Permissions of the token
	Permissions *TokenPermissions `json:"permissions,omitempty"`
	// ExpiresAt is absolute expiration timestamp when configured.
	ExpiresAt *string `json:"expires_at,omitempty"`
	// TTL is inactivity expiration in seconds.
	TTL *uint64 `json:"ttl,omitempty"`
	// LastAccess is the last successful access timestamp.
	LastAccess *string `json:"last_access,omitempty"`
	// IPAllowlist limits allowed client IPs/CIDRs.
	IPAllowlist []string `json:"ip_allowlist,omitempty"`
	// IsExpired indicates whether token is currently expired.
	IsExpired bool `json:"is_expired,omitempty"`
}
