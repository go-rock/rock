package rock

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/doabit/rock/trie"
)

type contextKeyType struct{}

// Params represents named parameter values
type Params map[string]interface{}

// contextKey is the context key for the params.
var contextKey = contextKeyType{}

// newContext returns a new Context that carries a provided params value.
func newContext(ctx context.Context, params Params) context.Context {
	return context.WithValue(ctx, contextKey, params)
}

// fromContext extracts params from a Context.
func fromContext(ctx context.Context) (Params, bool) {
	values, ok := ctx.Value(contextKey).(Params)

	return values, ok
}

// HandlerFunc is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (variables).
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Mux is a tire base HTTP request router which can be used to
// dispatch requests to different handler functions.
type Mux struct {
	trie      *trie.Trie
	otherwise HandlerFunc
}

// New returns a Mux instance.
func New(opts ...trie.Options) *Mux {
	return &Mux{trie: trie.New(opts...)}
}

// Get registers a new GET route for a path with matching handler in the Mux.
func (m *Mux) Get(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodGet, pattern, handler)
}

// Head registers a new HEAD route for a path with matching handler in the Mux.
func (m *Mux) Head(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodHead, pattern, handler)
}

// Post registers a new POST route for a path with matching handler in the Mux.
func (m *Mux) Post(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodPost, pattern, handler)
}

// Put registers a new PUT route for a path with matching handler in the Mux.
func (m *Mux) Put(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodPut, pattern, handler)
}

// Patch registers a new PATCH route for a path with matching handler in the Mux.
func (m *Mux) Patch(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodPatch, pattern, handler)
}

// Delete registers a new DELETE route for a path with matching handler in the Mux.
func (m *Mux) Delete(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodDelete, pattern, handler)
}

// Options registers a new OPTIONS route for a path with matching handler in the Mux.
func (m *Mux) Options(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodOptions, pattern, handler)
}

// Otherwise registers a new handler in the Mux
// that will run if there is no other handler matching.
func (m *Mux) Otherwise(handler HandlerFunc) {
	m.otherwise = handler
}

// Handle registers a new handler with method and path in the Mux.
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (m *Mux) Handle(method, pattern string, handler HandlerFunc) {
	if method == "" {
		panic(fmt.Errorf("invalid method"))
	}
	m.trie.Define(pattern).Handle(strings.ToUpper(method), handler)
}

// Handler is an adapter which allows the usage of an http.Handler as a
// request handle.
func (m *Mux) Handler(method, path string, handler http.Handler) {
	m.Handle(method, path, func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req)
	})
}

// HandlerFunc is an adapter which allows the usage of an http.HandlerFunc as a
// request handle.
func (m *Mux) HandlerFunc(method, path string, handler http.HandlerFunc) {
	m.Handler(method, path, handler)
}

// ServeHTTP implemented http.Handler interface
func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler HandlerFunc
	path := req.URL.Path
	method := req.Method
	res := m.trie.Match(path)

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

		if m.otherwise == nil {
			http.Error(w, fmt.Sprintf(`"%s" not implemented`, path), 501)
			return
		}
		handler = m.otherwise
	} else {
		ok := false
		if handler, ok = res.Node.GetHandler(method).(HandlerFunc); !ok {
			// OPTIONS support
			if method == http.MethodOptions {
				w.Header().Set("Allow", res.Node.GetAllow())
				w.WriteHeader(204)
				return
			}

			if m.otherwise == nil {
				// If no route handler is returned, it's a 405 error
				w.Header().Set("Allow", res.Node.GetAllow())
				http.Error(w, fmt.Sprintf(`"%s" not allowed in "%s"`, method, path), 405)
				return
			}
			handler = m.otherwise
		}
	}
	if len(res.Params) != 0 {
		ctx := newContext(req.Context(), res.Params)
		req = req.WithContext(ctx)
	}
	handler(w, req)
}

// GetParams returns params stored in the request.
func GetParams(r *http.Request) (Params, bool) {
	return fromContext(r.Context())
}
