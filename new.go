package golog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var ShowBasePath bool

type Log struct {
	Create time.Time
	// Label             map[string]string
	Deep        int
	Color       []color.Attribute
	Mu          *sync.RWMutex
	Line        string
	Out         bool
	Dir         string
	Size        int64
	EveryDay    bool
	Name        string
	Expire      int
	Format      func(ctime time.Time, hostname, line, msg string, label map[string]string) string
	cancel      context.CancelFunc
	level       Level
	task        *task
	logPriority bool
	duplicates  duplicate
	LogHandler  func(level Level, ctime time.Time, line, msg string, label map[string]string)
}

// 递归遍历文件夹
func walkDir() error {
	names := getNames()
	name := make([]string, 0, len(names))
	for k := range names {
		name = append(name, k)
	}
	return filepath.Walk(_dir, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 如果是文件，打印文件路径和修改时间
		if !info.IsDir() && containsSlice(info.Name(), name) {
			modTime := info.ModTime()
			if time.Since(modTime) > _expireClean {
				os.Remove(fp)
			}
		}
		return nil
	})
}

func containsSlice(str string, ss []string) bool {
	for _, v := range ss {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}

// 默认false  也就是性能优先
func (l *Log) SetLogPriority(logPriority bool, duplicates int, dd ...time.Duration) {
	l.logPriority = logPriority
	if l.logPriority && duplicates > 0 {
		cleanDuplicate := time.Minute
		if len(dd) > 0 {
			cleanDuplicate = dd[0]
		}
		l.duplicates.initDuplicate(duplicates, cleanDuplicate)
		return
	}
	l.logPriority = false
}

// name : filename, size: mb,
func NewLog(name string, size int64, everyday bool) *Log {
	if name != "" {
		fi, err := os.Stat(_dir)
		if err == nil && !fi.IsDir() {
			// 如果存在这个文件， 直接跳过
			fmt.Printf("%s is not a directory, will input log to the console \n", _dir)
			_name = ""
		}
		if err != nil {
			// 目录不存在就创建
			if err = os.MkdirAll(_dir, 0755); err != nil {
				fmt.Println(err)
				name = ""
			}

		}

	}
	name = filepath.Base(name)
	l := &Log{
		// Label:    make(map[string]string),
		Mu:       &sync.RWMutex{},
		Size:     size,
		Dir:      _dir,
		EveryDay: everyday,
		Name:     name,
		level:    _level,
		task: &task{
			cache: make(chan msgLog, 1000),
			exit:  make(chan struct{}),
			wg:    &sync.WaitGroup{},
		},
	}
	go l.task.write()
	addClean(_name)
	return l
}

func (l *Log) SetLogHandler(eh func(Level, time.Time, string, string, map[string]string)) {
	l.LogHandler = eh
}

// 关闭log
func (l *Log) Close() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	l.cancel()
	l = nil
}

// func (l *Log) SetLabel(key, value string) *Log {
// 	l.Mu.Lock()
// 	defer l.Mu.Unlock()
// 	l.Label[key] = value
// 	return l
// }

// func (l *Log) DelLabel(key string) *Log {
// 	l.Mu.RLock()
// 	defer l.Mu.RUnlock()
// 	delete(l.Label, key)
// 	return l
// }

// func (l *Log) GetLabel() map[string]string {
// 	l.Mu.Lock()
// 	defer l.Mu.Unlock()
// 	return l.Label
// }

// open file，  所有日志默认前面加了时间，
func (l *Log) Trace(msg ...interface{}) {
	// Access,
	if l.level <= TRACE {
		l.s(TRACE, arrToString(msg...))
	}
}

// open file，  所有日志默认前面加了时间，
func (l *Log) Tracef(format string, msg ...interface{}) {
	// Access,
	l.Trace(fmt.Sprintf(format, msg...))
}

// open file，  所有日志默认前面加了时间，
func (l *Log) Debug(msg ...interface{}) {
	// debug,
	if l.level <= DEBUG {
		l.s(DEBUG, arrToString(msg...))
	}
}

func (l *Log) SetLevel(lv Level) {
	// Access,
	l.level = lv
}

func (l *Log) Level() Level {
	// Access,
	return l.level
}

// open file，  所有日志默认前面加了时间，
func (l *Log) Debugf(format string, msg ...interface{}) {
	// Access,
	if l.level <= DEBUG {
		l.s(DEBUG, arrToString(msg...))
	}
}

// open file，  所有日志默认前面加了时间，
func (l *Log) Info(msg ...interface{}) {
	if l.level <= INFO {
		l.s(INFO, arrToString(msg...))
	}
}
func (l *Log) Infof(format string, msg ...interface{}) {
	// Access,
	if l.level <= INFO {
		l.s(INFO, arrToString(msg...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func (l *Log) Warn(msg ...interface{}) {
	// error日志，添加了错误函数，
	if l.level <= WARN {
		l.s(WARN, arrToString(msg...))
	}
}

func (l *Log) Warnf(format string, msg ...interface{}) {
	// Access,
	if l.level <= WARN {
		l.s(WARN, arrToString(msg...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func (l *Log) Error(msg ...interface{}) {
	// error日志，添加了错误函数，
	if l.level <= ERROR {
		l.s(ERROR, arrToString(msg...))
	}
}

func (l *Log) Errorf(format string, msg ...interface{}) {
	// Access,
	if l.level <= ERROR {
		l.s(ERROR, arrToString(msg...))
	}
}

func (l *Log) Fatal(msg ...interface{}) {
	// error日志，添加了错误函数，
	if l.level <= FATAL {
		l.s(FATAL, arrToString(msg...))
	}
	os.Exit(1)
}

func (l *Log) Fatalf(format string, msg ...interface{}) {
	// Access,
	if l.level <= FATAL {
		l.s(FATAL, arrToString(msg...))
	}
}

func (l *Log) UpFunc(deep int, msg ...interface{}) {
	// deep打印函数的深度， 相对于当前位置向外的深度
	if l.level <= DEBUG {
		l.s(DEBUG, arrToString(msg...), deep)
	}
}

func (l *Log) s(level Level, msg string, deep ...int) {
	if len(deep) > 0 && deep[0] > 0 {
		if ShowBasePath {
			msg = fmt.Sprintf("caller from %s -- %v", printBaseFileline(deep[0]), msg)
		} else {
			msg = fmt.Sprintf("caller from %s -- %v", printFileline(deep[0]), msg)
		}

	}

	ml := msgLog{}
	ml.Msg = msg
	ml.Level = level
	ml.out = l.Name == "." || l.Name == ""
	ml.dir = l.Dir
	ml.Ctime = time.Now()
	ml.name = l.Name
	ml.size = l.Size
	if _formatFunc == nil {
		ml.format = defaultFormat
	} else {
		ml.format = _formatFunc
	}
	ml.everyDay = l.EveryDay
	// ml.Label = l.GetLabel()

	if ShowBasePath {
		ml.Line = printBaseFileline(0)
	} else {
		ml.Line = printFileline(0)
	}

	if l.logPriority {
		key := ml.Line + ml.Msg
		if !l.duplicates.addMsg(key) {
			return
		}
	}

	if LogHandler != nil {
		go LogHandler(ml.Level, ml.Ctime, ml.Line, ml.Msg)
	}

	if l.logPriority {
		l.task.cache <- ml
	} else {
		select {
		case l.task.cache <- ml:
		default:
		}
	}

	// ml.control()

}
