package golog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	logPath   string        // 文件路径
	fileSize  int64         // 切割的文件大小
	everyDay  bool          // 每天一个来切割文件 （这个比上面个优先级高）
	cleanTime time.Duration = 0
	dir       string
)

// 文件名
var name string

// hostname
var hostname = ""

func init() {
	hostname, _ = os.Hostname()
}

func InitLogger(path string, size int64, everyday bool, ct ...time.Duration) {
	if path == "" {
		logPath = "."
		return
	}
	name = filepath.Base(path)
	dir = filepath.Dir(path)
	logPath = filepath.Clean(path)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}
	fileSize = size
	everyDay = everyday
	if len(ct) > 0 {
		cleanTime = ct[0]
	}
	go clean(dir, name)
}

// open file，  所有日志默认前面加了时间，
func Tracef(format string, args ...interface{}) {
	if Level <= TRACE {
		s(TRACE, fmt.Sprintf(format, args...))
	}
}

// open file，  所有日志默认前面加了时间，
func Debugf(format string, args ...interface{}) {
	if Level <= DEBUG {
		s(DEBUG, fmt.Sprintf(format, args...))
	}
}

// open file，  所有日志默认前面加了时间，
func Infof(format string, args ...interface{}) {
	if Level <= INFO {
		s(INFO, fmt.Sprintf(format, args...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Warnf(format string, args ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= WARN {
		s(WARN, fmt.Sprintf(format, args...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Errorf(format string, args ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= ERROR {
		s(ERROR, fmt.Sprintf(format, args...))
	}
}

func Fatalf(format string, args ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= FATAL {
		s(FATAL, fmt.Sprintf(format, args...))
	}
}

func UpFuncf(deep int, format string, args ...interface{}) {
	// deep打印函数的深度， 相对于当前位置向外的深度
	if Level <= DEBUG {
		s(DEBUG, fmt.Sprintf(format, args...), deep)
	}
}

// open file，  所有日志默认前面加了时间，
func Trace(msg ...interface{}) {
	// Access,
	if Level <= TRACE {
		s(TRACE, arrToString(msg...))
	}
}

// open file，  所有日志默认前面加了时间，
func Debug(msg ...interface{}) {
	// debug,
	if Level <= DEBUG {
		s(DEBUG, arrToString(msg...))
	}
}

// open file，  所有日志默认前面加了时间，
func Info(msg ...interface{}) {
	if Level <= INFO {
		s(INFO, arrToString(msg...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Warn(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= WARN {
		s(WARN, arrToString(msg...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Error(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= ERROR {
		s(ERROR, arrToString(msg...))
	}
}

func Fatal(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= FATAL {
		s(FATAL, arrToString(msg...))
	}
	Sync()
	os.Exit(1)
}

func UpFunc(deep int, msg ...interface{}) {
	// deep打印函数的深度， 相对于当前位置向外的深度
	if Level <= DEBUG {
		s(DEBUG, arrToString(msg...), deep)
	}
}

func arrToString(msg ...interface{}) string {
	ll := make([]string, 0, len(msg))
	for range msg {
		ll = append(ll, "%v")
	}
	return fmt.Sprintf(strings.Join(ll, ""), msg...)
}

func s(level level, msg string, deep ...int) {
	if len(deep) > 0 && deep[0] > 0 {
		msg = fmt.Sprintf("caller from %s -- %v", printFileline(deep[0]), msg)
	}
	cache <- msgLog{
		msg:     msg,
		level:   level,
		name:    name,
		create:  time.Now(),
		color:   GetColor(level),
		line:    printFileline(0),
		out:     logPath == "." || logPath == "",
		path:    dir,
		logPath: logPath,
	}

}
