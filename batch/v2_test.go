package batch

import (
	"strings"
	"testing"
)

func TestParseHeaderList(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		want        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "empty header returns nil",
			header:  "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "whitespace only header returns nil",
			header:  "   ",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "single entry",
			header:  "entry1",
			want:    []string{"entry1"},
			wantErr: false,
		},
		{
			name:    "multiple entries",
			header:  "entry1,entry2,entry3",
			want:    []string{"entry1", "entry2", "entry3"},
			wantErr: false,
		},
		{
			name:    "entries with spaces",
			header:  " entry1 , entry2 , entry3 ",
			want:    []string{"entry1", "entry2", "entry3"},
			wantErr: false,
		},
		{
			name:    "entries with percent encoding",
			header:  "entry%201,entry%202",
			want:    []string{"entry 1", "entry 2"},
			wantErr: false,
		},
		{
			name:        "empty entry in list",
			header:      "entry1,,entry3",
			want:        nil,
			wantErr:     true,
			errContains: "invalid entries/labels header",
		},
		{
			name:        "invalid percent encoding",
			header:      "entry%ZZ",
			want:        nil,
			wantErr:     true,
			errContains: "invalid URL escape",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeaderList(tt.header)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseHeaderList() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseHeaderList() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("parseHeaderList() unexpected error = %v", err)
				return
			}
			if !equalStringSlices(got, tt.want) {
				t.Errorf("parseHeaderList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
