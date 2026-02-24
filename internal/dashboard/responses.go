package dashboard

// API response structures

// HealthResponse is the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// DashboardSummary is the summary response
type DashboardSummary struct {
	TotalEpochs      int `json:"total_epochs"`
	TotalValidators  int `json:"total_validators"`
	TotalSlots       int `json:"total_slots"`
	ActiveValidators int `json:"active_validators"`
	CurrentDay       string `json:"current_day"`
}

// NetworkTopology is the network graph data response (PRIORITY)
type NetworkTopology struct {
	Nodes []TopologyNode `json:"nodes"`
	Links []TopologyLink `json:"links"`
}

// TopologyNode represents a node in the network graph
type TopologyNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "validator", "slot", "project"
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// TopologyLink represents a connection in the network graph
type TopologyLink struct {
	Source     string                 `json:"source"`
	Target     string                 `json:"target"`
	Type       string                 `json:"type"` // "votes_for", "submits_to", "validates"
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// EpochsList is the response for listing epochs
type EpochsList struct {
	Epochs []EpochSummary `json:"epochs"`
}

// EpochSummary is a summary of an epoch
type EpochSummary struct {
	EpochID            uint64 `json:"epoch_id"`
	Timestamp          int64  `json:"timestamp"`
	TotalValidators    int    `json:"total_validators"`
	EligibleNodesCount int    `json:"eligible_nodes_count"`
	SlotCount          int    `json:"slot_count"`
	AggregatedProjects int    `json:"aggregated_projects"`
}

// EpochDetail is the detailed response for a single epoch
type EpochDetail struct {
	EpochID            uint64            `json:"epoch_id"`
	DataMarket         string            `json:"data_market"`
	Timestamp          int64             `json:"timestamp"`
	TotalValidators    int               `json:"total_validators"`
	EligibleNodesCount int               `json:"eligible_nodes_count"`
	SubmissionCounts   map[string]int    `json:"submission_counts"`
	AggregatedProjects map[string]string `json:"aggregated_projects"`
	ValidatorBatches   []ValidatorBatch  `json:"validator_batches,omitempty"`
}

// ValidatorBatch represents a validator's batch
type ValidatorBatch struct {
	ValidatorID string `json:"validator_id"`
	BatchCID    string `json:"batch_cid"`
}

// ValidatorsList is the response for listing validators
type ValidatorsList struct {
	Validators []ValidatorSummary `json:"validators"`
}

// ValidatorSummary is a summary of a validator
type ValidatorSummary struct {
	ValidatorID    string `json:"validator_id"`
	TotalEpochs    int    `json:"total_epochs"`
	TotalBatches   int    `json:"total_batches"`
	LastActive     int64  `json:"last_active"`
	RecentEpochs   []uint64 `json:"recent_epochs,omitempty"`
}

// ValidatorDetail is the detailed response for a single validator
type ValidatorDetail struct {
	ValidatorID     string            `json:"validator_id"`
	TotalEpochs     int               `json:"total_epochs"`
	BatchesByEpoch  map[uint64]string `json:"batches_by_epoch"`
	Projects        []string          `json:"projects"`
	LastActive      int64             `json:"last_active"`
}

// SlotsList is the response for listing slots
type SlotsList struct {
	Slots []SlotSummary `json:"slots"`
}

// SlotSummary is a summary of a slot
type SlotSummary struct {
	SlotID         string `json:"slot_id"`
	TotalDays      int    `json:"total_days"`
	TotalSubmits   int    `json:"total_submits"`
	EligibleCount  int    `json:"eligible_count"`
	LastActive     int64  `json:"last_active"`
}

// SlotDetail is the detailed response for a single slot
type SlotDetail struct {
	SlotID          string            `json:"slot_id"`
	SubmissionsByDay map[string]int   `json:"submissions_by_day"`
	EligibleByDay   map[string]int    `json:"eligible_by_day"`
	TotalSubmits    int               `json:"total_submits"`
	TotalEligible   int               `json:"total_eligible"`
}

// ProjectsList is the response for listing projects
type ProjectsList struct {
	Projects []ProjectSummary `json:"projects"`
}

// ProjectSummary is a summary of a project
type ProjectSummary struct {
	ProjectID   string `json:"project_id"`
	VoteCount   int    `json:"vote_count"`
	Epochs      int    `json:"epochs"`
	LastCid     string `json:"last_cid,omitempty"`
}

// Timeline is the response for timeline events
type Timeline struct {
	Events []TimelineEvent `json:"events"`
}

// TimelineEvent represents an event in the timeline
type TimelineEvent struct {
	Type      string `json:"type"` // "epoch", "day_transition", "batch_aggregated"
	Timestamp int64  `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}
