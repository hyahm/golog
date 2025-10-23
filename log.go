package golog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// _logPath   string // 文件路径
	_fileSize int64          // 切割的文件大小默认单位M
	_everyDay bool           // 每天一个来切割文件 （这个比上面个优先级高）
	_dir      string = "log" // 文件目录
	_name     string
	// label             = make(map[string]string)
	// labelLock         = sync.RWMutex{}
	_logPriority      bool
	_duplicates       int
	_duplicateskey    map[string]int
	_duplicatesLocker sync.Mutex
)

var once = sync.Once{}

var LogHandler func(level Level, ctime time.Time, line, msg string)

// var Format string = "{{ .Ctime }} - [{{ .Level }}]{{ if .Label }} - {{ range $k,$v := .Label}}[{{$k}}:{{$v}}]{{end}}{{end}} - {{.Hostname}} - {{.Line}} - {{.Msg}}"

// var cancel context.CancelFunc

func SetDir(dir string) {
	_dir = filepath.Clean(dir)
	err := os.MkdirAll(_dir, 0755)
	if err != nil {
		fmt.Println(err)
		_dir = "."
	}
}

// 默认false  也就是日志优先,  类似 zap 开发模式， 打印所有日志， 设置true 的话， 类似 zap 生成模式，
// 后面的duplicates 是重复多少条值打印一条,   如果小于等于0 相当于logPriority 为false
func SetLogPriority(logPriority bool, duplicates int) {
	if duplicates > 0 {
		_logPriority = logPriority
		_duplicates = duplicates
		_duplicateskey = make(map[string]int)
		_duplicatesLocker = sync.Mutex{}
	}

}

// name : filename, size: mb,
func InitLogger(name string, size int64, everyday bool) {

	_name = filepath.Base(name)
	_fileSize = size
	_everyDay = everyday
	addClean(_name)
	// once.Do(func() {
	// 	var ctx context.Context
	// 	if _dir != "." && name != "" && _expire > 0 && (size > 0 || everyday) {
	// 		ctx, cancel = context.WithCancel(context.Background())
	// 		go clean(ctx, _dir, _name, time.Duration(_expire)*defaultUnit)
	// 	}
	// })

}

// 清理日志， 请在写入文件初始化后调用即可， 已经是异步处理
// func Clean(names ...string) {
// 	if len(names) == 0 {
// 		return
// 	}
// 	once.Do(func() {
// 		go clean(_dir, _expireClean, names...)
// 	})
// }

// func AddLabel(key, value string) {
// 	labelLock.RLock()
// 	defer labelLock.RUnlock()
// 	label[key] = value
// }

// func SetLabel(key, value string) {
// 	labelLock.RLock()
// 	defer labelLock.RUnlock()
// 	label[key] = value
// }

// func DelLabel(key string) {
// 	labelLock.Lock()
// 	defer labelLock.Unlock()
// 	delete(label, key)
// }

// func GetLabel() map[string]string {
// 	labelLock.RLock()
// 	defer labelLock.RUnlock()
// 	return label
// }

// open file，  所有日志默认前面加了时间，
func Tracef(format string, args ...interface{}) {
	if _level <= TRACE {
		s(TRACE, fmt.Sprintf(format, args...))
	}
}

// open file，  所有日志默认前面加了时间，
func Debugf(format string, args ...interface{}) {
	if _level <= DEBUG {
		s(DEBUG, fmt.Sprintf(format, args...))
	}
}

// open file，  所有日志默认前面加了时间，
func Infof(format string, args ...interface{}) {
	if _level <= INFO {
		s(INFO, fmt.Sprintf(format, args...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Warnf(format string, args ...interface{}) {
	// error日志，添加了错误函数，
	if _level <= WARN {
		s(WARN, fmt.Sprintf(format, args...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Errorf(format string, args ...interface{}) {
	// error日志，添加了错误函数，
	if _level <= ERROR {
		s(ERROR, fmt.Sprintf(format, args...))
	}
}

func Fatalf(format string, args ...interface{}) {
	// error日志，添加了错误函数，
	if _level <= FATAL {
		s(FATAL, fmt.Sprintf(format, args...))
	}
}

func UpFuncf(deep int, format string, args ...interface{}) {
	// deep打印函数的深度， 相对于当前位置向外的深度
	if _level <= DEBUG {
		s(DEBUG, fmt.Sprintf(format, args...), deep)
	}
}

// open file，  所有日志默认前面加了时间，
func Trace(msg ...interface{}) {
	// Access,
	if _level <= TRACE {
		s(TRACE, fmt.Sprint(msg...))
	}
}

// open file，  所有日志默认前面加了时间，
func Debug(msg ...interface{}) {
	// debug,
	if _level <= DEBUG {
		s(DEBUG, fmt.Sprint(msg...))
	}
}

// open file，  所有日志默认前面加了时间，
func Info(msg ...interface{}) {
	if _level <= INFO {
		s(INFO, fmt.Sprint(msg...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Warn(msg ...interface{}) {
	// error日志，添加了错误函数，
	if _level <= WARN {
		s(WARN, arrToString(msg...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func Error(msg ...interface{}) {
	// error日志，添加了错误函数，
	if _level <= ERROR {
		s(ERROR, arrToString(msg...))
	}
}

func Fatal(msg ...interface{}) {
	// error日志，添加了错误函数，
	if _level <= FATAL {
		s(FATAL, arrToString(msg...))
	}
	os.Exit(1)
}

func UpFunc(deep int, msg ...interface{}) {
	// deep打印函数的深度， 相对于当前位置向外的深度
	if _level <= DEBUG {
		s(DEBUG, arrToString(msg...), deep)
	}
}

func arrToString(msg ...interface{}) string {
	ll := make([]string, 0, len(msg))
	for _, v := range msg {
		ll = append(ll, fmt.Sprintf("%v", v))
	}
	return strings.Join(ll, " ")
}

func s(level Level, msg string, deep ...int) {

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
	ml.out = _name == "." || _name == ""
	ml.dir = _dir
	ml.size = _fileSize
	ml.everyDay = _everyDay
	if _formatFunc == nil {
		ml.format = defaultFormat
	} else {
		ml.format = _formatFunc
	}

	ml.Level = level
	ml.Msg = msg
	ml.Ctime = time.Now()
	// ml.Label = GetLabel()
	if ShowBasePath {
		ml.Line = printBaseFileline(0)
	} else {
		ml.Line = printFileline(0)
	}
	if _duplicateskey != nil {
		key := ml.Line + ml.Msg
		_duplicatesLocker.Lock()
		if _, ok := _duplicateskey[key]; ok {
			_duplicateskey[key] = _duplicateskey[key] + 1
			if _duplicateskey[key] == _duplicates {
				delete(_duplicateskey, key)
			}
			_duplicatesLocker.Unlock()
			return
		}
		_duplicateskey[key] = 0
		_duplicatesLocker.Unlock()
	}
	if LogHandler != nil {
		go LogHandler(ml.Level, ml.Ctime, ml.Line, ml.Msg)
	}
	// if ml.out {
	// 	// 控制台才添加颜色， 否则不添加颜色
	// 	ml.Color = GetColor(ml.Level)
	// }

	// logMsg, _ := ml.formatText()
	// ml.Msg = logMsg.String()
	// // ml.printLine()
	// // fmt.Print(ml.Msg)

	// // ml.control()

	if _logPriority {
		t.cache <- ml
	} else {
		select {
		case t.cache <- ml:
		default:
		}
	}

	// ml = nil
	// ml.reset()
	// PutPool(ml)

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
