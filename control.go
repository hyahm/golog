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

const BLOCKSIZE = 4 << 10

func (cl msgLog) control() {
	// format = printFileline() + format // printfileline()打印出错误的文件和行数
	// 判断是输出控制台 还是写入文件

	if cl.out {
		// 如果是输出到控制台，直接执行就好了
		cl.printLine()
		return
	}
	// // 写入文件
	if cl.size > 0 {
		f, err := os.OpenFile(filepath.Join(cl.dir, cl.name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		// 如果大于设定值， 那么
		fi, err := f.Stat()
		if err != nil {
			fmt.Println(err)
			return
		}
		if fi.Size() >= (cl.size-1)*1<<20-BLOCKSIZE {

			_, err := f.WriteString(cl.Msg)
			if err != nil {
				fmt.Println(err)
				return
			}
			f.Close()

			err = os.Rename(filepath.Join(cl.dir, cl.name), filepath.Join(cl.dir, fmt.Sprintf("%s_%s", cl.Ctime.Format("2006-01-02_15_04_05"), cl.name)))
			if err != nil {
				log.Println(err)
				// _dir = ""
				return
			}

		} else {
			_, err := f.WriteString(cl.Msg)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		return
	}
	// size 大小 分割优先
	if cl.size == 0 && cl.everyDay {

		// 不存在就移动创建
		if cl.Ctime.Format("20060102") != time.Now().Format("20060102") {
			oldfile := filepath.Join(cl.dir, cl.Ctime.Format("2006-01-02")+"_"+cl.name)
			// 如果每天备份的话， 文件名需要更新
			// 重命名
			_, err := os.Stat(oldfile)
			if err != nil {
				// 如果

				if err := os.Rename(filepath.Join(cl.dir, cl.name), filepath.Join(cl.dir, oldfile)); err != nil {
					log.Println(err)
					cl.out = true
				}

			}
			f, err := os.OpenFile(oldfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				// 如果失败，切换到控制台输出
				cl.out = true
				cl.printLine()
				return
			}
			defer f.Close()
			f.WriteString(cl.Msg)
			return
		}
	}
	// 如果按照文件大小判断的话，名字不变
	cl.writeToFile()

}

func (task *task) control(cl msgLog) {
	// format = printFileline() + format // printfileline()打印出错误的文件和行数
	// 判断是输出控制台 还是写入文件

	if cl.out {
		// 如果是输出到控制台，直接执行就好了
		cl.printLine()
		return
	}
	// // 写入文件
	if cl.size > 0 {
		f, err := os.OpenFile(filepath.Join(cl.dir, cl.name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		// 如果大于设定值， 那么
		fi, err := f.Stat()
		if err != nil {
			fmt.Println(err)
			return
		}
		if fi.Size() >= (cl.size-1)*1<<20-BLOCKSIZE {

			_, err := f.WriteString(cl.Msg)
			if err != nil {
				fmt.Println(err)
				return
			}
			f.Close()

			err = os.Rename(filepath.Join(cl.dir, cl.name), filepath.Join(cl.dir, fmt.Sprintf("%s_%s", cl.Ctime.Format("2006-01-02_15_04_05"), cl.name)))
			if err != nil {
				log.Println(err)
				// _dir = ""
				return
			}

		} else {
			_, err := f.WriteString(cl.Msg)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		return
	}
	// size 大小 分割优先
	if cl.size == 0 && cl.everyDay {

		// 不存在就移动创建
		if cl.Ctime.Format("20060102") != time.Now().Format("20060102") {
			oldfile := filepath.Join(cl.dir, cl.Ctime.Format("2006-01-02")+"_"+cl.name)
			// 如果每天备份的话， 文件名需要更新
			// 重命名
			_, err := os.Stat(oldfile)
			if err != nil {
				// 如果

				if err := os.Rename(filepath.Join(cl.dir, cl.name), filepath.Join(cl.dir, oldfile)); err != nil {
					log.Println(err)
					cl.out = true
				}

			}
			f, err := os.OpenFile(oldfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				// 如果失败，切换到控制台输出
				cl.out = true
				cl.printLine()
				return
			}
			defer f.Close()
			f.WriteString(cl.Msg)
			return
		}
	}
	// 如果按照文件大小判断的话，名字不变
	cl.writeToFile()

}

func (lm *msgLog) writeToFile() {
	//
	//不存在就新建
	f, err := os.OpenFile(filepath.Join(lm.dir, lm.name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// 如果失败，切换到控制台输出
		lm.out = true
		lm.printLine()
		return
	}
	defer f.Close()
	f.Write([]byte(lm.Msg))

}

func (lm *msgLog) printLine() {
	color.New(lm.Color...).Print(lm.Msg)

}

var tml *template.Template

func init() {
	var err error
	tml, err = template.New("golog").Parse(Format)
	if err != nil {
		panic(err)
	}
}

func (lm *msgLog) formatText() (*bytes.Buffer, error) {

	buf := bytes.NewBuffer(nil)
	err := tml.Execute(buf, lm)
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
