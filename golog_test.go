package golog

import (
	"testing"
	"time"
)

func TestInitLogger(t *testing.T) {
	Clean("test.log")
	SetDir("log")

	SetExpireDuration(time.Second * 10)
	// time.Sleep(10 * time.Second)
	l := NewLog("cl.log", 0, true)
	l2 := NewLog("c2.log", 0, true)
	defer l.Sync()
	defer l2.Sync()
	SetDevelop()
	// NewLog("test.log", 10, false, 7)
	// ShowBasePath = true

	// WarnHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
	// 	fmt.Println(msg)
	// }
	// ErrorHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
	// 	fmt.Println(msg)
	// }
	l.Info("消息")
	l2.Info("消息")
	l.Error("失败")
	l2.Error("失败")
	l.Error("失败")
	l.Error("失败")
	// time.Sleep(1 * time.Second)
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
