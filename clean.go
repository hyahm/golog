package golog

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var _names = make(map[string]struct{})
var mu = sync.RWMutex{}

// once 确保清理协程只启动一次。
var once = sync.Once{}

// getNames 返回已注册待清理文件名的快照（副本），避免遍历时数据竞争。
func getNames() map[string]struct{} {
	mu.RLock()
	defer mu.RUnlock()
	cp := make(map[string]struct{}, len(_names))
	for k := range _names {
		cp[k] = struct{}{}
	}
	return cp
}

// countNames 返回已注册的文件名数量。
func countNames() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(_names)
}

// addClean 注册需要过期清理的日志文件名，并确保清理协程已启动。
func addClean(names ...string) {
	mu.Lock()
	for _, v := range names {
		if v == "" {
			continue
		}
		_names[v] = struct{}{}
	}
	mu.Unlock()
	once.Do(func() {
		go clean()
	})
}

// clean 每 24 小时清理一次过期日志文件。
func clean() {
	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()
	walkDir()
	for range ticker.C {
		if countNames() == 0 {
			continue
		}
		walkDir()
	}
}

// walkDir 递归遍历日志目录，删除超过过期时间的已注册日志文件（含 .gz 归档）。
func walkDir() error {
	names := getNames()
	name := make([]string, 0, len(names))
	for k := range names {
		name = append(name, k)
	}

	return filepath.Walk(_dir, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 如果是文件，判断修改时间是否过期
		if !info.IsDir() && containsSlice(name, info.Name()) {
			if time.Since(info.ModTime()) > _expireClean {
				err = os.Remove(fp)
				if err != nil {
					log.Println(err)
				}
			}
		}
		return nil
	})
}

// containsSlice 判断 str 是否包含 ss 中任一元素。
func containsSlice(ss []string, str string) bool {
	for _, v := range ss {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}
