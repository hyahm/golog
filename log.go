// Package golog 是一个异步、简单易用的日志库，全程无需关闭操作，开箱即用。
//
// 常用全局方法：Info/Infof/Infow、Error 等；写文件调用 InitLogger；
// 程序退出前 defer Sync() 可避免日志丢失。
// 需要独立输出目标时使用 NewLog 创建实例，实例停止服务时调用其 Sync。
package golog

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// _dir 全局日志目录。
var _dir = "log"

// LogHandler 日志回调：每条日志写入前异步调用，可用于报警等二次处理。
// 应在首次打印日志前设置，运行中修改存在数据竞争风险。
var LogHandler func(level Level, ctime time.Time, line, msg string)

// exitFunc 退出函数，测试中可替换以避免 os.Exit 终止测试进程。
var exitFunc = os.Exit

// SetDir 设置全局日志目录并创建（影响全局实例与后续 NewLog 创建的实例）。
func SetDir(dir string) {
	_dir = filepath.Clean(dir)
	if err := os.MkdirAll(_dir, 0755); err != nil {
		_dir = "."
	}
	getDefaultLog().setDir(_dir)
}

// InitLogger 初始化全局日志写入文件。
// name: 日志文件名；size: 按大小切割阈值(MB)，0 不按大小切割；
// everyday: 是否按天切割（size > 0 时按大小切割优先）。
func InitLogger(name string, size int64, everyday bool) {
	// 如果要写入文件
	if name != "" {
		fi, err := os.Stat(_dir)
		if err == nil && !fi.IsDir() {
			// 如果存在这个文件， 直接跳过
			log.Printf("%s is not a directory, will input log to the console \n", _dir)
			name = ""
		} else if err != nil {
			// 目录不存在就创建
			if err = os.MkdirAll(_dir, 0755); err != nil {
				log.Println(err)
				return
			}
		}
	}
	getDefaultLog().initFile(filepath.Base(name), size, everyday)
	addClean(filepath.Base(name))
}

// SetConsole 设置全局写文件时是否同时输出到控制台，默认 false。
func SetConsole(b bool) {
	getDefaultLog().SetConsole(b)
}

// SetCompress 设置全局归档旧日志是否异步 gzip 压缩，默认 false。
func SetCompress(b bool) {
	getDefaultLog().SetCompress(b)
}

// SetRateLimit 开启全局限流采样：每个 window 窗口内最多记录 max 条日志。
// max <= 0 时关闭限流。
func SetRateLimit(max int, window time.Duration) {
	getDefaultLog().SetRateLimit(max, window)
}

// SetOutput 设置全局控制台输出目标（默认 os.Stdout），
// 用于测试捕获日志或对接自定义 Writer。
func SetOutput(w io.Writer) {
	t.cw.SetOutput(w)
}

// SetLogPriority 设置全局重复日志采样（参见 (*Log).SetLogPriority）。
func SetLogPriority(logPriority bool, duplicates int, dd ...time.Duration) {
	getDefaultLog().SetLogPriority(logPriority, duplicates, dd...)
}

// Trace 打印 TRACE 级别日志（全局）。
func Trace(msg ...any) { getDefaultLog().Trace(msg...) }

// Tracef 打印 TRACE 级别格式化日志（全局）。
func Tracef(format string, args ...any) { getDefaultLog().Tracef(format, args...) }

// Tracew 打印 TRACE 级别日志并附加结构化字段（全局）。
func Tracew(msg string, fields ...Field) { getDefaultLog().Tracew(msg, fields...) }

// Debug 打印 DEBUG 级别日志（全局）。
func Debug(msg ...any) { getDefaultLog().Debug(msg...) }

// Debugf 打印 DEBUG 级别格式化日志（全局）。
func Debugf(format string, args ...any) { getDefaultLog().Debugf(format, args...) }

// Debugw 打印 DEBUG 级别日志并附加结构化字段（全局）。
func Debugw(msg string, fields ...Field) { getDefaultLog().Debugw(msg, fields...) }

// Info 打印 INFO 级别日志（全局）。
func Info(msg ...any) { getDefaultLog().Info(msg...) }

// Infof 打印 INFO 级别格式化日志（全局）。
func Infof(format string, args ...any) { getDefaultLog().Infof(format, args...) }

