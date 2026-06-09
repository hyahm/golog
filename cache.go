package golog

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
)

type msgLog struct {
	// Prev    string    // 深度对于的路径
	Msg   string // 日志信息
	Level Level  // 日志级别
	Ctime time.Time
	// deep     int               // 向外的深度，  Upfunc 才会用到
	Color    []color.Attribute // 颜色
	Line     string            // 行号
	out      bool              // 文件还是控制台
	dir      string
	name     string
	size     int64 // 默认单位M
	everyDay bool
	format   func(level Level, ctime time.Time, line, msg string) string
	day      int
}

type cacheName struct {
	name map[string]struct{}
	mu   sync.RWMutex
}

var cn *cacheName

func init() {
	cn = &cacheName{
		name: make(map[string]struct{}),
		mu:   sync.RWMutex{},
	}
}

func checkName(name string) {
	fmt.Println(name)
	if name == "" || name == "." {
		return
	}
	cn.mu.Lock()
	defer cn.mu.Unlock()
	if _, ok := cn.name[name]; ok {
		panic("Repeated Sync() invocations on log instances of the same name: " + name)
	}
	cn.name[name] = struct{}{}
}
