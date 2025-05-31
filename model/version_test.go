package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expected    *Version
		expectError bool
	}{
		{
			name:    "valid version with patch",
			version: "1.2.3",
			expected: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name:    "valid version without patch",
			version: "1.2",
			expected: &Version{
				Major: 1,
				Minor: 2,
				Patch: 0,
			},
		},
		{
			name:    "version with v prefix",
			version: "v1.2.3",
			expected: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name:        "invalid version format",
			version:     "1",
			expectError: true,
		},
		{
			name:        "invalid major version",
			version:     "a.2.3",
			expectError: true,
		},
		{
			name:        "invalid minor version",
			version:     "1.b.3",
			expectError: true,
		},
		{
			name:        "invalid patch version",
			version:     "1.2.c",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ParseVersion(tt.version)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, version)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{
			name: "version with patch",
			version: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			expected: "1.2.3",
		},
		{
			name: "version without patch",
			version: &Version{
				Major: 1,
				Minor: 2,
				Patch: 0,
			},
			expected: "1.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.version.String())
		})
	}
}

func TestVersion_IsOlderThan(t *testing.T) {
	tests := []struct {
		name           string
		version        *Version
		other          *Version
		minorDiff      int
		expectedResult bool
	}{
		{
			name:           "older by major version",
			version:        &Version{Major: 1, Minor: 2},
			other:          &Version{Major: 2, Minor: 0},
			minorDiff:      3,
			expectedResult: true,
		},
		{
			name:           "newer by major version",
			version:        &Version{Major: 2, Minor: 0},
			other:          &Version{Major: 1, Minor: 8},
			minorDiff:      3,
			expectedResult: false,
		},
		{
			name:           "older by exactly minorDiff versions",
			version:        &Version{Major: 1, Minor: 2},
			other:          &Version{Major: 1, Minor: 5},
			minorDiff:      3,
			expectedResult: true,
		},
		{
			name:           "older by less than minorDiff versions",
			version:        &Version{Major: 1, Minor: 3},
			other:          &Version{Major: 1, Minor: 5},
			minorDiff:      3,
			expectedResult: false,
		},
		{
			name:           "newer version",
			version:        &Version{Major: 1, Minor: 5},
			other:          &Version{Major: 1, Minor: 3},
			minorDiff:      3,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.IsOlderThan(tt.other, tt.minorDiff)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCheckServerAPIVersion(t *testing.T) {
	sdkVersion := GetVersion() // Use the actual SDK version

	tests := []struct {
		name          string
		serverVersion string
		wantError     bool
		errorMessage  string
	}{
		{
			name:          "matching versions",
			serverVersion: sdkVersion,
			wantError:     false,
		},
		{
			name:          "server minor version older but within limit",
			serverVersion: "1.6.0",
			wantError:     false,
		},
		{
			name:          "server minor version too old",
			serverVersion: "1.5.0",
			wantError:     false,
		},
		{
			name:          "incompatible major versions",
			serverVersion: "2.0.0",
			wantError:     true,
			errorMessage:  "Incompatible server API version: 2.0.0. Client version: " + sdkVersion + ". Please update your client.",
		},
		{
			name:          "invalid server version",
			serverVersion: "invalid",
			wantError:     true,
			errorMessage:  "failed to parse server version: invalid version format: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckServerAPIVersion(tt.serverVersion, sdkVersion)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
