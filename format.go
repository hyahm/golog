package golog

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

// FormatFunc 自定义日志格式化函数：将一条日志渲染为单行字符串（须包含行尾换行）。
type FormatFunc func(level Level, ctime time.Time, line, msg string, fields []Field) string

var _formatFunc atomic.Value // 存储 FormatFunc

// SetFormatFunc 设置自定义日志格式，可在运行时并发安全调用。
func SetFormatFunc(f FormatFunc) {
	if f != nil {
		_formatFunc.Store(f)
	}
}

// loadFormatFunc 返回当前格式化函数，未设置时返回 defaultFormat。
func loadFormatFunc() FormatFunc {
	if f, ok := _formatFunc.Load().(FormatFunc); ok && f != nil {
		return f
	}
	return defaultFormat
}

var (
	_timeLayout atomic.Value // string，默认时间布局
	_utc        atomic.Bool  // 是否使用 UTC 时间输出
)

// SetTimeLayout 设置默认格式与 JSON 格式的时间布局（Go time 布局语法），
// 默认 "2006-01-02 15:04:05.000"。
func SetTimeLayout(layout string) {
	if layout != "" {
		_timeLayout.Store(layout)
	}
}

// SetUTC 设置日志时间是否使用 UTC 输出，默认使用本地时区。
func SetUTC(b bool) {
	_utc.Store(b)
}

// formatTime 按 layout/UTC 配置渲染时间。
func formatTime(t time.Time) string {
	if _utc.Load() {
		t = t.UTC()
	}
	layout, ok := _timeLayout.Load().(string)
	if !ok || layout == "" {
		layout = "2006-01-02 15:04:05.000"
	}
	return t.Format(layout)
}

// JsonFormat 以单行 JSON 输出，自动转义 msg/line 中的特殊字符，
// 结构化字段会合并为 JSON 顶层键。
func JsonFormat(level Level, ctime time.Time, line, msg string, fields []Field) string {
	m := make(map[string]any, len(fields)+4)
	m["createTime"] = formatTime(ctime)
	m["level"] = level.String()
	m["line"] = line
	m["msg"] = msg
	for _, f := range fields {
		m[f.Key] = fieldValAny(f.Val)
	}
	b, err := json.Marshal(m)
	if err != nil {
		// 存在无法序列化的字段值时降级为基础字段
		return fmt.Sprintf(`{"createTime": %q, "level": %q, "line": %q, "msg": %q}`+"\n",
			formatTime(ctime), level.String(), line, msg)
	}
	return string(b) + "\n"
}

// defaultFormat 默认文本格式：时间 -- [级别] -- 位置 -- 消息 k=v ...
func defaultFormat(level Level, ctime time.Time, line, msg string, fields []Field) string {
	return formatTime(ctime) + " -- [" + level.String() + "] -- " + line + " -- " + msg + renderFields(fields) + "\n"
}
