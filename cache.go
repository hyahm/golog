package golog

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

type msgLog struct {
	logPath string
	// Prev    string    // 深度对于的路径
	Msg    string    // 日志信息
	Level  level     // 日志级别
	create time.Time // 创建日志的时间
	Ctime  string
	// deep     int               // 向外的深度，  Upfunc 才会用到
	Color    []color.Attribute // 颜色
	Line     string            // 行号
	out      bool              // 文件还是控制台
	path     string
	name     string
	size     int64 // 文件大小
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

func clean(ctx context.Context, dir, name string) {
	for {
		select {
		case <-time.After(time.Duration(cleanTime) * time.Hour * 24):
			fs, err := ioutil.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, f := range fs {
				if strings.Contains(f.Name(), name) {
					os.Remove(filepath.Join(logPath, f.Name()))
				}
			}
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
