package golog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ShowBasePath 控制日志调用位置只显示文件名而非完整路径。
// 应在首次打印日志前设置，运行中修改存在数据竞争风险。
var ShowBasePath bool

// Log 独立日志实例，通过 NewLog 创建，与全局方法并存，
// 各实例拥有独立的异步任务、文件与配置。
type Log struct {
	mu       sync.RWMutex // 保护下方可运行时修改的配置
	name     string       // 日志文件名，空表示输出控制台
	dir      string
	size     int64 // 按大小切割阈值(MB)，0 不切割
	everyDay bool  // 按天切割
	console  bool  // 写文件的同时输出控制台
	compress bool  // 归档文件 gzip 压缩

	level      atomic.Int32 // 当前日志级别
	fields     []Field      // With 附加的结构化字段
	loggerName string       // Named 设置的模块名
	ctx        context.Context
	extraSkip  int // 全局包装函数额外增加的调用栈深度

	task        *task
	logPriority bool
	duplicates  *duplicate
	sampler     *sampler
}

var (
	defaultLogOnce sync.Once
	defaultLogIns  *Log
)

// getDefaultLog 返回全局默认 logger（懒初始化，避免 init 顺序依赖）。
func getDefaultLog() *Log {
	defaultLogOnce.Do(func() {
		defaultLogIns = &Log{
			mu:        sync.RWMutex{},
			dir:       _dir,
			task:      t,
			extraSkip: 1, // 全局函数比方法多一层调用栈
		}
		defaultLogIns.level.Store(int32(INFO))
	})
	return defaultLogIns
}

// NewLog 创建独立日志实例。
// name: 日志文件名（空则输出控制台）；size: 按大小切割阈值(MB)，0 不切割；
// everyday: 是否按天切割（size > 0 时按大小切割优先）。
func NewLog(name string, size int64, everyday bool) *Log {
	if name != "" {
		fi, err := os.Stat(_dir)
		if err == nil && !fi.IsDir() {
			// 存在同名文件，回退控制台输出
			log.Printf("%s is not a directory, will input log to the console \n", _dir)
			name = ""
		} else if err != nil {
			if err = os.MkdirAll(_dir, 0755); err != nil {
				log.Println(err)
				name = ""
			}
		}
	}
	l := &Log{
		mu:       sync.RWMutex{},
		size:     size,
		dir:      _dir,
		everyDay: everyday,
		name:     filepath.Base(name),
		task:     newTask(),
	}
	l.level.Store(int32(getDefaultLog().Level()))
	go l.task.write()
	addClean(l.name)
	return l
}

// clone 复制当前配置生成子 logger（供 With/Named/Ctx 使用），共享同一输出任务。
func (l *Log) clone() *Log {
	l.mu.RLock()
	defer l.mu.RUnlock()
	nl := &Log{
		mu:          sync.RWMutex{},
		name:        l.name,
		dir:         l.dir,
		size:        l.size,
		everyDay:    l.everyDay,
		console:     l.console,
		compress:    l.compress,
		fields:      append([]Field(nil), l.fields...),
		loggerName:  l.loggerName,
		ctx:         l.ctx,
		extraSkip:   l.extraSkip,
		task:        l.task,
		logPriority: l.logPriority,
		duplicates:  l.duplicates,
		sampler:     l.sampler,
	}
	nl.level.Store(l.level.Load())
	return nl
}

// With 返回携带新增结构化字段的子 logger，字段随后续所有日志输出。
func (l *Log) With(fields ...Field) *Log {
	nl := l.clone()
	if len(fields) > 0 {
		nl.mu.Lock()
		nl.fields = append(nl.fields, fields...)
		nl.mu.Unlock()
	}
	return nl
}

// Named 返回带模块名的子 logger，输出附加 logger=名称 字段。
func (l *Log) Named(name string) *Log {
	nl := l.clone()
	nl.mu.Lock()
	nl.loggerName = name
	nl.mu.Unlock()
	return nl
}

// Ctx 返回绑定 context 的子 logger，自动携带注入 context 的字段（如 trace_id）。
func (l *Log) Ctx(ctx context.Context) *Log {
	nl := l.clone()
	nl.mu.Lock()
	nl.ctx = ctx
	nl.mu.Unlock()
	return nl
}

