package golog

import "time"

const (
	Second time.Duration = 1 * time.Second
	Minute time.Duration = 60 * Second
	Hour   time.Duration = 60 * Minute
	Day    time.Duration = 24 * Hour
	Week   time.Duration = 7 * Day
)

// 默认的清理日志时间单位
var DefaultUnit time.Duration = Day
