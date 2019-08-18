package rock

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-chi/chi"
	"github.com/plimble/utils/pool"
)

var bufPool = pool.NewBufferPool(100)

type (
	App struct {
		*Router
		mux    *chi.Mux
		pool   sync.Pool
		render Renderer
		// panicFunc    PanicHandler
		notfoundFunc HandlerFunc

		// add
		contextFunc         ContextFunc
		customHandlersFuncs customHandlers
	}
)

func GetPool() *pool.BufferPool {
	return bufPool
}
func New() *App {
	app := &App{}
	app.Router = &Router{
		handlers: nil,
		prefix:   "/",
		app:      app,
	}
	app.mux = chi.NewMux()
	app.pool.New = func() interface{} {
		// c := &Ctx{}
		// c.index = -1
		// c.response = &c.writercache
		// return c
		c := app.contextFunc(app)
		b := c.BaseContext()
		b.index = -1
		b.parent = c
		return b
	}

	return app
}

func Default() *App {
	a := New()
	a.Use(Logger())
	return a
}

func DefaultWithout() *App {
	a := New()
	a.Use(Logger())
	return a
}

//Run server with specific address and port
func (app *App) Run(addr string) {
	if err := http.ListenAndServe(addr, app); err != nil {
		panic(err)
	}
}

//ServeHTTP implement http.Handler
func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// a.mux.ServeHTTP(w, req)
	c := a.pool.Get().(Context)
	c.RequestStart(w, req)
	// c.Next()
	// c.parent.RequestEnd()
	a.mux.ServeHTTP(c.Response(), c.Request())

	a.pool.Put(c)
}

func (app *App) GetMux() *chi.Mux {
	return app.mux
}

// RegisterCustomHandler registers a custom handler that gets wrapped by HandlerFunc
func (l *App) RegisterCustomHandler(customType interface{}, fn CustomHandlerFunc) {

	if l.customHandlersFuncs == nil {
		l.customHandlersFuncs = make(customHandlers)
	}

	t := reflect.TypeOf(customType)

	if _, ok := l.customHandlersFuncs[t]; ok {
		panic(fmt.Sprint("Custom Type + CustomHandlerFunc already declared: ", t))
	}

	l.customHandlersFuncs[t] = fn
}

// RegisterContext registers a custom Context function for creation
// and resetting of a global object passed per http request
func (l *App) RegisterContext(fn ContextFunc) {
	l.contextFunc = fn
}
