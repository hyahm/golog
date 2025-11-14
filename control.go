package golog

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fatih/color"
)

const BLOCKSIZE = 4 << 10

// func (cl msgLog) control() {
// 	// format = printFileline() + format // printfileline()打印出错误的文件和行数
// 	// 判断是输出控制台 还是写入文件

// 	if cl.out {
// 		// 如果是输出到控制台，直接执行就好了
// 		cl.printLine()
// 		return
// 	}
// 	// // 写入文件
// 	if cl.size > 0 {
// 		f, err := os.OpenFile(filepath.Join(cl.dir, cl.name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		defer f.Close()
// 		// 如果大于设定值， 那么
// 		fi, err := f.Stat()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		if fi.Size() >= (cl.size-1)*1<<20-BLOCKSIZE {

// 			_, err := f.WriteString(cl.Msg)
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			f.Close()

// 			err = os.Rename(filepath.Join(cl.dir, cl.name), filepath.Join(cl.dir, fmt.Sprintf("%s_%s", cl.Ctime.Format("2006-01-02_15_04_05"), cl.name)))
// 			if err != nil {
// 				log.Println(err)
// 				// _dir = ""
// 				return
// 			}

// 		} else {
// 			_, err := f.WriteString(cl.Msg)
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 		}
// 		return
// 	}
// 	// size 大小 分割优先
// 	if cl.size == 0 && cl.everyDay {

// 		// 不存在就移动创建
// 		if cl.Ctime.Format("20060102") != time.Now().Format("20060102") {
// 			oldfile := filepath.Join(cl.dir, cl.Ctime.Format("2006-01-02")+"_"+cl.name)
// 			// 如果每天备份的话， 文件名需要更新
// 			// 重命名
// 			_, err := os.Stat(oldfile)
// 			if err != nil {
// 				// 如果

// 				if err := os.Rename(filepath.Join(cl.dir, cl.name), filepath.Join(cl.dir, oldfile)); err != nil {
// 					log.Println(err)
// 					cl.out = true
// 				}

// 			}
// 			f, err := os.OpenFile(oldfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
// 			if err != nil {
// 				// 如果失败，切换到控制台输出
// 				cl.out = true
// 				cl.printLine()
// 				return
// 			}
// 			defer f.Close()
// 			f.WriteString(cl.Msg)
// 			return
// 		}
// 	}
// 	// 如果按照文件大小判断的话，名字不变
// 	cl.writeToFile()

// }

// 记录当前的日志偏移时间

func (task *task) control(cl msgLog) {
	// format = printFileline() + format // printfileline()打印出错误的文件和行数
	// 判断是输出控制台 还是写入文件
	if cl.out {
		// 如果是输出到控制台，直接执行就好了
		cl.printLine()
		return
	}
	// 写入文件
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
		// 当前日志的时间不是今天，需要切割，
		//  如果这个日志的时间不是今天， 那么应该写入到昨天的文件
		//   如果昨天的文件不存在并且今天的日志的更新时间是昨天的， 那么就重命名文件为昨天的日志， 然后将昨天的日志内容写入到昨天的日志文件
		// 今天的日志文件还是写入到今天的文件里面
		fi, err := os.Stat(filepath.Join(cl.dir, cl.name))
		var oldfile string
		if err == nil {
			// 用文件实际修改时间作为旧文件的日期（比创建时间更可靠）
			oldfile = fi.ModTime().Format("2006-01-02") + "_" + cl.name
		} else {
			// 若文件不存在，用创建时间作为旧文件日期（兜底）
			oldfile = cl.Ctime.Format("2006-01-02") + "_" + cl.name
		}
		oldfilePath := filepath.Join(cl.dir, oldfile)
		currentFilePath := filepath.Join(cl.dir, cl.name)

		// 核心逻辑：判断当前文件是否需要切割
		if err == nil { // 当前日志文件存在
			// 判断文件的修改时间是不是今天， 如果不是今天的， 就移动文件, 主要是为了移动文件，不做处理
			if fi.ModTime().Day() != time.Now().Day() {

				if err := os.Rename(currentFilePath, oldfilePath); err != nil {
					log.Println("重命名旧日志失败:", err)
					cl.out = true
					// return // 重命名失败，避免后续错误写入
				}
				// 切割后，创建新的今日日志文件并写入
				// 	cl.writeToFile(cl.name)
				// } else {
				// 1.2 旧文件已存在：直接向旧文件追加（避免覆盖已有历史日志）
				// cl.writeToFile(oldfile)
				// return
				// }
			}
		}
		if cl.Ctime.Day() == time.Now().Day() {
			cl.writeToFile(cl.name)
		} else {
			cl.writeToFile(cl.Ctime.Format("2006-01-02") + "_" + cl.name)
		}
		return
	}
	// 如果是空的如果按照文件大小判断的话，名字不变
	cl.writeToFile(cl.name)

}

func (lm *msgLog) writeToFile(filename string) {
	//
	//不存在就新建
	f, err := os.OpenFile(filepath.Join(lm.dir, filename), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
	// if !isatty.IsTerminal(os.Stdout.Fd()) {
	// 	panic("not in terminal")
	// }

	// 1. 获取控制台输出句柄
	// h := syscall.Handle(os.Stdout.Fd())

	// 2. 设置蓝底白字
	// syscall.set
	// syscall.SetConsoleTextAttribute(h, backBlue|foreWhite)

	// 3. 直接写 string —— 零逃逸

	// 4. 还原默认色（白字黑底）
	// syscall.SetConsoleTextAttribute(h, foreWhite)
	// ① 先避开接口
	// buf := []byte(lm.Msg)
	// // if len(lm.Color) > 0 {
	// color.New(color.BgBlue).Fprint(colorable.NewColorable(os.Stdout), buf)
	// 	return
	// }
	color.New(lm.Color...).Print(lm.Msg)
	// fmt.Print(lm.Msg)
}

// var tml *template.Template

// func init() {
// 	var err error
// 	// tml, err = template.New("golog").Parse(Format)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// }

func FormatLog(ml *msgLog) (string, error) {
	return "", nil
}

// func (lm *msgLog) formatText() (string, error) {

// 	buf := bytes.NewBuffer(nil)
// 	err := tml.Execute(buf, lm)
// 	if err != nil {
// 		return "", err
// 	}
// 	return buf.String(), nil
// }

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
