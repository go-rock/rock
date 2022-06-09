package rock

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-rock/rock/trie"
)

// Mux is a tire base HTTP request router which can be used to
// dispatch requests to different handler functions.
type Router struct {
	trie      *trie.Trie
	otherwise HandlerFunc
	noRoute   HandlerFunc
	noMethod  HandlerFunc
	// prefix    string
}

// New returns a Mux instance.
func NewRouter(opts ...trie.Options) *Router {
	return &Router{trie: trie.New(opts...)}
}

// Get registers a new GET route for a path with matching handler in the Router.
// func (r *Router) Get(pattern string, handler HandlerFunc) {
// 	r.Handle(http.MethodGet, pattern, handler)
// }

func (r *Router) Handle(method, pattern string, handler HandlerFunc) {
	if method == "" {
		panic(fmt.Errorf("invalid method"))
	}
	// if r.prefix != "" {
	// 	pattern = r.prefix + pattern
	// }
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
		if r.noRoute == nil {
			http.Error(w, fmt.Sprintf(`"%s" not implemented`, path), 501)
			return
		}
		handler = r.noRoute
	} else {
		// ok := false
		hd := res.Node.GetHandler(method)
		handler, _ = hd.(HandlerFunc)
		// handler = r.wrapHandler(hd)
		// if !ok {
		// 	panic("handler error")
		// }
		if handler == nil {
			// OPTIONS support
			if method == http.MethodOptions {
				w.Header().Set("Allow", res.Node.GetAllow())
				w.WriteHeader(204)
				return
			}

			if r.noMethod == nil {
				// If no route handler is returned, it's a 405 error
				w.Header().Set("Allow", res.Node.GetAllow())
				http.Error(w, fmt.Sprintf(`"%s" not allowed in "%s"`, method, path), 405)
				return
			}
			handler = r.noMethod
		}
	}

	if len(res.Params) != 0 {
		c.params = res.Params
	}
	c.handlers = append(c.handlers, handler)
	c.Next()
}
