package rock

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

func (m *Router) wrapHandler(h interface{}) (HandlerFunc, error) {
	switch h := h.(type) {
	case HandlerFunc:
		return h, nil
	case func(Context):
		return h, nil
	case func(http.ResponseWriter, *http.Request):
		return func(c Context) {
			h(c.Writer(), c.Request())
		}, nil
	default:
		return nil, fmt.Errorf("unknown handler type: %T", h)
	}
}

// Get filename by viewEngine
func EnsureTemplateName(s string, v ViewEngine) string {
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
