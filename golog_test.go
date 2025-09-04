package golog

import (
	"fmt"
	"testing"
	"time"
)

func TestInitLogger(t *testing.T) {
	defer Sync()
	InitLogger("test.log", 1024*10, true, 7)
	NewLog("test.log", 1024*10, true, 7)
	ShowBasePath = true
	DefaultUnit = Hour
	WarnHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
		fmt.Println(msg)
	}
	ErrorHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
		fmt.Println(msg)
	}
	Warn("警告")
	Error("失败")
	// golog.InitLogger("log/a.log", 1024, false, 10)
	// a := NewLog("log/a.log", 1024, true, 10)
	// for range 100 {
	// 	a.Info("foo", "aaaa", "bb")
	// }
	// a.Warn(color.New(color.BgYellow).Sprint("aaaa"), color.New(color.BgBlue).Sprint("bbbb"))
	// Level = DEBUG
	// // test()
	// a.Error("bar")

}
