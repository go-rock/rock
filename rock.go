package rock

import (
	"net/http"
	"strings"
	"sync"
	"text/template"
)

type App struct {
	*RouterGroup
	router *Router
	pool   sync.Pool
	groups []*RouterGroup

	// template
	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render

	render HTMLRender
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
	c.statusCode = http.StatusOK
	c.index = -1
	c.render = app.render
	return c
}

// Run defines the method to start a http server
func (app *App) Run(args ...string) (err error) {
	addr := ":8989"
	if len(args) > 0 {
		addr = args[0]
	}
	debugPrint("Rock running on %s", addr)
	return http.ListenAndServe(addr, app)
}

func (app *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range app.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := app.createContext(w, req)

	c.handlers = middlewares

	app.router.handle(c)
}
