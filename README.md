# golog simple easy log library

异步简单易用的日志库, 全程不需要关闭操作, 开箱即用
go version >= 1.25.0

### 安装
```
 go get github.com/hyahm/golog@main
```

### 最简单的同步打印控制台
```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	// 这一行主要是防止退出时日志没有写完， 导致看不到日志，  如果对日志要求没那么高的话， 可以不加上这条
	defer golog.Sync()

	golog.Info("one")           // stdout: 2022-03-04 10:19:31.000 -- [INFO] -- C:/work/golog/example/example.go:9 -- one
	golog.Info("adf", "cander") // stdout: ... -- adf cander
}
```

### 按照日志级别打印
```go
// ALL / TRACE / DEBUG / INFO(默认) / WARN / ERROR / FATAL / PANIC
golog.Debug("foo") // 默认 info 级别, debug 不打印
golog.SetLevel(golog.DEBUG)
golog.Debug("bar")
golog.ShowBasePath = true // 只显示文件名而不是完整路径（应在首次打日志前设置）
```

### 日志颜色设置(控制台打印才有效， 写入文件无效)
```go
infoColor := []color.Attribute{color.FgBlue, color.BgGreen}
golog.SetColor(golog.INFO, infoColor)
```

### 日志写入文件
```go
golog.SetDir("log")                    // 默认 log 目录
golog.InitLogger("test.log", 0, true)  // name, 按大小切割(MB, 0 不切割), 按天切割
golog.Infof("adf%s", "cander")
```

### 归档压缩 / 过期清理
```go
golog.SetCompress(true)                  // 切割归档后的旧文件异步 gzip 压缩（.gz）
golog.SetExpireDuration(time.Hour * 24 * 7) // 过期清理，默认 7 天，过期文件（含 .gz）自动删除
```

### 写文件的同时输出控制台（多路输出）
```go
golog.SetConsole(true) // 写文件的同时打印到控制台（带颜色）
```

### 结构化字段（logfmt / JSON 均支持）
```go
golog.Infow("user login", golog.Str("user", "tom"), golog.Int("age", 18))
// 2022-03-04 10:19:31.000 -- [INFO] -- xx.go:9 -- user login user=tom age=18

// With 返回携带固定字段的子 logger，可链式叠加
orderLog := golog.With(golog.Str("svc", "order"))
orderLog.Infow("create order", golog.Int64("oid", 1001))
```

内置字段构造器：`Any / Str / Int / Int64 / Bool / Float64 / Duration / Err`

### 模块命名子 logger
```go
dbLog := golog.Named("db")
dbLog.Info("query") // ... -- query logger=db
```

### context 透传（trace_id 等）
```go
ctx := golog.ContextWithTraceID(ctx, "t-abc-123")
ctx = golog.ContextWithFields(ctx, golog.Str("uid", "u1"))

golog.Ctx(ctx).Info("handle request") // 自动携带 uid=u1 trace_id=t-abc-123
```

### log/slog 适配器
```go
// 所有依赖标准库 slog 的库（GORM 等）可无缝走 golog
logger := slog.New(golog.NewSlogHandler(nil)) // nil 使用全局 logger
logger.Info("hello", "uid", 42)
```

### 自定义格式化
```go
// 格式化函数签名（fields 为结构化字段）
golog.SetFormatFunc(func(level golog.Level, ctime time.Time, line, msg string, fields []golog.Field) string {
	return fmt.Sprintf("%s [%s] %s %s %v\n", ctime.Format(time.RFC3339), level, line, msg, fields)
})

// 内置 JSON 格式：自动转义特殊字符，字段合并为 JSON 顶层键
golog.SetFormatFunc(golog.JsonFormat)
```

### 时间格式与时区
```go
golog.SetTimeLayout("2006-01-02 15:04:05.000") // 默认布局
golog.SetUTC(true)                             // 默认本地时区
```

### 限流采样（日志洪峰保护）
```go
golog.SetRateLimit(100, time.Second) // 每秒最多记录 100 条，超出丢弃
```

### 重复日志采样（借鉴 zap）
```go
golog.SetLogPriority(true, 100, time.Minute) // 每分钟 100 条重复日志只打印一条
```

### Fatal / Panic / Stack
```go
// Fatal/Fatalf/Fatalw: 同步写盘（保证落盘）后 os.Exit(1)
golog.Fatalf("bad thing: %v", err)

// Panic/Panicf: 同步写盘后 panic
golog.Panic("unexpected state")

// Stack: ERROR 级别附带当前 goroutine 堆栈
golog.Stack("here")
```

### 多文件操作
```go
logger1 := golog.NewLog("test1.log", 0, false) // 独立实例，用法与全局方法一致
defer logger1.Sync()                           // 实例停止服务时必须调用

logger2 := golog.NewLog("test2.log", 0, false)
defer logger2.Sync()
logger1.Info("foo")
logger2.Info("foo")
```

### 自定义输出目标（可测试性）
```go
var buf bytes.Buffer
golog.SetOutput(&buf) // 捕获控制台输出，用于单测断言或对接 syslog/kafka 等
```

### 回调函数
```go
// 每条日志异步回调，可用于报警等二次处理（应在首次打日志前设置）
golog.LogHandler = func(level golog.Level, ctime time.Time, line, msg string) {
	fmt.Println("你的代码出问题了")
}
```

### 接口方法调试， 可以知道是那一行调用了这个方法
```go
golog.SetLevel(golog.DEBUG)
func test() {
	golog.UpFunc(1, "who call me") // ... -- caller from example.go:11 -- who call me
}
```

### 错误日志源头追踪
```go
// 同步函数，包装 error 并记录来源
golog.Wrap(err) error
golog.Wraps("msg") error
golog.Unwrap(err) error // 解出原始 error，非 golog 包装时返回 nil

func (p project) DeleteProject(projectId int64) error {
	return golog.Wrap(db.Gorm.Table(p.table).Where("pid=?", projectId).Delete(nil).Error)
}

// 最外层打印自动追加来源路径
// D:/cander/ITflow/go/app/repo/bugs.go:23 -- Error 1054 (42S22): Unknown column 'pid' in 'where clause'
```

### 溢出保护
异步通道写满时日志会被丢弃（不阻塞业务），并通过控制台周期性输出
`[golog] N log messages dropped (async buffer full)` 告警。
