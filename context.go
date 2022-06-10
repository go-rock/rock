package rock

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-rock/rock/binding"
)

const (
	contentType    = "Content-Type"
	acceptLanguage = "Accept-Language"
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

		// binding
		Decode(v interface{}, args ...interface{}) (err error)
		ShouldBindJSON(v interface{}, args ...interface{}) (err error)
	}

	Ctx struct {
		app     *App
		request *http.Request
		writer  http.ResponseWriter
		params  Map
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
		// form
		formParsed          bool
		multipartFormParsed bool
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

func (c *Ctx) newContext(w http.ResponseWriter, r *http.Request) *Ctx {
	c.writer = w
	c.request = r
	c.Path = r.URL.Path
	c.Method = r.Method
	c.statusCode = http.StatusOK
	c.index = -1
	c.formParsed = false
	c.multipartFormParsed = false
	return c
}

func (c *Ctx) Request() *http.Request {
	return c.request
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
	return c.request.URL.Query().Get(key)
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

// Form

// ParseForm calls the underlying http.Request ParseForm
// but also adds the URL params to the request Form as if
// they were defined as query params i.e. ?id=13&ok=true but
// does not add the params to the http.Request.URL.RawQuery
// for SEO purposes
func (c *Ctx) ParseForm() error {
	if c.formParsed {
		return nil
	}

	if err := c.request.ParseForm(); err != nil {
		return err
	}

	for key, value := range c.params {
		c.request.Form.Add(key, value.(string))
	}

	c.formParsed = true

	return nil
}

// ParseMultipartForm calls the underlying http.Request ParseMultipartForm
// but also adds the URL params to the request Form as if they were defined
// as query params i.e. ?id=13&ok=true but does not add the params to the
// http.Request.URL.RawQuery for SEO purposes
func (c *Ctx) ParseMultipartForm(maxMemory int64) error {
	if c.multipartFormParsed {
		return nil
	}

	if err := c.request.ParseMultipartForm(maxMemory); err != nil {
		return err
	}

	for key, value := range c.params {
		c.request.Form.Add(key, value.(string))
	}

	c.multipartFormParsed = true

	return nil
}

// Binding

// Decode takes the request and attempts to discover it's content type via
// the http headers and then decode the request body into the provided struct.
// Example if header was "application/json" would decode using
// json.NewDecoder(io.LimitReader(c.request.Body, maxMemory)).Decode(v).
func (c *Ctx) Decode(v interface{}, args ...interface{}) (err error) {
	var maxMemory int64 = 10000
	var includeFormQueryParams bool = false
	if len(args) > 0 {
		result, ok := args[0].(bool)
		if ok {
			includeFormQueryParams = result
		}
	}
	if len(args) > 1 {
		result, ok := args[1].(int)
		if ok {
			maxMemory = int64(result)
		}
	}

	initFormDecoder()

	typ := c.request.Header.Get(ContentType)

	if idx := strings.Index(typ, ";"); idx != -1 {
		typ = typ[:idx]
	}

	switch typ {

	case ApplicationJSON:
		err = json.NewDecoder(io.LimitReader(c.request.Body, maxMemory)).Decode(v)

	case ApplicationXML:
		err = xml.NewDecoder(io.LimitReader(c.request.Body, maxMemory)).Decode(v)

	case ApplicationForm:

		if err = c.ParseForm(); err == nil {
			if includeFormQueryParams {
				err = formDecoder.Decode(v, c.request.Form)
			} else {
				err = formDecoder.Decode(v, c.request.PostForm)
			}
		}

	case MultipartForm:

		if err = c.ParseMultipartForm(maxMemory); err == nil {
			if includeFormQueryParams {
				err = formDecoder.Decode(v, c.request.Form)
			} else {
				err = formDecoder.Decode(v, c.request.MultipartForm.Value)
			}
		}
	}
	return
}

func (c *Ctx) ShouldBindJSON(v interface{}, args ...interface{}) (err error) {
	err = c.Decode(v, args...)
	if err != nil {
		return err
	}
	err = binding.Validate(v)
	return err
}
