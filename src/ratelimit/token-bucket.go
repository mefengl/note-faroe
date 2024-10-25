package ratelimit

import (
	"math"
	"sync"
	"time"
)

func NewTokenBucketRateLimit(max int, refillInterval time.Duration) TokenBucketRateLimit {
	ratelimit := TokenBucketRateLimit{
		mu:                         &sync.Mutex{},
		storage:                    map[string]refillingTokenBucket{},
		max:                        max,
		refillIntervalMilliseconds: refillInterval.Milliseconds(),
	}
	return ratelimit
}

type TokenBucketRateLimit struct {
	mu                         *sync.Mutex
	storage                    map[string]refillingTokenBucket
	max                        int
	refillIntervalMilliseconds int64
}

func (rl *TokenBucketRateLimit) Check(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if _, ok := rl.storage[key]; !ok {
		return true
	}
	now := time.Now()
	refill := int((now.UnixMilli() - rl.storage[key].refilledAtUnixMilliseconds) / rl.refillIntervalMilliseconds)
	count := int(math.Min(float64(rl.storage[key].count+refill), float64(rl.max)))
	return count > 0
}

func (rl *TokenBucketRateLimit) Consume(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	if _, ok := rl.storage[key]; !ok {
		rl.storage[key] = refillingTokenBucket{rl.max - 1, now.UnixMilli()}
		return true
	}
	refill := int((now.UnixMilli() - rl.storage[key].refilledAtUnixMilliseconds) / rl.refillIntervalMilliseconds)
	count := int(math.Min(float64(rl.storage[key].count+refill), float64(rl.max)))
	if count < 1 {
		return false
	}
	rl.storage[key] = refillingTokenBucket{count - 1, now.UnixMilli()}
	return true
}

func (rl *TokenBucketRateLimit) AddTokenIfEmpty(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket, ok := rl.storage[key]
	if !ok {
		return
	}
	now := time.Now()
	refill := int((now.UnixMilli() - bucket.refilledAtUnixMilliseconds) / rl.refillIntervalMilliseconds)
	count := int(math.Min(float64(bucket.count+refill), float64(rl.max)))
	if count < 1 {
		rl.storage[key] = refillingTokenBucket{1, now.UnixMilli()}
	}
}

func (rl *TokenBucketRateLimit) Reset(key string) {
	rl.mu.Lock()
	delete(rl.storage, key)
	rl.mu.Unlock()
}

func (rl *TokenBucketRateLimit) Clear() {
	rl.mu.Lock()
	size := len(rl.storage)
	rl.storage = make(map[string]refillingTokenBucket, size/2)
	rl.mu.Unlock()
}

type refillingTokenBucket struct {
	count                      int
	refilledAtUnixMilliseconds int64
}

func NewExpiringTokenBucketRateLimit(max int, expiresIn time.Duration) ExpiringTokenBucketRateLimit {
	ratelimit := ExpiringTokenBucketRateLimit{
		mu:                    &sync.Mutex{},
		storage:               map[string]expiringTokenBucket{},
		max:                   max,
		expiresInMilliseconds: expiresIn.Milliseconds(),
	}
	return ratelimit
}

type ExpiringTokenBucketRateLimit struct {
	mu                    *sync.Mutex
	storage               map[string]expiringTokenBucket
	max                   int
	expiresInMilliseconds int64
}

func (rl *ExpiringTokenBucketRateLimit) Check(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	if _, ok := rl.storage[key]; !ok {
		return true
	}
	expiresAtMilliseconds := rl.storage[key].createdAtUnixMilliseconds + rl.expiresInMilliseconds
	if now.UnixMilli() >= expiresAtMilliseconds {
		return true
	}
	return rl.storage[key].count > 0
}

func (rl *ExpiringTokenBucketRateLimit) Consume(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	if _, ok := rl.storage[key]; !ok {
		rl.storage[key] = expiringTokenBucket{rl.max - 1, now.UnixMilli()}
		return true
	}
	expiresAtMilliseconds := rl.storage[key].createdAtUnixMilliseconds + rl.expiresInMilliseconds
	if now.UnixMilli() >= expiresAtMilliseconds {
		rl.storage[key] = expiringTokenBucket{rl.max - 1, now.UnixMilli()}
		return true
	}
	if rl.storage[key].count < 1 {
		return false
	}
	rl.storage[key] = expiringTokenBucket{rl.storage[key].count - 1, now.UnixMilli()}
	return true
}

func (rl *ExpiringTokenBucketRateLimit) AddTokenIfEmpty(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, ok := rl.storage[key]
	if !ok {
		return
	}
	count := int(math.Max(float64(bucket.count), 1))
	rl.storage[key] = expiringTokenBucket{count, bucket.createdAtUnixMilliseconds}
}

func (rl *ExpiringTokenBucketRateLimit) Reset(key string) {
	rl.mu.Lock()
	delete(rl.storage, key)
	rl.mu.Unlock()
}

func (rl *ExpiringTokenBucketRateLimit) Clear() {
	rl.mu.Lock()
	size := len(rl.storage)
	rl.storage = make(map[string]expiringTokenBucket, size/2)
	rl.mu.Unlock()
}

type expiringTokenBucket struct {
	count                     int
	createdAtUnixMilliseconds int64
}
