package rock

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/plimble/utils/pool"
)

var bufPool = pool.NewBufferPool(100)

type (
	App struct {
		*Router
		mux          *chi.Mux
		pool         sync.Pool
		render       Renderer
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
		c := &Ctx{}
		c.index = -1
		c.Writer = &c.writercache
		return c
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
	a.mux.ServeHTTP(w, req)
}

func (app *App) GetMux() *chi.Mux {
	return app.mux
}
