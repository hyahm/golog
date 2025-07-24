package golog

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	"github.com/fatih/color"
)

func (lm *msgLog) control() {
	// format = printFileline() + format // printfileline()打印出错误的文件和行数
	// 判断是输出控制台 还是写入文件
	if lm.Level <= ERROR && lm.ErrorHandler != nil {
		lm.ErrorHandler(lm.Ctime, lm.Hostname, lm.Line, lm.Msg, lm.Label)
	}
	if lm.Level <= INFO && lm.InfoHandler != nil {
		lm.InfoHandler(lm.Ctime, lm.Hostname, lm.Line, lm.Msg, lm.Label)
	}
	if lm.Level <= WARN && lm.WarnHandler != nil {
		lm.WarnHandler(lm.Ctime, lm.Hostname, lm.Line, lm.Msg, lm.Label)
	}
	if lm.out {
		// 如果是输出到控制台，直接执行就好了
		lm.printLine()
		return
	} else {
		// 写入文件
		if lm.size > 0 {
			f, err := os.OpenFile(lm.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				defer f.Close()
				// 如果大于设定值， 那么
				fi, err := f.Stat()
				if err == nil && fi.Size() >= lm.size*1024 {
					err = os.Rename(lm.filepath, filepath.Join(lm.dir, fmt.Sprintf("%s_%s", lm.create.Format("2006-01-02_15_04_05"), lm.name)))
					if err != nil {
						log.Println(err)
						lm.out = true
						return
					}

				}
			}

		}
		// size 大小 分割优先
		if lm.size == 0 && lm.everyDay {

			// 不存在就移动创建
			if lm.create.Format("20060102") != time.Now().Format("20060102") {
				oldfile := filepath.Join(lm.dir, lm.create.Format("2006-01-02")+"_"+lm.name)
				// 如果每天备份的话， 文件名需要更新
				// 重命名
				_, err := os.Stat(oldfile)
				if err != nil {
					// 如果
					if err := os.Rename(lm.filepath, filepath.Join(lm.dir, oldfile)); err != nil {
						log.Println(err)
						lm.out = true
					}

				}
				f, err := os.OpenFile(oldfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					// 如果失败，切换到控制台输出
					lm.out = true
					lm.printLine()
					return
				}
				defer f.Close()
				buf, err := lm.formatText()
				if err != nil {
					return
				}
				// buf.WriteString("\n")
				// logMsg := fmt.Sprintf("%s - [%s] - %s - %s - %s - %v\n", lm.Ctime, lm.Level, lm.Prev, lm.Hostname, lm.Line, lm.Msg)
				f.Write([]byte(buf.Bytes()))
				return
			}
		}
		// 如果按照文件大小判断的话，名字不变
		lm.writeToFile()

	}
}

func (lm *msgLog) writeToFile() {
	//
	//不存在就新建
	f, err := os.OpenFile(lm.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// 如果失败，切换到控制台输出
		lm.out = true
		lm.printLine()
		return
	}
	defer f.Close()
	buf, err := lm.formatText()
	if err != nil {
		return
	}
	// buf.WriteString("\n")
	// logMsg := fmt.Sprintf("%s - [%s] - %s - %s - %s - %v\n", lm.Ctime, lm.Level, lm.Prev, lm.Hostname, lm.Line, lm.Msg)
	f.Write([]byte(buf.Bytes()))

}

func (lm *msgLog) printLine() {
	buf, err := lm.formatText()
	if err != nil {
		return
	}
	color.New(lm.Color...).Print(buf.String())
	// color.New(lm.Color...).Printf("%s - [%s] - %s - %s - %s - %v\n", lm.Ctime, lm.Level, lm.Prev, lm.Hostname, lm.Line, lm.Msg)
}

func (lm *msgLog) formatText() (*bytes.Buffer, error) {
	tml, err := template.New(lm.name).Parse(lm.format)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var b []byte
	buf := bytes.NewBuffer(b)
	err = tml.Execute(buf, lm)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return buf, nil
}

func printFileline(c int) string {
	c += 3
	_, file, line, ok := runtime.Caller(c)
	if !ok {
		file = "???"
		line = 0
	}
	return fmt.Sprintf("%s:%d", file, line)
}

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
