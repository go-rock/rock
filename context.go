package rock

import (
	"net/http"

	"github.com/go-chi/chi"
)

type Ctx struct {
	writercache         Response
	request             *http.Request
	response            *Response
	parent              Context
	index               int
	handlers            []HandlerFunc
	data                map[string]interface{}
	render              Renderer
	formParsed          bool
	multipartFormParsed bool
}

// BaseContext returns the underlying context object LARS uses internally.
// used when overriding the context object
func (c *Ctx) BaseContext() *Ctx {
	return c
}

func (c *Ctx) Request() *http.Request {
	return c.request
}
func (c *Ctx) Response() *Response {
	return c.response
}

//Next next middleware
func (c *Ctx) Next() {
	c.index++
	s := int(len(c.handlers))
	if c.index < s {
		// pp.Println(s, c)
		c.handlers[c.index](c.parent)
		// c.handlers[c.index](c.parent)
	}
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in github.
// func (c *Ctx) Next() {
// 	c.index++
// 	c.handlers[c.index](c.parent)
// }

// BaseContext returns the underlying context object LARS uses internally.
// used when overriding the context object
func (c *Ctx) Param(name string) string {
	return chi.URLParam(c.request, name)
}
