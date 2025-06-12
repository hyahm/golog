package golog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	logPath   string // 文件路径
	fileSize  int64  // 切割的文件大小
	everyDay  bool   // 每天一个来切割文件 （这个比上面个优先级高）
	cleanTime int    = 0
	dir       string
)

// 文件名
var name string
var Format string = "{{ .Ctime }} - [{{ .Level }}]{{ if .Label }} - {{ range $k,$v := .Label}}[{{$k}}:{{$v}}]{{end}}{{end}} - {{.Hostname}} - {{.Line}} - {{.Msg}}"
var label map[string]string
var labelLock sync.RWMutex

// hostname
var hostname = ""
var cancel context.CancelFunc

func init() {
	hostname, _ = os.Hostname()
	label = make(map[string]string)
	labelLock = sync.RWMutex{}
}

// size: kb
func InitLogger(path string, size int64, everyday bool, ct ...int) {
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
	var ctx context.Context
	if logPath != "." && cleanTime > 0 {
		ctx, cancel = context.WithCancel(context.Background())
		go clean(ctx, dir, name)
	}

}

func Close() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("No need to close")
		}
	}()
	cancel()
}

func AddLabel(key, value string) {
	labelLock.RLock()
	defer labelLock.RUnlock()
	label[key] = value
}

func SetLabel(key, value string) {
	labelLock.RLock()
	defer labelLock.RUnlock()
	label[key] = value
}

func DelLabel(key string) {
	labelLock.Lock()
	defer labelLock.Unlock()
	delete(label, key)
}

func GetLabel() map[string]string {
	labelLock.RLock()
	defer labelLock.RUnlock()
	return label
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
		s(TRACE, arrToString(msg...)+"\n")
	}
}

// open file，  所有日志默认前面加了时间，
func Debug(msg ...interface{}) {
	// debug,
	if Level <= DEBUG {
		s(DEBUG, arrToString(msg...)+"\n")
	}
}

// open file，  所有日志默认前面加了时间，
func Info(msg ...interface{}) {
	if Level <= INFO {
		s(INFO, arrToString(msg...)+"\n")
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Warn(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= WARN {
		s(WARN, arrToString(msg...)+"\n")
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Error(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= ERROR {
		s(ERROR, arrToString(msg...)+"\n")
	}
}

func Fatal(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= FATAL {
		s(FATAL, arrToString(msg...)+"\n")
	}
	Sync()
	os.Exit(1)
}

func UpFunc(deep int, msg ...interface{}) {
	// deep打印函数的深度， 相对于当前位置向外的深度
	if Level <= DEBUG {
		s(DEBUG, arrToString(msg...)+"\n", deep)
	}
}

func arrToString(msg ...interface{}) string {
	ll := make([]string, 0, len(msg))
	for _, v := range msg {
		ll = append(ll, fmt.Sprintf("%v", v))
	}
	return strings.Join(ll, " ")
}

func s(level level, msg string, deep ...int) {
	if len(deep) > 0 && deep[0] > 0 {
		if ShowBasePath {
			msg = fmt.Sprintf("caller from %s -- %v", printBaseFileline(deep[0]), msg)
		} else {
			msg = fmt.Sprintf("caller from %s -- %v", printFileline(deep[0]), msg)
		}

	}

	now := time.Now()
	ml := msgLog{
		Msg:      msg,
		Level:    level,
		name:     name,
		create:   now,
		Ctime:    now.Format("2006-01-02 15:04:05"),
		Color:    GetColor(level),
		Line:     printFileline(0),
		out:      logPath == "." || logPath == "",
		path:     dir,
		logPath:  logPath,
		size:     fileSize,
		Hostname: hostname,
		format:   Format,
		Label:    GetLabel(),
	}
	if ShowBasePath {
		ml.Line = printBaseFileline(0)
	}

	cache <- ml

}
