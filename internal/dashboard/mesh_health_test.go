package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetMeshHealthFromFiles(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().Unix()
	dm := "0xmarket"

	writeTally := func(epochID uint64, ts int64, validator string, submissions map[string]int) {
		t.Helper()
		dump := TallyDump{
			EpochID:            epochID,
			DataMarket:         dm,
			Timestamp:          ts,
			SubmissionCounts:   submissions,
			EligibleNodesCount: 1,
			TotalValidators:    1,
			AggregatedProjects: map[string]string{},
			ValidatorBatchCIDs: map[string]string{validator: "QmX"},
		}
		raw, err := json.Marshal(dump)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Join(dir, "epoch_"+formatUint(epochID)+"_market.json")
		if err := os.WriteFile(name, raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeTally(100, now-3600, "0xval1", map[string]int{"1": 3, "2": 2})
	writeTally(99, now-7200, "0xval2", map[string]int{"3": 1})

	r := NewReader(dm, dir)
	h, err := r.GetMeshHealth(context.Background(), 1440)
	if err != nil {
		t.Fatal(err)
	}
	if h.PeerCount24h != 2 {
		t.Fatalf("peer_count_24h: got %d want 2", h.PeerCount24h)
	}
	if h.Heartbeats24h != 6 {
		t.Fatalf("heartbeats_24h: got %d want 6", h.Heartbeats24h)
	}
	if h.Source != "watcher" {
		t.Fatalf("source: got %q", h.Source)
	}
}

func formatUint(u uint64) string {
	return fmt.Sprintf("%d", u)
}
