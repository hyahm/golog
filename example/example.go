package main

import (
	"github.com/fatih/color"
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	golog.ShowBasePath = true
	golog.DefaultUnit = golog.Second
	// golog.InitLogger("log/a.log", 1024, false, 10)
	a := golog.NewLog("log/a.log", 1024, false, 10)
	a.Debugf("foo", "aaaa", "bb")
	a.Warn(color.New(color.BgYellow).Sprint("aaaa"), color.New(color.BgBlue).Sprint("bbbb"))
	golog.Level = golog.DEBUG
	// test()
	a.Error("bar")
	// engine, err := xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
	// // See "Important settings" section.
	// db.SetConnMaxLifetime(time.Minute * 3)
	// db.SetMaxOpenConns(10)
	// db.SetMaxIdleConns(10)
	// for {
	// 	golog.Debugf("foo", "aaaa", "bb")
	// 	golog.Warn(color.New(color.BgYellow).Sprint("aaaa"), color.New(color.BgBlue).Sprint("bbbb"))
	// 	golog.Level = golog.DEBUG
	// 	test()
	// 	golog.Error("bar")
	// }
}

// func test() {
// 	// 此方法的日志级别是DEBUG， 所以调试的时候必须将日志级别设置成DEBUG，不然不会显示
// 	a.UpFunc(1, "who call me") // 2022-03-04 10:49:38 - [DEBUG] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:16 - caller from C:/work/golog/example/example.go:11 -- who call me
// }
