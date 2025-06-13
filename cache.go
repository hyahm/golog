package golog

import (
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
	Msg    string    // 日志信息
	Level  level     // 日志级别
	create time.Time // 创建日志的时间
	Ctime  string
	// deep     int               // 向外的深度，  Upfunc 才会用到
	Color    []color.Attribute // 颜色
	Line     string            // 行号
	out      bool              // 文件还是控制台
	filepath string
	dir      string
	name     string
	size     int64 // 文件大小
	everyDay bool
	format   string
	Hostname string
	Label    map[string]string
}

var cache chan msgLog
var exit chan bool

func init() {
	cache = make(chan msgLog, 100000)
	exit = make(chan bool)
	go write()

}

// 递归遍历文件夹
func walkDir() error {
	fmt.Println(time.Now())
	return filepath.Walk(_dir, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			Error(err)
			return err
		}

		// 如果是文件，打印文件路径和修改时间
		if !info.IsDir() && strings.Contains(fp, _name) {

			modTime := info.ModTime()
			fmt.Println(fp, "-------", time.Since(modTime), "remove time:", time.Duration(_expire)*DefaultUnit)

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
	for c := range cache {
		c.control()
	}
	exit <- true
}

func Sync() {
	// 等待日志写完
	close(cache)
	<-exit
}
