package golog

import (
	"fmt"
	"sync"
	"time"
)

var duplicateVal duplicate

type msgCache struct {
	count int
	start time.Time
}

type duplicate struct {
	count  int
	key    map[string]*msgCache
	locker sync.RWMutex
}

func (d *duplicate) initDuplicate(count int) {
	d.count = count
	d.key = make(map[string]*msgCache)
	d.locker = sync.RWMutex{}
	go d.cleanDuplicate()
}

// 返回是否要写入
func (d *duplicate) addMsg(key string) bool {
	d.locker.Lock()
	defer d.locker.Unlock()
	if _, ok := d.key[key]; !ok {
		d.key[key] = &msgCache{
			start: time.Now(),
		}
		return true
	}
	d.key[key].count += 1
	if d.key[key].count == d.count {
		delete(d.key, key)
	}
	return false
}

func (d *duplicate) cleanDuplicate() {
	for {
		d.locker.Lock()
		for k := range d.key {
			fmt.Println(d.key[k].start.Sub(time.Now()))
			// 如果
		}
		time.Sleep(time.Minute)
	}
}
