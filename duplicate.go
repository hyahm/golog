package golog

import (
	"sync"
	"time"
)

var duplicateVal *duplicate

type msgCache struct {
	count int
	start time.Time
}

type duplicate struct {
	count  int
	key    map[string]*msgCache
	max    int
	locker sync.RWMutex
}

func newDuplicate(count int, dd time.Duration) *duplicate {
	d := &duplicate{
		max:    count,
		key:    make(map[string]*msgCache),
		locker: sync.RWMutex{},
	}

	go d.cleanDuplicate(dd)
	return d
}

// 返回是否要写入
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

func (d *duplicate) cleanDuplicate(dd time.Duration) {
	for {
		d.locker.Lock()
		for k := range d.key {
			if time.Since(d.key[k].start) > dd {
				delete(d.key, k)
			}
		}
		d.locker.Unlock()
		time.Sleep(dd)
	}
}
