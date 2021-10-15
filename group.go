package rock

import (
	"net/http"
)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	app         *App          // all groups share a Engine instance
}

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

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}
