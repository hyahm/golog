package golog

import (
	"errors"
	"fmt"
	"testing"
)

func TestInitLogger(t *testing.T) {
	defer Sync()
	// InitLogger("aa.log", 10, false)
	// SetExpireDuration(time.Second * 5)
	SetLevel(DEBUG)
	s := a()

	Info(s)

	m := Unwrap(s)

	Info(m)

	// time.Sleep(10 * time.Second)
	// ShowBasePath = true
	// l2.SetLogPriority(true, 100, time.Minute)
	// WarnHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
	// 	fmt.Println(msg)
	// }
	// ErrorHandler = func(ctime time.Time, hostname, line, msg string, label map[string]string) {
	// 	fmt.Println(msg)
	// }

	fmt.Println(Wrap(a()))

	// time.Sleep(1 * time.Second)
	// golog.InitLogger("log/a.log", 1024, false, 10)
	// a := NewLog("log/a.log", 1024, true, 10)
	// for range 100 {
	// 	a.Info("foo", "aaaa", "bb")
	// }
	// a.Warn(color.New(color.BgYellow).Sprint("aaaa"), color.New(color.BgBlue).Sprint("bbbb"))
	// Level = DEBUG
	// // test()
	// a.Error("bar")
	// time.Sleep(time.Second * 100)
}

func a() error {
	return Wrap(errors.New("aaaaa"))
}

func TestWrapUnwrap(t *testing.T) {
	origin := errors.New("origin")
	w := Wrap(origin)
	if !errors.Is(w, origin) {
		t.Fatalf("Wrap should preserve errors.Is: %v", w)
	}
	if Unwrap(w) != origin {
		t.Fatalf("Unwrap(Wrap(err)) should return origin, got %v", Unwrap(w))
	}
	if Unwrap(errors.New("other")) != nil {
		t.Fatalf("Unwrap should return nil for non-golog error")
	}

	ws := Wraps("boom")
	if ws.(*gologError).err.Error() != "boom" {
		t.Fatalf("Wraps should hold string error, got %v", ws)
	}
	if Unwrap(ws) == nil {
		t.Fatalf("Unwrap(Wraps(str)) should return the string error")
	}
}
