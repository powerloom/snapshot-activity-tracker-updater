// Package epochconfig derives limits from DATA_MARKET_EPOCHS_PER_DAY and integer factors
// so operators configure multiples of one protocol "day" instead of raw epoch counts.
package epochconfig

import (
	"os"
	"strconv"
)

const defaultEpochsPerDay = 7200

// EpochsPerDay is the number of epochs in one protocol day for this data market
// (e.g. ~7200 at ~12s per epoch). Override with DATA_MARKET_EPOCHS_PER_DAY.
func EpochsPerDay() int {
	v := os.Getenv("DATA_MARKET_EPOCHS_PER_DAY")
	if v == "" {
		return defaultEpochsPerDay
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultEpochsPerDay
	}
	return n
}

// TallyRetentionEpochs is the max number of epoch tallies to retain (files + Redis index).
// If TALLY_RETENTION_FILES is set, it wins. Otherwise: TALLY_RETENTION_EPOCHS_FACTOR * EpochsPerDay().
func TallyRetentionEpochs() int {
	if v := os.Getenv("TALLY_RETENTION_FILES"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n >= 0 {
			return n
		}
	}
	f := intFromEnv("TALLY_RETENTION_EPOCHS_FACTOR", 1)
	if f < 1 {
		f = 1
	}
	return EpochsPerDay() * f
}

// TallyRetentionDays is wall-clock age pruning (independent of epoch length).
func TallyRetentionDays() int {
	d := intFromEnv("TALLY_RETENTION_DAYS", 7)
	if d < 0 {
		return 0
	}
	return d
}

// PruneIntervalHours for periodic tally prune ticks.
func PruneIntervalHours() int {
	h := intFromEnv("TALLY_PRUNE_INTERVAL_HOURS", 1)
	if h < 1 {
		return 1
	}
	return h
}

// DashboardMaxEpochPageLimit caps ?limit= on /api/epochs and /api/timeline.
// DASHBOARD_MAX_EPOCH_PAGE_FACTOR * EpochsPerDay() (default factor 2 = two protocol days of headroom).
func DashboardMaxEpochPageLimit() int {
	f := intFromEnv("DASHBOARD_MAX_EPOCH_PAGE_FACTOR", 2)
	if f < 1 {
		f = 1
	}
	return EpochsPerDay() * f
}

// DefaultEpochPageLimit is the default page size when the client omits ?limit=.
func DefaultEpochPageLimit() int {
	f := intFromEnv("DASHBOARD_DEFAULT_EPOCH_PAGE_FACTOR", 0)
	if f < 1 {
		return 500
	}
	n := EpochsPerDay() * f
	if n < 1 {
		return 500
	}
	return n
}

// NetworkTopologyRecentEpochLimit is how many recent epochs to load for the topology graph.
// NETWORK_TOPOLOGY_EPOCH_FACTOR * EpochsPerDay() / NETWORK_TOPOLOGY_DAY_FRACTION_DENOMINATOR
// (defaults: factor 1, denom 240 → 30 epochs when EpochsPerDay=7200).
func NetworkTopologyRecentEpochLimit() int {
	denom := intFromEnv("NETWORK_TOPOLOGY_DAY_FRACTION_DENOMINATOR", 240)
	if denom < 1 {
		denom = 240
	}
	f := intFromEnv("NETWORK_TOPOLOGY_EPOCH_FACTOR", 1)
	if f < 1 {
		f = 1
	}
	n := EpochsPerDay() * f / denom
	if n < 1 {
		n = 1
	}
	max := EpochsPerDay()
	if n > max {
		return max
	}
	return n
}

func intFromEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
