package golog

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// _names 记录需过期清理的文件名及其过期时长，时长为 0 表示跟随全局 _expireClean。
var _names = make(map[string]time.Duration)
var mu = sync.RWMutex{}

// once 确保清理协程只启动一次。
var once = sync.Once{}

// getNames 返回已注册待清理文件名及过期时长的快照（副本），避免遍历时数据竞争。
func getNames() map[string]time.Duration {
	mu.RLock()
	defer mu.RUnlock()
	cp := make(map[string]time.Duration, len(_names))
	for k, v := range _names {
		cp[k] = v
	}
	return cp
}

// countNames 返回已注册的文件名数量。
func countNames() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(_names)
}

// addClean 注册需要过期清理的日志文件名及其过期时长（0 表示跟随全局值），
// 并确保清理协程已启动。
func addClean(expire time.Duration, names ...string) {
	mu.Lock()
	for _, v := range names {
		if v == "" {
			continue
		}
		_names[v] = expire
	}
	mu.Unlock()
	once.Do(func() {
		go clean()
	})
}

// setCleanExpire 更新已注册文件名的过期时长，未注册时忽略。
func setCleanExpire(name string, expire time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := _names[name]; ok {
		_names[name] = expire
	}
}

// _cleanInterval 过期清理协程的执行间隔（纳秒），默认 24 小时。
var _cleanInterval atomic.Int64

func init() {
	_cleanInterval.Store(int64(time.Hour * 24))
}

// SetCleanInterval 设置过期清理协程的执行间隔，默认 24 小时。
// 可在运行中随时调用，下一轮清理生效；主要用于测试与演示。
func SetCleanInterval(d time.Duration) {
	if d > 0 {
		_cleanInterval.Store(int64(d))
	}
}

// CleanNow 立即执行一次过期清理并返回结果，无需等待后台协程的下个轮询。
func CleanNow() error {
	return walkDir()
}

// clean 按设定间隔执行过期清理，间隔可在运行中通过 SetCleanInterval 修改。
func clean() {
	walkDir()
	for {
		d := time.Duration(_cleanInterval.Load())
		if d <= 0 {
			d = time.Hour * 24
		}
		t := time.NewTimer(d)
		<-t.C
		if countNames() == 0 {
			continue
		}
		walkDir()
	}
}

// walkDir 递归遍历日志目录，删除超过各自过期时长的已注册日志文件（含 .gz 归档），
// 注册时长为 0 的文件跟随全局 _expireClean。
func walkDir() error {
	names := getNames()

	return filepath.Walk(_dir, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 如果是文件，判断修改时间是否过期
		if !info.IsDir() {
			for name, expire := range names {
				if !strings.Contains(info.Name(), name) {
					continue
				}
				if expire <= 0 {
					expire = _expireClean
				}
				if time.Since(info.ModTime()) > expire {
					if err = os.Remove(fp); err != nil {
						log.Println(err)
					}
				}
				break
			}
		}
		return nil
	})
}