// SetLevel 设置日志级别，可在运行时并发安全调用。
func (l *Log) SetLevel(lv Level) {
	l.level.Store(int32(lv))
}

// Level 返回当前日志级别。
func (l *Log) Level() Level {
	return Level(l.level.Load())
}

// enabled 判断给定级别是否允许输出。
func (l *Log) enabled(lv Level) bool {
	return Level(l.level.Load()) <= lv
}

// SetConsole 设置写文件时是否同时输出到控制台，默认 false。
func (l *Log) SetConsole(b bool) {
	l.mu.Lock()
	l.console = b
	l.mu.Unlock()
}

// SetCompress 设置归档的旧日志文件是否异步 gzip 压缩，默认 false。
func (l *Log) SetCompress(b bool) {
	l.mu.Lock()
	l.compress = b
	l.mu.Unlock()
}

// SetOutput 设置控制台输出目标（默认 os.Stdout），
// 用于测试捕获日志或对接 syslog/kafka 等自定义 Writer。
func (l *Log) SetOutput(w io.Writer) {
	l.task.cw.SetOutput(w)
}

// SetRateLimit 开启限流采样：每个 window 窗口内最多记录 max 条日志，超出部分丢弃。
// max <= 0 时关闭限流。可与 SetLogPriority 叠加使用。
func (l *Log) SetRateLimit(max int, window time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if max <= 0 {
		l.sampler = nil
		return
	}
	l.sampler = newSampler(max, window)
}

