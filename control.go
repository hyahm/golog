package golog

import (
	"fmt"
	"log"
	"path"
	"runtime"
)

// BLOCKSIZE 批量写文件的缓冲阈值（字节）。
const BLOCKSIZE = 4 << 10

// getFW 懒创建任务对应的文件写入器。
func (task *task) getFW(ml msgLog) *fileWriter {
	task.fwMu.Lock()
	defer task.fwMu.Unlock()
	if task.fw == nil {
		task.fw = &fileWriter{
			dir:      ml.dir,
			name:     ml.name,
			size:     ml.size,
			everyDay: ml.everyDay,
			compress: ml.compress,
		}
	}
	return task.fw
}

// closeFW 关闭并刷新文件写入器，用于退出收尾。
func (task *task) closeFW() {
	task.fwMu.Lock()
	defer task.fwMu.Unlock()
	if task.fw != nil {
		task.fw.close()
	}
}

// control 将一批日志写入目标：控制台或文件（失败时回退控制台）。
func (task *task) control(cl msgLog) {
	if cl.out {
		task.cw.print(cl)
		return
	}
	fw := task.getFW(cl)
	if err := fw.write(cl); err != nil {
		log.Println("golog: write file failed:", err)
		cl.out = true
		task.cw.print(cl)
	}
}

// syncWrite 绕过异步通道直接写入，用于 Fatal/Panic 等必须立即落盘的日志。
func (task *task) syncWrite(ml msgLog) {
	if ml.out {
		task.cw.print(ml)
		return
	}
	fw := task.getFW(ml)
	if err := fw.write(ml); err != nil {
		log.Println("golog: write file failed:", err)
		ml.out = true
		task.cw.print(ml)
		return
	}
	if ml.console {
		task.cw.print(ml)
	}
}

// printFileline 返回调用者位置，c 为向外追加的栈深度。
func printFileline(c int) string {
	c += 3
	_, file, line, ok := runtime.Caller(c)
	if !ok {
		file = "???"
		line = 0
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// printBaseFileline 返回调用者位置（仅文件名），c 为向外追加的栈深度。
func printBaseFileline(c int) string {
	c += 3
	_, file, line, ok := runtime.Caller(c)
	if !ok {
		file = "???"
		line = 0
	}
	fileBase := path.Base(file)
	return fmt.Sprintf("%s:%d", fileBase, line)
}
