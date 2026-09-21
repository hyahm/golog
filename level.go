package golog

// Level 日志级别，值越小级别越低。
type Level int32

// 支持的日志级别，通过 SetLevel 设置输出阈值。
const (
	ALL Level = iota
	TRACE
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

// String 返回级别名称字符串。
func (l Level) String() string {
	switch l {
	case ALL:
		return "ALL"
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	case PANIC:
		return "PANIC"
	default:
		return "INFO"
	}
}

// Int 返回级别的整数值。
func (l Level) Int() int {
	return int(l)
}

// SetLevel 设置全局日志级别，可在运行时并发安全调用。
func SetLevel(l Level) {
	getDefaultLog().SetLevel(l)
}

// GetLevel 返回当前全局日志级别。
func GetLevel() Level {
	return getDefaultLog().Level()
}
