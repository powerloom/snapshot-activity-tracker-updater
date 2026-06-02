package dashboard

import (
	"context"
	"time"

	"p2p-debugger/internal/epochconfig"
)

// GetMeshHealth aggregates participation signals from epoch tallies in the time window.
// peer_count_24h = unique validator IDs seen; heartbeats_24h = sum of slot submission counts.
func (r *Reader) GetMeshHealth(ctx context.Context, windowMinutes int) (*MeshHealthResponse, error) {
	if windowMinutes <= 0 || windowMinutes > 1440 {
		windowMinutes = 1440
	}
	cutoff := time.Now().Add(-time.Duration(windowMinutes) * time.Minute).Unix()

	limit := epochconfig.DashboardMaxEpochPageLimit()
	if limit <= 0 {
		limit = 500
	}

	tallies, _, err := r.loadTallies(ctx, 0, limit)
	if err != nil {
		return nil, err
	}

	peers := make(map[string]struct{})
	var submissionObservations int

	for _, tally := range tallies {
		if tally == nil {
			continue
		}
		if tally.Timestamp > 0 && tally.Timestamp < cutoff {
			break
		}
		for _, vid := range validatorIDsFromTally(tally) {
			peers[vid] = struct{}{}
		}
		for _, n := range tally.SubmissionCounts {
			if n > 0 {
				submissionObservations += n
			}
		}
	}

	return &MeshHealthResponse{
		PeerCount24h:  len(peers),
		Heartbeats24h: submissionObservations,
		WindowMinutes: windowMinutes,
		Timestamp:     time.Now().UTC(),
		Source:        "watcher",
	}, nil
}
