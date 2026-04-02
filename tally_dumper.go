package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"p2p-debugger/internal/epochconfig"
	"p2p-debugger/redis"
)

// ValidatorEpochSummary is per-validator visibility for an epoch (includes rows without batch IPFS CID).
type ValidatorEpochSummary struct {
	ValidatorID         string `json:"validator_id"`
	BatchCID              string `json:"batch_cid,omitempty"`
	HasBatchCID           bool   `json:"has_batch_cid"`
	ProjectIDsCount       int    `json:"project_ids_count"`
	SubmissionRowsCount int    `json:"submission_rows_count"`
	ProjectVotesCount     int    `json:"project_votes_count"`
}

// TallyDump represents a per-epoch tally dump
type TallyDump struct {
	EpochID              uint64                  `json:"epoch_id"`
	DataMarket           string                  `json:"data_market"`
	Timestamp            int64                   `json:"timestamp"`
	SubmissionCounts     map[string]int          `json:"submission_counts"` // slotID -> count
	EligibleNodesCount   int                     `json:"eligible_nodes_count"`
	TotalValidators      int                     `json:"total_validators"`
	AggregatedProjects   map[string]string       `json:"aggregated_projects"`  // projectID -> winning CID
	ValidatorBatchCIDs   map[string]string       `json:"validator_batch_cids"` // validatorID -> batch_ipfs_cid (when CID present)
	ValidatorSummaries   []ValidatorEpochSummary `json:"validator_summaries,omitempty"`
}

// TallyDumper handles generating and managing per-epoch tally dumps
type TallyDumper struct {
	dumpDir            string
	retentionFiles     int
	retentionDays      int
	pruneIntervalHours int
	enabled            bool
	enableFileDumps    bool
	enableRedis        bool
	dataMarket         string
	mu                 sync.Mutex
}

// NewTallyDumper creates a new tally dumper
func NewTallyDumper() *TallyDumper {
	dumpDir := os.Getenv("TALLY_DUMP_DIR")
	if dumpDir == "" {
		dumpDir = "./tallies"
	}

	retentionFiles := epochconfig.TallyRetentionEpochs()
	retentionDays := epochconfig.TallyRetentionDays()
	pruneInterval := epochconfig.PruneIntervalHours()

	enabled := os.Getenv("ENABLE_TALLY_DUMPS") != "false"
	enableFileDumps := os.Getenv("ENABLE_TALLY_FILE_DUMPS") != "false"
	enableRedis := os.Getenv("ENABLE_TALLY_REDIS") != "false"

	dataMarket := os.Getenv("DATA_MARKET_ADDRESS")

	return &TallyDumper{
		dumpDir:            dumpDir,
		retentionFiles:     retentionFiles,
		retentionDays:      retentionDays,
		pruneIntervalHours: pruneInterval,
		enabled:            enabled,
		enableFileDumps:    enableFileDumps,
		enableRedis:        enableRedis,
		dataMarket:         dataMarket,
	}
}

// Initialize creates the dump directory and starts pruning routine
func (td *TallyDumper) Initialize(ctx context.Context) error {
	if !td.enabled {
		log.Printf("Tally dumps disabled (ENABLE_TALLY_DUMPS=false)")
		return nil
	}

	if td.enableFileDumps {
		if err := os.MkdirAll(td.dumpDir, 0755); err != nil {
			return fmt.Errorf("failed to create tally dump directory: %w", err)
		}
	}

	log.Printf("Tally dumps: dir=%s, epochsPerDay=%d, retentionEpochs=%d, retentionDays=%d, pruneIntervalH=%d, file=%v, redis=%v",
		td.dumpDir, epochconfig.EpochsPerDay(), td.retentionFiles, td.retentionDays, td.pruneIntervalHours, td.enableFileDumps, td.enableRedis)

	if err := td.Prune(ctx); err != nil {
		log.Printf("Warning: Failed to prune on startup: %v", err)
	}

	go td.periodicPrune(ctx)

	return nil
}

