package golog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// newBufLog 创建输出到 buffer 的控制台实例。
func newBufLog(buf *bytes.Buffer) *Log {
	l := NewLog("", 0, false)
	l.SetOutput(buf)
	return l
}

// newTmpDir 创建临时目录并注册恢复全局 _dir。
func newTmpDir(t *testing.T) string {
	t.Helper()
	old := _dir
	td := t.TempDir()
	SetDir(td)
	t.Cleanup(func() { SetDir(old) })
	return td
}

func TestJsonFormatEscape(t *testing.T) {
	msg := "he said \"hi\"\nsecond line\ttab"
	line := "a_b.go:1 \"quoted\""
	out := JsonFormat(ERROR, time.Now(), line, msg, []Field{Str("k", "v\"x\ny")})
	if !json.Valid([]byte(strings.TrimRight(out, "\n"))) {
		t.Fatalf("JsonFormat produced invalid JSON: %q", out)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["msg"] != msg {
		t.Fatalf("msg mismatch: %v", m["msg"])
	}
	if m["k"] != "v\"x\ny" {
		t.Fatalf("field mismatch: %v", m["k"])
	}
}

func TestDefaultFormatFields(t *testing.T) {
	out := defaultFormat(INFO, time.Now(), "l.go:9", "hello", []Field{Str("a", "b"), Int("n", 3)})
	if !strings.Contains(out, "hello a=b n=3") {
		t.Fatalf("fields not rendered: %q", out)
	}
	q := defaultFormat(INFO, time.Now(), "l", "m", []Field{Str("sp", "has space")})
	if !strings.Contains(q, `sp="has space"`) {
		t.Fatalf("spaced value not quoted: %q", q)
	}
}

func TestSetOutputCapture(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	l.Info("captured msg")
	l.Sync()
	s := buf.String()
	if !strings.Contains(s, "captured msg") {
		t.Fatalf("output not captured: %q", s)
	}
	if !strings.Contains(s, "golog_ext_test.go") {
		t.Fatalf("caller line missing: %q", s)
	}
}

func TestLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	l.SetLevel(WARN)
	l.Info("should not appear")
	l.Error("should appear")
	l.Sync()
	s := buf.String()
	if strings.Contains(s, "should not appear") {
		t.Fatalf("info leaked below threshold: %q", s)
	}
	if !strings.Contains(s, "should appear") {
		t.Fatalf("error missing: %q", s)
	}
	if l.Level() != WARN {
		t.Fatalf("Level() mismatch: %v", l.Level())
	}
}

func TestFatalSyncFlushBeforeExit(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	code := 0
	exitFunc = func(c int) { code = c }
	defer func() { exitFunc = os.Exit }()

	l.Fatal("boom fatal")
	// 同步写路径：无需 Sync 即应可见
	if !strings.Contains(buf.String(), "boom fatal") {
		t.Fatalf("fatal log not flushed before exit: %q", buf.String())
	}
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestFatalfExits(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	called := false
	exitFunc = func(c int) { called = true }
	defer func() { exitFunc = os.Exit }()

	l.Fatalf("fatal %s", "fmt")
	if !strings.Contains(buf.String(), "fatal fmt") {
		t.Fatalf("fatalf not flushed: %q", buf.String())
	}
	if !called {
		t.Fatal("Fatalf did not exit")
	}
}

func TestPanicLogs(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Panic should panic")
		}
		if !strings.Contains(buf.String(), "panicking") {
			t.Fatalf("panic log not written: %q", buf.String())
		}
	}()
	l.Panic("panicking")
}

func TestPanicfLogs(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	defer func() {
		_ = recover()
		if !strings.Contains(buf.String(), "pf 42") {
			t.Fatalf("panicf log not written: %q", buf.String())
		}
	}()
	l.Panicf("pf %d", 42)
}

func TestStack(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	l.Stack("stack here")
	l.Sync()
	if !strings.Contains(buf.String(), "goroutine") {
		t.Fatalf("stack trace missing: %q", buf.String())
	}
}

