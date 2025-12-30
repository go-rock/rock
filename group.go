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

// UseFunc 添加函数类型的中间件
func (group *RouterGroup) UseFunc(middlewares ...func(Context)) {
	for _, middleware := range middlewares {
		group.middlewares = append(group.middlewares, HandlerFunc(middleware))
	}
}

// UseWithPriority 添加带优先级的中间件
func (group *RouterGroup) UseWithPriority(priority int, middleware HandlerFunc) {
	// 简单的优先级实现：在指定位置插入
	if priority <= 0 {
		group.middlewares = append([]HandlerFunc{middleware}, group.middlewares...)
	} else if priority >= len(group.middlewares) {
		group.middlewares = append(group.middlewares, middleware)
	} else {
		// 在指定位置插入
		group.middlewares = append(group.middlewares[:priority], append([]HandlerFunc{middleware}, group.middlewares[priority:]...)...)
	}
}

// RemoveMiddleware 移除指定索引的中间件
func (group *RouterGroup) RemoveMiddleware(index int) {
	if index >= 0 && index < len(group.middlewares) {
		group.middlewares = append(group.middlewares[:index], group.middlewares[index+1:]...)
	}
}

// ClearMiddleware 清除所有中间件
func (group *RouterGroup) ClearMiddleware() {
	group.middlewares = []HandlerFunc{}
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
		file, ok := c.Param("filepath").(string)
		if !ok {
			c.Status(http.StatusBadRequest)
			return
		}
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
