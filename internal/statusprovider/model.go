package statusprovider

import (
	"encoding/json"
	"time"
)

const SchemaVersion = 1

type RuntimePresence string

const (
	RuntimePresent RuntimePresence = "present"
	RuntimeAbsent  RuntimePresence = "absent"
	RuntimeStale   RuntimePresence = "stale"
	RuntimeUnknown RuntimePresence = "unknown"
)

type AgentState string

const (
	AgentWorking AgentState = "working"
	AgentReady   AgentState = "ready"
	AgentBlocked AgentState = "blocked"
	AgentDone    AgentState = "done"
	AgentUnknown AgentState = "unknown"
)

type Request struct {
	SchemaVersion int     `json:"schema_version"`
	Queries       []Query `json:"queries"`
}

type Query struct {
	QueryID   string          `json:"query_id"`
	Path      string          `json:"path,omitempty"`
	RepoRoot  string          `json:"repo_root,omitempty"`
	GitBranch string          `json:"git_branch,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

type Response struct {
	SchemaVersion int              `json:"schema_version"`
	ObservedAt    time.Time        `json:"observed_at"`
	FreshUntil    time.Time        `json:"fresh_until"`
	Providers     []ProviderStatus `json:"providers"`
	Results       []Result         `json:"results"`
}

type ProviderStatus struct {
	Name       string    `json:"name"`
	Available  bool      `json:"available"`
	ObservedAt time.Time `json:"observed_at"`
	FreshUntil time.Time `json:"fresh_until"`
	Error      string    `json:"error,omitempty"`
}

type Result struct {
	QueryID         string               `json:"query_id"`
	Metadata        json.RawMessage      `json:"metadata,omitempty"`
	RuntimePresence RuntimePresence      `json:"runtime_presence"`
	AgentState      AgentState           `json:"agent_state"`
	Match           *MatchEvidence       `json:"match,omitempty"`
	Runtimes        []RuntimeObservation `json:"runtimes"`
	Errors          []QueryError         `json:"errors"`
}

type MatchEvidence struct {
	Type        string `json:"type"`
	MatchedPath string `json:"matched_path,omitempty"`
}

type RuntimeObservation struct {
	ObservationID   string          `json:"observation_id"`
	Backend         string          `json:"backend"`
	NativeID        string          `json:"native_id"`
	Path            string          `json:"path,omitempty"`
	RepoRoot        string          `json:"repo_root,omitempty"`
	GitBranch       string          `json:"git_branch,omitempty"`
	RuntimePresence RuntimePresence `json:"runtime_presence"`
	AgentState      AgentState      `json:"agent_state"`
	ObservedAt      time.Time       `json:"observed_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
}

type QueryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
