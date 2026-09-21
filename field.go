package golog

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Field 结构化日志字段，随日志一起输出（logfmt/JSON 两种格式均支持）。
type Field struct {
	Key string
	Val any
}

// Any 构造任意类型字段。
func Any(key string, val any) Field {
	return Field{Key: key, Val: val}
}

// Str 构造字符串字段。
func Str(key, val string) Field {
	return Field{Key: key, Val: val}
}

// Int 构造 int 字段。
func Int(key string, val int) Field {
	return Field{Key: key, Val: val}
}

// Int64 构造 int64 字段。
func Int64(key string, val int64) Field {
	return Field{Key: key, Val: val}
}

// Bool 构造 bool 字段。
func Bool(key string, val bool) Field {
	return Field{Key: key, Val: val}
}

// Float64 构造 float64 字段。
func Float64(key string, val float64) Field {
	return Field{Key: key, Val: val}
}

// Duration 构造 time.Duration 字段，按字符串渲染。
func Duration(key string, d time.Duration) Field {
	return Field{Key: key, Val: d}
}

// Err 构造 error 字段，key 固定为 error。
func Err(err error) Field {
	return Field{Key: "error", Val: err}
}

// fieldValString 将字段值渲染为 logfmt 风格字符串。
func fieldValString(v any) string {
	switch x := v.(type) {
	case string:
		// 含空白或引号等特殊字符时加引号，保证可解析
		if strings.ContainsAny(x, ` "=`) {
			return strconv.Quote(x)
		}
		return x
	case error:
		return fieldValString(x.Error())
	case time.Duration:
		return x.String()
	case time.Time:
		return x.Format("2006-01-02 15:04:05.000")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// fieldValAny 将字段值转换为 JSON 可序列化类型。
func fieldValAny(v any) any {
	switch x := v.(type) {
	case error:
		return x.Error()
	case time.Duration:
		return x.String()
	default:
		return v
	}
}

// renderFields 渲染为 " k1=v1 k2=v2" 形式，无字段时返回空串。
func renderFields(fields []Field) string {
	if len(fields) == 0 {
		return ""
	}
	var b strings.Builder
	for _, f := range fields {
		b.WriteByte(' ')
		b.WriteString(f.Key)
		b.WriteByte('=')
		b.WriteString(fieldValString(f.Val))
	}
	return b.String()
}
