// expire_demo 以真实服务方式验证 golog 实例级过期清理功能。
//
// 启动：go run .
// 观察点：
//  1. 三个日志文件持续写入，体积不断增长
//  2. 各实例停止写入后，文件按"各自设置的过期时长"先后消失
//     a.log 10 秒过期（实例单独设置）
//     b.log 20 秒过期（实例单独设置）
//     c.log 30 秒过期（跟随全局 SetExpireDuration）
//  3. 全程由后台清理协程自动完成（演示中间隔调为 3 秒，默认 24 小时）
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hyahm/golog"
)

const dir = "log_demo"

type instance struct {
	name    string
	expire  time.Duration // 0 表示跟随全局
	write   time.Duration // 持续写入时长
	log     *golog.Log
	stopped bool
}

// listDir 打印目录内文件的名称、大小与老化时间。
func listDir() {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("读取目录失败:", err)
		return
	}
	if len(entries) == 0 {
		fmt.Println("                                          [目录已空]")
		return
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var sb strings.Builder
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		fmt.Fprintf(&sb, "%s(%d字节,老化%s) ", e.Name(), info.Size(),
			time.Since(info.ModTime()).Round(time.Second))
	}
	fmt.Println("   ", sb.String())
}

func main() {
	golog.SetDir(dir)
	golog.SetExpireDuration(30 * time.Second) // 全局默认 30 秒
	golog.SetCleanInterval(3 * time.Second)   // 演示用：3 秒清理一次，默认 24 小时

	// for 循环创建实例，各自设置不同的过期时长
	ins := []*instance{
		{name: "a.log", expire: 10 * time.Second, write: 8 * time.Second},
		{name: "b.log", expire: 20 * time.Second, write: 15 * time.Second},
		{name: "c.log", expire: 0, write: 25 * time.Second}, // 0 = 跟随全局 30 秒
	}
	for _, it := range ins {
		it.log = golog.NewLog(it.name, 0, false)
		if it.expire > 0 {
			it.log.SetExpireDuration(it.expire)
		}
		exp := it.expire
		if exp == 0 {
			exp = 30 * time.Second
		}
		fmt.Printf("[%s] 创建 %-6s 过期时长=%s 写入持续=%s\n",
			time.Now().Format("15:04:05"), it.name, exp, it.write)
	}
	fmt.Printf("[%s] 初始状态:", time.Now().Format("15:04:05"))
	listDir()

	start := time.Now()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for range tick.C {
		el := time.Since(start).Round(time.Second)

		// 各实例在写入期内持续打日志，到期后 Sync 关闭句柄等待清理
		for _, it := range ins {
			if !it.stopped {
				if el <= it.write {
					it.log.Infof("%s 正常运行 已写入 %s", it.name, el)
				} else {
					it.log.Sync() // 关闭文件句柄（Windows 下删除前必须关闭）
					it.stopped = true
					fmt.Printf("[%s] %-6s 停止写入，等待过期清理...\n",
						time.Now().Format("15:04:05"), it.name)
				}
			}
		}

		fmt.Printf("[%s] 已运行 %3v:", time.Now().Format("15:04:05"), el)
		listDir()

		// 全部文件清理完毕，演示结束
		if entries, _ := os.ReadDir(dir); len(entries) == 0 && el > 30*time.Second {
			fmt.Printf("[%s] 全部文件已按各自过期时长清理完毕，演示结束\n",
				time.Now().Format("15:04:05"))
			return
		}
	}
}
