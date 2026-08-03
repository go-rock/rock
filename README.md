# Rock

[![CI](https://github.com/go-rock/rock/actions/workflows/ci.yml/badge.svg)](https://github.com/go-rock/rock/actions)

一个轻量、易用的 Go Web 框架，基于标准库 `net/http`，受 [gin](https://github.com/gin-gonic/gin)、[iris](https://github.com/kataras/iris)、[lars](https://github.com/go-playground/lars) 启发。

## 特性

- 基于 trie 的高性能路由：命名参数 `:id`、通配符 `:name*`、正则参数 `:id(\d+)`、尾斜杠/固定路径自动重定向
- 路由分组 + **per-group 中间件与 404/405 处理**
- 请求上下文（`Context`）统一封装：响应、参数、绑定、上传、日志
- 内置 `ShouldBind` 请求绑定与 go-playground/validator 校验
- 文件上传（大小/扩展名/MIME 校验、唯一文件名、请求体上限防护）
- 统一错误响应格式，生产环境不泄露内部细节
- 可插拔的视图引擎（默认配合 [`rock-pongo2`](https://github.com/go-rock/rock-pongo2)）
- 优雅关闭与 HTTPS（TLS）
- 基于 `sync.Pool` 的对象复用，零额外依赖的性能设计

## 安装

```bash
go get github.com/go-rock/rock
```

Go 版本要求：>= 1.17

## 快速开始

```go
package main

import "github.com/go-rock/rock"

func main() {
	app := rock.New()
	app.Use(rock.Recovery())

	app.Get("/", func(c rock.Context) {
		c.JSON(200, rock.M{"message": "hello rock"})
	})

	app.Run(":8989") // 阻塞运行，Ctrl-C 优雅退出
}
```

## 路由

### 基础路由

```go
app.Get("/users", listUsers)
app.Post("/users", createUser)
app.Put("/users/:id", updateUser)
app.Delete("/users/:id", deleteUser)
app.Patch("/users/:id", patchUser)
app.Options("/users", preflight)
```

### 参数路由

| 语法 | 说明 | 示例 |
|---|---|---|
| `:name` | 命名参数 | `/users/:id` 匹配 `/users/42` |
| `:name*` | 通配符（捕获剩余路径） | `/files/:path*` 匹配 `/files/a/b.txt` |
| `:name(regexp)` | 正则参数 | `/users/:id(\d+)` 只匹配数字 id |
| `::name` | 字面量 `:name` | `/x/::id` 只匹配 `/x/:id` |
| `:name+suffix` | 后缀参数 | `/a/:file+.json` 匹配 `/a/data.json` |

```go
app.Get("/users/:id", func(c rock.Context) {
	id := c.Param("id")          // "42"
	uid := c.MustParamInt("id", 0) // 42，失败用默认值 0
})
```

### 静态文件

```go
app.Static("/assets", "./static")  // 目录
```

### 路由分组

分组可以嵌套，前缀自动拼接；**分组级中间件和 404/405 只作用于本分组路径**。

```go
admin := app.Group("/admin")
admin.Use(authMiddleware())        // 只作用于 /admin 前缀
admin.Get("/login", adminLogin)

// per-group 404：/admin 下未匹配的路径返回 JSON
admin.NoRoute(func(c rock.Context) {
	c.JSON(404, rock.M{"msg": "not found"})
})

// 根分组 404（作用于全局）
app.NoRoute(func(c rock.Context) {
	c.HTML("404")
})
```

## 中间件

中间件是 `func(rock.Context)`，通过 `c.Next()` 继续链路；`c.Abort()` 终止链路。

```go
func Logger() rock.HandlerFunc {
	return func(c rock.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[%d] %s in %v", c.StatusCode(), c.Request().URL.Path, time.Since(start))
	}
}

app.Use(rock.Recovery(), Logger())
```

分组中间件按"外层分组先、内层分组后"执行。额外工具方法：

- `group.UseFunc(...)` —— 添加 `func(Context)` 类型中间件
- `group.UseWithPriority(priority, mw)` —— 指定位置插入
- `group.RemoveMiddleware(i)` / `group.ClearMiddleware()` —— 移除/清空

内置 `rock.Recovery()` 恢复 panic 并返回统一的 500 响应。

### 路由级中间件

中间件可以只绑定到某一条路由（在处理器之前执行），适合"同一分组下有公有有私有"的场景：

```go
g.POST("/login", Login)                            // 公开路由：不挂鉴权
g.GET("/users", ListUser, JWTAuth(secret), RequirePermission(enf)) // 只对该路由挂鉴权
```

与分组中间件叠加时，执行顺序为：**分组中间件 → 路由级中间件 → 处理器**。

## 请求上下文

`rock.Context` 封装了完整的请求/响应能力。

### 响应

```go
c.JSON(200, rock.M{"ok": true})          // Content-Type: application/json
c.XML(200, obj)
c.String(200, "Hello %s", name)          // text/plain
c.Status(204)                            // 只设状态码（懒写头，可先设后改）
c.SetHeader("X-Custom", "1")
c.Attachment(fileReader, "a.txt")        // 下载
c.Inline(fileReader, "view.txt")         // 内联
```

### 请求参数

```go
c.Param("id")            // 路径参数
c.Query("page")          // 查询参数
c.GetQuery("page")       // (值, 是否存在)，?page= 视为存在但为空
c.QueryInt("page")       // 查询参数转 int
c.MustPostInt("age", 0)  // 表单参数，带默认值
```

### 绑定与校验

```go
type CreateUser struct {
	Name  string `binding:"required"`
	Email string `binding:"required,email"`
}

func createUser(c rock.Context) {
	var req CreateUser
	if err := c.ShouldBind(&req); err != nil { // 按 Content-Type 自动选择 JSON/XML/Form
		c.JSON(400, rock.M{"error": err.Error()})
		return
	}
	// 使用 req...
}
```

`ShouldBind`/`Decode` 同一请求内可重复调用；默认 body 上限 10MB，可通过 `c.ShouldBind(&req, false, 5<<20)` 调整。

### 视图数据与渲染

```go
c.Set("user", user)        // 键值数据
c.SetData(rock.M{...})     // 整体数据
c.ViewData("title", "首页") // 视图数据（供模板使用）
c.HTML("home", rock.M{"post": p}) // 渲染模板
```

## 文件上传

```go
func upload(c rock.Context) {
	config := &rock.FileUploadConfig{
		MaxFileSize:       10 << 20,          // 单文件 10MB
		MaxTotalSize:      11 << 20,          // 请求体总上限（防 DoS）
		AllowedExtensions: []string{".jpg", ".jpeg", ".png"},
		AllowedMimeTypes:  []string{"image/jpeg", "image/png"},
		SaveDir:           "./uploads",
		GenerateUniqueName: true,
		FilenamePrefix:     "upload_",
	}

	info, err := c.SaveSingleFile("file", config)
	if err != nil {
		c.JSON(400, rock.M{"error": err.Error()})
		return
	}
	c.JSON(200, rock.H{"filename": info.Filename, "path": info.SavedPath})
}
```

快捷方法：`c.UploadSingleImage` / `c.UploadSingleDocument` / `c.UploadMultipleImages` / `c.SaveMultipleFiles`。
上传目录如果通过 `app.Static` 对外服务，注意白名单不要放开 `.html`/`.svg` 等可被浏览器执行的类型。

## 错误处理

统一的错误响应结构：`{"success": false, "error": {"code": 400, "message": "...", "detail": "..."}}`。

```go
rock.WriteError(c, 404, rock.NewAppError(rock.ErrNotFound, "not found"))
rock.WriteSuccess(c, data) // 200 + {"success": true, "data": ...}

// 错误码常量
rock.ErrBadRequest   // 400
rock.ErrUnauthorized // 401
rock.ErrForbidden    // 403
rock.ErrNotFound     // 404
rock.ErrMethodNotAllow // 405
rock.ErrUnprocessable  // 422
rock.ErrInternalServer // 500
```

`ShouldBind` 的校验错误会由 `WriteError` 自动映射为 400 "Validation failed"；未知内部错误在生产环境只返回通用文案，细节仅在调试模式（`rock.SetDebug(true)`）下进入 `detail`。

## 日志

内置 `RockLogger`，默认开启请求日志：

```go
app.SetLogLevel(rock.LevelDebug)              // Debug < Info < Warn < Error < Fatal
app.EnableRequestLog(false)                   // 关闭请求日志
app.SetLoggerOutput(os.Stdout, file)          // 多输出

// Context 内日志
c.LogInfo("processing %s", c.GetPath())
c.LogError("something failed")

// 全局便捷函数
rock.Info("app started")
rock.Errorf("error: %v", err)
```

## 视图引擎

核心不内置模板引擎，通过 `ViewEngine` 接口插拔。推荐 [`rock-pongo2`](https://github.com/go-rock/rock-pongo2)（基于 pongo2/Django 语法）：

```go
import render "github.com/go-rock/rock-pongo2"

app.RegisterView(render.New(render.ViewConfig{
	ViewDir:   "./views/",
	Extension: ".html",
}))
app.Get("/", func(c rock.Context) {
	c.HTML("home", rock.M{"title": "Home"})
})
```

## 配置

```go
app.SetTrustProxy(true)  // 部署在可信反向代理（nginx 等）后时开启，让 ClientIP 读取代理头；默认关闭防伪造
app.SetDebug(true)       // 开启调试输出（路由表、错误细节）；生产环境保持关闭
app.SetLogLevel(rock.LevelInfo)
```

## 服务器

```go
app.Run()                // 默认 :8989
app.Run(":8080")
app.RunTLS(":8443", "server.crt", "server.key") // HTTPS
```

`Run`/`RunTLS` 阻塞运行，收到 `SIGINT`/`SIGTERM` 后等待在途请求最多 5 秒完成再退出。

## 测试

```bash
go test ./...           # 全部测试
go test -race ./...     # 竞态检测
go test -cover ./...    # 覆盖率
```

## 项目结构

```
rock/
├── rock.go        # App：ServeHTTP、中间件收集、服务器（Run/RunTLS）
├── context.go     # Context/Ctx：请求-响应上下文、绑定、上传、视图数据
├── router.go      # Router：路由分发、per-group NoRoute/NoMethod
├── trie/          # trie 前缀树路由引擎（参数/通配符/正则/重定向）
├── group.go       # RouterGroup：分组、中间件管理、路由注册
├── binding/       # 请求绑定与 go-playground/validator 校验
├── upload.go      # 文件上传
├── errors.go      # 统一错误模型
├── logger.go      # 日志
├── recovery.go    # panic 恢复中间件
├── store.go       # 并发安全的键值存储（ctx.Values）
├── config.go      # 配置项
└── debug.go       # 调试开关（SetDebug/IsDebugging）
```

## 致谢

- [gin](https://github.com/gin-gonic/gin)
- [iris](https://github.com/kataras/iris)
- [lars](https://github.com/go-playground/lars)
- [trie-mux](https://github.com/teambition/trie-mux)（trie 路由引擎的原型）

## 许可证

MIT
