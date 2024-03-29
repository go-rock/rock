package rock

import (
	"net/http"
	"path"
)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	app         *App          // all groups share a Engine instance
}

func (group *RouterGroup) RegisterView(viewEngine ViewEngine) {
	// group.app.view.Register(viewEngine)

	handler := func(ctx Context) {
		ctx.ViewEngine(viewEngine)
		ctx.Next()
	}
	group.Use(handler)
	// api.UseError(handler)
}
func (group *RouterGroup) SetRender(render ViewEngine) {}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	app := group.app
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		app:    app,
	}
	app.groups = append(app.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	// log.Printf("Route %4s - %s", method, pattern)
	group.app.router.Handle(method, pattern, handler)
}

func (group *RouterGroup) Get(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodGet, pattern, handler)
}

func (group *RouterGroup) Post(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodPost, pattern, handler)
}

func (group *RouterGroup) Put(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodPut, pattern, handler)
}

func (group *RouterGroup) Patch(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodPatch, pattern, handler)
}

func (group *RouterGroup) Delete(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodDelete, pattern, handler)
}

func (group *RouterGroup) Options(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodOptions, pattern, handler)
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) NoRoute(handler HandlerFunc) {
	group.app.router.noRoute = handler
}

func (group *RouterGroup) NoMethod(handler HandlerFunc) {
	group.app.router.noMethod = handler
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c Context) {
		file := c.Param("filepath").(string)
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer(), c.Request())
	}
}

// serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/:filepath*")
	// Register GET handlers
	group.Get(urlPattern, handler)
}
