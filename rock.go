package rock

import (
	"fmt"
	"net/http"
	"sync"
)

type App struct {
	*RouterGroup
	router *Router
	pool   sync.Pool
	groups []*RouterGroup
}

func New() *App {
	app := &App{}
	app.router = NewRouter(app)
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
	c.writer = w
	c.req = r
	c.Path = r.URL.Path
	c.Method = r.Method
	c.index = -1
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

func (app *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := app.createContext(w, req)

	app.router.handle(c)
}
