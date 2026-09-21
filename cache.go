package golog

import (
	"time"

	"github.com/fatih/color"
)

// msgLog 单条日志在异步通道与写线程之间传递的内部载体。
type msgLog struct {
	Msg      string // 已格式化的日志文本（进入写线程前为原始消息）
	Level    Level
	Ctime    time.Time
	Color    []color.Attribute // 控制台颜色
	Line     string            // 调用位置
	fields   []Field           // 结构化字段
	format   FormatFunc        // 格式化函数
	out      bool              // true 输出控制台，false 写文件
	console  bool              // 写文件时是否同时输出控制台
	compress bool              // 切割归档后是否 gzip 压缩
	dir      string
	name     string
	size     int64 // 按大小切割阈值(MB)
	everyDay bool  // 按天切割
	day      int   // 批量缓冲里最后一条日志的日期（天），用于跨天切割
}
