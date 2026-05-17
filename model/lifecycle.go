package model

// LifecycleType defines lifecycle action type.
type LifecycleType string

const (
	LifecycleTypeDelete LifecycleType = "delete"
)

// IsValid returns true when the lifecycle type matches a known value.
func (t LifecycleType) IsValid() bool {
	switch t {
	case LifecycleTypeDelete:
		return true
	default:
		return false
	}
}

// LifecycleMode defines lifecycle operating mode.
type LifecycleMode string

const (
	LifecycleModeEnabled  LifecycleMode = "enabled"
	LifecycleModeDisabled LifecycleMode = "disabled"
	LifecycleModeDryRun   LifecycleMode = "dry_run"
)

// IsValid returns true when the lifecycle mode matches a known value.
func (m LifecycleMode) IsValid() bool {
	switch m {
	case LifecycleModeEnabled, LifecycleModeDisabled, LifecycleModeDryRun:
		return true
	default:
		return false
	}
}

// LifecycleSettings represents lifecycle policy settings.
type LifecycleSettings struct {
	// Lifecycle action type.
	LifecycleType LifecycleType `json:"type,omitempty"`
	// Bucket to apply lifecycle to.
	Bucket string `json:"bucket"`
	// List of entries to process. If empty, all matching entries are used.
	Entries []string `json:"entries,omitempty"`
	// Maximum record age.
	MaxAge string `json:"max_age"`
	// Interval between lifecycle runs.
	Interval string `json:"interval,omitempty"`
	// Conditional query.
	When any `json:"when,omitempty"`
	// Lifecycle mode.
	Mode LifecycleMode `json:"mode,omitempty"`
}

// LifecycleInfo represents basic information about a lifecycle policy.
type LifecycleInfo struct {
	// Name of the lifecycle policy.
	Name string `json:"name"`
	// Whether the lifecycle policy is provisioned.
	IsProvisioned bool `json:"is_provisioned"`
	// Whether the lifecycle worker is currently running.
	IsRunning bool `json:"is_running"`
	// Current lifecycle mode.
	Mode LifecycleMode `json:"mode"`
}

// LifecycleModePayload represents the payload to update lifecycle mode.
type LifecycleModePayload struct {
	Mode LifecycleMode `json:"mode"`
}

// FullLifecycleInfo represents complete information about a lifecycle policy.
type FullLifecycleInfo struct {
	// Basic lifecycle information.
	Info *LifecycleInfo `json:"info"`
	// Lifecycle settings.
	Settings *LifecycleSettings `json:"settings"`
}