// SetLogPriority 设置重复日志采样（借鉴 zap 生产模式）：
// logPriority 为 true 且 duplicates > 0 时，窗口期内相同日志只记录第一条，
// 重复达到 duplicates 条后再放行一条用于统计。logPriority 为 false 时记录全部日志。
func (l *Log) SetLogPriority(logPriority bool, duplicates int, dd ...time.Duration) {
	cleanDuplicate := time.Minute
	if len(dd) > 0 {
		cleanDuplicate = dd[0]
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if logPriority && duplicates > 0 {
		if l.duplicates != nil {
			l.duplicates.close()
		}
		l.duplicates = newDuplicate(duplicates, cleanDuplicate)
		l.logPriority = true
		return
	}
	l.logPriority = false
}

// SetDir 设置当前实例的日志输出目录并创建目录。
func (l *Log) SetDir(dir string) {
	dir = filepath.Clean(dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	l.mu.Lock()
	l.dir = dir
	l.mu.Unlock()
}

// initFile 由全局 InitLogger 调用，设置全局实例的文件输出参数。
func (l *Log) initFile(name string, size int64, everyday bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.name, l.size, l.everyDay = name, size, everyday
}

// setDir 由全局 SetDir 调用，同步全局实例目录。
func (l *Log) setDir(dir string) {
	l.mu.Lock()
	l.dir = dir
	l.mu.Unlock()
}

// prepare 构造单条日志的公共部分，返回载体、重复采样器与限流采样器。
func (l *Log) prepare(level Level, msg, line string, fields []Field) (msgLog, *duplicate, *sampler) {
	l.mu.RLock()
	all := make([]Field, 0, len(l.fields)+len(fields)+2)
	if l.loggerName != "" {
		all = append(all, Str("logger", l.loggerName))
	}
	all = append(all, l.fields...)
	ml := msgLog{
		name:     l.name,
		dir:      l.dir,
		size:     l.size,
		everyDay: l.everyDay,
		console:  l.console,
		compress: l.compress,
	}
	dup, smp, ctxv := l.duplicates, l.sampler, l.ctx
	l.mu.RUnlock()

	if ctxv != nil {
		all = append(all, FieldsFromContext(ctxv)...)
	}
	all = append(all, fields...)

	ml.out = ml.name == "" || ml.name == "."
	ml.format = loadFormatFunc()
	ml.Level = level
	ml.Msg = msg
	ml.Ctime = time.Now()
	ml.fields = all
	ml.Line = line
	return ml, dup, smp
}

// dispatch 完成采样过滤、回调分发并投递到异步通道。
func (l *Log) dispatch(level Level, msg, line string, fields []Field) {
	ml, dup, smp := l.prepare(level, msg, line, fields)
	if dup != nil && !dup.addMsg(line+msg) {
		return
	}
	if smp != nil && !smp.allow() {
		return
	}
	if h := LogHandler; h != nil {
		l.task.handlerWg.Go(func() {
			h(ml.Level, ml.Ctime, ml.Line, ml.Msg)
		})
	}
	l.task.send(ml)
}

// logm 记录日志并自动计算调用位置，deep 为向外追加的栈深度，
// deep > 0 时额外标注调用者的调用位置（caller from）。
func (l *Log) logm(level Level, msg string, deep int, fields []Field) {
	skip := l.extraSkip
	if deep > 0 {
		if ShowBasePath {
			msg = fmt.Sprintf("caller from %s -- %v", printBaseFileline(skip+deep), msg)
		} else {
			msg = fmt.Sprintf("caller from %s -- %v", printFileline(skip+deep), msg)
		}
	}
	var line string
	if ShowBasePath {
		line = printBaseFileline(skip)
	} else {
		line = printFileline(skip)
	}
	l.dispatch(level, msg, line, fields)
}

// logAt 使用显式调用位置记录日志（供 slog 适配器等外部调用使用）。
func (l *Log) logAt(level Level, msg, line string, fields []Field) {
	l.dispatch(level, msg, line, fields)
}

// logSync 绕过异步通道直接写盘，用于 Fatal/Panic 等必须立即落盘的场景。
func (l *Log) logSync(level Level, msg string, fields []Field) {
	var line string
	if ShowBasePath {
		line = printBaseFileline(l.extraSkip)
	} else {
		line = printFileline(l.extraSkip)
	}
	ml, _, _ := l.prepare(level, msg, line, fields)
	ml.Msg = ml.format(ml.Level, ml.Ctime, ml.Line, ml.Msg, ml.fields)
	ml.Color = GetColor(level)
	l.task.syncWrite(ml)
}

// Sync 等待当前实例的异步日志全部落盘并停止内部协程，可安全重复调用。
// 实例停止服务时必须调用。
func (l *Log) Sync() {
	l.mu.RLock()
	dup := l.duplicates
	l.mu.RUnlock()
	if dup != nil {
		dup.close()
	}
	l.task.syncClose()
}

// Trace 打印 TRACE 级别日志。
func (l *Log) Trace(msg ...any) {
	if l.enabled(TRACE) {
		l.logm(TRACE, arrToString(msg...), 0, nil)
	}
}

// Tracef 打印 TRACE 级别格式化日志。
func (l *Log) Tracef(format string, args ...any) {
	if l.enabled(TRACE) {
		l.logm(TRACE, fmt.Sprintf(format, args...), 0, nil)
	}
}

// Tracew 打印 TRACE 级别日志并附加结构化字段。
func (l *Log) Tracew(msg string, fields ...Field) {
	if l.enabled(TRACE) {
		l.logm(TRACE, msg, 0, fields)
	}
}

// Debug 打印 DEBUG 级别日志。
func (l *Log) Debug(msg ...any) {
	if l.enabled(DEBUG) {
		l.logm(DEBUG, arrToString(msg...), 0, nil)
	}
}

// Debugf 打印 DEBUG 级别格式化日志。
func (l *Log) Debugf(format string, args ...any) {
	if l.enabled(DEBUG) {
		l.logm(DEBUG, fmt.Sprintf(format, args...), 0, nil)
	}
}

// Debugw 打印 DEBUG 级别日志并附加结构化字段。
func (l *Log) Debugw(msg string, fields ...Field) {
	if l.enabled(DEBUG) {
		l.logm(DEBUG, msg, 0, fields)
	}
}

// Info 打印 INFO 级别日志。
func (l *Log) Info(msg ...any) {
	if l.enabled(INFO) {
		l.logm(INFO, arrToString(msg...), 0, nil)
	}
}

// Infof 打印 INFO 级别格式化日志。
func (l *Log) Infof(format string, args ...any) {
	if l.enabled(INFO) {
		l.logm(INFO, fmt.Sprintf(format, args...), 0, nil)
	}
}

// Infow 打印 INFO 级别日志并附加结构化字段。
func (l *Log) Infow(msg string, fields ...Field) {
	if l.enabled(INFO) {
		l.logm(INFO, msg, 0, fields)
	}
}

// Warn 打印 WARN 级别日志。
func (l *Log) Warn(msg ...any) {
	if l.enabled(WARN) {
		l.logm(WARN, arrToString(msg...), 0, nil)
	}
}

// Warnf 打印 WARN 级别格式化日志。
func (l *Log) Warnf(format string, args ...any) {
	if l.enabled(WARN) {
		l.logm(WARN, fmt.Sprintf(format, args...), 0, nil)
	}
}

// Warnw 打印 WARN 级别日志并附加结构化字段。
func (l *Log) Warnw(msg string, fields ...Field) {
	if l.enabled(WARN) {
		l.logm(WARN, msg, 0, fields)
	}
}

// Error 打印 ERROR 级别日志。
func (l *Log) Error(msg ...any) {
	if l.enabled(ERROR) {
		l.logm(ERROR, arrToString(msg...), 0, nil)
	}
}

// Errorf 打印 ERROR 级别格式化日志。
func (l *Log) Errorf(format string, args ...any) {
	if l.enabled(ERROR) {
		l.logm(ERROR, fmt.Sprintf(format, args...), 0, nil)
	}
}

// Errorw 打印 ERROR 级别日志并附加结构化字段。
func (l *Log) Errorw(msg string, fields ...Field) {
	if l.enabled(ERROR) {
		l.logm(ERROR, msg, 0, fields)
	}
}

// Fatal 同步写入 FATAL 级别日志（保证落盘）并以退出码 1 结束进程。
func (l *Log) Fatal(msg ...any) {
	if l.enabled(FATAL) {
		l.logSync(FATAL, arrToString(msg...), nil)
	}
	exitFunc(1)
}

// Fatalf 同步写入 FATAL 级别格式化日志并结束进程。
func (l *Log) Fatalf(format string, args ...any) {
	if l.enabled(FATAL) {
		l.logSync(FATAL, fmt.Sprintf(format, args...), nil)
	}
	exitFunc(1)
}

// Fatalw 同步写入 FATAL 级别日志（含结构化字段）并结束进程。
func (l *Log) Fatalw(msg string, fields ...Field) {
	if l.enabled(FATAL) {
		l.logSync(FATAL, msg, fields)
	}
	exitFunc(1)
}

// Panic 同步写入 PANIC 级别日志（保证落盘）后 panic。
func (l *Log) Panic(msg ...any) {
	if l.enabled(PANIC) {
		l.logSync(PANIC, arrToString(msg...), nil)
	}
	panic(arrToString(msg...))
}

// Panicf 同步写入 PANIC 级别格式化日志后 panic。
func (l *Log) Panicf(format string, args ...any) {
	if l.enabled(PANIC) {
		l.logSync(PANIC, fmt.Sprintf(format, args...), nil)
	}
	panic(fmt.Sprintf(format, args...))
}

// Stack 打印 ERROR 级别日志并附带当前 goroutine 的调用堆栈。
func (l *Log) Stack(msg ...any) {
	if l.enabled(ERROR) {
		buf := make([]byte, 4096)
		for {
			n := runtime.Stack(buf, false)
			if n < len(buf) {
				buf = buf[:n]
				break
			}
			buf = make([]byte, 2*len(buf))
		}
		l.logm(ERROR, arrToString(msg...)+"\n"+string(buf), 0, nil)
	}
}

// UpFunc 打印 DEBUG 级别日志并标注调用者的调用位置，deep 为向外追加的栈深度。
func (l *Log) UpFunc(deep int, msg ...any) {
	if l.enabled(DEBUG) {
		l.logm(DEBUG, arrToString(msg...), deep, nil)
	}
}

// UpFuncf 打印 DEBUG 级别格式化日志并标注调用者的调用位置。
func (l *Log) UpFuncf(deep int, format string, args ...any) {
	if l.enabled(DEBUG) {
		l.logm(DEBUG, fmt.Sprintf(format, args...), deep, nil)
	}
}
