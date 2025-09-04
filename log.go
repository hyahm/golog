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
	// _logPath   string // 文件路径
	_fileSize int64  // 切割的文件大小默认单位M
	_everyDay bool   // 每天一个来切割文件 （这个比上面个优先级高）
	_dir      string // 文件目录
	_filePath string
	_name     string
	_expire   int // 过期时间
)

var once = sync.Once{}

var ErrorHandler func(ctime time.Time, hostname, line, msg string, label map[string]string)
var InfoHandler func(ctime time.Time, hostname, line, msg string, label map[string]string)
var WarnHandler func(ctime time.Time, hostname, line, msg string, label map[string]string)

// 文件名

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

// size: mb
func InitLogger(path string, size int64, everyday bool, ct ...int) {
	if path == "" {
		_filePath = "."
		return
	}
	_name = filepath.Base(path)

	_dir = filepath.Dir(path)
	_filePath = filepath.Clean(path)
	err := os.MkdirAll(_dir, 0755)
	if err != nil {
		panic(err)
	}
	_fileSize = size
	_everyDay = everyday
	if len(ct) > 0 {
		_expire = ct[0]
	}
	var ctx context.Context
	once.Do(func() {
		if _filePath != "." && _expire > 0 {
			ctx, cancel = context.WithCancel(context.Background())
			go clean(ctx, _filePath, time.Duration(_expire))
		}
	})

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
		s(TRACE, fmt.Sprint(msg...)+"\n")
	}
}

// open file，  所有日志默认前面加了时间，
func Debug(msg ...interface{}) {
	// debug,
	if Level <= DEBUG {
		s(DEBUG, fmt.Sprint(msg...)+"\n")
	}
}

// open file，  所有日志默认前面加了时间，
func Info(msg ...interface{}) {
	if Level <= INFO {
		s(INFO, fmt.Sprint(msg...)+"\n")
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

	// atomic.StoreInt64(&lastTime, time.Now().Unix())
	// 写入缓存, 增加一个4096 是放置打日志导致丢失
	// ml := GetPool()
	ml := msgLog{}
	ml.name = _name
	ml.out = _dir == "." || _dir == ""
	ml.dir = _dir
	ml.size = _fileSize
	ml.filepath = _filePath
	ml.everyDay = _everyDay
	ml.Hostname = hostname
	ml.format = Format
	ml.Level = level
	ml.Msg = msg
	ml.Ctime = time.Now()
	ml.Label = GetLabel()
	if ShowBasePath {
		ml.Line = printBaseFileline(0)
	} else {
		ml.Line = printFileline(0)
	}

	if level == ERROR && ErrorHandler != nil {
		go ErrorHandler(ml.Ctime, ml.Hostname, ml.Line, ml.Msg, ml.Label)
	}
	if level == INFO && InfoHandler != nil {
		go InfoHandler(ml.Ctime, ml.Hostname, ml.Line, ml.Msg, ml.Label)
	}
	if level == WARN && WarnHandler != nil {
		go WarnHandler(ml.Ctime, ml.Hostname, ml.Line, ml.Msg, ml.Label)
	}
	go func() {
		ml.Color = GetColor(level)
		logMsg, _ := ml.formatText()
		ml.Msg = logMsg.String()
		// ml.printLine()
		// fmt.Print(ml.Msg)
		// fmt.Println(111)
		// ml.control()
		t.cache <- ml
		// ml = nil
		// ml.reset()
		// PutPool(ml)
	}()

	// if ml.BufCache.Len() > 1<<20 {
	// 	fmt.Println("write bytes")
	// 	ml.BufCache.Write(logMsg.Bytes())
	// 	ml.Msg = ml.BufCache.String()
	// 	ml.BufCache.Reset()
	// 	cache <- ml
	// } else {
	// 	ml.BufCache.Write(logMsg.Bytes())

	// }

}

// 保留上次写入chan 的时间
// var lastTime int64
