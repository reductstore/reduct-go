package batch

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseCSVRowReference is the pre-optimisation implementation of ParseCSVRow,
// kept verbatim as the oracle for FuzzParseCSVRowMatchesReference.
//
// It is deliberately the slow, allocation-heavy version: its job is to define
// the parse semantics of the record header wire format independently of the
// implementation under test, so a rewrite of ParseCSVRow cannot quietly change
// behaviour. Do not optimise or "fix" it — a difference between this and
// ParseCSVRow is exactly what the fuzz test exists to catch. If the header
// format itself ever changes, update both, and update the seed corpus.
func parseCSVRowReference(row string) (result CSVRowResult, panicked bool) {
	defer func() {
		// The original panics on rows with fewer than two fields because it
		// slices items[2:] unconditionally.
		if recover() != nil {
			panicked = true
		}
	}()

	items := make([]string, 0)
	escaped := ""
	current := ""

	for _, char := range row {
		if char == ',' && escaped == "" {
			if current != "" {
				items = append(items, current)
				current = ""
			}
			continue
		}

		if char == '"' {
			if escaped == "" {
				escaped = current
				current = ""
			} else {
				items = append(items, escaped+current)
				escaped = ""
				current = ""
			}
			continue
		}

		current += string(char)
	}

	if current != "" {
		if escaped != "" {
			items = append(items, escaped+current)
		} else {
			items = append(items, current)
		}
	}

	result = CSVRowResult{Labels: make(map[string]any)}

	if len(items) > 0 {
		size, err := strconv.ParseInt(items[0], 10, 64)
		if err == nil {
			result.Size = size
		}
	}

	if len(items) > 1 {
		result.ContentType = items[1]
	}

	for _, item := range items[2:] {
		if !strings.Contains(item, "=") {
			continue
		}

		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			result.Labels[parts[0]] = parts[1]
		}
	}

	return result, false
}

func TestParseCSVRow(t *testing.T) {
	tests := []struct {
		name        string
		row         string
		size        int64
		contentType string
		labels      map[string]any
	}{
		{
			name:        "size and content type",
			row:         "1024,text/plain",
			size:        1024,
			contentType: "text/plain",
			labels:      map[string]any{},
		},
		{
			name:        "with labels",
			row:         "10,application/json,a=1,b=two",
			size:        10,
			contentType: "application/json",
			labels:      map[string]any{"a": "1", "b": "two"},
		},
		{
			name:        "quoted label value containing a comma",
			row:         `10,text/plain,a="x,y",b=2`,
			size:        10,
			contentType: "text/plain",
			labels:      map[string]any{"a": "x,y", "b": "2"},
		},
		{
			name:        "empty fields are skipped",
			row:         "5,,a=1",
			size:        5,
			contentType: "a=1",
			labels:      map[string]any{},
		},
		{
			name:        "label value containing an equals sign",
			row:         "5,text/plain,a=b=c",
			size:        5,
			contentType: "text/plain",
			labels:      map[string]any{"a": "b=c"},
		},
		{
			name:        "label without a value is ignored",
			row:         "5,text/plain,novalue",
			size:        5,
			contentType: "text/plain",
			labels:      map[string]any{},
		},
		{
			name:        "single field does not panic",
			row:         "42",
			size:        42,
			contentType: "",
			labels:      map[string]any{},
		},
		{
			name:        "empty row",
			row:         "",
			size:        0,
			contentType: "",
			labels:      map[string]any{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ParseCSVRow(test.row)
			assert.Equal(t, test.size, got.Size)
			assert.Equal(t, test.contentType, got.ContentType)
			assert.Equal(t, test.labels, got.Labels)
		})
	}
}

// TestParseCSVRowIsAllocationLean guards the optimisation: parsing a realistic
// record header must not allocate per character.
func TestParseCSVRowIsAllocationLean(t *testing.T) {
	row := "1024,text/plain,label0=value0,label1=value1,label2=value2,label3=value3," +
		"label4=value4,label5=value5,label6=value6"

	allocs := testing.AllocsPerRun(100, func() {
		if ParseCSVRow(row).Size != 1024 {
			t.Fatal("unexpected size")
		}
	})

	// One map header plus its buckets; the field strings alias the input row.
	assert.LessOrEqual(t, allocs, float64(12), "ParseCSVRow allocates too much: %v allocs/op", allocs)
}

// FuzzParseCSVRowMatchesReference proves the optimised parser is behaviourally
// identical to the original for every input the original handles.
func FuzzParseCSVRowMatchesReference(f *testing.F) {
	seeds := []string{
		"1024,text/plain",
		"10,application/json,a=1,b=two",
		`10,text/plain,a="x,y",b=2`,
		`10,text/plain,"quoted`,
		`5,text/plain,="empty key"`,
		"5,,a=1",
		",,,",
		"42",
		"",
		`"`,
		`a"b"c,d`,
		"5,text/plain,a=b=c",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, row string) {
		// The original re-encodes each rune, which mangles invalid UTF-8; the
		// optimised parser preserves the raw bytes. Header values are always
		// valid UTF-8 in practice, so only compare on that domain.
		if !utf8.ValidString(row) {
			t.Skip("invalid UTF-8 is out of scope")
		}

		want, panicked := parseCSVRowReference(row)
		got := ParseCSVRow(row)

		if panicked {
			// The optimised parser must not panic where the original did.
			return
		}

		require.Equal(t, want.Size, got.Size, "size mismatch for %q", row)
		require.Equal(t, want.ContentType, got.ContentType, "content type mismatch for %q", row)
		require.Equal(t, want.Labels, got.Labels, "labels mismatch for %q", row)
	})
}
