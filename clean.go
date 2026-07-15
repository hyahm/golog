package golog

import (
	"sync"
	"time"
)

var _names = make(map[string]struct{})
var mu = sync.RWMutex{}

func getNames() map[string]struct{} {
	mu.RLock()
	defer mu.RUnlock()
	return _names
}

func addClean(names ...string) {
	mu.Lock()
	defer mu.Unlock()
	for _, v := range names {
		_names[v] = struct{}{}
	}
	once.Do(func() {
		go clean()
	})
}

func clean() {
	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()
	walkDir()
	for range ticker.C {
		if len(_names) == 0 {
			continue
		}
		walkDir()

	}
}
