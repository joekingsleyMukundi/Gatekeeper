package internals

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, time.Duration, error)
}

type SlidingWindowLimiter struct {
	client     *redis.Client
	limit      int
	windowSize time.Duration
}

func NewSlidingWindowLimiter(client *redis.Client, limit int, window time.Duration) RateLimiter {
	return &SlidingWindowLimiter{
		client:     client,
		limit:      limit,
		windowSize: window,
	}
}
func (rl *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, time.Duration, error) {
	now := time.Now().Unix()
	windowStart := now - int64(rl.windowSize.Seconds())
	member := fmt.Sprintf("%d-%d", now, rand.Intn(10000))
	pipeline := rl.client.Pipeline()
	pipeline.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: member,
	})
	pipeline.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	countCmd := pipeline.ZCard(ctx, key)
	pipeline.Expire(ctx, key, rl.windowSize)
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return false, 0, err
	}
	count := countCmd.Val()
	fmt.Printf("Rate limiter key: %s, Count: %d, Limit: %d\n", key, count, rl.limit)
	if count > int64(rl.limit) {
		oldestCmd := rl.client.ZRangeWithScores(ctx, key, 0, 0)
		oldestItems, err := oldestCmd.Result()
		if err != nil || len(oldestItems) == 0 {
			return false, 0, fmt.Errorf("rate limited, but couldn't fetch oldest timestamp")
		}
		oldestTime := int64(oldestItems[0].Score)
		retryAfter := time.Duration((oldestTime+int64(rl.windowSize.Seconds()))-now) * time.Second
		if retryAfter < 0 {
			retryAfter = 1 * time.Second
		}

		fmt.Printf("Rate limited! Retry after: %v\n", retryAfter)
		return false, retryAfter, nil
	}

	return true, 0, nil
}
