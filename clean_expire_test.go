package golog

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// listDir 打印目录内剩余的日志文件及各自距今的老化时间，便于观察每轮清理的变化。
func listDir(t *testing.T, dir, title string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) == 0 {
		t.Logf("%s ==> 目录已空\n", title)
		return
	}
	left := make([]string, 0, len(entries))
	for _, e := range entries {
		info, err1 := e.Info()
		if err1 != nil {
			continue
		}
		left = append(left, fmt.Sprintf("%s(老化%s)", e.Name(), time.Since(info.ModTime()).Round(time.Second)))
	}
	t.Logf("%s ==> 剩余 %d 个文件: %v\n", title, len(entries), left)
}

// TestNewLogExpireCleanLoop 循环创建多个 NewLog 实例并多轮模拟时间推进，
// 观察日志文件被逐轮清理的变化：
// 偶数序号实例各自设置不同的过期时长（1h/3h/5h/7h/9h），
// 奇数序号实例不设置（跟随全局默认 7 天），
// 每轮将文件修改时间回拨指定小时后调用 walkDir 清理，直至全部过期删除。
func TestNewLogExpireCleanLoop(t *testing.T) {
	td := newTmpDir(t)

	// 第一轮循环：创建 10 个实例，偶数序号单独设置过期时长
	const n = 10
	names := make([]string, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("loop_%d.log", i)
		l := NewLog(name, 0, false)
		if i%2 == 0 {
			d := time.Duration(i+1) * time.Hour
			l.SetExpireDuration(d)
			t.Logf("创建 %-14s 过期时长 = %v（实例单独设置）", name, d)
		} else {
			t.Logf("创建 %-14s 过期时长 = %v（跟随全局默认）", name, _expireClean)
		}
		l.Infof("hello %d", i)
		l.Sync()
		names = append(names, name)

		// 校验清理注册表中记录的过期时长
		mu.RLock()
		got := _names[name]
		mu.RUnlock()
		want := time.Duration(0)
		if i%2 == 0 {
			want = time.Duration(i+1) * time.Hour
		}
		if got != want {
			t.Fatalf("expire of %s = %v, want %v", name, got, want)
		}
	}
	// 补一个同名 .gz 归档，验证归档同样会被清理
	gzFile := filepath.Join(td, names[0]+".20261124.gz")
	if err := os.WriteFile(gzFile, []byte("fake"), 0644); err != nil {
		t.Fatalf("write %s: %v", gzFile, err)
	}
	listDir(t, td, "初始状态")

	// 第二轮循环：按小时推进模拟时间，每轮回拨文件修改时间并清理一次
	// 1~10 小时逐轮推进（偶数序号实例依次过期），最后跳到 169 小时（7 天+1 小时，奇数序号全部过期）
	rounds := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 7*24 + 1}
	for _, hour := range rounds {
		past := time.Now().Add(-time.Duration(hour) * time.Hour)
		for _, name := range names {
			fp := filepath.Join(td, name)
			if _, err := os.Stat(fp); err == nil {
				_ = os.Chtimes(fp, past, past)
			}
		}
		if _, err := os.Stat(gzFile); err == nil {
			_ = os.Chtimes(gzFile, past, past)
		}
		before := countFiles(t, td)
		if err := walkDir(); err != nil {
			t.Fatalf("walkDir: %v", err)
		}
		after := countFiles(t, td)
		title := fmt.Sprintf("第 %2d 轮：所有文件老化 %2d 小时后执行清理（删除 %d 个）", index(rounds, hour)+1, hour, before-after)
		listDir(t, td, title)
	}

	// 最终验证：全部文件与归档均被删除
	if c := countFiles(t, td); c != 0 {
		t.Fatalf("expect empty dir, still %d files", c)
	}
	t.Logf("全部实例的日志文件均已按各自过期时长清理完毕")
}

// countFiles 统计目录内文件数量。
func countFiles(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	return len(entries)
}

// index 返回 v 在 s 中的下标（测试辅助，找不到返回 -1）。
func index(s []int, v int) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

// TestNewLogExpireCleanRealAging 用真实时间流逝验证：过期时长 1 秒的实例，
// 文件在超过 1 秒后被清理，未设置实例的文件不受影响。
func TestNewLogExpireCleanRealAging(t *testing.T) {
	td := newTmpDir(t)

	exp := NewLog("aging_expire.log", 0, false)
	exp.SetExpireDuration(time.Second)
	keep := NewLog("aging_keep.log", 0, false)
	t.Logf("创建 aging_expire.log 过期时长 = 1s（实例单独设置）")
	t.Logf("创建 aging_keep.log  过期时长 = %v（跟随全局默认）", _expireClean)
	exp.Infof("expire me")
	keep.Infof("keep me")
	exp.Sync()
	keep.Sync()
	listDir(t, td, "初始状态")

	// 等待超过 1 秒的过期时长
	time.Sleep(1100 * time.Millisecond)

	if err := walkDir(); err != nil {
		t.Fatalf("walkDir: %v", err)
	}
	listDir(t, td, "等待 1.1 秒后执行清理")

	if _, err := os.Stat(filepath.Join(td, "aging_expire.log")); !os.IsNotExist(err) {
		t.Fatalf("aging_expire.log should be removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(td, "aging_keep.log")); err != nil {
		t.Fatalf("aging_keep.log should remain: %v", err)
	}
}
