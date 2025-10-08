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
			if len(cl.Msg) > 0 && time.Since(cl.now).Milliseconds() > 100 {
				t.control(cl)
				cl.Msg = ""

			}

		case c, ok := <-t.cache:
			if !ok {
				t.exit <- struct{}{}
				return
			}
			if len(c.Color) > 0 {
				// 有带颜色日志要实时打印
				t.control(c)
				continue
			}
			cl.dir = c.dir
			cl.out = c.out
			cl.now = time.Now()
			cl.name = c.name
			cl.everyDay = c.everyDay
			cl.Ctime = c.Ctime
			cl.size = c.size
			if len(cl.Msg) < BLOCKSIZE {
				cl.Msg += c.Msg
				continue
			}
			cl.Msg += c.Msg
			t.control(cl)
			cl.Msg = ""

		}
	}

}

var _expireClean time.Duration = time.Hour * 24 * 365

// 设置清理时间 默认365天
func SetExpireDuration(d time.Duration) {

}

func Sync() {
	// 等待所有通道写完日志写完,  可以不写，
	// time.Sleep(1 * time.Millisecond * 300)

	t.wg.Wait()
	close(t.cache)
	<-t.exit
	// time.Sleep(time.Millisecond * 200)

}

func (l *Log) Sync() {
	// 等待所有通道写完日志写完,  可以不写，
	// time.Sleep(1 * time.Millisecond * 300)
	l.task.wg.Wait()
	close(l.task.cache)
	<-l.task.exit
	// time.Sleep(time.Millisecond * 200)

}
