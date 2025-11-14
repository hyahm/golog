package golog

import (
	"sync"
	"time"
)

type task struct {
	cache chan msgLog
	exit  chan struct{}
	wg    *sync.WaitGroup
}

var t *task

func init() {
	t = &task{
		cache: make(chan msgLog, 1000),
		exit:  make(chan struct{}),
		wg:    &sync.WaitGroup{},
	}
	// exit = make(chan bool)
	// 增加1024字节的缓存， 也就是假设每一条日志的最大长度是1024
	go t.write()

}

func (t *task) write() {
	cl := msgLog{
		// buf: bytes.NewBuffer(nil),
		// now: time.Now(),
	}
	// go SecondCache()
	ticker := time.NewTicker(200 * time.Millisecond)

	defer ticker.Stop() // 主函数退出前停止 Ticker，防止 goroutine 泄漏
	for {
		select {
		case <-ticker.C:
			// 不管有没有写满， 200毫秒必须
			if len(cl.Msg) > 0 {
				t.control(cl)
				cl.Msg = ""
				cl.day = 0
			}

		case c, ok := <-t.cache:

			// fmt.Println("--------------", c.Msg)
			if !ok {
				// 关闭通道 的操作
				if len(cl.Msg) > 0 {
					t.control(cl)
					cl.Msg = ""
					cl.day = 0
				}
				t.exit <- struct{}{}
				return
			}
			if c.out {
				// 控制台才添加颜色， 否则不添加颜色
				c.Color = GetColor(c.Level)
			}
			c.Msg = c.format(c.Level, c.Ctime, c.Line, c.Msg)
			if c.out {
				// 有带颜色日志要实时打印
				t.control(c)
				cl.Msg = ""
				cl.day = 0
				continue
			}

			if cl.day != 0 && c.Ctime.Day() != cl.day {
				t.control(c)
				continue
			}
			cl.dir = c.dir
			cl.out = c.out
			cl.name = c.name
			cl.day = c.Ctime.Day()
			cl.everyDay = c.everyDay
			cl.Ctime = c.Ctime
			cl.size = c.size
			cl.Color = c.Color
			cl.Msg += c.Msg

			if len(cl.Msg) < BLOCKSIZE {
				// 缓存没满就继续
				continue
			}
			t.control(cl)
			cl.Msg = ""
			cl.day = 0

		}
	}

}

var _expireClean time.Duration = time.Hour * 24 * 365

// 设置清理时间 默认365天
func SetExpireDuration(d time.Duration) {
	if d > 0 {
		_expireClean = d
	}

}

func Sync() {
	// 等待所有通道写完日志写完,  可以不写，
	// time.Sleep(1 * time.Millisecond * 300)

	t.wg.Wait()
	close(t.cache)
	<-t.exit

}

func (l *Log) Sync() {
	// 等待所有通道写完日志写完,  可以不写，
	// time.Sleep(1 * time.Millisecond * 300)
	l.task.wg.Wait()
	close(l.task.cache)
	<-l.task.exit

}
