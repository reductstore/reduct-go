package model

import "fmt"

type QueryType string

const (
	QueryTypeQuery  QueryType = "QUERY"
	QueryTypeRemove QueryType = "REMOVE"
)

type QueryEntry struct {
	QueryType    string            `json:"query_type"`
	Start        *int64            `json:"start,omitempty"`
	Stop         *int64            `json:"stop,omitempty"`
	Include      map[string]string `json:"include,omitempty"`
	Exclude      map[string]string `json:"exclude,omitempty"`
	EachS        *float64          `json:"each_s,omitempty"`
	EachN        *int32            `json:"each_n,omitempty"`
	Limit        *int32            `json:"limit,omitempty"`
	TTL          *int32            `json:"ttl,omitempty"`
	OnlyMetadata *bool             `json:"only_metadata,omitempty"`
	Continuous   *bool             `json:"continuous,omitempty"`
	When         any               `json:"when,omitempty"`
	Strict       *bool             `json:"strict,omitempty"`
	Ext          map[string]any    `json:"ext,omitempty"`
}

type LabelMap map[string]any

type QueryOptions struct {
	TTL          *int32         `json:"ttl,omitempty"`
	Include      LabelMap       `json:"include,omitempty"`
	Exclude      LabelMap       `json:"exclude,omitempty"`
	EachS        *float64       `json:"each_s,omitempty"`
	EachN        *int32         `json:"each_n,omitempty"`
	Limit        *int32         `json:"limit,omitempty"`
	Continuous   *bool          `json:"continuous,omitempty"`
	PollInterval *int32         `json:"poll_interval,omitempty"`
	Head         *bool          `json:"head,omitempty"`
	When         map[string]any `json:"when,omitempty"`
	Strict       *bool          `json:"strict,omitempty"`
	Ext          map[string]any `json:"ext,omitempty"`
}

func SerializeQueryOptions(queryType QueryType, opts QueryOptions, start, stop *int64) QueryEntry {
	return QueryEntry{
		QueryType:    string(queryType),
		Start:        start,
		Stop:         stop,
		TTL:          opts.TTL,
		Include:      toStringMap(opts.Include),
		Exclude:      toStringMap(opts.Exclude),
		EachS:        opts.EachS,
		EachN:        opts.EachN,
		Limit:        opts.Limit,
		Continuous:   opts.Continuous,
		When:         opts.When,
		Strict:       opts.Strict,
		OnlyMetadata: opts.Head,
		Ext:          opts.Ext,
	}
}

func toStringMap(m LabelMap) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string)
	for k, v := range m {
		result[k] = fmt.Sprint(v)
	}
	return result
}
