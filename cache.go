package golog

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

type msgLog struct {
	// Prev    string    // 深度对于的路径
	Msg   string // 日志信息
	Level level  // 日志级别
	Ctime time.Time
	// deep     int               // 向外的深度，  Upfunc 才会用到
	Color        []color.Attribute // 颜色
	Line         string            // 行号
	out          bool              // 文件还是控制台
	filepath     string
	dir          string
	name         string
	size         int64 // 文件大小
	everyDay     bool
	format       string
	Hostname     string
	Label        map[string]string
	ErrorHandler func(time.Time, string, string, string, map[string]string)
	InfoHandler  func(time.Time, string, string, string, map[string]string)
	WarnHandler  func(time.Time, string, string, string, map[string]string)
}

var cache chan msgLog
var exit chan bool

func init() {
	cache = make(chan msgLog, 1000)
	exit = make(chan bool)
	// 增加1024字节的缓存， 也就是假设每一条日志的最大长度是1024
	b := make([]byte, 0, 1<<20+1024)
	cacheBuf = bytes.NewBuffer(b)
	go write()

}

var cacheBuf *bytes.Buffer

// 递归遍历文件夹
func walkDir() error {
	return filepath.Walk(_dir, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			Error(err)
			return err
		}

		// 如果是文件，打印文件路径和修改时间
		if !info.IsDir() && strings.Contains(fp, _name) {
			modTime := info.ModTime()
			if time.Since(modTime) > time.Duration(_expire)*DefaultUnit {
				os.Remove(fp)
			}
		}
		return nil
	})
}

func clean(ctx context.Context) {
	for {
		select {
		case <-time.After(time.Duration(_expire) * DefaultUnit):
			fmt.Println("clean log")
			walkDir()
		case <-ctx.Done():
			return
		}

	}
}

func write() {
	var c msgLog
	ticker := time.NewTicker(1 * time.Millisecond * 100)

	defer ticker.Stop() // 主函数退出前停止 Ticker，防止 goroutine 泄漏
	for {
		select {
		case <-ticker.C:
			// 如果写入文件的操作是空闲的， 那么就写入文件
			if cacheBuf.Len() > 0 {
				c.control(cacheBuf.Bytes())
				cacheBuf.Reset()
			}
		case c = <-cache:
			b, err := c.formatText()
			if err == nil {
				cacheBuf.Write(b.Bytes())
			}

		}

	}

}

func Sync() {
	// 等待日志写完
	close(cache)
	<-exit
}