func TestWithAndFields(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	base := l.With(Str("svc", "order"))
	base.Infow("doing", Int("id", 7))
	child := base.With(Str("sub", "x"))
	child.Infow("more")
	l.Sync()
	s := buf.String()
	if !strings.Contains(s, "doing svc=order id=7") {
		t.Fatalf("with fields missing: %q", s)
	}
	if !strings.Contains(s, "more svc=order sub=x") {
		t.Fatalf("nested with fields missing: %q", s)
	}
	// 父实例不应被污染
	var buf2 bytes.Buffer
	l2 := newBufLog(&buf2)
	l2.Info("clean")
	l2.Sync()
	if strings.Contains(buf2.String(), "svc=order") {
		t.Fatalf("parent polluted by child: %q", buf2.String())
	}
}

func TestNamed(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	l.Named("db").Info("query")
	l.Sync()
	if !strings.Contains(buf.String(), "logger=db") {
		t.Fatalf("named field missing: %q", buf.String())
	}
}

func TestCtxFields(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	ctx := ContextWithTraceID(ContextWithFields(context.Background(), Str("uid", "u1")), "t-123")
	l.Ctx(ctx).Info("with ctx")
	l.Ctx(context.Background()).Info("no ctx fields")
	l.Sync()
	s := buf.String()
	if !strings.Contains(s, "uid=u1 trace_id=t-123") {
		t.Fatalf("ctx fields missing: %q", s)
	}
	if strings.Count(s, "no ctx fields") != 1 || !strings.Contains(s, "no ctx fields\n") {
		t.Fatalf("plain ctx log polluted: %q", s)
	}
	if strings.Contains(strings.Split(s, "no ctx fields")[1], "trace_id") {
		t.Fatalf("ctx fields leaked to plain log: %q", s)
	}
}

func TestSlogAdapter(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	sl := slog.New(NewSlogHandler(l))
	sl.Info("hello slog", "uid", 42)
	sl.Warn("warned")
	sl.Debug("dropped by level")

	// WithAttrs / WithGroup / Enabled
	h := NewSlogHandler(l).WithAttrs([]slog.Attr{slog.String("pre", "1")}).(slog.Handler)
	h2 := h.WithGroup("g").(*SlogHandler)
	sl2 := slog.New(h2)
	if !sl2.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("Enabled should be true for error")
	}
	l.SetLevel(ERROR)
	if sl2.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("Enabled should be false for info at ERROR threshold")
	}
	l.SetLevel(DEBUG)
	sl2.Error("grouped", "b", 2)
	l.Sync() // 单次 Sync 收尾，避免任务提前关闭
	s := buf.String()
	if !strings.Contains(s, "hello slog") || !strings.Contains(s, "uid=42") {
		t.Fatalf("slog basic missing: %q", s)
	}
	if !strings.Contains(s, "warned") {
		t.Fatalf("slog warn missing: %q", s)
	}
	if strings.Contains(s, "dropped by level") {
		t.Fatalf("slog debug leaked: %q", s)
	}
	if !strings.Contains(s, "pre=1 g.b=2") {
		t.Fatalf("slog group/attrs missing: %q", s)
	}
}

func TestRateLimit(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	l.SetRateLimit(2, time.Minute)
	for i := 0; i < 5; i++ {
		l.Info("limited msg")
	}
	l.SetRateLimit(0, 0) // 关闭限流
	l.Info("after off")
	l.Sync() // 单次 Sync 收尾
	s := buf.String()
	if got := strings.Count(s, "limited msg"); got != 2 {
		t.Fatalf("rate limit allowed %d logs, want 2", got)
	}
	if !strings.Contains(s, "after off") {
		t.Fatal("rate limit off failed")
	}
}

