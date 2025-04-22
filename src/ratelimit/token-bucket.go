package ratelimit

import (
	"math"
	"sync"
	"time"
)

// --- Refilling Token Bucket (补充型令牌桶) ---
// 特点：令牌按固定间隔自动补充，有容量上限。

// NewTokenBucketRateLimit 创建补充型令牌桶限流器。
// max: 桶容量。
// refillInterval: 令牌补充间隔。
func NewTokenBucketRateLimit(max int, refillInterval time.Duration) TokenBucketRateLimit {
	ratelimit := TokenBucketRateLimit{
		mu:                         &sync.Mutex{},
		storage:                    map[string]refillingTokenBucket{},
		max:                        max,
		refillIntervalMilliseconds: refillInterval.Milliseconds(),
	}
	return ratelimit
}

// TokenBucketRateLimit 补充型令牌桶限流器结构。
type TokenBucketRateLimit struct {
	mu                         *sync.Mutex                  // 并发锁
	storage                    map[string]refillingTokenBucket // key -> 令牌桶状态
	max                        int                          // 最大容量
	refillIntervalMilliseconds int64                        // 补充间隔(ms)
}

// Check 检查是否有可用令牌 (不消耗)。
// 返回 true 表示有令牌或首次访问。
func (rl *TokenBucketRateLimit) Check(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket, ok := rl.storage[key]
	if !ok {
		return true // 首次访问，总是有令牌
	}
	now := time.Now()
	// 计算应补充的令牌
	refill := int((now.UnixMilli() - bucket.refilledAtUnixMilliseconds) / rl.refillIntervalMilliseconds)
	// 当前有效令牌数 (不超过 max)
	count := int(math.Min(float64(bucket.count+refill), float64(rl.max)))
	return count > 0 // 有令牌则返回 true
}

// Consume 尝试消耗一个令牌。
// 返回 true 表示成功消耗。
func (rl *TokenBucketRateLimit) Consume(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	bucket, ok := rl.storage[key]
	if !ok {
		// 首次消耗，创建新桶 (容量 max-1)
		rl.storage[key] = refillingTokenBucket{rl.max - 1, now.UnixMilli()}
		return true
	}
	// 计算应补充和当前有效令牌数
	refill := int((now.UnixMilli() - bucket.refilledAtUnixMilliseconds) / rl.refillIntervalMilliseconds)
	count := int(math.Min(float64(bucket.count+refill), float64(rl.max)))
	if count < 1 {
		return false // 无可用令牌
	}
	// 消耗一个令牌，更新状态
	rl.storage[key] = refillingTokenBucket{count - 1, now.UnixMilli()}
	return true
}

// AddTokenIfEmpty 如果桶为空，则添加一个令牌。
// 用于特殊场景，允许空桶后进行一次操作。
func (rl *TokenBucketRateLimit) AddTokenIfEmpty(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket, ok := rl.storage[key]
	if !ok {
		return // key 不存在
	}
	now := time.Now()
	// 计算当前有效令牌数
	refill := int((now.UnixMilli() - bucket.refilledAtUnixMilliseconds) / rl.refillIntervalMilliseconds)
	count := int(math.Min(float64(bucket.count+refill), float64(rl.max)))
	if count < 1 {
		// 桶空，添加一个令牌
		rl.storage[key] = refillingTokenBucket{1, now.UnixMilli()}
	}
}

// Reset 删除指定 key 的令牌桶记录。
func (rl *TokenBucketRateLimit) Reset(key string) {
	rl.mu.Lock()
	delete(rl.storage, key)
	rl.mu.Unlock()
}

// Clear 清空所有 key 的记录。
func (rl *TokenBucketRateLimit) Clear() {
	rl.mu.Lock()
	size := len(rl.storage)
	// 创建新 map (尝试回收内存)
	rl.storage = make(map[string]refillingTokenBucket, size/2)
	rl.mu.Unlock()
}

// refillingTokenBucket 补充型令牌桶状态。
type refillingTokenBucket struct {
	count                      int   // 当前令牌数
	refilledAtUnixMilliseconds int64 // 上次记录时间(ms)
}

