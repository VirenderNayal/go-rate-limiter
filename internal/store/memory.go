// /internal/store/memory.go

package store

import (
	"context"
	"sync"
)

type windowData struct {
	bucket int64
	count  int64
}

type MemoryFixedWindowStore struct {
	data map[string]windowData
	mu   sync.Mutex
}

func (s *MemoryFixedWindowStore) IncrementAndGet(ctx context.Context, key string, window int64, bucket int64) (int64, error) {
	s.mu.Lock()

	defer s.mu.Unlock()

	entry := s.data[key]

	if entry.bucket != bucket {
		entry.bucket = bucket
		entry.count = 1

		s.data[key] = entry

		return 1, nil
	}

	entry.count++
	s.data[key] = entry

	return entry.count, nil
}

func NewMemoryFixedWindowStore() *MemoryFixedWindowStore {
	return &MemoryFixedWindowStore{data: make(map[string]windowData)}
}

type MemorySlidingWindowStore struct {
	timestamps map[string][]int64
	mu         sync.Mutex
}

func (s *MemorySlidingWindowStore) AddAndCountTimestamps(ctx context.Context, key string, now int64, windowStart int64) (int64, error) {
	s.mu.Lock()

	defer s.mu.Unlock()

	allTimestamps := s.timestamps[key]

	var validTimestamps []int64
	for _, ts := range allTimestamps {
		if ts > windowStart {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	validTimestamps = append(validTimestamps, now)

	s.timestamps[key] = validTimestamps

	return int64(len(validTimestamps)), nil
}

func NewMemorySlidingWindowStore() *MemorySlidingWindowStore {
	return &MemorySlidingWindowStore{timestamps: make(map[string][]int64)}
}

type bucketData struct {
	tokens    int64
	timestamp int64
}

type MemoryTokenBucketStore struct {
	data map[string]bucketData
	mu   sync.Mutex
}

func (s *MemoryTokenBucketStore) AllowAndUpdate(ctx context.Context, key string, limit int64, window int64, now int64) (bool, error) {
	s.mu.Lock()

	defer s.mu.Unlock()

	entry := s.data[key]

	if entry.timestamp == 0 {
		entry.timestamp = now
		entry.tokens = limit
	}

	elapsed := now - entry.timestamp
	refill := (elapsed * limit) / window

	entry.tokens = min(entry.tokens+refill, limit)

	if entry.tokens < 1 {
		s.data[key] = entry
		return false, nil
	}

	entry.tokens--
	entry.timestamp = now

	s.data[key] = entry

	return true, nil
}

func NewMemoryTokenBucketStore() *MemoryTokenBucketStore {
	return &MemoryTokenBucketStore{data: make(map[string]bucketData)}
}
