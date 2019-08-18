package rock

import (
	"net/http"
	"reflect"
	"runtime"
)

// GetContext is a helper method for retrieving the Context object from
// the ResponseWriter when using native go hanlders.
// NOTE: this will panic if fed an http.ResponseWriter not provided by lars's
// chaining.
func GetContext(w http.ResponseWriter) Context {
	return w.(*Response).context
}

var NativeChainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	c := GetContext(w)
	b := c.BaseContext()

	b.request = r

	if b.index+1 < len(b.handlers) {
		c.Next()
	}
})

// wrapHandler wraps Handler type
func (l *App) wrapHandler(h Handler) HandlerFunc {

	switch h := h.(type) {
	case HandlerFunc:
		return h

	case func(Context):
		return h

	case http.Handler, http.HandlerFunc:
		return func(c Context) {

			ctx := c.BaseContext()

			if h.(http.Handler).ServeHTTP(ctx.response, ctx.request); ctx.response.status != http.StatusOK || ctx.response.committed {
				return
			}

			if ctx.index+1 < len(ctx.handlers) {
				c.Next()
			}
		}

	case func(http.ResponseWriter, *http.Request):
		return func(c Context) {

			ctx := c.BaseContext()

			if h(ctx.response, ctx.request); ctx.response.status != http.StatusOK || ctx.response.committed {
				return
			}

			if ctx.index+1 < len(ctx.handlers) {
				c.Next()
			}
		}

	case func(http.ResponseWriter, *http.Request, http.Handler):

		return func(c Context) {
			ctx := c.BaseContext()

			h(ctx.response, ctx.request, NativeChainHandler)
		}

	case func(http.Handler) http.Handler:

		hf := h(NativeChainHandler)

		return func(c Context) {
			ctx := c.BaseContext()

			hf.ServeHTTP(ctx.response, ctx.request)
		}

	default:
		if fn, ok := l.customHandlersFuncs[reflect.TypeOf(h)]; ok {
			return func(c Context) {
				fn(c, h)
			}
		}

		panic("unknown handler")
	}
}

// wrapHandlerWithName wraps Handler type and returns it including it's name before wrapping
func (l *App) wrapHandlerWithName(h Handler) (chain HandlerFunc, handlerName string) {

	chain = l.wrapHandler(h)
	handlerName = runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	return
}
