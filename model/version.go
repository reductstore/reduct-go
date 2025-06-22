package model

import (
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major int
	Minor int
}

//go:embed VERSION
var version string

func init() {
	version = strings.TrimSpace(version)
}

// GetVersion returns the current version of the SDK.
func GetVersion() string {
	return version
}

// ParseVersion parses a version string into a Version struct.
func ParseVersion(version string) (*Version, error) {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", version)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", version)
	}

	return &Version{
		Major: major,
		Minor: minor,
	}, nil
}

// String returns the string representation of the version.
func (v *Version) String() string {
	if v.Major == 0 && v.Minor == 0 {
		return "dev"
	}
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// IsOlderThan returns true if this version is older than the other version by at least minorVersionDiff minor versions.
func (v *Version) IsOlderThan(other *Version, minorVersionDiff int) bool {
	if v.Major < other.Major {
		return true
	}
	if v.Major > other.Major {
		return false
	}
	return (other.Minor - v.Minor) >= minorVersionDiff
}

// CheckServerAPIVersion checks if the server API version is compatible with the client version
// It returns an error if the major versions don't match, and logs a warning if the server
// is more than 2 minor versions behind the client.
func CheckServerAPIVersion(serverVersion, clientVersion string) error {
	server, err := ParseVersion(serverVersion)
	if err != nil {
		return fmt.Errorf("failed to parse server version: %w", err)
	}

	client, err := ParseVersion(clientVersion)
	if err != nil {
		return fmt.Errorf("failed to parse client version: %w", err)
	}

	if server.Major != client.Major {
		return &APIError{
			Message: fmt.Sprintf("Incompatible server API version: %s. Client version: %s. Please update your client.",
				serverVersion, clientVersion),
		}
	}

	if client.Minor-server.Minor > 2 {
		log.Printf("WARNING: Server API version %s is too old for this client version %s. Please update your server.",
			serverVersion, clientVersion)
	}

	return nil
}
