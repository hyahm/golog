package golog

import (
	"fmt"
	"testing"
)

func TestInitLogger(t *testing.T) {
	defer Sync()

	SetDir("log")

	// time.Sleep(10 * time.Second)
	// InitLogger("aaa/test.log", 10, false, 7)
	// NewLog("test.log", 10, false, 7)
	// ShowBasePath = true

	// WarnHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
	// 	fmt.Println(msg)
	// }
	// ErrorHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
	// 	fmt.Println(msg)
	// }
	Info("消息")
	Warn("警告")
	Error("失败")
	fmt.Println(111)
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
