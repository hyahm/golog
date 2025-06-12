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

	"github.com/fatih/color"
)

var nowday string

func (lm *msgLog) control() {
	// format = printFileline() + format // printfileline()打印出错误的文件和行数
	// 判断是输出控制台 还是写入文件
	if lm.out {
		// 如果是输出到控制台，直接执行就好了
		lm.printLine()
		return
	} else {
		// 写入文件
		if everyDay {
			// 如果每天备份的话， 文件名需要更新
			thisDay := fmt.Sprintf("%d-%d-%d", lm.create.Year(), lm.create.Month(), lm.create.Day())
			if nowday == "" {
				nowday = thisDay
			}
			if thisDay != nowday {
				// 重命名
				if err := os.Rename(lm.logPath, filepath.Join(lm.path, nowday+"_"+lm.name)); err != nil {
					log.Println(err)
					lm.out = true
					return
				}
				nowday = thisDay
			}

		}
		if lm.size > 0 {
			f, err := os.Open(lm.logPath)
			if err == nil {
				// 如果大于设定值， 那么
				fi, err := f.Stat()
				if err == nil && fi.Size() >= lm.size*1024 {
					f.Close()
					err = os.Rename(lm.logPath, filepath.Join(lm.path, fmt.Sprintf("%d_%s", lm.create.Unix(), lm.name)))
					if err != nil {
						log.Println(err)
					}

				}
				f.Close()
			}

		}
		// 如果按照文件大小判断的话，名字不变
		lm.writeToFile()

	}
}

func (lm *msgLog) writeToFile() {
	//
	//不存在就新建
	f, err := os.OpenFile(lm.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// 如果失败，切换到控制台输出
		color.Red("Permission denied,  auto change to Stdout")
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
