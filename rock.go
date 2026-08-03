package rock

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-playground/form/v4"
	log "github.com/kataras/golog"
)

var (
	formDecoder     *form.Decoder
	formDecoderInit sync.Once
)

// App 是 rock 框架的核心实例，实现了 http.Handler 接口。
// 通过 New 创建，用于注册路由、中间件、视图引擎并启动服务。
type App struct {
	*RouterGroup
	router *Router
	pool   sync.Pool
	groups []*RouterGroup

	// template
	config *Configuration
	view   View

	// logging
	logger *RockLogger
}

// GetView 返回视图引擎持有者（View）。
func (app *App) GetView() View {
	return app.view
}

// Logger 配置日志系统
func (app *App) Logger() *RockLogger {
	return app.logger
}

// SetLogLevel 设置日志级别
func (app *App) SetLogLevel(level LogLevel) {
	if app.logger != nil {
		app.logger.SetLevel(level)
	}
}

// SetLoggerOutput 设置日志输出目标
func (app *App) SetLoggerOutput(outputs ...io.Writer) {
	if app.logger != nil {
		app.logger.SetOutputs(outputs...)
	}
}

// EnableRequestLog 启用或禁用请求日志
func (app *App) EnableRequestLog(enabled bool) {
	if app.logger != nil {
		app.logger.EnableRequestLog(enabled)
	}
}

func New() *App {
	config := DefaultConfiguration()
	app := &App{
		config: &config,
		router: NewRouter(),
		logger: NewLogger(),
	}
	app.RouterGroup = &RouterGroup{app: app}
	app.groups = []*RouterGroup{app.RouterGroup}
	app.pool.New = func() interface{} {
		return app.allocateContext()
	}
	return app
}

func (app *App) allocateContext() *Ctx {
	return &Ctx{app: app}
}

func (app *App) createContext(w http.ResponseWriter, r *http.Request) *Ctx {
	c := app.pool.Get().(*Ctx)
	c.newContext(w, r)
	return c
}

// serverOptions 服务器启动选项
type serverOptions struct {
	certFile string
	keyFile  string
}

// shutdownTimeout 优雅关闭时等待在途请求完成的最长时间
const shutdownTimeout = 5 * time.Second

// Run 启动 HTTP 服务并阻塞，直到收到 SIGINT/SIGTERM 优雅退出。
// args[0] 可选，为监听地址，默认 ":8989"。
func (app *App) Run(args ...string) error {
	addr := ":8989"
	if len(args) > 0 {
		addr = args[0]
	}
	return app.serve(addr, serverOptions{})
}

// RunTLS 以 HTTPS 启动服务（certFile/keyFile 为 PEM 文件路径），
// 同样支持 SIGINT/SIGTERM 优雅退出。
func (app *App) RunTLS(addr, certFile, keyFile string) error {
	return app.serve(addr, serverOptions{certFile: certFile, keyFile: keyFile})
}

// serve 启动 HTTP(S) 服务，处理优雅退出。
func (app *App) serve(addr string, opts serverOptions) error {
	scheme := "http"
	if opts.certFile != "" {
		scheme = "https"
	}
	if app.logger != nil {
		app.logger.Infof("Rock running on %s://localhost%s", scheme, addr)
	} else {
		debugPrint("Rock running on %s://localhost%s", scheme, addr)
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: app,
	}

	// 监听退出信号做优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		var err error
		if opts.certFile != "" {
			err = srv.ListenAndServeTLS(opts.certFile, opts.keyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		if app.logger != nil {
			app.logger.Infof("Received %s, shutting down...", sig)
		}
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}

func (app *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := app.createContext(w, req)
	// 无论 handler 是否 panic，都要把 Context 归还对象池
	defer app.pool.Put(c)

	// 记录请求开始时间
	startTime := time.Now()

	// 获取客户端IP
	ip := c.ClientIP()
	userAgent := req.UserAgent()

	// 收集并按优先级排序中间件
	middlewares := app.collectMiddlewares(req.URL.Path)
	c.handlers = middlewares

	// 执行路由处理
	app.router.handle(c)

	// 记录请求日志
	if app.logger != nil {
		statusCode := c.StatusCode()
		latency := time.Since(startTime)
		app.logger.RequestLog(req.Method, req.URL.Path, ip, userAgent, statusCode, latency)
	}
}

// collectMiddlewares 收集匹配路径的所有中间件。
// app.groups 已按 prefix 长度有序（见 RouterGroup.Group），
// 因此直接按序拼接即可得到"外层分组先、内层分组后"的执行顺序，无需每请求排序。
func (app *App) collectMiddlewares(path string) []HandlerFunc {
	var middlewares []HandlerFunc

	// 按路径段匹配：组前缀 "/admin" 只匹配 "/admin" 或 "/admin/..."，
	// 避免把 "/admin" 组的中间件误套到 "/administrator" 这类同前缀路径上。
	for _, group := range app.groups {
		if len(group.middlewares) > 0 && groupMatchesPath(group.prefix, path) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	return middlewares
}

// ConfigurationReadOnly returns an object which doesn't allow field writing.
func (app *App) ConfigurationReadOnly() *Configuration {
	return app.config
}

// SetTrustProxy 控制是否信任反向代理设置的头（X-Real-IP / X-Forwarded-For）。
// 仅当应用部署在可信反向代理之后时才应开启，默认关闭。
func (app *App) SetTrustProxy(enabled bool) {
	app.config.TrustProxyHeaders = enabled
}

// SetDebug 开启/关闭全局调试输出（路由表、debugPrint，以及 WriteError 的
// 内部错误细节）。生产环境建议保持关闭。测试环境下始终为调试模式。
func (app *App) SetDebug(enabled bool) {
	SetDebug(enabled)
}

func (app *App) View(writer io.Writer, filename string, bindingData interface{}) error {
	if !app.view.Registered() {
		err := errors.New("view engine is missing, use `RegisterView`")
		// app.logger.Error(err)
		log.Error(err)
		return err
	}

	return app.view.ExecuteWriter(writer, filename, bindingData)
}

func (app *App) RegisterView(viewEngine ViewEngine) {
	log.Info("register view from app")
	app.view.Register(viewEngine)
}

// Bidning

func initFormDecoder() {
	formDecoderInit.Do(func() {
		formDecoder = form.NewDecoder()
	})
}
