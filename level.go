package golog

// type level int
type Level int

const (
	ALL Level = iota
	TRACE
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

// 日志级别
var _level Level = INFO

func SetLevel(l Level) {
	_level = l
}

func (l Level) String() string {
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

func (l Level) Int() int {
	return int(l)
}
