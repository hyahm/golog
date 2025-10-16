package golog

import (
	"time"

	"github.com/fatih/color"
)

type msgLog struct {
	// Prev    string    // 深度对于的路径
	Msg   string // 日志信息
	Level level  // 日志级别
	Ctime time.Time
	// deep     int               // 向外的深度，  Upfunc 才会用到
	Color    []color.Attribute // 颜色
	Line     string            // 行号
	out      bool              // 文件还是控制台
	dir      string
	name     string
	size     int64 // 默认单位M
	everyDay bool
	format   string
	Hostname string
	// now      time.Time
	Label map[string]string
	// ErrorHandler func(time.Time, string, string, string, map[string]string)
	// InfoHandler  func(time.Time, string, string, string, map[string]string)
	// WarnHandler  func(time.Time, string, string, string, map[string]string)
	// buf *bytes.Buffer
}

// var exit chan bool

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
