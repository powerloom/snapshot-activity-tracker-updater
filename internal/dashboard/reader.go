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

	"p2p-debugger/internal/epochconfig"
	"p2p-debugger/redis"
)

// Reader reads tally data from Redis (preferred) or local tally JSON files.
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

// TallyDump matches JSON from tally_dumper / Redis.
type TallyDump struct {
	EpochID              uint64                  `json:"epoch_id"`
	DataMarket           string                  `json:"data_market"`
	Timestamp            int64                   `json:"timestamp"`
	SubmissionCounts     map[string]int          `json:"submission_counts"`
	EligibleNodesCount   int                     `json:"eligible_nodes_count"`
	TotalValidators      int                     `json:"total_validators"`
	AggregatedProjects     map[string]string       `json:"aggregated_projects"`
	ValidatorBatchCIDs   map[string]string       `json:"validator_batch_cids"`
	ValidatorSummaries   []ValidatorEpochSummary `json:"validator_summaries,omitempty"`
}

// ValidatorEpochSummary mirrors tally_dumper (for JSON decode).
type ValidatorEpochSummary struct {
	ValidatorID           string `json:"validator_id"`
	BatchCID              string `json:"batch_cid,omitempty"`
	HasBatchCID           bool   `json:"has_batch_cid"`
	ProjectIDsCount       int    `json:"project_ids_count"`
	SubmissionRowsCount int    `json:"submission_rows_count"`
	ProjectVotesCount     int    `json:"project_votes_count"`
}

