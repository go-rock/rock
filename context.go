package rock

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	Context interface {
		Request() *http.Request
		Writer() http.ResponseWriter
		Next()
		// response method
		StatusCode() int
		Status(code int)
		SetHeader(key string, value string)
		Fail(code int, err string)
		String(code int, format string, values ...interface{})
		JSON(code int, obj interface{})
		// request method
		Param(key string) interface{}
	}

	Ctx struct {
		app    *App
		req    *http.Request
		writer http.ResponseWriter
		params Map
		// request info
		Path   string
		Method string
		// response info
		statusCode int
		// middleware
		handlers []HandlerFunc
		index    int
	}
)

func newContext(w http.ResponseWriter, req *http.Request) *Ctx {
	return &Ctx{
		writer: w,
		req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

func (c *Ctx) Request() *http.Request {
	return c.req
}
func (c *Ctx) Writer() http.ResponseWriter {
	return c.writer
}

func (c *Ctx) Next() {

	c.index++
	s := len(c.handlers)
	fmt.Println(c.index, s)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Ctx) StatusCode() int {
	return c.statusCode
}

func (c *Ctx) Status(code int) {
	c.statusCode = code
	c.writer.WriteHeader(code)
}

func (c *Ctx) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (c *Ctx) SetHeader(key string, value string) {
	c.writer.Header().Set(key, value)
}

func (c *Ctx) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Ctx) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.writer, err.Error(), 500)
	}
}

func (c *Ctx) Param(key string) interface{} {
	return c.params[key]
}
