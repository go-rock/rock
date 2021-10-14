package rock

import (
	"fmt"
	"net/http"
	"sync"
)

type App struct {
	router *Router
	pool   sync.Pool
}

func New() *App {
	app := &App{}
	app.router = NewRouter(app)
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
	c.writer = w
	c.req = r
	c.Path = r.URL.Path
	c.Method = r.Method
	return c
}

// Run defines the method to start a http server
func (app *App) Run(args ...string) (err error) {
	addr := ":8989"
	if len(args) > 0 {
		addr = args[0]
	}
	fmt.Println("rock running on " + addr)
	return http.ListenAndServe(addr, app)
}

func (app *App) Get(pattern string, handler HandlerFunc) {
	app.router.Handle(http.MethodGet, pattern, handler)
}

func (app *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	app.router.ServeHTTP(w, req)
}