// --- Expiring Token Bucket (过期型令牌桶) ---
// 特点：令牌有固定有效期，不自动补充。桶过期后下次请求会重置。

// NewExpiringTokenBucketRateLimit 创建过期型令牌桶限流器。
// max: 桶容量。
// expiresIn: 桶的有效期。
func NewExpiringTokenBucketRateLimit(max int, expiresIn time.Duration) ExpiringTokenBucketRateLimit {
	ratelimit := ExpiringTokenBucketRateLimit{
		mu:                    &sync.Mutex{},
		storage:               map[string]expiringTokenBucket{},
		max:                   max,
		expiresInMilliseconds: expiresIn.Milliseconds(),
	}
	return ratelimit
}

// ExpiringTokenBucketRateLimit 过期型令牌桶限流器结构。
type ExpiringTokenBucketRateLimit struct {
	mu                    *sync.Mutex                 // 并发锁
	storage               map[string]expiringTokenBucket // key -> 令牌桶状态
	max                   int                         // 最大容量
	expiresInMilliseconds int64                       // 有效期(ms)
}

// Check 检查是否有可用且未过期的令牌 (不消耗)。
// 返回 true 表示有令牌、首次访问或桶已过期(下次会重置)。
func (rl *ExpiringTokenBucketRateLimit) Check(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	bucket, ok := rl.storage[key]
	if !ok {
		return true // 首次访问
	}
	// 计算过期时间点
	expiresAtMilliseconds := bucket.createdAtUnixMilliseconds + rl.expiresInMilliseconds
	if now.UnixMilli() >= expiresAtMilliseconds {
		return true // 已过期 (下次 consume 会重置)
	}
	// 未过期，检查令牌数
	return bucket.count > 0
}

// Consume 尝试消耗一个令牌。
// 返回 true 表示成功消耗。
func (rl *ExpiringTokenBucketRateLimit) Consume(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	bucket, ok := rl.storage[key]
	if !ok {
		// 首次消耗，创建新桶
		rl.storage[key] = expiringTokenBucket{rl.max - 1, now.UnixMilli()}
		return true
	}
	// 计算过期时间点
	expiresAtMilliseconds := bucket.createdAtUnixMilliseconds + rl.expiresInMilliseconds
	if now.UnixMilli() >= expiresAtMilliseconds {
		// 已过期，重置桶并消耗一个
		rl.storage[key] = expiringTokenBucket{rl.max - 1, now.UnixMilli()}
		return true
	}
	// 未过期
	if bucket.count < 1 {
		return false // 无可用令牌
	}
	// 消耗一个令牌 (创建时间不变)
	rl.storage[key] = expiringTokenBucket{bucket.count - 1, bucket.createdAtUnixMilliseconds}
	return true
}

// AddTokenIfEmpty 如果桶为空 (且理论上未过期)，则将令牌数设置为 1。
// 注意：原代码逻辑未严格检查是否过期，可能需要审视。
func (rl *ExpiringTokenBucketRateLimit) AddTokenIfEmpty(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket, ok := rl.storage[key]
	if !ok {
		return // key 不存在
	}
	// 确保令牌数至少为 1 (创建时间不变)
	count := int(math.Max(float64(bucket.count), 1))
	rl.storage[key] = expiringTokenBucket{count, bucket.createdAtUnixMilliseconds}
}

// Reset 删除指定 key 的令牌桶记录。
func (rl *ExpiringTokenBucketRateLimit) Reset(key string) {
	rl.mu.Lock()
	delete(rl.storage, key)
	rl.mu.Unlock()
}

// Clear 清空所有 key 的记录。
func (rl *ExpiringTokenBucketRateLimit) Clear() {
	rl.mu.Lock()
	size := len(rl.storage)
	rl.storage = make(map[string]expiringTokenBucket, size/2)
	rl.mu.Unlock()
}

// expiringTokenBucket 过期型令牌桶状态。
type expiringTokenBucket struct {
	count                     int   // 当前令牌数
	createdAtUnixMilliseconds int64 // 创建时间(ms)，用于判断过期
}
