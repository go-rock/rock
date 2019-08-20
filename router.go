package rock

import (
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
func (r *Router) Use(middlewares ...Handler) {
	for _, handler := range middlewares {
		r.handlers = append(r.handlers, r.app.wrapHandler(handler))
	}
}

//GET handle GET method
func (r *Router) GET(path string, handlers ...Handler) {
	r.Handle(GET, path, handlers)
}

//POST handle POST method
func (r *Router) POST(path string, handlers ...Handler) {
	r.Handle(POST, path, handlers)
}

//PATCH handle PATCH method
func (r *Router) PATCH(path string, handlers ...Handler) {
	r.Handle(PATCH, path, handlers)
}

//PUT handle PUT method
func (r *Router) PUT(path string, handlers ...Handler) {
	r.Handle(PUT, path, handlers)
}

//DELETE handle DELETE method
func (r *Router) DELETE(path string, handlers ...Handler) {
	r.Handle(DELETE, path, handlers)
}

//HEAD handle HEAD method
func (r *Router) HEAD(path string, handlers ...Handler) {
	r.Handle(HEAD, path, handlers)
}

//OPTIONS handle OPTIONS method
func (r *Router) OPTIONS(path string, handlers ...Handler) {
	r.Handle(OPTIONS, path, handlers)
}

//Group group route
func (r *Router) Group(path string, handlers ...Handler) *Router {
	hs := r.combineHandlers(handlers)
	return &Router{
		handlers: hs,
		prefix:   r.path(path),
		app:      r.app,
	}
}

// Panic call when panic was called
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

func (r *Router) combineHandlers(handlers []Handler) []HandlerFunc {
	aLen := len(r.handlers)
	hLen := len(handlers)
	h := make([]HandlerFunc, aLen+hLen)
	copy(h, r.handlers)
	for i := 0; i < hLen; i++ {
		h[aLen+i] = r.app.wrapHandler(handlers[i])
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
func (r *Router) Handle(method, path string, handlers []Handler) {
	hs := r.combineHandlers(handlers)
	r.app.mux.MethodFunc(method, r.path(path), func(w http.ResponseWriter, req *http.Request) {
		// c := r.app.createContext(w, req)
		c := r.app.pool.Get().(*Ctx)
		c.RequestStart(w, req)
		// c.handlers = hs
		c.SetHandlers(hs)
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
	r.app.mux.MethodFunc("GET", p, func(w http.ResponseWriter, req *http.Request) {
		// c := r.app.createContext(w, req)
		c := r.app.pool.Get().(*Ctx)
		c.RequestStart(w, req)

		// c.handlers = hs
		c.SetHandlers(handlers)
		// c.handlers = handlers
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

func (app *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// app.mux.ServeHTTP(w, req)
	// c := &Context{mux.rock, w, r, make(map[string]interface{}), bpool.NewBufferPool(48)}
	c := app.pool.Get().(*Ctx)
	c.RequestStart(w, req)
	// c.handlers = hs
	// // c.handlers =
	// pp.Println("start")
	c.Next()
	// pp.Println(len(c.handlers))
	app.mux.ServeHTTP(c.response, c.request)

	// pp.Println("end")
	// // c.response.Write([]byte("xiao"))
	app.pool.Put(c)
}
