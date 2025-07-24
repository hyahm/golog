# golog  simple easy log library

异步简单易用的日志库, 全程不需要关闭操作, 开箱即用

### 安装
```
 go get github.com/hyahm/golog@main
```

### 日志格式化(text/template模板语法)
> 通过 golog.Format 设置输出格式，默认的输出格式如下

`{{ .Ctime }} - [{{ .Level }}]{{ if .Label }} - {{ range $k,$v := .Label}}[{{$k}}:{{$v}}]{{end}}{{end}} - {{.Hostname}} - {{.Line}} - {{.Msg}}`

### 最简单的同步打印控制台
```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	// 这一行主要是防止退出时日志没有写完， 导致看不到日志，  如果对日志要求没那么高的话， 可以不加上这条
	defer golog.Sync()
	golog.Info("one") // stdout: 2022-03-04 10:19:31 - [INFO] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:9 - one
	golog.Info("adf", "cander") // stdout: 2022-03-04 10:19:31 - [INFO] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:9 - adf cander
	golog.ShowBasePath = true
	golog.Info("adf", "cander") // stdout: 2022-03-04 10:19:31 - [INFO] - DESKTOP-NENB5CA - example.go:9 - adf cander
}
```


### 格式化打印

```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	// 虽然是可视化输出，
	golog.Infof("adf%s\n", "cander") // stdout: 2022-03-04 10:19:31 - [INFO] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:11 - adfcander
	// 默认的日志级别是info， 所以debug级别不会打印出来,
	golog.Debug("foo") // stdout: nothing
	// 通过 golog.Level = golog.DEBUG 可以设置级别为DEBUG
	golog.Level = golog.DEBUG //
	golog.Debug("bar")        // stdout: 2022-03-04 10:21:00 - [DEBUG] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:14 - bar
}
```

### 按照日志级别打印

```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	// 默认的日志级别是info， 所以debug级别不会打印出来,
	golog.Debug("foo") // stdout: nothing
	// 通过 golog.Level = golog.DEBUG 可以设置级别为DEBUG
	golog.Level = golog.DEBUG //
	golog.Debug("bar")        // stdout: 2022-03-04 10:21:00 - [DEBUG] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:14 - bar
}
```

```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	// All level = iota * 10
	// TRACE
	// DEBUG
	// INFO  默认info级别
	// WARN
	// ERROR
	// FATAL

	defer golog.Sync()
	// 虽然是可视化输出， 但是不需要增加\n换行
	golog.Debug("foo") // stdout: nothing
	// 通过 golog.Level = golog.DEBUG 可以设置级别为DEBUG
	golog.Level = golog.DEBUG //
	
	golog.Debug("bar")        // stdout: 2022-03-04 10:21:00 - [DEBUG] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:14 - bar
	golog.ShowBasePath = true  // 显示基本文件，而不是完整路径
	golog.Debug("baz")         // stdout: 2022-03-04 10:21:00 - [DEBUG] - DESKTOP-NENB5CA - example.go:14 - bar
}
```

### 日志颜色设置(控制台打印才有效， 写入文件无效)
```go
package main

import (
	"github.com/fatih/color"
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	infoColor := make([]color.Attribute, 0)
	infoColor = append(infoColor, color.FgBlue, color.BgGreen) // 文字颜色蓝色， 背景色绿色
	golog.SetColor(golog.INFO, infoColor)                      // 设置为info级别的日志颜色
	golog.Infof("adf%s", "cander")                             // stdout: 2022-03-04 10:19:31 - [INFO] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:9 - adfcander
}
```

### 日志写入文件
```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	// 第一个参数是设置日志目录 ， 如果为空，默认显示再控制台, 强烈建议单独放在一个目录里面， 不然第四个参数有可能误删除文件
	// 第二个参数是设置日志切割的大小，0 表示不按照大小切割， 默认单位M，
	//  第三个事是否每天切割，
	// 第四个是删除多少天以前的日志，当前设置的是7天， 根据设置的name 来匹配， 不写或者0表示不删除
	// 没错，写入文件就是只需要增加这一行即可
	golog.InitLogger("log/test.log", 0, true)
	
	golog.Infof("adf%s", "cander") 
	// log/test.log: 2022-03-04 10:19:31 - [INFO] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:9 - adfcander
}

```

### 日志文件自动清除
```go
package main

import (
	"time"

	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	// 设置清除7小时以前的日志
	golog.DefaultUnit = golog.Hour  // 设置单为小时
	golog.InitLogger("log/test.log", 0, true, 7)
	golog.Infof("adf%s", "cander")
	// log/test.log: 2022-03-04 10:19:31 - [INFO] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:13 - adfcander
}
```

### 增加Label标签
```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	golog.AddLabel("key1", "value1")
	golog.AddLabel("key2", "value2")
	golog.Info("foo") // stdout: 2022-03-04 10:32:51 - [INFO] - [key1:value1][key2:value2] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:11 - foo
}
```

### 多文件操作
```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	logger1 := golog.NewLog("log/test1.log", 0, false) // 这是操作log/test1.log的实例， 用法与golog的方法使用一致
	logger2 := golog.NewLog("log/test2.log", 0, false) // 这是操作log/test2.log的实例， 用法与golog的方法使用一致
	logger1.Info("foo")
	logger2.Info("foo")
	// 如果这些日志实例在服务器运行中可能会停止，则必须在此日志服务停止时必须关闭
}
```

### 增加ErrorHandler , InfoHandler的回调函数，  方便报警, 只有在调用golog.Error\[f\]()的时候才会调用
```go
	// 为什么只有info 和 error  ， 因为只有这2个开发最常用   warn 和 debug 建议做本地调试使用， 如果要更新细致的处理，建议搭配 label
	defer golog.Sync()
	
	golog.ErrorHandler = func(ctime, hostname, line, msg string, label map[string]string) {
		// 可以自定义报警信息， 方便及时知道运行中代码内的错误
		fmt.Println("你的代码出问题了")
	}
	golog.ErrorHandler = func(ctime, hostname, line, msg string, label map[string]string) {
		// 可以对info 信息做处理，
		fmt.Println("你的代码出问题了")
	}
	golog.Error("aaaaaa")
```


### 接口方法调试， 可以知道是那一行调用了这个方法
```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	defer golog.Sync()
	golog.Info("foo")
	golog.Level = golog.DEBUG
	test()
	golog.Info("bar")
}

func test() {
	// 此方法的日志级别是DEBUG， 所以调试的时候必须将日志级别设置成DEBUG，不然不会显示
	golog.UpFunc(1, "who call me") // 2022-03-04 10:49:38 - [DEBUG] - DESKTOP-NENB5CA - C:/work/golog/example/example.go:17 - caller from C:/work/golog/example/example.go:11 -- who call me
}
```


