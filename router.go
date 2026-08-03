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
	trie *trie.Trie
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
		// 在 req.URL 的副本上计算重定向目标，不改动原始路径，
		// 保证中间件通过 c.Request().URL.Path 看到的仍是原始请求路径
		targetPath := res.TSR
		if res.FPR != "" {
			targetPath = res.FPR
		}
		targetURL := *req.URL
		targetURL.Path = targetPath
		redirectURL := targetURL.String()

		code := 301
		if method != "GET" {
			code = 307
		}
		handler = func(ctx Context) {
			http.Redirect(w, req, redirectURL, code)
		}
	} else if res.Node == nil {
		// 无匹配路由：按分组就近选择 404 处理函数
		if h := r.resolveNoRoute(c, path); h != nil {
			handler = h
		} else {
			handler = func(ctx Context) {
				WriteError(ctx, 404, NewAppError(ErrNotFound, "Route Not Found"))
			}
		}
	} else {
		hd := res.Node.GetHandler(method)
		// HEAD 未注册时自动回退到 GET（响应体由 net/http 在服务端剥离）
		if hd == nil && method == http.MethodHead {
			hd = res.Node.GetHandler(http.MethodGet)
		}
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
			} else if noMethod := r.resolveNoMethod(c, path); noMethod != nil {
				handler = noMethod
			} else {
				handler = func(ctx Context) {
					WriteError(ctx, 405, NewAppError(ErrMethodNotAllow, fmt.Sprintf(`Method "%s" not allowed in "%s"`, method, path)))
				}
			}
		}
	}

	if len(res.Params) != 0 {
		c.params = res.Params
	}
	c.handlers = append(c.handlers, handler)
	c.Next()
}

// resolveNoRoute 找出对 path 最具体且注册了 NoRoute 的分组。
// 未找到时回退到根分组（app.NoRoute），再返回 nil 由调用方使用默认 404。
func (r *Router) resolveNoRoute(c *Ctx, path string) HandlerFunc {
	return r.resolveNotFound(c, path, func(g *RouterGroup) HandlerFunc { return g.noRoute })
}

// resolveNoMethod 同 resolveNoRoute，针对 405。
func (r *Router) resolveNoMethod(c *Ctx, path string) HandlerFunc {
	return r.resolveNotFound(c, path, func(g *RouterGroup) HandlerFunc { return g.noMethod })
}

// resolveNotFound 在匹配 path 的分组中选取 prefix 最长且注册了处理器的那一个，
// 保证最内层的 NoRoute/NoMethod 优先。
func (r *Router) resolveNotFound(c *Ctx, path string, pick func(*RouterGroup) HandlerFunc) HandlerFunc {
	if c.app == nil {
		return nil
	}
	best := HandlerFunc(nil)
	bestLen := -1
	for _, g := range c.app.groups {
		if !groupMatchesPath(g.prefix, path) {
			continue
		}
		if h := pick(g); h != nil && len(g.prefix) > bestLen {
			best = h
			bestLen = len(g.prefix)
		}
	}
	return best
}
