package rock

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/doabit/rock/trie"
)

// Mux is a tire base HTTP request router which can be used to
// dispatch requests to different handler functions.
type Router struct {
	app         *App
	trie        *trie.Trie
	otherwise   HandlerFunc
	middlewares []HandlerFunc
	prefix      string
}

// New returns a Mux instance.
func NewRouter(app *App, opts ...trie.Options) *Router {
	return &Router{trie: trie.New(opts...), prefix: "", app: app}
}

// Get registers a new GET route for a path with matching handler in the Router.
func (r *Router) Get(pattern string, handler HandlerFunc) {
	r.Handle(http.MethodGet, pattern, handler)
}

func (r *Router) Handle(method, pattern string, handler HandlerFunc) {
	if method == "" {
		panic(fmt.Errorf("invalid method"))
	}
	if r.prefix != "" {
		pattern = r.prefix + pattern
	}
	hds := []HandlerFunc{}
	hds = append(hds, handler)
	debugPrintRoute(method, pattern, hds)
	r.trie.Define(pattern).Handle(strings.ToUpper(method), handler)
}

// func(r *Router) handle(c htt)
// ServeHTTP implemented http.Handler interface
func (r *Router) handle(c *Ctx) {
	var handler HandlerFunc
	req := c.Request()
	w := c.Writer()
	path := req.URL.Path
	method := req.Method
	res := r.trie.Match(path)

	if res.Node == nil {
		// FixedPathRedirect or TrailingSlashRedirect
		if res.TSR != "" || res.FPR != "" {
			req.URL.Path = res.TSR
			if res.FPR != "" {
				req.URL.Path = res.FPR
			}
			code := 301
			if method != "GET" {
				code = 307
			}
			http.Redirect(w, req, req.URL.String(), code)
			return
		}

		if r.otherwise == nil {
			http.Error(w, fmt.Sprintf(`"%s" not implemented`, path), 501)
			return
		}
		handler = r.otherwise
	} else {
		ok := false
		hd := res.Node.GetHandler(method)
		handler, ok = hd.(HandlerFunc)
		if !ok {
			panic("handler error")
		}
		handler = r.wrapHandler(hd)
		if handler == nil {
			// OPTIONS support
			if method == http.MethodOptions {
				w.Header().Set("Allow", res.Node.GetAllow())
				w.WriteHeader(204)
				return
			}

			if r.otherwise == nil {
				// If no route handler is returned, it's a 405 error
				w.Header().Set("Allow", res.Node.GetAllow())
				http.Error(w, fmt.Sprintf(`"%s" not allowed in "%s"`, method, path), 405)
				return
			}
			handler = r.otherwise
		}
	}

	if len(res.Params) != 0 {
		c.params = res.Params
	}
	fmt.Println(c.Path, c.params, c.handlers)
	handler(c)
}
