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

func (r *Router) Handle(method, pattern string, handler HandlerFunc) error {
	if method == "" {
		return fmt.Errorf("invalid method")
	}
	// if r.prefix != "" {
	// 	pattern = r.prefix + pattern
	// }
	hds := []HandlerFunc{}
	hds = append(hds, handler)
	debugPrintRoute(method, pattern, hds)
	r.trie.Define(pattern).Handle(strings.ToUpper(method), handler)
	return nil
}

// ServeHTTP implemented http.Handler interface
// func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
// 	// 创建一个简单的Context用于测试
// 	c := &Ctx{}
// 	c.newContext(w, req)
// 	r.handle(c)
// }

// func(r *Router) handle(c htt)
func (r *Router) handle(c *Ctx) {
	var handler HandlerFunc
	req := c.Request()
	w := c.Writer()
	path := req.URL.Path
	method := req.Method
	res := r.trie.Match(path)

	// 所有响应（包括 404/405/OPTIONS/重定向）都作为链上的最后一个 handler，
	// 保证全局中间件（日志、鉴权、Recovery 等）对未匹配路径同样生效。
	if res.TSR != "" || res.FPR != "" {
		// FixedPathRedirect or TrailingSlashRedirect
		req.URL.Path = res.TSR
		if res.FPR != "" {
			req.URL.Path = res.FPR
		}
		code := 301
		if method != "GET" {
			code = 307
		}
		redirectURL := req.URL.String()
		handler = func(ctx Context) {
			http.Redirect(w, req, redirectURL, code)
		}
	} else if res.Node == nil {
		// 无匹配路由：默认 404 或自定义 noRoute
		if r.noRoute == nil {
			handler = func(ctx Context) {
				WriteError(ctx, 404, NewAppError(ErrNotFound, "Route Not Found"))
			}
		} else {
			handler = r.noRoute
		}
	} else {
		hd := res.Node.GetHandler(method)
		if hf, ok := hd.(HandlerFunc); ok {
			handler = hf
		} else if hd != nil {
			// 尝试包装其他类型的处理器
			wrappedHandler, err := r.wrapHandler(hd)
			if err != nil {
				handler = func(ctx Context) {
					WriteError(ctx, 500, NewAppError(ErrInternalServer, fmt.Sprintf("Invalid handler for %s %s: %v", method, path, err)))
				}
			} else {
				handler = wrappedHandler
			}
		}
		if handler == nil {
			// 节点存在但没有匹配的 method
			if method == http.MethodOptions {
				// OPTIONS support
				handler = func(ctx Context) {
					ctx.SetHeader("Allow", res.Node.GetAllow())
					ctx.Status(http.StatusNoContent)
					ctx.Write(nil)
				}
			} else if r.noMethod == nil {
				handler = func(ctx Context) {
					WriteError(ctx, 405, NewAppError(ErrMethodNotAllow, fmt.Sprintf(`Method "%s" not allowed in "%s"`, method, path)))
				}
			} else {
				handler = r.noMethod
			}
		}
	}

	if len(res.Params) != 0 {
		c.params = res.Params
	}
	c.handlers = append(c.handlers, handler)
	c.Next()
}
