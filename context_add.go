package rock

import (
	"net/http"
)

// NewContext returns a new default lars Context object.
func NewContext(l *App) *Ctx {

	c := &Ctx{}

	c.response = newResponse(nil, c)
	return c
}

// RequestStart resets the Context to it's default request state
func (c *Ctx) RequestStart(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response.reset(w)
	// c.params = c.params[0:0]
	// c.queryParams = nil
	c.index = -1
	c.handlers = nil
	c.formParsed = false
	c.multipartFormParsed = false
}

// Text returns the provided string with status code
func (c *Ctx) Text(code int, s string) error {
	return c.TextBytes(code, []byte(s))
}

// TextBytes returns the provided response with status code
func (c *Ctx) TextBytes(code int, b []byte) (err error) {

	// c.response.Header().Set(ContentType, TextPlainCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}
