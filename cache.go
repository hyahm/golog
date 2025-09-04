package golog

import (
	"time"

	"github.com/fatih/color"
)

type task struct {
	cache chan msgLog
}

type msgLog struct {
	// Prev    string    // 深度对于的路径
	Msg   string // 日志信息
	Level level  // 日志级别
	Ctime time.Time
	// deep     int               // 向外的深度，  Upfunc 才会用到
	Color    []color.Attribute // 颜色
	Line     string            // 行号
	out      bool              // 文件还是控制台
	filepath string
	dir      string
	name     string
	size     int64 // 默认单位M
	everyDay bool
	format   string
	Hostname string
	now      time.Time
	Label    map[string]string
	// ErrorHandler func(time.Time, string, string, string, map[string]string)
	// InfoHandler  func(time.Time, string, string, string, map[string]string)
	// WarnHandler  func(time.Time, string, string, string, map[string]string)
	// buf *bytes.Buffer
}

func (ml *msgLog) reset() {
	ml.Level = 0
	ml.Msg = ""
	ml.Ctime = time.Time{}
	ml.Color = nil
	ml.Label = nil
}

// var exit chan bool

var t *task

func init() {
	t = &task{
		cache: make(chan msgLog, 500),
	}
	// exit = make(chan bool)
	// 增加1024字节的缓存， 也就是假设每一条日志的最大长度是1024
	go t.write()

}

// 递归遍历文件夹
// func walkDir() error {
// 	return filepath.Walk(_dir, func(fp string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			Error(err)
// 			return err
// 		}

// 		// 如果是文件，打印文件路径和修改时间
// 		if !info.IsDir() && strings.Contains(fp, _name) {
// 			modTime := info.ModTime()
// 			if time.Since(modTime) > time.Duration(_expire)*DefaultUnit {
// 				os.Remove(fp)
// 			}
// 		}
// 		return nil
// 	})
// }

// func clean(ctx context.Context) {
// 	for {
// 		select {
// 		case <-time.After(time.Duration(_expire) * DefaultUnit):
// 			fmt.Println("clean log")
// 			walkDir()
// 		case <-ctx.Done():
// 			return
// 		}

// 	}
// }

// func SecondCache() {

// 	for c := range cache {
// 		c.control()
// 	}
// }

func (t *task) write() {
	cl := msgLog{
		// buf: bytes.NewBuffer(nil),
		// now: time.Now(),
	}
	// go SecondCache()
	ticker := time.NewTicker(1 * time.Second * 1)

	defer ticker.Stop() // 主函数退出前停止 Ticker，防止 goroutine 泄漏
	for {
		select {
		case <-ticker.C:
			if len(cl.Msg) > 0 && time.Since(cl.now).Milliseconds() > 100 {
				t.control(cl)
				cl.Msg = ""

			}

		case c := <-t.cache:
			cl.dir = c.dir
			cl.out = c.out
			cl.now = time.Now()
			cl.filepath = c.filepath
			cl.name = c.name
			cl.everyDay = c.everyDay
			cl.Ctime = c.Ctime
			cl.size = c.size
			if len(cl.Msg) < BLOCKSIZE {
				cl.Msg += c.Msg
			} else {
				cl.Msg += c.Msg
				t.control(cl)
				cl.Msg = ""
			}
		}
	}

}

func Sync() {
	// 等待所有通道写完日志写完, 如果日志量太大， 建议换成zap， zap 的速度是本日志库的约2倍
	time.Sleep(1 * time.Millisecond * 300)
	close(t.cache)
	time.Sleep(1 * time.Millisecond * 200)

}
