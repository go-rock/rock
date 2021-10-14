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
		Status(code int)
		SetHeader(key string, value string)
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
		StatusCode int
	}
)

func newContext(w http.ResponseWriter, req *http.Request) *Ctx {
	return &Ctx{
		writer: w,
		req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

func (c *Ctx) Request() *http.Request {
	return c.req
}
func (c *Ctx) Writer() http.ResponseWriter {
	return c.writer
}

func (c *Ctx) Next() {

}

func (c *Ctx) Status(code int) {
	c.StatusCode = code
	c.writer.WriteHeader(code)
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