// Dump writes a tally for an epoch (JSON file and/or Redis).
func (td *TallyDumper) Dump(ctx context.Context, tally *TallyDump) error {
	if !td.enabled || tally == nil {
		return nil
	}

	td.mu.Lock()
	defer td.mu.Unlock()

	jsonBytes, err := json.MarshalIndent(tally, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tally: %w", err)
	}

	if td.enableFileDumps {
		filename := fmt.Sprintf("epoch_%d_%s.json", tally.EpochID, strings.ToLower(strings.TrimPrefix(tally.DataMarket, "0x")))
		path := filepath.Join(td.dumpDir, filename)
		if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
			return fmt.Errorf("failed to write tally file: %w", err)
		}
	}

	if td.enableRedis && redis.RedisClient != nil && tally.DataMarket != "" {
		if err := redis.StoreTally(ctx, tally.DataMarket, tally.EpochID, jsonBytes, td.retentionFiles); err != nil {
			log.Printf("❌ Redis tally store failed: %v", err)
		}
	}

	validatorInfo := ""
	if len(tally.ValidatorBatchCIDs) > 0 {
		validatorList := make([]string, 0, len(tally.ValidatorBatchCIDs))
		for validatorID, batchCID := range tally.ValidatorBatchCIDs {
			validatorList = append(validatorList, fmt.Sprintf("%s:%s", validatorID, batchCID))
		}
		validatorInfo = fmt.Sprintf(", validatorBatches=%v", validatorList)
	}
	log.Printf("📊 TALLY: epoch=%d, dataMarket=%s, slots=%d, eligibleNodes=%d, validators=%d, summaries=%d%s",
		tally.EpochID, tally.DataMarket, len(tally.SubmissionCounts), tally.EligibleNodesCount, tally.TotalValidators, len(tally.ValidatorSummaries), validatorInfo)

	return nil
}

// Prune removes old tally files and aligns Redis.
func (td *TallyDumper) Prune(ctx context.Context) error {
	if !td.enabled {
		return nil
	}

	if td.enableFileDumps {
		if err := td.pruneFiles(); err != nil {
			return err
		}
	}

	if td.enableRedis && redis.RedisClient != nil && td.dataMarket != "" && td.retentionFiles > 0 {
		if err := redis.PruneTallyEpochs(ctx, td.dataMarket, td.retentionFiles); err != nil {
			log.Printf("Warning: Redis tally prune: %v", err)
		}
	}

	return nil
}

func (td *TallyDumper) pruneFiles() error {
	entries, err := os.ReadDir(td.dumpDir)
	if err != nil {
		return fmt.Errorf("failed to read dump directory: %w", err)
	}

	type fileInfo struct {
		path    string
		modTime time.Time
	}

	files := make([]fileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, fileInfo{
			path:    filepath.Join(td.dumpDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	cutoffTime := time.Now().Add(-time.Duration(td.retentionDays) * 24 * time.Hour)
	deletedCount := 0

	for i, file := range files {
		shouldDelete := false

		if td.retentionFiles > 0 && len(files)-i > td.retentionFiles {
			shouldDelete = true
		}

		if td.retentionDays > 0 && file.modTime.Before(cutoffTime) {
			shouldDelete = true
		}

		if shouldDelete {
			if err := os.Remove(file.path); err != nil {
				log.Printf("Warning: Failed to delete old tally file %s: %v", file.path, err)
			} else {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		log.Printf("Pruned %d old tally files (retention: %d files, %d days)", deletedCount, td.retentionFiles, td.retentionDays)
	}

	return nil
}

// periodicPrune runs pruning periodically
func (td *TallyDumper) periodicPrune(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(td.pruneIntervalHours) * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := td.Prune(ctx); err != nil {
				log.Printf("Error during periodic prune: %v", err)
			}
		}
	}
}

// buildValidatorEpochSummaries builds per-validator visibility from Level 2 batches.
func buildValidatorEpochSummaries(batches []*FinalizedBatch) []ValidatorEpochSummary {
	out := make([]ValidatorEpochSummary, 0, len(batches))
	for _, b := range batches {
		if b == nil || b.SequencerId == "" {
			continue
		}
		rows := 0
		for _, subs := range b.SubmissionDetails {
			rows += len(subs)
		}
		cid := b.BatchIPFSCID
		out = append(out, ValidatorEpochSummary{
			ValidatorID:         b.SequencerId,
			BatchCID:              cid,
			HasBatchCID:           cid != "",
			ProjectIDsCount:       len(b.ProjectIds),
			SubmissionRowsCount: rows,
			ProjectVotesCount:     len(b.ProjectVotes),
		})
	}
	return out
}