func TestSamplerUnit(t *testing.T) {
	s := newSampler(2, time.Minute)
	if !s.allow() || !s.allow() {
		t.Fatal("first two should pass")
	}
	if s.allow() {
		t.Fatal("third should be limited")
	}
	if s.Dropped() != 1 {
		t.Fatalf("dropped = %d, want 1", s.Dropped())
	}
}

func TestDuplicateClose(t *testing.T) {
	d := newDuplicate(3, 30*time.Millisecond)
	if !d.addMsg("k") {
		t.Fatal("first should pass")
	}
	if d.addMsg("k") {
		t.Fatal("second should be suppressed")
	}
	d.close()
	d.close() // 重复 close 应安全
}

func TestDroppedCounter(t *testing.T) {
	tk := newTask()
	tk.cache = make(chan msgLog, 1) // 缩小缓冲以便触发丢弃
	tk.send(msgLog{Msg: "1"})
	tk.send(msgLog{Msg: "2"}) // 缓冲为 1，第二条应被丢弃计数
	if got := tk.dropped.Load(); got != 1 {
		t.Fatalf("dropped = %d, want 1", got)
	}
}

func TestFileWrite(t *testing.T) {
	td := newTmpDir(t)
	l := NewLog("wf.log", 0, false)
	l.Infof("file message %d", 1)
	l.Sync()
	b, err := os.ReadFile(filepath.Join(td, "wf.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(b), "file message 1") {
		t.Fatalf("file content missing: %q", b)
	}
}

func TestMultiOutput(t *testing.T) {
	td := newTmpDir(t)
	var buf bytes.Buffer
	l := NewLog("multi.log", 0, false)
	l.SetConsole(true)
	l.SetOutput(&buf)
	l.Info("both places")
	l.Sync()
	if !strings.Contains(buf.String(), "both places") {
		t.Fatalf("console copy missing: %q", buf.String())
	}
	b, err := os.ReadFile(filepath.Join(td, "multi.log"))
	if err != nil || !strings.Contains(string(b), "both places") {
		t.Fatalf("file copy missing: %q, err=%v", b, err)
	}
}

func TestSizeRotation(t *testing.T) {
	td := t.TempDir()
	fw := &fileWriter{dir: td, name: "r.log", size: 1}
	big := strings.Repeat("a", 600<<10)
	if err := fw.write(msgLog{dir: td, name: "r.log", size: 1, Ctime: time.Now(), Msg: big}); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := fw.write(msgLog{dir: td, name: "r.log", size: 1, Ctime: time.Now(), Msg: big}); err != nil {
		t.Fatalf("second write: %v", err)
	}
	fw.close()
	archived, _ := filepath.Glob(filepath.Join(td, "*_r.log"))
	if len(archived) != 1 {
		t.Fatalf("archived files = %v, want 1", archived)
	}
	if _, err := os.Stat(filepath.Join(td, "r.log")); err != nil {
		t.Fatalf("current file missing: %v", err)
	}
}

func TestDayRotation(t *testing.T) {
	td := t.TempDir()
	fw := &fileWriter{dir: td, name: "d.log", everyDay: true}
	yesterday := time.Now().AddDate(0, 0, -1)
	if err := fw.write(msgLog{dir: td, name: "d.log", everyDay: true, Ctime: yesterday, Msg: "old\n"}); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if err := fw.write(msgLog{dir: td, name: "d.log", everyDay: true, Ctime: time.Now(), Msg: "new\n"}); err != nil {
		t.Fatalf("write new: %v", err)
	}
	fw.close()
	archived := filepath.Join(td, yesterday.Format("2006-01-02")+"_d.log")
	if _, err := os.Stat(archived); err != nil {
		t.Fatalf("day-archived file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(td, "d.log")); err != nil {
		t.Fatalf("current file missing: %v", err)
	}
}

func TestCompressRotation(t *testing.T) {
	td := t.TempDir()
	fw := &fileWriter{dir: td, name: "c.log", size: 1, compress: true}
	big := strings.Repeat("b", 600<<10)
	_ = fw.write(msgLog{dir: td, name: "c.log", size: 1, compress: true, Ctime: time.Now(), Msg: big})
	_ = fw.write(msgLog{dir: td, name: "c.log", size: 1, compress: true, Ctime: time.Now(), Msg: big})
	fw.close()

	deadline := time.Now().Add(3 * time.Second)
	var gz string
	for time.Now().Before(deadline) {
		if m, _ := filepath.Glob(filepath.Join(td, "*_c.log.gz")); len(m) == 1 {
			// 压缩为异步流程，.gz 生成后原文件才被删除，两者都就绪才算完成
			if old, _ := filepath.Glob(filepath.Join(td, "*_c.log")); len(old) == 0 {
				gz = m[0]
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if gz == "" {
		t.Fatal("compressed archive not created or original not removed")
	}
}

func TestTimeLayoutUTC(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	defer func() {
		SetTimeLayout("2006-01-02 15:04:05.000")
		SetUTC(false)
	}()
	SetTimeLayout(time.RFC3339)
	SetUTC(true)
	l.Info("utc msg")
	l.Sync()
	if !strings.Contains(buf.String(), "Z ") {
		t.Fatalf("UTC/RFC3339 layout not applied: %q", buf.String())
	}
}

func TestSyncIdempotent(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	l.Info("before sync")
	l.Sync()
	l.Sync() // 重复调用不应死锁/panic
	if !strings.Contains(buf.String(), "before sync") {
		t.Fatalf("log lost: %q", buf.String())
	}
}

func TestUpFunc(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	l.SetLevel(DEBUG) // UpFunc 为 DEBUG 级别
	upFuncHelper(l)
	l.Sync()
	if !strings.Contains(buf.String(), "caller from") {
		t.Fatalf("caller info missing: %q", buf.String())
	}
}

func upFuncHelper(l *Log) {
	l.UpFunc(1, "who called me")
}

func TestWrapFileline(t *testing.T) {
	err := Wrap(os.ErrNotExist)
	if err == nil || !strings.Contains(err.Error(), "golog_ext_test.go") {
		t.Fatalf("Wrap should record fileline: %v", err)
	}
	if Unwrap(err) != os.ErrNotExist {
		t.Fatalf("Unwrap mismatch: %v", Unwrap(err))
	}
}

func TestConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLog(&buf)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				l.Infow("conlog", Int("g", g), Int("i", i))
			}
		}(g)
	}
	// 并发修改级别，验证无 data race
	for i := 0; i < 50; i++ {
		l.SetLevel(DEBUG)
		l.SetLevel(INFO)
	}
	wg.Wait()
	l.Sync()
	if got := strings.Count(buf.String(), "conlog"); got == 0 {
		t.Fatal("no concurrent logs written")
	}
}

func TestGetLevelRoundTrip(t *testing.T) {
	SetLevel(DEBUG)
	if GetLevel() != DEBUG {
		t.Fatalf("GetLevel = %v, want DEBUG", GetLevel())
	}
	SetLevel(INFO)
}

func TestCompressFileUnit(t *testing.T) {
	td := t.TempDir()
	src := filepath.Join(td, "unit.log")
	if err := os.WriteFile(src, []byte("hello gzip"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := compressFile(src); err != nil {
		t.Fatalf("compressFile: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("source should be removed after compress")
	}
	if fi, err := os.Stat(src + ".gz"); err != nil || fi.Size() == 0 {
		t.Fatalf("gz file missing or empty: %v", err)
	}
}

func TestUniqPath(t *testing.T) {
	td := t.TempDir()
	p := filepath.Join(td, "x.log")
	os.WriteFile(p, []byte("1"), 0644)
	os.WriteFile(filepath.Join(td, "x_1.log"), []byte("2"), 0644)
	got := uniqPath(p)
	if got != filepath.Join(td, "x_2.log") {
		t.Fatalf("uniqPath = %s", got)
	}
}
