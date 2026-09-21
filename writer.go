package golog

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

// _fwBufSize 文件写缓冲大小。
const _fwBufSize = 64 << 10

// compressSem 限制并发压缩的协程数。
var compressSem = make(chan struct{}, 2)

// consoleWriter 控制台（或自定义 io.Writer）输出器，写操作互斥，
// 输出目标可通过 SetOutput 替换（测试捕获、对接 syslog/kafka 等）。
type consoleWriter struct {
	mu sync.Mutex
	w  io.Writer
}

// newConsoleWriter 创建默认输出到 stdout 的控制台输出器。
func newConsoleWriter() *consoleWriter {
	return &consoleWriter{w: colorable.NewColorable(os.Stdout)}
}

// SetOutput 替换输出目标，w 为 nil 时忽略。
func (c *consoleWriter) SetOutput(w io.Writer) {
	if w == nil {
		return
	}
	c.mu.Lock()
	c.w = w
	c.mu.Unlock()
}

// writeString 互斥写入字符串。
func (c *consoleWriter) writeString(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, _ = io.WriteString(c.w, s)
}

// print 输出一条日志，按级别颜色着色（无颜色属性时原样输出）。
func (c *consoleWriter) print(ml msgLog) {
	if len(ml.Color) > 0 {
		c.writeString(color.New(ml.Color...).Sprint(ml.Msg))
		return
	}
	c.writeString(ml.Msg)
}

// fileWriter 长期持有文件句柄的文件写入器，负责按大小/按天切割与 gzip 归档。
type fileWriter struct {
	mu       sync.Mutex
	dir      string
	name     string
	size     int64 // 按大小切割阈值(MB)，0 不切割
	everyDay bool  // 按天切割
	compress bool  // 归档文件是否异步 gzip 压缩

	f       *os.File
	w       *bufio.Writer
	curSize int64  // 当前文件已写字节数
	curDay  string // 当前文件对应的日期 2006-01-02
}

// reconfig 检测配置变化（目录/文件名/切割策略），变化时关闭旧句柄。
func (fw *fileWriter) reconfig(ml msgLog) {
	if fw.dir == ml.dir && fw.name == ml.name && fw.size == ml.size &&
		fw.everyDay == ml.everyDay && fw.compress == ml.compress {
		return
	}
	fw.flushLocked()
	fw.dir, fw.name, fw.size = ml.dir, ml.name, ml.size
	fw.everyDay, fw.compress = ml.everyDay, ml.compress
}

// write 写入一批日志（ml.Msg 为已格式式的完整文本）。
func (fw *fileWriter) write(ml msgLog) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.reconfig(ml)
	day := ml.Ctime.Format("2006-01-02")
	if fw.f == nil {
		if err := fw.open(day); err != nil {
			return err
		}
	}
	// 按天切割：跨天时把当前文件归档为 日期_文件名
	if fw.everyDay && day != fw.curDay {
		fw.rotateLocked(fw.curDay)
		if err := fw.open(day); err != nil {
			return err
		}
	}
	// 按大小切割：写满阈值后归档为 时间戳_文件名
	if fw.size > 0 && fw.curSize+int64(len(ml.Msg)) > fw.size<<20 {
		fw.rotateLocked(ml.Ctime.Format("2006-01-02_15_04_05"))
		if err := fw.open(day); err != nil {
			return err
		}
	}
	n, err := fw.w.WriteString(ml.Msg)
	fw.curSize += int64(n)
	if err != nil {
		return err
	}
	return fw.w.Flush()
}

// open 打开/创建当前日志文件；按天切割模式下若存在历史文件先归档。
func (fw *fileWriter) open(day string) error {
	path := filepath.Join(fw.dir, fw.name)
	if fw.everyDay {
		if fi, err := os.Stat(path); err == nil && !sameDay(fi.ModTime(), time.Now()) {
			old := fi.ModTime().Format("2006-01-02") + "_" + fw.name
			if err := os.Rename(path, filepath.Join(fw.dir, old)); err == nil {
				fw.compressAsync(filepath.Join(fw.dir, old))
			}
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	fw.f = f
	fw.w = bufio.NewWriterSize(f, _fwBufSize)
	fw.curSize = fi.Size()
	fw.curDay = day
	return nil
}

// rotateLocked 关闭并归档当前文件（调用方须持有锁）。
func (fw *fileWriter) rotateLocked(stamp string) {
	fw.flushLocked()
	path := filepath.Join(fw.dir, fw.name)
	if fi, err := os.Stat(path); err == nil && fi.Size() > 0 {
		target := uniqPath(filepath.Join(fw.dir, stamp+"_"+fw.name))
		if err := os.Rename(path, target); err != nil {
			log.Println("golog: rotate rename failed:", err)
		} else if fw.compress {
			fw.compressAsync(target)
		}
	}
	fw.curSize = 0
}

// flushLocked 刷新缓冲并关闭句柄（调用方须持有锁）。
func (fw *fileWriter) flushLocked() {
	if fw.w != nil {
		_ = fw.w.Flush()
		fw.w = nil
	}
	if fw.f != nil {
		_ = fw.f.Close()
		fw.f = nil
	}
}

// close 刷新缓冲并关闭文件，用于 Sync 收尾。
func (fw *fileWriter) close() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.flushLocked()
}

// compressAsync 异步 gzip 压缩归档文件（并发受 compressSem 限制）。
func (fw *fileWriter) compressAsync(path string) {
	go func() {
		compressSem <- struct{}{}
		defer func() { <-compressSem }()
		if err := compressFile(path); err != nil {
			log.Println("golog: compress log failed:", err)
		}
	}()
}

// compressFile 将文件压缩为 path.gz 并删除原文件。
func compressFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	out, err := os.Create(path + ".gz")
	if err != nil {
		in.Close()
		return err
	}
	gz := gzip.NewWriter(out)
	_, err = io.Copy(gz, in)
	in.Close() // Windows 下删除前必须先关闭句柄
	if err == nil {
		err = gz.Close()
	}
	if err == nil {
		err = out.Close()
	} else {
		out.Close()
	}
	if err != nil {
		_ = os.Remove(path + ".gz")
		return err
	}
	return os.Remove(path)
}

// sameDay 判断两个时间是否同一天。
func sameDay(a, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}

// uniqPath 若路径已存在则追加 _1/_2... 序号，避免覆盖。
func uniqPath(p string) string {
	if _, err := os.Stat(p); err != nil {
		return p
	}
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(p, ext)
	for i := 1; i < 100; i++ {
		cand := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(cand); err != nil {
			return cand
		}
	}
	return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
}
