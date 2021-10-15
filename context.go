package rock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
		Query(key string) string
		QueryInt(key string) int
		// render
		// HTML(code int, name string, data interface{})
		HTML(name string, status ...int)
		Data() M
		SetData(M)
		Set(key string, value interface{})
		Get(key string) (value interface{}, exists bool)
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
		// render
		render HTMLRender
		data   M
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

// Query

func (c *Ctx) Query(key string) string {
	return c.req.URL.Query().Get(key)
}

func (c *Ctx) QueryInt(key string) int {
	value := c.Query(key)
	result, _ := strconv.Atoi(value)
	return result
}

// Context data

func (c *Ctx) Data() M {
	return c.data
}

// set all data
func (c *Ctx) SetData(data M) {
	c.data = data
}

func (c *Ctx) Set(key string, value interface{}) {
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	c.data[key] = value
}

func (c *Ctx) Get(key string) (value interface{}, exists bool) {
	value, exists = c.data[key]
	return
}
