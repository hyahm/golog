package golog

import (
	"sync"
	"time"
)

// sampler 固定窗口限流采样器：每个窗口内最多放行 max 条日志，其余丢弃。
type sampler struct {
	mu       sync.Mutex
	max      int
	interval time.Duration
	count    int
	dropped  int
	window   time.Time
}

// newSampler 创建限流采样器，interval 小于等于 0 时按 1 秒处理。
func newSampler(max int, interval time.Duration) *sampler {
	if interval <= 0 {
		interval = time.Second
	}
	return &sampler{
		max:      max,
		interval: interval,
		window:   time.Now(),
	}
}

// allow 判断当前日志是否放行，窗口滚动时自动重置计数。
func (s *sampler) allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if now.Sub(s.window) >= s.interval {
		s.window = now
		s.count = 0
		s.dropped = 0
	}
	if s.count < s.max {
		s.count++
		return true
	}
	s.dropped++
	return false
}

// Dropped 返回当前窗口内因限流被丢弃的日志条数。
func (s *sampler) Dropped() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dropped
}
