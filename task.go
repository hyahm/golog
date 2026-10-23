package golog

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// task 独立的异步日志写入任务：一个缓存通道 + 一个写协程。
type task struct {
	cache     chan msgLog
	exit      chan struct{}
	handlerWg *sync.WaitGroup // LogHandler 回调的等待组

	cw *consoleWriter

	fwMu sync.Mutex
	fw   *fileWriter

	dropped     atomic.Int64 // 因通道满被丢弃的日志总数
	reportedDrp atomic.Int64 // 上次汇报时的丢弃数

	syncDo sync.Once
	mu     sync.RWMutex // 写锁保护关闭流程，读锁允许并发发送
	closed bool
}

// t 全局默认日志任务。
var t *task

// newTask 创建任务并初始化通道与控制台输出器（写协程由调用方启动）。
func newTask() *task {
	return &task{
		cache:     make(chan msgLog, 16384),
		exit:      make(chan struct{}),
		handlerWg: &sync.WaitGroup{},
		cw:        newConsoleWriter(),
	}
}

func init() {
	t = newTask()
	go t.write()
}

// write 消费通道：控制台日志实时输出，文件日志批量（4KB 或 200ms）落盘。
func (task *task) write() {
	var buf []byte // 文件日志批量拼接缓冲，避免逐条字符串拼接分配
	cl := msgLog{} // 当前批次的目录/文件名等元信息
	flush := func() {
		if len(buf) == 0 {
			return
		}
		cl.Msg = string(buf)
		task.control(cl)
		buf = buf[:0]
		cl = msgLog{}
	}
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop() // 主函数退出前停止 Ticker，防止 goroutine 泄漏
	for {
		select {
		case <-ticker.C:
			// 不管有没有写满，200 毫秒必须刷盘
			flush()
			task.flushFW()
			task.reportDropped()
		case c, ok := <-task.cache:
			if !ok {
				// 通道关闭，刷盘后退出
				flush()
				task.reportDropped()
				task.closeFW()
				task.exit <- struct{}{}
				return
			}
			// Msg 已在生产端完成格式化
			if c.out {
				// 控制台日志加颜色实时打印
				c.Color = GetColor(c.Level)
				task.cw.print(c)
				continue
			}
			if c.console {
				// 写文件同时输出控制台：逐条着色打印
				c.Color = GetColor(c.Level)
				task.cw.print(c)
			}
			if cl.day != 0 && c.Ctime.Day() != cl.day {
				// 跨天时前一批先落盘
				flush()
			}
			cl.dir = c.dir
			cl.out = c.out
			cl.name = c.name
			cl.day = c.Ctime.Day()
			cl.everyDay = c.everyDay
			cl.Ctime = c.Ctime
			cl.size = c.size
			cl.compress = c.compress
			buf = append(buf, c.Msg...)

			if len(buf) < BLOCKSIZE {
				// 缓存没满就继续
				continue
			}
			flush()
		}
	}
}

// reportDropped 汇报因通道满被丢弃的日志数量（至多每个 tick 一次）。
func (task *task) reportDropped() {
	cur := task.dropped.Load()
	if cur == task.reportedDrp.Load() {
		return
	}
	task.reportedDrp.Store(cur)
	msg := fmt.Sprintf("%s -- [WARN] -- golog -- [golog] %d log messages dropped (async buffer full)\n",
		formatTime(time.Now()), cur)
	task.cw.print(msgLog{Msg: msg})
}

var _expireClean time.Duration = time.Hour * 24 * 7

// SetExpireDuration 设置过期清理时间，默认 7 天，须在 InitLogger 前调用。
func SetExpireDuration(d time.Duration) {
	if d > 0 {
		_expireClean = d
	}
}

// send 在通道未关闭时非阻塞写入日志；通道满时丢弃并计数，避免阻塞业务。
// 持读锁可与并发发送共存，closeCache 持写锁等待所有在途发送完成。
func (task *task) send(ml msgLog) {
	task.mu.RLock()
	defer task.mu.RUnlock()
	if task.closed {
		return
	}
	select {
	case task.cache <- ml:
	default:
		task.dropped.Add(1)
	}
}

// closeCache 标记通道关闭并关闭 cache，写锁保证与 send 互斥，防止重复 close。
func (task *task) closeCache() {
	task.mu.Lock()
	defer task.mu.Unlock()
	if !task.closed {
		task.closed = true
		close(task.cache)
	}
}

// syncClose 等待全部异步日志落盘并结束写协程，可安全重复调用。
func (task *task) syncClose() {
	task.syncDo.Do(func() {
		task.handlerWg.Wait()
		task.closeCache()
		<-task.exit
	})
}

// Sync 等待全局异步日志全部落盘，可安全重复调用。
// 程序退出前 defer 调用可避免日志丢失。
func Sync() {
	getDefaultLog().Sync()
}
