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
	mu     sync.Mutex
	closed bool
}

// t 全局默认日志任务。
var t *task

// newTask 创建任务并初始化通道与控制台输出器（写协程由调用方启动）。
func newTask() *task {
	return &task{
		cache:     make(chan msgLog, 1000),
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
	cl := msgLog{}
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop() // 主函数退出前停止 Ticker，防止 goroutine 泄漏
	for {
		select {
		case <-ticker.C:
			// 不管有没有写满，200 毫秒必须刷盘
			if len(cl.Msg) > 0 {
				task.control(cl)
				cl = msgLog{}
			}
			task.reportDropped()
		case c, ok := <-task.cache:
			if !ok {
				// 通道关闭，刷盘后退出
				if len(cl.Msg) > 0 {
					task.control(cl)
					cl = msgLog{}
				}
				task.reportDropped()
				task.closeFW()
				task.exit <- struct{}{}
				return
			}
			c.Msg = c.format(c.Level, c.Ctime, c.Line, c.Msg, c.fields)
			if c.out {
				// 控制台日志加颜色实时打印
				c.Color = GetColor(c.Level)
				task.control(c)
				continue
			}
			if c.console {
				// 写文件同时输出控制台：逐条着色打印
				c.Color = GetColor(c.Level)
				task.cw.print(c)
			}
			if cl.day != 0 && c.Ctime.Day() != cl.day {
				task.control(cl)
				cl = msgLog{}
			}
			cl.dir = c.dir
			cl.out = c.out
			cl.name = c.name
			cl.day = c.Ctime.Day()
			cl.everyDay = c.everyDay
			cl.Ctime = c.Ctime
			cl.size = c.size
			cl.compress = c.compress
			cl.Msg += c.Msg

			if len(cl.Msg) < BLOCKSIZE {
				// 缓存没满就继续
				continue
			}
			task.control(cl)
			cl = msgLog{}
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
func (task *task) send(ml msgLog) {
	task.mu.Lock()
	if task.closed {
		task.mu.Unlock()
		return
	}
	select {
	case task.cache <- ml:
		task.mu.Unlock()
	default:
		task.mu.Unlock()
		task.dropped.Add(1)
	}
}

// closeCache 标记通道关闭并关闭 cache，保证与 send 互斥，防止重复 close。
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
