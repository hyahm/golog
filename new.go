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
	Create       time.Time
	Label        map[string]string
	Deep         int
	Color        []color.Attribute
	Mu           *sync.RWMutex
	Line         string
	Out          bool
	Dir          string
	Size         int64
	EveryDay     bool
	Name         string
	Expire       int
	Format       string
	cancel       context.CancelFunc
	level        level
	task         *task
	_logPriority bool
	ErrorHandler func(ctime time.Time, hostname, line, msg string, label map[string]string)
	InfoHandler  func(ctime time.Time, hostname, line, msg string, label map[string]string)
	WarnHandler  func(ctime time.Time, hostname, line, msg string, label map[string]string)
}

// 递归遍历文件夹
func walkDir(dir, name string, expire time.Duration) error {

	return filepath.Walk(dir, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 如果是文件，打印文件路径和修改时间
		if !info.IsDir() && strings.Contains(info.Name(), name) {

			modTime := info.ModTime()
			if time.Since(modTime) > expire {
				os.Remove(fp)
			}
		}
		return nil
	})
}

func clean(ctx context.Context, dir, name string, expire time.Duration) {
	for {
		select {

		case <-time.After(expire):
			walkDir(dir, name, expire)

			// fs, err := os.ReadDir(l.Dir)
			// if err != nil {
			// 	continue
			// }
			// for _, f := range fs {
			// 	if strings.Contains(f.Name(), l.Name) {
			// 		os.Remove(filepath.Join(logPath, f.Name()))
			// 	}
			// }
		case <-ctx.Done():
			return
		}

	}
}

// 默认false  也就是性能优先
func (l *Log) SetLogPriority(logPriority bool) {
	l._logPriority = logPriority
}

// name : filename, size: mb,
func NewLog(name string, size int64, everyday bool, ct ...int) *Log {
	var expire int
	name = filepath.Base(name)
	if len(ct) > 0 {
		expire = ct[0]
	}
	l := &Log{
		Label:    make(map[string]string),
		Mu:       &sync.RWMutex{},
		Dir:      _dir,
		Size:     size,
		EveryDay: everyday,
		Name:     name,
		Expire:   expire,
		level:    INFO,
		task: &task{
			cache: make(chan msgLog, 1000),
			exit:  make(chan struct{}),
			wg:    &sync.WaitGroup{},
		},
	}
	go l.task.write()

	once.Do(func() {
		var ctx context.Context
		if _dir != "." && name != "" && expire > 0 && (size > 0 || everyday) {
			ctx, cancel = context.WithCancel(context.Background())
			go clean(ctx, _dir, l.Name, time.Duration(expire)*defaultUnit)
		}
	})
	return l
}

func (l *Log) SetErrorHandler(eh func(time.Time, string, string, string, map[string]string)) {
	l.ErrorHandler = eh
}

func (l *Log) SetWarnHandler(eh func(time.Time, string, string, string, map[string]string)) {
	l.WarnHandler = eh
}
func (l *Log) SetInfoHandler(eh func(time.Time, string, string, string, map[string]string)) {
	l.InfoHandler = eh
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

func (l *Log) SetLabel(key, value string) *Log {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.Label[key] = value
	return l
}

func (l *Log) DelLabel(key string) *Log {
	l.Mu.RLock()
	defer l.Mu.RUnlock()
	delete(l.Label, key)
	return l
}

func (l *Log) GetLabel() map[string]string {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	return l.Label
}

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
		l.s(DEBUG, arrToString(msg...)+"\n")
	}
}

func (l *Log) SetLevel(lv LogLevel) {
	// Access,
	l.level = level(lv)
}

func (l *Log) Level() LogLevel {
	// Access,
	return LogLevel(l.level)
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
		l.s(INFO, arrToString(msg...)+"\n")
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
		l.s(WARN, arrToString(msg...)+"\n")
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
		l.s(ERROR, arrToString(msg...)+"\n")
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
		l.s(FATAL, arrToString(msg...)+"\n")
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
		l.s(DEBUG, arrToString(msg...)+"\n", deep)
	}
}

func (l *Log) s(level level, msg string, deep ...int) {
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
	ml.Hostname = hostname
	ml.name = l.Name
	ml.size = l.Size
	ml.format = Format
	ml.everyDay = l.EveryDay
	ml.Label = l.GetLabel()

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
	l.task.wg.Go(func() {
		if ml.out {
			// 控制台才添加颜色， 否则不添加颜色
			ml.Color = GetColor(ml.Level)
		}

		logMsg, _ := ml.formatText()
		ml.Msg = logMsg.String()
		if l._logPriority {
			l.task.cache <- ml
		} else {
			select {
			case l.task.cache <- ml:
			default:
			}
		}

		// ml.control()

	})
}
