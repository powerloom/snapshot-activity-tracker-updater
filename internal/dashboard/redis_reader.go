package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"p2p-debugger/redis"
)

// Reader reads data from Redis and tally files
type Reader struct {
	dataMarket string
	dumpDir    string
}

// NewReader creates a new dashboard data reader
func NewReader(dataMarket, dumpDir string) *Reader {
	if dumpDir == "" {
		dumpDir = "./tallies"
	}
	return &Reader{
		dataMarket: dataMarket,
		dumpDir:    dumpDir,
	}
}

// GetDashboardSummary returns the dashboard summary
func (r *Reader) GetDashboardSummary(ctx context.Context) (*DashboardSummary, error) {
	currentDay, err := redis.Get(ctx, redis.CurrentDayKey(r.dataMarket))
	if err != nil {
		return nil, fmt.Errorf("failed to get current day: %w", err)
	}

	// Count epochs from tally files
	tallyFiles, err := r.getTallyFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	totalValidators := 0
	activeValidators := 0
	uniqueSlots := make(map[string]bool)

	for _, tally := range tallyFiles {
		if tally.TotalValidators > 0 {
			totalValidators = max(totalValidators, tally.TotalValidators)
		}
		if tally.EligibleNodesCount > 0 {
			activeValidators = max(activeValidators, tally.EligibleNodesCount)
		}
		for slotID := range tally.SubmissionCounts {
			uniqueSlots[slotID] = true
		}
	}
	totalSlots := len(uniqueSlots)

	return &DashboardSummary{
		TotalEpochs:      len(tallyFiles),
		TotalValidators:  totalValidators,
		TotalSlots:       totalSlots,
		ActiveValidators: activeValidators,
		CurrentDay:       currentDay,
	}, nil
}

// GetNetworkTopology returns the network topology data (PRIORITY)
// Uses only the most recent epochs to avoid slow response and client timeouts (broken pipe)
func (r *Reader) GetNetworkTopology(ctx context.Context) (*NetworkTopology, error) {
	tallyFiles, err := r.getRecentTallyFiles(30)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	topology := &NetworkTopology{
		Nodes: []TopologyNode{},
		Links: []TopologyLink{},
	}

	validatorNodes := make(map[string]bool)
	slotNodes := make(map[string]bool)
	projectNodes := make(map[string]bool)

	// Process each tally file to build the network graph
	for _, tally := range tallyFiles {
		// Add validator nodes
		for validatorID := range tally.ValidatorBatchCIDs {
			if !validatorNodes[validatorID] {
				topology.Nodes = append(topology.Nodes, TopologyNode{
					ID:    validatorID,
					Type:  "validator",
					Label: r.shortenID(validatorID),
					Properties: map[string]any{
						"epoch_id": tally.EpochID,
					},
				})
				validatorNodes[validatorID] = true
			}
		}

		// Add slot nodes
		for slotID := range tally.SubmissionCounts {
			if !slotNodes[slotID] {
				topology.Nodes = append(topology.Nodes, TopologyNode{
					ID:    slotID,
					Type:  "slot",
					Label: fmt.Sprintf("Slot %s", slotID),
					Properties: map[string]any{
						"submission_count": tally.SubmissionCounts[slotID],
					},
				})
				slotNodes[slotID] = true
			}
		}

		// Add project nodes
		for projectID, cid := range tally.AggregatedProjects {
			if !projectNodes[projectID] {
				topology.Nodes = append(topology.Nodes, TopologyNode{
					ID:    projectID,
					Type:  "project",
					Label: r.shortenID(projectID),
					Properties: map[string]any{
						"latest_cid": cid,
					},
				})
				projectNodes[projectID] = true
			}

			// Add links from validators to projects (they validated this project in this epoch)
			for validatorID := range tally.ValidatorBatchCIDs {
				linkID := fmt.Sprintf("%s-%s-%d", validatorID, projectID, tally.EpochID)
				topology.Links = append(topology.Links, TopologyLink{
					Source: validatorID,
					Target: projectID,
					Type:   "validates",
					Properties: map[string]any{
						"link_id":  linkID,
						"epoch_id": tally.EpochID,
					},
				})
			}

			// Add links from slots to projects (slots submitted to this project)
			for slotID := range tally.SubmissionCounts {
				linkID := fmt.Sprintf("%s-%s-%d", slotID, projectID, tally.EpochID)
				topology.Links = append(topology.Links, TopologyLink{
					Source: slotID,
					Target: projectID,
					Type:   "submits_to",
					Properties: map[string]any{
						"link_id":  linkID,
						"epoch_id": tally.EpochID,
					},
				})
			}
		}
	}

	return topology, nil
}

