package rock

import (
	"fmt"
	"net/http"
)

type Router struct {
	handlers []HandlerFunc
	app      *App
	prefix   string
}

func (r *Router) Mount(prefix string, handler http.Handler) {
	r.app.mux.Mount(prefix, handler)
}

//Use register middleware
func (r *Router) Use(middlewares ...HandlerFunc) {
	for _, handler := range middlewares {
		r.handlers = append(r.handlers, handler)
	}
}

//GET handle GET method
func (r *Router) GET(path string, handlers ...HandlerFunc) {
	r.Handle("GET", path, handlers)
}

//POST handle POST method
func (r *Router) POST(path string, handlers ...HandlerFunc) {
	r.Handle("POST", path, handlers)
}

//PATCH handle PATCH method
func (r *Router) PATCH(path string, handlers ...HandlerFunc) {
	r.Handle("PATCH", path, handlers)
}

//PUT handle PUT method
func (r *Router) PUT(path string, handlers ...HandlerFunc) {
	r.Handle("PUT", path, handlers)
}

//DELETE handle DELETE method
func (r *Router) DELETE(path string, handlers ...HandlerFunc) {
	r.Handle("DELETE", path, handlers)
}

//HEAD handle HEAD method
func (r *Router) HEAD(path string, handlers ...HandlerFunc) {
	r.Handle("HEAD", path, handlers)
}

//OPTIONS handle OPTIONS method
func (r *Router) OPTIONS(path string, handlers ...HandlerFunc) {
	r.Handle("OPTIONS", path, handlers)
}

//Group group route
func (r *Router) Group(path string, handlers ...HandlerFunc) *Router {
	handlers = r.combineHandlers(handlers)
	return &Router{
		handlers: handlers,
		prefix:   r.path(path),
		app:      r.app,
	}
}

//Panic call when panic was called
// func (r *Router) Panic(h PanicHandler) {
// 	r.app.panicFunc = h
// }

//RouteNotFound call when route does not match
func (r *Router) RouteNotFound(h HandlerFunc) {
	r.app.notfoundFunc = h
}

//HandlerFunc convert http.HandlerFunc to ace.HandlerFunc
func (r *Router) HTTPHandlerFunc(h http.HandlerFunc) HandlerFunc {
	return func(c Context) {
		h(c.Response(), c.Request())
	}
}

func (r *Router) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	aLen := len(r.handlers)
	hLen := len(handlers)
	h := make([]HandlerFunc, aLen+hLen)
	copy(h, r.handlers)
	for i := 0; i < hLen; i++ {
		h[aLen+i] = handlers[i]
	}
	return h
}

func (r *Router) path(p string) string {
	if r.prefix == "/" {
		return p
	}

	return concat(r.prefix, p)
}

//Handle handle with specific method
func (r *Router) Handle(method, path string, handlers []HandlerFunc) {
	handlers = r.combineHandlers(handlers)
	r.app.mux.MethodFunc(method, r.path(path), func(w http.ResponseWriter, req *http.Request) {
		c := r.app.createContext(w, req)
		c.handlers = handlers
		c.Next()
		r.app.pool.Put(c)
	})
}

//Static server static file
//path is url path
//root is root directory
func (r *Router) Static(path string, root http.Dir, handlers ...HandlerFunc) {
	path = r.path(path)
	fileServer := http.StripPrefix(path, http.FileServer(root))

	handlers = append(handlers, func(c Context) {
		fileServer.ServeHTTP(c.Response(), c.Request())
	})

	p := r.staticPath(path)
	fmt.Println(p)
	r.app.mux.MethodFunc("GET", p, func(w http.ResponseWriter, req *http.Request) {
		c := r.app.createContext(w, req)
		fmt.Println("stat")
		c.handlers = handlers
		c.Next()
		r.app.pool.Put(c)
	})
}

func (r *Router) staticPath(p string) string {
	if p == "/" {
		return "/*"
	}

	return concat(p, "/*")
}
