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
	Create   time.Time
	Label    map[string]string
	Deep     int
	Color    []color.Attribute
	Mu       *sync.RWMutex
	Line     string
	Out      bool
	Path     string
	Dir      string
	Size     int64
	EveryDay bool
	Name     string
	Expire   int
	Format   string
	cancel   context.CancelFunc
}

func (l *Log) clean(ctx context.Context) {
	for {
		select {
		case <-time.After(time.Duration(l.Expire) * time.Hour * 24):
			fs, err := os.ReadDir(l.Dir)
			if err != nil {
				continue
			}
			for _, f := range fs {
				if strings.Contains(f.Name(), l.Name) {
					os.Remove(filepath.Join(logPath, f.Name()))
				}
			}
		case <-ctx.Done():
			return
		}

	}
}

// size: kb
func NewLog(path string, size int64, everyday bool, ct ...int) *Log {
	var expire int
	path = filepath.Clean(path)
	if len(ct) > 0 {
		expire = ct[0]
	}
	l := &Log{
		Label:    make(map[string]string),
		Mu:       &sync.RWMutex{},
		Path:     path,
		Size:     size,
		EveryDay: everyday,
		Expire:   expire,
	}
	l.Dir = filepath.Dir(path)
	err := os.MkdirAll(l.Dir, 0755)
	if err != nil {
		panic(err)
	}
	l.Name = filepath.Base(path)
	var ctx context.Context

	if l.Name != "." && l.Expire > 0 {
		os.OpenFile("cccc", os.O_CREATE, 0744)
		ctx, l.cancel = context.WithCancel(context.Background())
		go l.clean(ctx)
	}
	return l
}

// 关闭log
func (l *Log) Close() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Not need be close")
		}
	}()
	l.cancel()
	l = nil
}

func (l *Log) AddLabel(key, value string) *Log {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.Label[key] = value
	return l
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
	if Level <= TRACE {
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
	if Level <= DEBUG {
		l.s(DEBUG, arrToString(msg...)+"\n")
	}
}

// open file，  所有日志默认前面加了时间，
func (l *Log) Debugf(format string, msg ...interface{}) {
	// Access,
	if Level <= DEBUG {
		l.s(DEBUG, arrToString(msg...))
	}
}

// open file，  所有日志默认前面加了时间，
func (l *Log) Info(msg ...interface{}) {
	if Level <= INFO {
		l.s(INFO, arrToString(msg...)+"\n")
	}
}
func (l *Log) Infof(format string, msg ...interface{}) {
	// Access,
	if Level <= INFO {
		l.s(INFO, arrToString(msg...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func (l *Log) Warn(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= WARN {
		l.s(WARN, arrToString(msg...)+"\n")
	}
}

func (l *Log) Warnf(format string, msg ...interface{}) {
	// Access,
	if Level <= WARN {
		l.s(WARN, arrToString(msg...))
	}
}

// 可以根据下面格式一样，在format 后加上更详细的输出值
func (l *Log) Error(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= ERROR {
		l.s(ERROR, arrToString(msg...)+"\n")
	}
}

func (l *Log) Errorf(format string, msg ...interface{}) {
	// Access,
	if Level <= ERROR {
		l.s(ERROR, arrToString(msg...))
	}
}

func (l *Log) Fatal(msg ...interface{}) {
	// error日志，添加了错误函数，
	if Level <= FATAL {
		l.s(FATAL, arrToString(msg...)+"\n")
	}
	Sync()
	os.Exit(1)
}

func (l *Log) Fatalf(format string, msg ...interface{}) {
	// Access,
	if Level <= FATAL {
		l.s(FATAL, arrToString(msg...))
	}
}

func (l *Log) UpFunc(deep int, msg ...interface{}) {
	// deep打印函数的深度， 相对于当前位置向外的深度
	if Level <= DEBUG {
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
	// pre := ""
	// for k, v := range l.Label {
	// 	pre += fmt.Sprintf("[%s = %s]", k, v)
	// }
	if l.Format == "" {
		l.Format = Format
	}
	now := time.Now()
	ml := msgLog{
		// Prev:    pre,
		Msg:      msg,
		Level:    level,
		create:   now,
		Ctime:    now.Format("2006-01-02 15:04:05"),
		Color:    GetColor(level),
		Line:     printFileline(0),
		out:      l.Name == "." || l.Name == "",
		path:     l.Dir,
		logPath:  l.Path,
		Hostname: hostname,
		name:     l.Name,
		size:     l.Size,
		format:   l.Format,
		Label:    l.GetLabel(),
	}
	if ShowBasePath {
		ml.Line = printBaseFileline(0)
	}
	cache <- ml
}
