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
		// writer
		Write(rawBody []byte) (int, error)
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
		ViewEngine(engine ViewEngine)
		HTML(name string, viewData ...interface{})
		Data() M
		SetData(M)
		Set(key string, value interface{})
		Get(key string) (value interface{}, exists bool)

		GetView() View
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
		// render HTMLRender
		data   M
		values Store
	}
)

func (c *Ctx) GetView() View {
	return c.app.view
}

// func newContext(w http.ResponseWriter, req *http.Request) *Ctx {
// 	return &Ctx{
// 		writer: w,
// 		req:    req,
// 		Path:   req.URL.Path,
// 		Method: req.Method,
// 		index:  -1,
// 	}
// }

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
	_, err := c.Write([]byte(fmt.Sprintf(format, values...)))
	if err != nil {
		http.Error(c.writer, err.Error(), 500)
	}
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

//  +------------------------------------------------------------+
//  | Body (raw) Writers                                         |
//  +------------------------------------------------------------+

// Write writes the data to the connection as part of an HTTP reply.
//
// If WriteHeader has not yet been called, Write calls
// WriteHeader(http.StatusOK) before writing the data. If the Header
// does not contain a Content-Type line, Write adds a Content-Type set
// to the result of passing the initial 512 bytes of written data to
// DetectContentType.
//
// Depending on the HTTP protocol version and the client, calling
// Write or WriteHeader may prevent future reads on the
// Request.Body. For HTTP/1.x requests, handlers should read any
// needed request body data before writing the response. Once the
// headers have been flushed (due to either an explicit Flusher.Flush
// call or writing enough data to trigger a flush), the request body
// may be unavailable. For HTTP/2 requests, the Go HTTP server permits
// handlers to continue to read the request body while concurrently
// writing the response. However, such behavior may not be supported
// by all HTTP/2 clients. Handlers should read before writing if
// possible to maximize compatibility.
func (ctx *Ctx) Write(rawBody []byte) (int, error) {
	return ctx.writer.Write(rawBody)
}
