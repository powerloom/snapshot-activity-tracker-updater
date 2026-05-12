package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

// StoreTally writes the tally JSON blob and updates the epoch index. Prunes oldest epochs
// when retentionFiles > 0 and count exceeds the cap.
func StoreTally(ctx context.Context, dataMarket string, epochID uint64, jsonBytes []byte, retentionFiles int) error {
	if RedisClient == nil {
		return errors.New("Redis client not initialized")
	}
	dm := normalizeDataMarket(dataMarket)
	if dm == "" {
		return errors.New("data market address required for tally storage")
	}
	epochKey := TallyEpochKey(dm, epochID)
	indexKey := TallyEpochIndexZSet(dm)

	pipe := RedisClient.Pipeline()
	pipe.Set(ctx, epochKey, jsonBytes, 0)
	pipe.ZAdd(ctx, indexKey, redis.Z{Score: float64(epochID), Member: strconv.FormatUint(epochID, 10)})
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("store tally: %w", err)
	}

	if retentionFiles > 0 {
		if err := PruneTallyEpochs(ctx, dm, retentionFiles); err != nil {
			return err
		}
	}
	return nil
}

// GetTallyJSON returns raw JSON for one epoch, or nil if missing.
func GetTallyJSON(ctx context.Context, dataMarket string, epochID uint64) ([]byte, error) {
	if RedisClient == nil {
		return nil, errors.New("Redis client not initialized")
	}
	dm := normalizeDataMarket(dataMarket)
	if dm == "" {
		return nil, errors.New("data market address required")
	}
	epochKey := TallyEpochKey(dm, epochID)
	val, err := RedisClient.Get(ctx, epochKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

// TallyIndexCount returns how many epochs are indexed for the data market.
func TallyIndexCount(ctx context.Context, dataMarket string) (int64, error) {
	if RedisClient == nil {
		return 0, errors.New("Redis client not initialized")
	}
	dm := normalizeDataMarket(dataMarket)
	if dm == "" {
		return 0, nil
	}
	return ZCard(ctx, TallyEpochIndexZSet(dm))
}

// HasTallyData returns true if the ZSET index has at least one epoch.
func HasTallyData(ctx context.Context, dataMarket string) (bool, error) {
	n, err := TallyIndexCount(ctx, dataMarket)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ListEpochIDsDesc returns epoch IDs newest-first. offset/limit apply to the reversed order.
// limit <= 0 means all epochs (use with care).
func ListEpochIDsDesc(ctx context.Context, dataMarket string, offset, limit int) ([]uint64, error) {
	if RedisClient == nil {
		return nil, errors.New("Redis client not initialized")
	}
	dm := normalizeDataMarket(dataMarket)
	if dm == "" {
		return nil, nil
	}
	indexKey := TallyEpochIndexZSet(dm)
	var start, stop int64
	if limit <= 0 {
		start = int64(offset)
		stop = -1
	} else {
		start = int64(offset)
		stop = int64(offset + limit - 1)
	}
	members, err := ZRevRange(ctx, indexKey, start, stop)
	if err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(members))
	for _, m := range members {
		e, err := strconv.ParseUint(m, 10, 64)
		if err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

// FetchTallyJSONs loads tally JSON for the given epoch IDs (order preserved).
func FetchTallyJSONs(ctx context.Context, dataMarket string, epochIDs []uint64) ([][]byte, error) {
	if RedisClient == nil {
		return nil, errors.New("Redis client not initialized")
	}
	dmKey := strings.ToLower(strings.TrimSpace(dataMarket))
	if dmKey == "" {
		return nil, errors.New("data market address required")
	}
	if len(epochIDs) == 0 {
		return nil, nil
	}
	pipe := RedisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(epochIDs))
	for i, eid := range epochIDs {
		cmds[i] = pipe.Get(ctx, TallyEpochKey(dmKey, eid))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	out := make([][]byte, len(epochIDs))
	for i, cmd := range cmds {
		b, err := cmd.Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				out[i] = nil
				continue
			}
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

// PruneTallyEpochs removes the oldest epochs from Redis until at most keepNewest remain.
func PruneTallyEpochs(ctx context.Context, dataMarket string, keepNewest int) error {
	if RedisClient == nil || keepNewest <= 0 {
		return nil
	}
	dm := normalizeDataMarket(dataMarket)
	if dm == "" {
		return nil
	}
	indexKey := TallyEpochIndexZSet(dm)
	total, err := ZCard(ctx, indexKey)
	if err != nil || total <= int64(keepNewest) {
		return err
	}
	removeCount := int(total) - keepNewest
	if removeCount <= 0 {
		return nil
	}
	// Oldest epochs have lowest scores — ZRANGE ascending
	oldMembers, err := ZRange(ctx, indexKey, 0, int64(removeCount-1))
	if err != nil {
		return err
	}
	if len(oldMembers) == 0 {
		return nil
	}
	epochKeys := make([]string, 0, len(oldMembers))
	for _, m := range oldMembers {
		e, err := strconv.ParseUint(m, 10, 64)
		if err != nil {
			continue
		}
		epochKeys = append(epochKeys, TallyEpochKey(dm, e))
	}
	if len(epochKeys) > 0 {
		if _, err := Del(ctx, epochKeys...); err != nil {
			return err
		}
	}
	if _, err := ZRem(ctx, indexKey, oldMembers...); err != nil {
		return err
	}
	return nil
}

func normalizeDataMarket(addr string) string {
	return strings.ToLower(strings.TrimSpace(addr))
}