// GetEpochs returns a list of epochs with data
func (r *Reader) GetEpochs(ctx context.Context) (*EpochsList, error) {
	tallyFiles, err := r.getTallyFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	// Sort by epoch ID descending
	for i, j := 0, len(tallyFiles)-1; i < j; i, j = i+1, j-1 {
		tallyFiles[i], tallyFiles[j] = tallyFiles[j], tallyFiles[i]
	}

	epochs := make([]EpochSummary, 0, len(tallyFiles))
	for _, tally := range tallyFiles {
		epochs = append(epochs, EpochSummary{
			EpochID:            tally.EpochID,
			Timestamp:          tally.Timestamp,
			TotalValidators:    tally.TotalValidators,
			EligibleNodesCount: tally.EligibleNodesCount,
			SlotCount:          len(tally.SubmissionCounts),
			AggregatedProjects: len(tally.AggregatedProjects),
		})
	}

	return &EpochsList{Epochs: epochs}, nil
}

// GetEpochDetail returns detailed information about a specific epoch
func (r *Reader) GetEpochDetail(ctx context.Context, epochID uint64) (*EpochDetail, error) {
	tally, err := r.getTallyForEpoch(epochID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally for epoch %d: %w", epochID, err)
	}

	if tally == nil {
		return nil, fmt.Errorf("epoch %d not found", epochID)
	}

	batches := make([]ValidatorBatch, 0, len(tally.ValidatorBatchCIDs))
	for validatorID, batchCID := range tally.ValidatorBatchCIDs {
		batches = append(batches, ValidatorBatch{
			ValidatorID: validatorID,
			BatchCID:    batchCID,
		})
	}

	return &EpochDetail{
		EpochID:            tally.EpochID,
		DataMarket:         tally.DataMarket,
		Timestamp:          tally.Timestamp,
		TotalValidators:    tally.TotalValidators,
		EligibleNodesCount: tally.EligibleNodesCount,
		SubmissionCounts:   tally.SubmissionCounts,
		AggregatedProjects: tally.AggregatedProjects,
		ValidatorBatches:   batches,
	}, nil
}

// GetValidators returns a list of validators
func (r *Reader) GetValidators(ctx context.Context) (*ValidatorsList, error) {
	tallyFiles, err := r.getTallyFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	validatorData := make(map[string]*ValidatorSummary)

	for _, tally := range tallyFiles {
		for validatorID := range tally.ValidatorBatchCIDs {
			if data, exists := validatorData[validatorID]; exists {
				data.TotalEpochs++
				data.TotalBatches++
				if tally.Timestamp > data.LastActive {
					data.LastActive = tally.Timestamp
				}
				if len(data.RecentEpochs) < 10 {
					data.RecentEpochs = append(data.RecentEpochs, tally.EpochID)
				}
			} else {
				validatorData[validatorID] = &ValidatorSummary{
					ValidatorID:  validatorID,
					TotalEpochs:  1,
					TotalBatches: 1,
					LastActive:   tally.Timestamp,
					RecentEpochs: []uint64{tally.EpochID},
				}
			}
		}
	}

	validators := make([]ValidatorSummary, 0, len(validatorData))
	for _, data := range validatorData {
		validators = append(validators, *data)
	}

	return &ValidatorsList{Validators: validators}, nil
}

// GetValidatorDetail returns detailed information about a specific validator
func (r *Reader) GetValidatorDetail(ctx context.Context, validatorID string) (*ValidatorDetail, error) {
	tallyFiles, err := r.getTallyFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	batchesByEpoch := make(map[uint64]string)
	projects := make(map[string]bool)
	lastActive := int64(0)
	totalEpochs := 0

	for _, tally := range tallyFiles {
		if batchCID, exists := tally.ValidatorBatchCIDs[validatorID]; exists {
			batchesByEpoch[tally.EpochID] = batchCID
			totalEpochs++
			if tally.Timestamp > lastActive {
				lastActive = tally.Timestamp
			}
		}
		// Track projects from aggregated data
		for projectID := range tally.AggregatedProjects {
			projects[projectID] = true
		}
	}

	if totalEpochs == 0 {
		return nil, fmt.Errorf("validator %s not found", validatorID)
	}

	projectList := make([]string, 0, len(projects))
	for projectID := range projects {
		projectList = append(projectList, projectID)
	}

	return &ValidatorDetail{
		ValidatorID:    validatorID,
		TotalEpochs:    totalEpochs,
		BatchesByEpoch: batchesByEpoch,
		Projects:       projectList,
		LastActive:     lastActive,
	}, nil
}

