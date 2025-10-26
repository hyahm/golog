package golog

import (
	"testing"
	"time"
)

func TestInitLogger(t *testing.T) {
	defer Sync()
	// InitLogger("aa.log", 10, false)
	// SetExpireDuration(time.Second)
	SetLevel(DEBUG)
	SetLogPriority(true, 100, time.Minute)
	// SetExpireDuration(time.Second * 10)
	// time.Sleep(10 * time.Second)
	// l := NewLog("", 0, true)
	// l2 := NewLog("", 0, true)
	// defer l.Sync()
	// defer l2.Sync()
	// NewLog("test.log", 10, false, 7)
	// ShowBasePath = true

	// WarnHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
	// 	fmt.Println(msg)
	// }
	// ErrorHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
	// 	fmt.Println(msg)
	// }
	for {
		Infof("消息%s", "asdfasdf")
		time.Sleep(10 * time.Microsecond)
	}

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
	// time.Sleep(time.Second * 100)
}
