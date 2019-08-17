package rock

import "net/http"

type Ctx struct {
	writercache responseWriter
	request     *http.Request
	Writer      ResponseWriter
	index       int8
	handlers    []HandlerFunc
	data        map[string]interface{}
	render      Renderer
}

func (c *Ctx) Request() *http.Request {
	return c.request
}
func (c *Ctx) Response() ResponseWriter {
	return c.Writer
}

//Next next middleware
func (c *Ctx) Next() {
	c.index++
	s := int8(len(c.handlers))
	if c.index < s {
		c.handlers[c.index](c)
	}
}
