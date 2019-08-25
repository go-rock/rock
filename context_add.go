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
	// c.data = map[string]interface{}{}
	c.data = nil
	c.queryParams = nil
	c.index = -1
	c.handlers = nil
	c.formParsed = false
	c.multipartFormParsed = false
}
