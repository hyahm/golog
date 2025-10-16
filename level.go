package golog

// type level int
type LogLevel int
type level LogLevel

const (
	ALL level = iota * 10
	TRACE
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

// 日志级别
var _level level = INFO

func SetLevel(l level) {
	_level = l
}

func (l level) String() string {
	switch l {
	case 0:
		return "ALL"
	case 10:
		return "TRACE"
	case 20:
		return "DEBUG"
	case 30:
		return "INFO"
	case 40:
		return "WARN"
	case 50:
		return "ERROR"
	case 60:
		return "FATAL"
	default:
		return "INFO"
	}
}

func (l level) Int() int {
	return int(l)
}
