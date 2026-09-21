package golog

import (
	"sync"
	"time"
)

// duplicate 重复日志采样器：窗口期内相同的日志只放行第一条，
// 达到阈值后放行一条并重新计数（借鉴 zap 的采样思想）。
type duplicate struct {
	count  int
	key    map[string]*msgCache
	max    int
	locker sync.RWMutex

	stop     chan struct{}
	stopOnce sync.Once
}

type msgCache struct {
	count int
	start time.Time
}

// newDuplicate 创建采样器并启动过期清理协程，count 为窗口内允许重复的最大条数。
func newDuplicate(count int, dd time.Duration) *duplicate {
	d := &duplicate{
		max:      count,
		key:      make(map[string]*msgCache),
		locker:   sync.RWMutex{},
		stop:     make(chan struct{}),
		stopOnce: sync.Once{},
	}

	go d.cleanDuplicate(dd)
	return d
}

// close 停止内部清理协程，重复调用安全。
func (d *duplicate) close() {
	d.stopOnce.Do(func() {
		close(d.stop)
	})
}

// addMsg 返回该条日志是否允许写入。
func (d *duplicate) addMsg(key string) bool {

	d.locker.Lock()
	defer d.locker.Unlock()
	if _, ok := d.key[key]; !ok {
		d.key[key] = &msgCache{
			start: time.Now(),
			count: 1,
		}
		return true
	}
	d.key[key].count += 1
	if d.key[key].count >= d.max {
		delete(d.key, key)
	}
	return false
}

// cleanDuplicate 周期性清理过期的采样记录，close 后退出。
func (d *duplicate) cleanDuplicate(dd time.Duration) {
	timer := time.NewTimer(dd)
	defer timer.Stop()
	for {
		select {
		case <-d.stop:
			return
		case <-timer.C:
			d.locker.Lock()
			for k := range d.key {
				if time.Since(d.key[k].start) > dd {
					delete(d.key, k)
				}
			}
			d.locker.Unlock()
			timer.Reset(dd)
		}
	}
}
