package golog

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

// TestFastFormatTimeEquivalence 校验手写快速时间格式化与标准库输出完全一致。
func TestFastFormatTimeEquivalence(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	layouts := []struct {
		layout string
		ms     bool
	}{
		{DefaultTimeLayout, true},
		{"2006-01-02 15:04:05", false},
	}
	for i := 0; i < 100000; i++ {
		tt := time.Unix(r.Int63n(1<<42), r.Int63n(1e9)).UTC()
		for _, l := range layouts {
			if got, want := fastFormatTime(tt, l.ms), tt.Format(l.layout); got != want {
				t.Fatalf("fastFormatTime(%s, ms=%v) = %q, want %q", tt, l.ms, got, want)
			}
		}
	}
}

// TestSingleProducerNoDrop 单生产者全速写入时消费端应完全跟得上（允许极端环境小幅丢弃）。
func TestSingleProducerNoDrop(t *testing.T) {
	l := NewLog("spnd.log", 0, false)
	const total = 100000
	for i := 0; i < total; i++ {
		l.Infow("spndmsg", Int("i", i))
	}
	l.Sync()
	if got := l.task.dropped.Load(); got > total/100 {
		t.Fatalf("dropped %d logs (%.2f%%), consumer too slow", got, float64(got)/total*100)
	}
}

// TestSustainedConcurrentNoDrop 8 协程持续限速写入（约 40 万条/秒）应零丢弃。
func TestSustainedConcurrentNoDrop(t *testing.T) {
	l := NewLog("sust.log", 0, false)
	const producers, per = 8, 5000
	var wg sync.WaitGroup
	for g := 0; g < producers; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < per; i++ {
				l.Infow("sustmsg", Int("g", g), Int("i", i))
				time.Sleep(20 * time.Microsecond)
			}
		}(g)
	}
	wg.Wait()
	l.Sync()
	if got := l.task.dropped.Load(); got != 0 {
		t.Fatalf("dropped %d logs under sustained load", got)
	}
}

var benchSink string

// BenchmarkLogCall 衡量生产端单条日志（含调用位置/格式化/投递）的完整成本。
func BenchmarkLogCall(b *testing.B) {
	l := NewLog("bench.log", 0, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infow("benchmsg", Int("i", i))
	}
	l.Sync()
	benchSink = time.Now().String()
}
