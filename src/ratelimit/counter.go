package ratelimit

import "sync"

// NewLimitCounter 创建并返回一个新的 LimitCounter 实例。
// LimitCounter 是一个简单的基于计数的限流器。
// 它跟踪每个 key 的请求次数，并在达到最大限制 (max) 之前允许请求。
//
// 这个限流器是内存中的，并且是并发安全的。
//
// 例如：如果你想限制每个用户 ID 每分钟只能尝试登录 5 次，
// 你可以创建一个 LimitCounter，max 设置为 5。
// 每次用户尝试登录时，调用 Consume(userID)。如果返回 true，则允许登录尝试；
// 如果返回 false，则表示该用户已达到限制，应拒绝登录尝试。
// 你还需要一个机制来定期（例如每分钟）调用 Clear() 或针对特定用户调用 Delete(userID) 来重置计数器。
//
// 参数:
//   max (int): 每个 key 允许的最大请求次数。一旦计数达到 max，后续对该 key 的 Consume 调用将返回 false。
//
// 返回值:
//   LimitCounter: 初始化后的 LimitCounter 结构体实例。
func NewLimitCounter(max int) LimitCounter {
	// 初始化 LimitCounter 结构体
	counter := LimitCounter{
		mu:      &sync.Mutex{},              // 初始化互斥锁，用于保证并发安全
		storage: map[string]int{},          // 初始化存储计数器的 map，key 是限流对象标识符，value 是当前计数值
		max:     max,                       // 设置最大允许的计数值
	}
	return counter
}

// LimitCounter 结构体定义了一个基于计数的限流器。
// 它内部使用一个 map 来存储每个 key 的当前计数值，并使用互斥锁来保证并发访问的安全。
type LimitCounter struct {
	mu      *sync.Mutex    // mu 是一个互斥锁 (Mutex)，用于保护 storage 的并发访问。
	                        // 在多 goroutine 环境下，对 map 的读写操作需要加锁，防止数据竞争。
	storage map[string]int // storage 是一个 map，用于存储每个 key 当前的请求计数值。
	                        // key 是用来标识限流对象的字符串，例如用户 ID、IP 地址等。
	                        // value 是该 key 对应的当前计数值。
	max     int            // max 是每个 key 允许的最大计数值。当 storage[key] 达到 max 时，限流触发。
}

// Consume 方法尝试为指定的 key 消耗一个计数。
// 如果当前 key 的计数值小于最大限制 (max)，则计数值加 1 并返回 true，表示请求被允许。
// 如果当前 key 的计数值已经达到或超过最大限制，则删除该 key 的记录并返回 false，表示请求被拒绝（触发限流）。
// 删除记录是为了防止 map 无限增长，当一个 key 达到限制后，它的计数就没必要继续保存了，
// 等待下次 Clear() 或手动 Delete()。
// 这个方法是并发安全的。
//
// 参数:
//   key (string): 需要进行限流判断和计数的标识符。
//
// 返回值:
//   bool: 如果请求被允许（未达到限制），返回 true；如果请求被拒绝（已达到限制），返回 false。
func (lc *LimitCounter) Consume(key string) bool {
	lc.mu.Lock()         // 加锁，防止并发访问 storage
	defer lc.mu.Unlock() // 使用 defer 确保在函数退出时解锁

	// 检查当前 key 的计数值是否小于最大限制
	if lc.storage[key] < lc.max {
		// 如果小于限制，计数值加 1
		lc.storage[key]++
		// 返回 true，表示允许本次请求
		return true
	}
	// 如果已达到或超过限制，删除该 key 的记录
	delete(lc.storage, key)
	// 返回 false，表示拒绝本次请求
	return false
}

// Delete 方法从计数器存储中移除指定的 key。
// 这通常用于在某个事件发生后（例如密码重置成功）主动清除某个 key 的限流计数，
// 或者用于配合外部逻辑实现计数器的过期。
// 这个方法是并发安全的。
//
// 参数:
//   key (string): 需要从存储中删除的标识符。
func (lc *LimitCounter) Delete(key string) {
	lc.mu.Lock()         // 加锁
	delete(lc.storage, key) // 从 map 中删除指定的 key
	lc.mu.Unlock()       // 解锁
}

// Clear 方法清空整个计数器存储。
// 它会创建一个新的空 map 来替换旧的 map。
// 将新 map 的容量设置为旧 map 大小的一半是一种优化，
// 假设在清空后，活跃的 key 数量可能会减少。
// 这常用于定期重置所有限流计数，例如每分钟或每小时清空一次。
// 这个方法是并发安全的。
func (lc *LimitCounter) Clear() {
	lc.mu.Lock()         // 加锁
	size := len(lc.storage) // 获取当前 map 的大小
	// 创建一个新的 map，容量预设为原大小的一半（可以根据实际情况调整）
	// 这可以释放旧 map 占用的内存，并为后续使用提供一个较小的初始容量
	lc.storage = make(map[string]int, size/2)
	lc.mu.Unlock()       // 解锁
}
