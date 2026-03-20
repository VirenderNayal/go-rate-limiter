package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisFixedWindowStore struct {
	client *redis.Client
}

func (r *RedisFixedWindowStore) IncrementAndGet(ctx context.Context, key string, window int64, bucket int64) (int64, error) {
	redisKey := fmt.Sprintf("%s:%d", key, bucket)

	count, err := r.client.Incr(ctx, redisKey).Result()

	if err != nil {
		return 0, err
	}

	if count == 1 {
		if err := r.client.Expire(ctx, redisKey, time.Duration(window)*time.Second).Err(); err != nil {
			return 0, fmt.Errorf("failed to set expiry: %w", err)
		}
	}

	return count, nil
}

func NewRedisFixedWindowStore(client *redis.Client) *RedisFixedWindowStore {
	return &RedisFixedWindowStore{client}
}

type RedisSlidingWindowStore struct {
	client *redis.Client
}

func (r *RedisSlidingWindowStore) AddAndCountTimestamps(ctx context.Context, key string, now int64, windowStart int64) (int64, error) {

	if err := r.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10)).Err(); err != nil {
		return 0, fmt.Errorf("failed to remove expired entries: %w", err)
	}

	if err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d-%d", now, time.Now().UnixNano()),
	}).Err(); err != nil {
		return 0, fmt.Errorf("failed to add entry: %w", err)
	}

	if err := r.client.Expire(ctx, key, time.Duration(now-windowStart)*time.Second).Err(); err != nil {
		return 0, fmt.Errorf("failed to set expiry: %w", err)
	}

	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count entries: %w", err)
	}

	return count, nil
}

func NewRedisSlidingWindowStore(client *redis.Client) *RedisSlidingWindowStore {
	return &RedisSlidingWindowStore{client}
}

type RedisTokenBucketStore struct {
	client *redis.Client
}

var tokenBucketScript = redis.NewScript(`
    local tokens = tonumber(redis.call('HGET', KEYS[1], 'tokens'))
    local timestamp = tonumber(redis.call('HGET', KEYS[1], 'timestamp'))
    local capacity = tonumber(ARGV[1])
    local window = tonumber(ARGV[2])
    local now = tonumber(ARGV[3])

    if tokens == nil then
        tokens = capacity
        timestamp = now
    end

    local elapsed = now - timestamp
    local refill = math.floor((elapsed * capacity) / window)
    tokens = math.min(tokens + refill, capacity)

    if tokens < 1 then
        redis.call('HSET', KEYS[1], 'tokens', tokens, 'timestamp', timestamp)
        redis.call('EXPIRE', KEYS[1], window)
        return 0
    end

    redis.call('HSET', KEYS[1], 'tokens', tokens - 1, 'timestamp', now)
    redis.call('EXPIRE', KEYS[1], window)
    return 1
`)

func (r *RedisTokenBucketStore) AllowAndUpdate(ctx context.Context, key string, capacity int64, refillWindow int64, now int64) (bool, error) {
	result, err := tokenBucketScript.Run(ctx, r.client, []string{key}, capacity, refillWindow, now).Int64()
	if err != nil {
		return false, fmt.Errorf("token bucket script error: %w", err)
	}
	return result == 1, nil
}

func NewRedisTokenBucketStore(client *redis.Client) *RedisTokenBucketStore {
	return &RedisTokenBucketStore{client}
}
