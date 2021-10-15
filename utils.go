package rock

import (
	"net/http"
)

func (m *Router) wrapHandler(h interface{}) HandlerFunc {
	switch h := h.(type) {
	case HandlerFunc:
		return h
	case func(Context):
		return h
	case func(http.ResponseWriter, *http.Request):
		return func(c Context) {
			h(c.Writer(), c.Request())
		}
	default:
		panic("unknown handler")
	}
}