// Infow 打印 INFO 级别日志并附加结构化字段（全局）。
func Infow(msg string, fields ...Field) { getDefaultLog().Infow(msg, fields...) }

// Warn 打印 WARN 级别日志（全局）。
func Warn(msg ...any) { getDefaultLog().Warn(msg...) }

// Warnf 打印 WARN 级别格式化日志（全局）。
func Warnf(format string, args ...any) { getDefaultLog().Warnf(format, args...) }

// Warnw 打印 WARN 级别日志并附加结构化字段（全局）。
func Warnw(msg string, fields ...Field) { getDefaultLog().Warnw(msg, fields...) }

// Error 打印 ERROR 级别日志（全局）。
func Error(msg ...any) { getDefaultLog().Error(msg...) }

// Errorf 打印 ERROR 级别格式化日志（全局）。
func Errorf(format string, args ...any) { getDefaultLog().Errorf(format, args...) }

// Errorw 打印 ERROR 级别日志并附加结构化字段（全局）。
func Errorw(msg string, fields ...Field) { getDefaultLog().Errorw(msg, fields...) }

// Fatal 同步写入 FATAL 级别日志（保证落盘）并以退出码 1 结束进程。
func Fatal(msg ...any) { getDefaultLog().Fatal(msg...) }

// Fatalf 同步写入 FATAL 级别格式化日志并结束进程。
func Fatalf(format string, args ...any) { getDefaultLog().Fatalf(format, args...) }

// Fatalw 同步写入 FATAL 级别日志（含结构化字段）并结束进程。
func Fatalw(msg string, fields ...Field) { getDefaultLog().Fatalw(msg, fields...) }

// Panic 同步写入 PANIC 级别日志（保证落盘）后 panic。
func Panic(msg ...any) { getDefaultLog().Panic(msg...) }

// Panicf 同步写入 PANIC 级别格式化日志后 panic。
func Panicf(format string, args ...any) { getDefaultLog().Panicf(format, args...) }

// Stack 打印 ERROR 级别日志并附带当前 goroutine 调用堆栈（全局）。
func Stack(msg ...any) { getDefaultLog().Stack(msg...) }

// UpFunc 打印 DEBUG 级别日志并标注调用者的调用位置（全局）。
func UpFunc(deep int, msg ...any) { getDefaultLog().UpFunc(deep, msg...) }

// UpFuncf 打印 DEBUG 级别格式化日志并标注调用者的调用位置（全局）。
func UpFuncf(deep int, format string, args ...any) {
	getDefaultLog().UpFuncf(deep, format, args...)
}

// With 返回携带新增结构化字段的全局子 logger。
func With(fields ...Field) *Log { return getDefaultLog().With(fields...) }

// Named 返回带模块名的全局子 logger。
func Named(name string) *Log { return getDefaultLog().Named(name) }

// arrToString 将多个任意类型参数以空格连接为字符串。
func arrToString(msg ...interface{}) string {
	ll := make([]string, 0, len(msg))
	for _, v := range msg {
		ll = append(ll, fmt.Sprintf("%v", v))
	}
	return strings.Join(ll, " ")
}

// Wrap 包装 error 并记录包装处的文件行号，最外层打印时自动追加来源路径。
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return &gologError{fileline: printFileline(-1), err: err}
}

// Wraps 将字符串包装为带来源位置的 error。
func Wraps(err string) error {
	return &gologError{fileline: printFileline(-1), err: errors.New(err)}
}

// Unwrap 返回 Wrap/Wraps 包装的原始 error，无法解包时返回 nil。
func Unwrap(err error) error {
	if e, ok := err.(*gologError); ok {
		return e.err
	}
	return nil
}

// gologError 记录错误产生的文件行号及原始错误信息，方便通过 Unwrap 还原。
type gologError struct {
	fileline string
	err      error
}

// Error 实现 error 接口，输出格式：来源位置 -- 原始错误。
func (e *gologError) Error() string {
	return fmt.Sprintf("%s -- %s", e.fileline, e.err)
}

// Unwrap 实现标准库 errors 的解包接口，支持 errors.Is/errors.As。
func (e *gologError) Unwrap() error {
	return e.err
}