// loadTallies returns tallies from Redis when the index is non-empty; otherwise reads files.
// offset/limit apply to epoch ordering (newest first). limit 0 = all from offset.
func (r *Reader) loadTallies(ctx context.Context, offset, limit int) ([]*TallyDump, int64, error) {
	useRedis, err := redis.HasTallyData(ctx, r.dataMarket)
	if err != nil {
		return nil, 0, err
	}
	if useRedis {
		total, err := redis.TallyIndexCount(ctx, r.dataMarket)
		if err != nil {
			return nil, 0, err
		}
		ids, err := redis.ListEpochIDsDesc(ctx, r.dataMarket, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		raws, err := redis.FetchTallyJSONs(ctx, r.dataMarket, ids)
		if err != nil {
			return nil, 0, err
		}
		out := make([]*TallyDump, 0, len(raws))
		for _, raw := range raws {
			if raw == nil {
				continue
			}
			var t TallyDump
			if err := json.Unmarshal(raw, &t); err != nil {
				continue
			}
			out = append(out, &t)
		}
		return out, total, nil
	}

	tallies, err := r.getTallyFilesWithOffsetLimit(offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return tallies, int64(len(tallies)), nil
}

func (r *Reader) getTallyForEpoch(ctx context.Context, epochID uint64) (*TallyDump, error) {
	if redis.RedisClient != nil {
		raw, err := redis.GetTallyJSON(ctx, r.dataMarket, epochID)
		if err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			var t TallyDump
			if err := json.Unmarshal(raw, &t); err != nil {
				return nil, err
			}
			return &t, nil
		}
	}
	return r.getTallyForEpochFromFile(epochID)
}

func (r *Reader) getTallyForEpochFromFile(epochID uint64) (*TallyDump, error) {
	dataMarketLower := strings.ToLower(strings.TrimPrefix(r.dataMarket, "0x"))
	filename := fmt.Sprintf("epoch_%d_%s.json", epochID, dataMarketLower)
	p := filepath.Join(r.dumpDir, filename)

	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	var tally TallyDump
	if err := json.Unmarshal(data, &tally); err != nil {
		return nil, err
	}

	return &tally, nil
}

func (r *Reader) getTallyFilesWithOffsetLimit(offset, limit int) ([]*TallyDump, error) {
	entries, err := filepath.Glob(filepath.Join(r.dumpDir, "epoch_*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob tally files: %w", err)
	}

	type pathEpoch struct {
		path  string
		epoch uint64
	}
	pathEpochs := make([]pathEpoch, 0, len(entries))
	for _, entry := range entries {
		base := filepath.Base(entry)
		if strings.HasPrefix(base, "epoch_") && strings.HasSuffix(base, ".json") {
			mid := base[6 : len(base)-5]
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

	if offset > 0 && offset < len(pathEpochs) {
		pathEpochs = pathEpochs[offset:]
	} else if offset >= len(pathEpochs) {
		return []*TallyDump{}, nil
	}

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

// GetDashboardSummary returns the dashboard summary
func (r *Reader) GetDashboardSummary(ctx context.Context) (*DashboardSummary, error) {
	currentDay, err := redis.Get(ctx, redis.CurrentDayKey(r.dataMarket))
	if err != nil {
		return nil, fmt.Errorf("failed to get current day: %w", err)
	}

	var tallyFiles []*TallyDump
	var totalEpochs int64

	useRedis, _ := redis.HasTallyData(ctx, r.dataMarket)
	if useRedis {
		var n int64
		n, err = redis.TallyIndexCount(ctx, r.dataMarket)
		if err != nil {
			return nil, err
		}
		totalEpochs = n
		tallyFiles, _, err = r.loadTallies(ctx, 0, 0)
	} else {
		tallyFiles, err = r.getTallyFilesWithOffsetLimit(0, 0)
		totalEpochs = int64(len(tallyFiles))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load tallies: %w", err)
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
		TotalEpochs:      int(totalEpochs),
		TotalValidators:  totalValidators,
		TotalSlots:       totalSlots,
		ActiveValidators: activeValidators,
		CurrentDay:       currentDay,
	}, nil
}

// GetNetworkTopology uses the N most recent epochs.
func (r *Reader) GetNetworkTopology(ctx context.Context) (*NetworkTopology, error) {
	tallyFiles, _, err := r.loadTallies(ctx, 0, epochconfig.NetworkTopologyRecentEpochLimit())
	if err != nil {
		return nil, fmt.Errorf("failed to get tally data: %w", err)
	}

	topology := &NetworkTopology{
		Nodes: []TopologyNode{},
		Links: []TopologyLink{},
	}

	validatorNodes := make(map[string]bool)
	slotNodes := make(map[string]bool)
	projectNodes := make(map[string]bool)

	for _, tally := range tallyFiles {
		validatorsInEpoch := validatorIDsFromTally(tally)
		for _, vid := range validatorsInEpoch {
			if !validatorNodes[vid] {
				topology.Nodes = append(topology.Nodes, TopologyNode{
					ID:    vid,
					Type:  "validator",
					Label: r.shortenID(vid),
					Properties: map[string]any{
						"epoch_id": tally.EpochID,
					},
				})
				validatorNodes[vid] = true
			}
		}

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

			for _, validatorID := range validatorsInEpoch {
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

// GetEpochs returns epochs with optional pagination (?offset=&limit=).
func (r *Reader) GetEpochs(ctx context.Context, offset, limit int) (*EpochsList, error) {
	total, err := r.epochListTotal(ctx)
	if err != nil {
		return nil, err
	}

	tallyFiles, _, err := r.loadTallies(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get epochs: %w", err)
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

	return &EpochsList{
		Epochs: epochs,
		Total:  int(total),
		Offset: offset,
		Limit:  limit,
	}, nil
}

func (r *Reader) epochListTotal(ctx context.Context) (int64, error) {
	useRedis, err := redis.HasTallyData(ctx, r.dataMarket)
	if err != nil {
		return 0, err
	}
	if useRedis {
		return redis.TallyIndexCount(ctx, r.dataMarket)
	}
	entries, err := filepath.Glob(filepath.Join(r.dumpDir, "epoch_*.json"))
	if err != nil {
		return 0, err
	}
	return int64(len(entries)), nil
}

// GetEpochDetail returns detailed information about a specific epoch
func (r *Reader) GetEpochDetail(ctx context.Context, epochID uint64) (*EpochDetail, error) {
	tally, err := r.getTallyForEpoch(ctx, epochID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally for epoch %d: %w", epochID, err)
	}

	if tally == nil {
		return nil, fmt.Errorf("epoch %d not found", epochID)
	}

	batches := validatorBatchesFromTally(tally)

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

func validatorBatchesFromTally(tally *TallyDump) []ValidatorBatch {
	if len(tally.ValidatorSummaries) > 0 {
		out := make([]ValidatorBatch, 0, len(tally.ValidatorSummaries))
		for _, vs := range tally.ValidatorSummaries {
			out = append(out, ValidatorBatch{
				ValidatorID:           vs.ValidatorID,
				BatchCID:              vs.BatchCID,
				HasBatchCID:           vs.HasBatchCID,
				ProjectIDsCount:       vs.ProjectIDsCount,
				SubmissionRowsCount:   vs.SubmissionRowsCount,
				ProjectVotesCount:     vs.ProjectVotesCount,
			})
		}
		return out
	}
	batches := make([]ValidatorBatch, 0, len(tally.ValidatorBatchCIDs))
	for validatorID, batchCID := range tally.ValidatorBatchCIDs {
		batches = append(batches, ValidatorBatch{
			ValidatorID: validatorID,
			BatchCID:    batchCID,
			HasBatchCID: batchCID != "",
		})
	}
	return batches
}

func validatorIDsFromTally(tally *TallyDump) []string {
	if len(tally.ValidatorSummaries) > 0 {
		out := make([]string, len(tally.ValidatorSummaries))
		for i := range tally.ValidatorSummaries {
			out[i] = tally.ValidatorSummaries[i].ValidatorID
		}
		return out
	}
	out := make([]string, 0, len(tally.ValidatorBatchCIDs))
	for id := range tally.ValidatorBatchCIDs {
		out = append(out, id)
	}
	return out
}

// GetValidators aggregates over all retained epochs (may be heavy; prefer Redis-backed retention).
func (r *Reader) GetValidators(ctx context.Context) (*ValidatorsList, error) {
	tallyFiles, _, err := r.loadTallies(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally data: %w", err)
	}

	validatorData := make(map[string]*ValidatorSummary)

	for _, tally := range tallyFiles {
		seen := validatorIDsFromTally(tally)
		for _, validatorID := range seen {
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
	tallyFiles, _, err := r.loadTallies(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally data: %w", err)
	}

	batchesByEpoch := make(map[uint64]string)
	projects := make(map[string]bool)
	lastActive := int64(0)
	totalEpochs := 0

	for _, tally := range tallyFiles {
		var batchCID string
		found := false
		if len(tally.ValidatorSummaries) > 0 {
			for _, vs := range tally.ValidatorSummaries {
				if vs.ValidatorID == validatorID {
					batchCID = vs.BatchCID
					found = true
					break
				}
			}
		} else {
			if c, ok := tally.ValidatorBatchCIDs[validatorID]; ok {
				batchCID = c
				found = true
			}
		}
		if found {
			batchesByEpoch[tally.EpochID] = batchCID
			totalEpochs++
			if tally.Timestamp > lastActive {
				lastActive = tally.Timestamp
			}
		}
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
	tallyFiles, _, err := r.loadTallies(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally data: %w", err)
	}

	slotData := make(map[string]*SlotSummary)
	slotDays := make(map[string]map[string]bool)
	slotEligible := make(map[string]map[string]bool)

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
	tallyFiles, _, err := r.loadTallies(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally data: %w", err)
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
		SlotID:           slotID,
		SubmissionsByDay: submissionsByDay,
		EligibleByDay:    eligibleByDay,
		TotalSubmits:     totalSubmits,
		TotalEligible:    len(eligibleByDay),
	}, nil
}

// GetProjects returns a list of projects with vote counts
func (r *Reader) GetProjects(ctx context.Context) (*ProjectsList, error) {
	tallyFiles, _, err := r.loadTallies(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally data: %w", err)
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

// GetTimeline returns a timeline of events (optional ?offset=&limit=).
func (r *Reader) GetTimeline(ctx context.Context, offset, limit int) (*Timeline, error) {
	tallyFiles, _, err := r.loadTallies(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally data: %w", err)
	}
	total, err := r.epochListTotal(ctx)
	if err != nil {
		return nil, err
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

	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp > events[j].Timestamp {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return &Timeline{
		Events: events,
		Total:  int(total),
		Offset: offset,
		Limit:  limit,
	}, nil
}

func (r *Reader) dayFromTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02")
}

func (r *Reader) shortenID(id string) string {
	if len(id) <= 16 {
		return id
	}
	return id[:8] + "..." + id[len(id)-4:]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
