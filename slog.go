package golog

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

// SlogHandler 将 golog 适配为标准库 log/slog 的 Handler，
// 可通过 slog.New(golog.NewSlogHandler(l)) 接入所有依赖 slog 的库。
type SlogHandler struct {
	l      *Log
	attrs  []Field
	groups []string
}

// NewSlogHandler 返回基于 l 的 slog.Handler；l 为 nil 时使用全局 logger。
func NewSlogHandler(l *Log) slog.Handler {
	if l == nil {
		l = getDefaultLog()
	}
	return &SlogHandler{l: l}
}

// slogLevel 将 slog 级别映射为 golog 级别。
func slogLevel(sl slog.Level) Level {
	switch {
	case sl >= slog.LevelError:
		return ERROR
	case sl >= slog.LevelWarn:
		return WARN
	case sl >= slog.LevelInfo:
		return INFO
	default:
		return DEBUG
	}
}

// Enabled 判断该级别日志是否会被记录。
func (h *SlogHandler) Enabled(_ context.Context, sl slog.Level) bool {
	return h.l.enabled(slogLevel(sl))
}

// Handle 将一条 slog 记录转写为 golog 日志（含字段、分组前缀、调用位置与 context 字段）。
func (h *SlogHandler) Handle(ctx context.Context, r slog.Record) error {
	prefix := strings.Join(h.groups, ".")
	fields := make([]Field, 0, len(h.attrs)+int(r.NumAttrs())+2)
	fields = append(fields, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		key := a.Key
		if prefix != "" {
			key = prefix + "." + key
		}
		fields = append(fields, Any(key, a.Value.Resolve().Any()))
		return true
	})
	fields = append(fields, FieldsFromContext(ctx)...)

	line := ""
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		if f, more := fs.Next(); f.File != "" {
			_ = more
			line = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
	}
	h.l.logAt(slogLevel(r.Level), r.Message, line, fields)
	return nil
}

// WithAttrs 返回携带新增属性的子 Handler。
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := &SlogHandler{
		l:      h.l,
		attrs:  append([]Field(nil), h.attrs...),
		groups: append([]string(nil), h.groups...),
	}
	prefix := strings.Join(nh.groups, ".")
	for _, a := range attrs {
		key := a.Key
		if prefix != "" {
			key = prefix + "." + key
		}
		nh.attrs = append(nh.attrs, Any(key, a.Value.Resolve().Any()))
	}
	return nh
}

// WithGroup 返回增加分组的子 Handler，分组名作为字段前缀。
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	nh := &SlogHandler{
		l:      h.l,
		attrs:  append([]Field(nil), h.attrs...),
		groups: append(append([]string(nil), h.groups...), name),
	}
	return nh
}
