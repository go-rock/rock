package rock

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
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

// Get filename by viewEngine
func EnsureTemplateName(s string, v ViewEngine) string {
	log.Printf("name %s %s", s, v.Ext())
	if s == "" {
		return s
	}

	s = strings.TrimPrefix(s, "/")

	if ext := v.Ext(); ext != "" {
		if !strings.HasSuffix(s, ext) {
			return s + ext
		}
	}

	return s
}

func detectContentType(filename string) (t string) {
	if t = mime.TypeByExtension(filepath.Ext(filename)); t == "" {
		t = OctetStream
	}
	return
}
