package rock

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-playground/form/v4"
	log "github.com/kataras/golog"
)

var (
	formDecoder     *form.Decoder
	formDecoderInit sync.Once
)

type App struct {
	*RouterGroup
	router *Router
	pool   sync.Pool
	groups []*RouterGroup

	// template
	config *Configuration
	view   View
}

func (app *App) GetView() View {
	return app.view
}

func New() *App {
	config := DefaultConfiguration()
	app := &App{
		config: &config,
		router: NewRouter(),
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

// Run defines the method to start a http server
func (app *App) Run(args ...string) (err error) {
	addr := ":8989"
	if len(args) > 0 {
		addr = args[0]
	}
	debugPrint("Rock running on http://localhost%s", addr)
	return http.ListenAndServe(addr, app)
}

func (app *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := app.createContext(w, req)

	// 收集并按优先级排序中间件
	middlewares := app.collectMiddlewares(req.URL.Path)
	c.handlers = middlewares

	// 执行路由处理
	app.router.handle(c)

	// 释放Context到对象池
	app.pool.Put(c)
}

// collectMiddlewares 收集并排序中间件
func (app *App) collectMiddlewares(path string) []HandlerFunc {
	type middlewareWithPriority struct {
		handler HandlerFunc
		priority int
	}

	var middlewareList []middlewareWithPriority

	// 收集所有匹配的中间件
	for _, group := range app.groups {
		if strings.HasPrefix(path, group.prefix) && len(group.middlewares) > 0 {
			for i, handler := range group.middlewares {
				// 简单的优先级策略：group的深度和中间件在组中的位置
				priority := len(group.prefix) * 100 + i
				middlewareList = append(middlewareList, middlewareWithPriority{
					handler: handler,
					priority: priority,
				})
			}
		}
	}

	// 按优先级排序（从低到高）
	if len(middlewareList) > 1 {
		for i := 0; i < len(middlewareList); i++ {
			for j := i + 1; j < len(middlewareList); j++ {
				if middlewareList[i].priority > middlewareList[j].priority {
					middlewareList[i], middlewareList[j] = middlewareList[j], middlewareList[i]
				}
			}
		}
	}

	// 提取排序后的处理器
	middlewares := make([]HandlerFunc, len(middlewareList))
	for i, mw := range middlewareList {
		middlewares[i] = mw.handler
	}

	return middlewares
}

// ConfigurationReadOnly returns an object which doesn't allow field writing.
func (app *App) ConfigurationReadOnly() *Configuration {
	return app.config
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
