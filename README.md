# golog

异步、简单、易用的 Go 日志库，开箱即用，全程无需关心关闭操作。

- go version >= 1.25.0
- 零配置可直接使用，无任何强制初始化步骤
- 异步写盘（不阻塞业务），`Fatal/Panic` 同步落盘保证可见
- 支持文件切割（按大小 / 按天）、归档 gzip 压缩、过期自动清理
- 支持结构化字段（logfmt / JSON）、`context` 透传、`log/slog` 适配
- 内置限流采样与重复日志采样，日志洪峰保护
- 溢出保护：缓冲写满时丢弃日志而非阻塞业务

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [使用方法](#使用方法)
  - [初始化顺序与注意事项](#初始化顺序与注意事项)
  - [场景一开发环境纯控制台输出](#场景一开发环境纯控制台输出)
  - [场景二生产环境文件切割压缩清理](#场景二生产环境文件切割压缩清理)
  - [场景三Web-服务中配合-context-使用](#场景三web-服务中配合-context-使用)
  - [场景四对接-logslog-生态gorm-等](#场景四对接-logslog-生态gorm-等)
  - [程序退出注意事项](#程序退出注意事项)
- [日志级别](#日志级别)
- [日志颜色设置](#日志颜色设置控制台打印才有效写入文件无效)
- [日志写入文件](#日志写入文件)
- [归档压缩--过期清理](#归档压缩--过期清理)
- [多路输出](#多路输出)
- [结构化字段logfmt--json-均支持](#结构化字段logfmt--json-均支持)
- [模块命名子-logger](#模块命名子-logger)
- [context-透传trace_id-等](#context-透传trace_id-等)
- [logslog-适配器](#logslog-适配器)
- [自定义格式化](#自定义格式化)
- [时间格式与时区](#时间格式与时区)
- [限流采样日志洪峰保护](#限流采样日志洪峰保护)
- [重复日志采样借鉴-zap](#重复日志采样借鉴-zap)
- [fatal--panic--stack](#fatal--panic--stack)
- [多文件操作](#多文件操作)
- [自定义输出目标可测试性](#自定义输出目标可测试性)
- [回调函数](#回调函数)
- [接口方法调试](#接口方法调试)
- [错误日志源头追踪](#错误日志源头追踪)
- [溢出保护](#溢出保护)
- [默认值一览](#默认值一览)
- [异步架构说明](#异步架构说明)

## 安装

```
go get github.com/hyahm/golog@main
```

## 快速开始

最简单的同步打印控制台：

```go
package main

import (
	"github.com/hyahm/golog"
)

func main() {
	// 这一行主要是防止退出时日志没有写完，导致看不到日志。如果对日志要求没那么高，可以不加上这条
	defer golog.Sync()

	golog.Info("one")           // stdout: 2022-03-04 10:19:31.000 -- [INFO] -- C:/work/golog/example/example.go:9 -- one
	golog.Info("adf", "cander") // stdout: ... -- adf cander
}
```

## 使用方法

### 初始化顺序与注意事项

大部分配置项可任意时机调用（运行中修改并发安全），但以下几点有时序要求：

| 配置项 | 时机要求 |
| --- | --- |
| `SetExpireDuration` | 必须在 `InitLogger` / `NewLog` 之前调用 |
| `ShowBasePath`、`LogHandler` | 应在首次打印日志前设置 |
| `SetDir` | 应在 `InitLogger` / `NewLog` 之前调用才能生效到文件路径 |
| 其余 `SetLevel` / `SetConsole` 等 | 任意时机，运行中修改并发安全 |

### 场景一：开发环境（纯控制台输出）

零配置即可使用，只需要退出前刷盘：

```go
package main

import "github.com/hyahm/golog"

func main() {
	golog.SetLevel(golog.DEBUG) // 开发期放开 debug 日志
	defer golog.Sync()           // 防止退出时日志未写完

	golog.Debugf("debug message")
	golog.Infof("info message")
}
```

### 场景二：生产环境（文件+切割+压缩+清理）

典型的生产配置，全部在 `main` 函数开头完成：

```go
package main

import (
	"time"

	"github.com/hyahm/golog"
)

func main() {
	// ---- 初始化日志（注意：SetExpireDuration 须在 InitLogger 之前）----
	golog.ShowBasePath = true                    // 日志只显示文件名，不暴露部署路径
	golog.SetDir("/var/log/myapp")               // 日志目录，默认 log
	golog.SetExpireDuration(time.Hour * 24 * 30)  // 保留 30 天，默认 7 天
	golog.SetConsole(false)                      // 生产可不输出控制台
	golog.SetCompress(true)                      // 归档文件异步 gzip 压缩
	golog.SetLogPriority(true, 100, time.Minute) // 重复日志采样，防止刷屏
	golog.SetRateLimit(10000, time.Second)       // 限流采样，洪峰保护
	golog.InitLogger("app.log", 100, true)       // 写文件：超 100MB 切割 + 按天切割
	defer golog.Sync()                           // 退出前保证落盘

	// ---- 业务代码，直接调用全局方法 --------
	golog.Info("service started")
	if err := doSomething(); err != nil {
		golog.Errorf("doSomething failed: %v", err)
	}
}

func doSomething() error {
	// Wrap 包装错误并记录来源位置，最外层打印时自动追加文件行号
	return golog.Wraps("something wrong")
}
```

运行后目录结构：

```
/var/log/myapp/
├── app.log            # 当前日志
├── 2026-09-21_app.log # 按天归档
└── 2026-09-20_app.log.gz # 过期前异步压缩的归档
```

### 场景三：Web 服务中配合 context 使用

中间件注入 `trace_id`，业务日志自动携带，排查问题时按 `trace_id` 串联整条链路：

```go
package main

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hyahm/golog"
)

var requestID atomic.Int64

func traceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 生成 trace_id 注入 context
		id := time.Now().UnixNano() + requestID.Add(1)
		ctx := golog.ContextWithTraceID(r.Context(), strconv.FormatInt(id, 10))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Ctx(ctx) 返回绑定 context 的子 logger，自动携带 trace_id
	golog.Ctx(ctx).Infow("request in", golog.Str("path", r.URL.Path))

	// 也可以注入业务字段后向下传递
	ctx = golog.ContextWithFields(ctx, golog.Int64("uid", 1001))
	handleOrder(ctx, 20001)

	w.Write([]byte("ok"))
}

func handleOrder(ctx context.Context, orderID int64) {
	// 此处所有通过 Ctx(ctx) 打的日志都携带 uid + trace_id
	golog.Ctx(ctx).Infow("order created", golog.Int64("order_id", orderID))
}

func main() {
	golog.SetDir("log")
	golog.InitLogger("web.log", 50, true)
	golog.SetConsole(true) // 本地调试时文件+控制台双写
	defer golog.Sync()

	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8080", traceMiddleware(mux))
}
```

输出效果（文件中一行）：

```
2026-09-21 10:00:00.000 -- [INFO] -- home.go:31 -- order created uid=1001 trace_id=xxx order_id=20001
```

### 场景四：对接 log/slog 生态（GORM 等）

依赖标准库 `log/slog` 的第三方库无需改造，直接接管其日志输出：

```go
package main

import (
	"log/slog"

	"github.com/hyahm/golog"
)

func main() {
	golog.InitLogger("db.log", 50, true)
	defer golog.Sync()

	// golog 适配器作为 slog.Handler，传 nil 使用全局 logger
	// 也可以传入子 logger：golog.NewSlogHandler(golog.Named("gorm"))
	slogLogger := slog.New(golog.NewSlogHandler(nil))

	// 任何接受 slog.Logger 的库（GORM slog 模式、第三方 SDK 等）直接接入
	slogLogger.Info("query executed", "rows", 42, "cost_ms", 3)
	// 级别映射：slog Error/Warn/Info -> golog ERROR/WARN/INFO，其余 -> DEBUG
}
```

### 程序退出注意事项

- `defer golog.Sync()`：等待异步日志全部落盘，可重复调用；纯控制台输出模式日志即时可见，可不调用
- `NewLog` 创建的独立实例：实例停止服务（而非进程退出）时必须调用其 `Sync()` 停止内部协程
- `Fatal` 系列自带同步落盘并 `os.Exit(1)`，无需额外处理
- 异常 panic 场景可配合 `recover` 使用 `golog.Stack()` 保留现场堆栈：

```go
func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				golog.Stack(r) // ERROR 级别附带 goroutine 堆栈
			}
		}()
		fn()
	}()
}
```

## 日志级别

```go
// ALL / TRACE / DEBUG / INFO(默认) / WARN / ERROR / FATAL / PANIC
golog.Debug("foo")            // 默认 info 级别, debug 不打印
golog.SetLevel(golog.DEBUG)   // 运行时可安全切换
golog.Debug("bar")
golog.Info(golog.GetLevel())  // 获取当前级别

golog.ShowBasePath = true // 只显示文件名而不是完整路径（应在首次打日志前设置）
```

每个级别均有三种形式：

```go
golog.Trace / Tracef / Tracew
golog.Debug / Debugf / Debugw
golog.Info  / Infof  / Infow
golog.Warn  / Warnf  / Warnw
golog.Error / Errorf / Errorw
golog.Fatal / Fatalf / Fatalw  // 同步写盘（保证落盘）后 os.Exit(1)
golog.Panic / Panicf           // 同步写盘后 panic
```

## 日志颜色设置(控制台打印才有效， 写入文件无效)

```go
infoColor := []color.Attribute{color.FgBlue, color.BgGreen}
golog.SetColor(golog.INFO, infoColor)
```

依赖 `github.com/fatih/color`。默认颜色：ERROR/FATAL 红色、WARN 黄色、DEBUG 绿色、PANIC 洋红色，其余无颜色。

## 日志写入文件

```go
golog.SetDir("log")                    // 默认 log 目录
golog.InitLogger("test.log", 0, true)  // name, 按大小切割(MB, 0 不切割), 按天切割
golog.Infof("adf%s", "cander")
```

说明：

- `name` 传空字符串则只输出控制台
- 按大小切割优先于按天切割；归档文件名为 `时间戳_name` / `日期_name`，重名自动追加 `_1/_2`
- `SetDir` 失败（无法创建目录）时回退到当前目录

## 归档压缩 / 过期清理

```go
golog.SetCompress(true)                      // 切割归档后的旧文件异步 gzip 压缩（.gz）
golog.SetExpireDuration(time.Hour * 24 * 7)  // 过期清理，默认 7 天，过期文件（含 .gz）自动删除
```

过期清理每 24 小时执行一次，需在 `InitLogger` 之前设置。

## 多路输出

```go
golog.SetConsole(true) // 写文件的同时打印到控制台（带颜色）
```

## 结构化字段（logfmt / JSON 均支持）

```go
golog.Infow("user login", golog.Str("user", "tom"), golog.Int("age", 18))
// 2022-03-04 10:19:31.000 -- [INFO] -- xx.go:9 -- user login user=tom age=18

// With 返回携带固定字段的子 logger，可链式叠加
orderLog := golog.With(golog.Str("svc", "order"))
orderLog.Infow("create order", golog.Int64("oid", 1001))
```

内置字段构造器：`Any / Str / Int / Int64 / Bool / Float64 / Duration / Err`

## 模块命名子 logger

```go
dbLog := golog.Named("db")
dbLog.Info("query") // ... -- query logger=db
```

## context 透传（trace_id 等）

```go
ctx := golog.ContextWithTraceID(ctx, "t-abc-123")
ctx = golog.ContextWithFields(ctx, golog.Str("uid", "u1"))

golog.Ctx(ctx).Info("handle request") // 自动携带 uid=u1 trace_id=t-abc-123

// 也可以手动取出 context 中的字段
fields := golog.FieldsFromContext(ctx)
```

## log/slog 适配器

```go
// 所有依赖标准库 slog 的库（GORM 等）可无缝走 golog
logger := slog.New(golog.NewSlogHandler(nil)) // nil 使用全局 logger
logger.Info("hello", "uid", 42)
```

级别映射：slog 的 Error/Warn/Info 对应 ERROR/WARN/INFO，其余映射为 DEBUG；`WithGroup` 的组名会作为 `group.key` 前缀。

## 自定义格式化

```go
// 格式化函数签名（fields 为结构化字段）
golog.SetFormatFunc(func(level golog.Level, ctime time.Time, line, msg string, fields []golog.Field) string {
	return fmt.Sprintf("%s [%s] %s %s %v\n", ctime.Format(time.RFC3339), level, line, msg, fields)
})

// 内置 JSON 格式：自动转义特殊字符，字段合并为 JSON 顶层键
golog.SetFormatFunc(golog.JsonFormat)
```

## 时间格式与时区

```go
golog.SetTimeLayout("2006-01-02 15:04:05.000") // 默认布局
golog.SetUTC(true)                             // 默认本地时区
```

## 限流采样（日志洪峰保护）

```go
golog.SetRateLimit(100, time.Second) // 每秒最多记录 100 条，超出丢弃
```

固定窗口采样，`max <= 0` 表示禁用。

## 重复日志采样（借鉴 zap）

```go
golog.SetLogPriority(true, 100, time.Minute) // 每分钟 100 条重复日志只打印一条
```

以 `文件行+内容` 为 key，重复日志只打印第一条；累计 `duplicates` 次后放行一条并重新计数。可选第三个参数为窗口时长，默认 1 分钟。

## Fatal / Panic / Stack

```go
// Fatal/Fatalf/Fatalw: 同步写盘（保证落盘）后 os.Exit(1)
golog.Fatalf("bad thing: %v", err)

// Panic/Panicf: 同步写盘后 panic
golog.Panic("unexpected state")

// Stack: ERROR 级别附带当前 goroutine 堆栈
golog.Stack("here")
```

## 多文件操作

```go
logger1 := golog.NewLog("test1.log", 0, false) // 独立实例，用法与全局方法一致
defer logger1.Sync()                           // 实例停止服务时必须调用

logger2 := golog.NewLog("test2.log", 0, false)
defer logger2.Sync()
logger1.Info("foo")
logger2.Info("foo")
```

独立实例继承当前全局级别与全局目录，并自动注册过期清理；实例上也有 `SetLevel / SetDir / SetOutput / SetConsole / SetCompress / SetRateLimit / SetLogPriority` 等方法，只影响该实例。

## 自定义输出目标（可测试性）

```go
var buf bytes.Buffer
golog.SetOutput(&buf) // 捕获控制台输出，用于单测断言或对接 syslog/kafka 等
```

## 回调函数

```go
// 每条日志异步回调，可用于报警等二次处理（应在首次打日志前设置）
golog.LogHandler = func(level golog.Level, ctime time.Time, line, msg string) {
	fmt.Println("你的代码出问题了")
}
```

## 接口方法调试

可以知道是哪一行调用了这个方法：

```go
golog.SetLevel(golog.DEBUG)
func test() {
	golog.UpFunc(1, "who call me")      // ... -- caller from example.go:11 -- who call me
	golog.UpFuncf(1, "who call me %s", "tom") // 格式化版本
}
```

## 错误日志源头追踪

```go
// 同步函数，包装 error 并记录来源，兼容 errors.Is / errors.As
golog.Wrap(err) error
golog.Wraps("msg") error
golog.Unwrap(err) error // 解出原始 error，非 golog 包装时返回 nil

func (p project) DeleteProject(projectId int64) error {
	return golog.Wrap(db.Gorm.Table(p.table).Where("pid=?", projectId).Delete(nil).Error)
}

// 最外层打印自动追加来源路径
// D:/cander/ITflow/go/app/repo/bugs.go:23 -- Error 1054 (42S22): Unknown column 'pid' in 'where clause'
```

## 溢出保护

异步通道写满时日志会被丢弃（不阻塞业务），并通过控制台周期性输出
`[golog] N log messages dropped (async buffer full)` 告警。

## 默认值一览

| 配置项 | 默认值 | 说明 |
| --- | --- | --- |
| 日志级别 | `INFO` | `SetLevel` 修改 |
| 日志目录 | `log` | `SetDir` 修改 |
| 过期清理 | 7 天 | `SetExpireDuration` 修改 |
| 时间布局 | `2006-01-02 15:04:05.000` | `SetTimeLayout` 修改 |
| 时区 | 本地时区 | `SetUTC(true)` 切换 UTC |
| 异步缓冲 | 1000 条 | 写满丢弃，不阻塞业务 |
| 文件刷盘 | 4KB 或 200ms 批量 | 控制台输出立即打印 |
| 归档压缩并发 | 2 | gzip 信号量限制 |
| 过期清理周期 | 24 小时 | 自动执行 |

## 异步架构说明

- 普通 日志走异步缓冲通道，由后台 goroutine 消费写盘，业务调用零阻塞
- `Fatal / Panic` 系列绕过异步通道同步写盘，保证进程退出前日志可见
- 归档压缩为异步执行，并发受信号量（2）限制，不抢占写盘
- `Sync()` 可重复调用，用于退出前强制刷盘（控制台模式无需调用）