// GetSlots returns a list of slots
func (r *Reader) GetSlots(ctx context.Context) (*SlotsList, error) {
	tallyFiles, err := r.getTallyFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	slotData := make(map[string]*SlotSummary)
	slotDays := make(map[string]map[string]bool)    // slotID -> days
	slotEligible := make(map[string]map[string]bool) // slotID -> days

	for _, tally := range tallyFiles {
		day := r.dayFromTimestamp(tally.Timestamp)

		for slotID, count := range tally.SubmissionCounts {
			if _, exists := slotData[slotID]; !exists {
				slotData[slotID] = &SlotSummary{
					SlotID:       slotID,
					TotalSubmits: count,
					LastActive:   tally.Timestamp,
				}
				slotDays[slotID] = make(map[string]bool)
			} else {
				slotData[slotID].TotalSubmits += count
				if tally.Timestamp > slotData[slotID].LastActive {
					slotData[slotID].LastActive = tally.Timestamp
				}
			}
			slotDays[slotID][day] = true
		}

		// Track eligible submissions by day
		if tally.EligibleNodesCount > 0 {
			for slotID := range tally.SubmissionCounts {
				if slotEligible[slotID] == nil {
					slotEligible[slotID] = make(map[string]bool)
				}
				slotEligible[slotID][day] = true
			}
		}
	}

	slots := make([]SlotSummary, 0, len(slotData))
	for slotID, data := range slotData {
		data.TotalDays = len(slotDays[slotID])
		data.EligibleCount = len(slotEligible[slotID])
		slots = append(slots, *data)
	}

	return &SlotsList{Slots: slots}, nil
}

