package main

import (
	"time"

	"github.com/fatih/color"
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	golog.ShowBasePath = true
	a := golog.NewLog("", 0, false, 10)
	a.Info("foo", "aaaa", "bb")
	a.Info(color.New(color.BgYellow).Sprint("aaaa"), color.New(color.BgBlue).Sprint("bbbb"))
	golog.Level = golog.DEBUG
	test()
	golog.Info("bar")
	time.Sleep(10 * time.Second)

}

func test() {
	// 此方法的日志级别是DEBUG， 所以调试的时候必须将日志级别设置成DEBUG，不然不会显示
	golog.UpFunc(1, "who call me") // 2022-03-04 10:49:38 - [DEBUG] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:16 - caller from C:/work/golog/example/example.go:11 -- who call me
}
