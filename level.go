package golog

// type level int
type LogLevel int
type level LogLevel

const (
	ALL level = iota
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
	case 1:
		return "TRACE"
	case 2:
		return "DEBUG"
	case 3:
		return "INFO"
	case 4:
		return "WARN"
	case 5:
		return "ERROR"
	case 6:
		return "FATAL"
	default:
		return "INFO"
	}
}

func (l level) Int() int {
	return int(l)
}