// GetSlotDetail returns detailed information about a specific slot
func (r *Reader) GetSlotDetail(ctx context.Context, slotID string) (*SlotDetail, error) {
	tallyFiles, err := r.getTallyFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	submissionsByDay := make(map[string]int)
	eligibleByDay := make(map[string]int)
	found := false

	for _, tally := range tallyFiles {
		day := r.dayFromTimestamp(tally.Timestamp)
		if count, exists := tally.SubmissionCounts[slotID]; exists {
			submissionsByDay[day] += count
			if tally.EligibleNodesCount > 0 {
				eligibleByDay[day]++
			}
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("slot %s not found", slotID)
	}

	totalSubmits := 0
	for _, count := range submissionsByDay {
		totalSubmits += count
	}

	return &SlotDetail{
		SlotID:          slotID,
		SubmissionsByDay: submissionsByDay,
		EligibleByDay:   eligibleByDay,
		TotalSubmits:    totalSubmits,
		TotalEligible:   len(eligibleByDay),
	}, nil
}

// GetProjects returns a list of projects with vote counts
func (r *Reader) GetProjects(ctx context.Context) (*ProjectsList, error) {
	tallyFiles, err := r.getTallyFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	projectData := make(map[string]*ProjectSummary)

	for _, tally := range tallyFiles {
		for projectID, cid := range tally.AggregatedProjects {
			if data, exists := projectData[projectID]; exists {
				data.VoteCount++
				data.Epochs++
				data.LastCid = cid
			} else {
				projectData[projectID] = &ProjectSummary{
					ProjectID: projectID,
					VoteCount: 1,
					Epochs:    1,
					LastCid:   cid,
				}
			}
		}
	}

	projects := make([]ProjectSummary, 0, len(projectData))
	for _, data := range projectData {
		projects = append(projects, *data)
	}

	return &ProjectsList{Projects: projects}, nil
}

// GetTimeline returns a timeline of events
func (r *Reader) GetTimeline(ctx context.Context) (*Timeline, error) {
	tallyFiles, err := r.getTallyFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tally files: %w", err)
	}

	events := make([]TimelineEvent, 0, len(tallyFiles))

	for _, tally := range tallyFiles {
		events = append(events, TimelineEvent{
			Type:      "epoch",
			Timestamp: tally.Timestamp,
			Data: map[string]any{
				"epoch_id":            tally.EpochID,
				"total_validators":    tally.TotalValidators,
				"eligible_nodes":      tally.EligibleNodesCount,
				"slot_count":          len(tally.SubmissionCounts),
				"aggregated_projects": len(tally.AggregatedProjects),
			},
		})
	}

	// Sort by timestamp ascending
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp > events[j].Timestamp {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return &Timeline{Events: events}, nil
}

// TallyDump represents the tally dump file structure
type TallyDump struct {
	EpochID            uint64            `json:"epoch_id"`
	DataMarket         string            `json:"data_market"`
	Timestamp          int64             `json:"timestamp"`
	SubmissionCounts   map[string]int    `json:"submission_counts"`
	EligibleNodesCount int               `json:"eligible_nodes_count"`
	TotalValidators    int               `json:"total_validators"`
	AggregatedProjects map[string]string `json:"aggregated_projects"`
	ValidatorBatchCIDs map[string]string `json:"validator_batch_cids"`
}

// getTallyFiles reads all tally dump files
func (r *Reader) getTallyFiles() ([]*TallyDump, error) {
	return r.getTallyFilesWithLimit(0)
}

// getRecentTallyFiles reads the N most recent tally files (by epoch ID descending)
// limit 0 means no limit (all files). Used by topology to avoid processing 1000+ files.
func (r *Reader) getRecentTallyFiles(limit int) ([]*TallyDump, error) {
	return r.getTallyFilesWithLimit(limit)
}

func (r *Reader) getTallyFilesWithLimit(limit int) ([]*TallyDump, error) {
	entries, err := filepath.Glob(filepath.Join(r.dumpDir, "epoch_*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob tally files: %w", err)
	}

	// Parse epoch IDs from filenames and sort descending (newest first)
	type pathEpoch struct {
		path  string
		epoch uint64
	}
	pathEpochs := make([]pathEpoch, 0, len(entries))
	for _, entry := range entries {
		base := filepath.Base(entry)
		// Format: epoch_24569676_4198bf81b55ee4af6f9ddc176f8021960813f641.json
		if strings.HasPrefix(base, "epoch_") && strings.HasSuffix(base, ".json") {
			mid := base[6 : len(base)-5] // strip "epoch_" and ".json"
			parts := strings.SplitN(mid, "_", 2)
			if len(parts) >= 1 {
				if epoch, err := strconv.ParseUint(parts[0], 10, 64); err == nil {
					pathEpochs = append(pathEpochs, pathEpoch{entry, epoch})
				}
			}
		}
	}
	sort.Slice(pathEpochs, func(i, j int) bool {
		return pathEpochs[i].epoch > pathEpochs[j].epoch
	})

	if limit > 0 && len(pathEpochs) > limit {
		pathEpochs = pathEpochs[:limit]
	}

	tallies := make([]*TallyDump, 0, len(pathEpochs))
	for _, pe := range pathEpochs {
		data, err := os.ReadFile(pe.path)
		if err != nil {
			continue
		}

		var tally TallyDump
		if err := json.Unmarshal(data, &tally); err != nil {
			continue
		}

		tallies = append(tallies, &tally)
	}

	return tallies, nil
}

// getTallyForEpoch reads a specific tally file
func (r *Reader) getTallyForEpoch(epochID uint64) (*TallyDump, error) {
	dataMarketLower := strings.ToLower(strings.TrimPrefix(r.dataMarket, "0x"))
	filename := fmt.Sprintf("epoch_%d_%s.json", epochID, dataMarketLower)
	filepath := filepath.Join(r.dumpDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var tally TallyDump
	if err := json.Unmarshal(data, &tally); err != nil {
		return nil, err
	}

	return &tally, nil
}

// dayFromTimestamp converts a timestamp to a day string
func (r *Reader) dayFromTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02")
}

// shortenID shortens an ID for display
func (r *Reader) shortenID(id string) string {
	if len(id) <= 16 {
		return id
	}
	return id[:8] + "..." + id[len(id)-4:]
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
