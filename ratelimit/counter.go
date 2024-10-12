package ratelimit

import "sync"

func NewLimitCounter(max int) LimitCounter {
	counter := LimitCounter{
		mu:      &sync.Mutex{},
		storage: map[string]int{},
		max:     max,
	}
	return counter
}

type LimitCounter struct {
	mu      *sync.Mutex
	storage map[string]int
	max     int
}

func (lc *LimitCounter) Consume(key string) bool {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if lc.storage[key] < lc.max {
		lc.storage[key]++
		return true
	}
	delete(lc.storage, key)
	return false
}

func (lc *LimitCounter) Delete(key string) {
	lc.mu.Lock()
	delete(lc.storage, key)
	lc.mu.Unlock()
}

func (lc *LimitCounter) Clear() {
	lc.mu.Lock()
	size := len(lc.storage)
	lc.storage = make(map[string]int, size/2)
	lc.mu.Unlock()
}
