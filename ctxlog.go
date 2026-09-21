package golog

import "context"

// ctxKey context 中携带日志字段的键类型。
type ctxKey struct{}

// ContextWithFields 将结构化字段注入 context，
// 后续通过 Ctx(ctx) 打印的日志会自动携带这些字段。
func ContextWithFields(ctx context.Context, fields ...Field) context.Context {
	old, _ := ctx.Value(ctxKey{}).([]Field)
	merged := make([]Field, 0, len(old)+len(fields))
	merged = append(merged, old...)
	merged = append(merged, fields...)
	return context.WithValue(ctx, ctxKey{}, merged)
}

// ContextWithTraceID 注入 trace_id 字段，用于请求链路追踪。
func ContextWithTraceID(ctx context.Context, id string) context.Context {
	return ContextWithFields(ctx, Str("trace_id", id))
}

// FieldsFromContext 取出 context 中携带的全部日志字段，无字段时返回 nil。
func FieldsFromContext(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}
	f, _ := ctx.Value(ctxKey{}).([]Field)
	return f
}

// Ctx 返回绑定 context 的全局子 logger，自动携带注入 context 的字段（如 trace_id）。
func Ctx(ctx context.Context) *Log {
	return getDefaultLog().Ctx(ctx)
}
