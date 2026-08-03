package rock

import (
	"net/http"
	"path"
	"strings"
)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	app         *App          // all groups share a Engine instance
	noRoute     HandlerFunc   // 该分组路径下的 404 处理（per-group）
	noMethod    HandlerFunc   // 该分组路径下的 405 处理（per-group）
}

// groupMatchesPath 判断分组的 prefix 是否按路径段匹配 path。
// 根分组（prefix==""）匹配所有路径；"/admin" 只匹配 "/admin" 或 "/admin/..."，
// 不会误匹配 "/administrator"。
func groupMatchesPath(prefix, path string) bool {
	return prefix == "" || path == prefix || strings.HasPrefix(path, prefix+"/")
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
	// 按 prefix 长度有序插入，保持 app.groups 始终按（前缀长度, 注册顺序）排列，
	// 这样 collectMiddlewares 直接按序拼接即可，无需每请求排序
	insertPos := len(app.groups)
	for i, g := range app.groups {
		if len(g.prefix) > len(newGroup.prefix) {
			insertPos = i
			break
		}
	}
	app.groups = append(app.groups, nil)
	copy(app.groups[insertPos+1:], app.groups[insertPos:])
	app.groups[insertPos] = newGroup
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

// NoRoute 为当前分组及其子路径注册 404 处理函数（per-group）。
// 与全局 NoRoute 不同，它只作用于匹配该分组 prefix 的路径；
// 未命中任何注册了 NoRoute 的分组时，回退到根分组（app.NoRoute）。
func (group *RouterGroup) NoRoute(handler HandlerFunc) {
	group.noRoute = handler
}

// NoMethod 为当前分组注册 405 处理函数，作用范围同 NoRoute。
func (group *RouterGroup) NoMethod(handler HandlerFunc) {
	group.noMethod = handler
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c Context) {
		file, ok := c.Param("filepath").(string)
		if !ok {
			c.Status(http.StatusBadRequest)
			// Status 为懒写入，此处直接发送响应头
			c.Writer().WriteHeader(http.StatusBadRequest)
			return
		}
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			// Status 为懒写入，此处直接发送响应头
			c.Writer().WriteHeader(http.StatusNotFound)
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
