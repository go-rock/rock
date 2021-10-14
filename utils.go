package rock

import (
	"fmt"
	"net/http"

	"github.com/k0kubun/pp"
)

func (m *Router) wrapHandler(h interface{}) HandlerFunc {
	switch h := h.(type) {
	case HandlerFunc:
		fmt.Println("handler func")
		return h
	case func(Context):
		fmt.Println("handler context func")
		return h
	case func(http.ResponseWriter, *http.Request):
		return func(c Context) {
			h(c.Writer(), c.Request())
		}
	default:
		pp.Println(h.(HandlerFunc))
		panic("unknown handler")
	}
}
