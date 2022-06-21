package rock

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-rock/rock/binding"
)

const (
	abortIndex = math.MaxInt8 / 2
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
		XML(int, interface{}) error
		// request method
		Param(key string) interface{}
		Query(key string) string
		GetQuery(name string) (string, bool)
		QueryInt(key string) int

		ParseForm() error
		ParseMultipartForm(maxMemory int64) error

		ClientIP() (clientIP string)
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
		ShouldBind(v interface{}, args ...interface{}) (err error)

		// Methods
		Abort()
		AbortWithStatusJSON(code int, jsonObj interface{})
		Redirect(url string)

		// Post body

		MustPostInt(key string, d int) int
		MustPostString(key string, d string) string

		MustParamInt(name string, d int) int

		MustQueryInt(name string, d int) int
		MustQueryString(name string, d string) string

		FormFile(name string) (*multipart.FileHeader, error)
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

// XML marshals provided interface + returns XML + status code
func (c *Ctx) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}

	c.writer.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.writer.WriteHeader(code)

	if _, err = c.writer.Write([]byte(xml.Header)); err == nil {
		_, err = c.writer.Write(b)
	}

	return
}

func (c *Ctx) Param(key string) interface{} {
	return c.params[key]
}

// Query

func (c *Ctx) Query(key string) string {
	return c.request.URL.Query().Get(key)
}

func (c *Ctx) GetQuery(name string) (string, bool) {
	if val := c.Query(name); val != "" {
		return val, true
	} else {
		return "", false
	}
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

//  Body (raw) Writers
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

func (c *Ctx) ShouldBind(v interface{}, args ...interface{}) (err error) {
	err = c.Decode(v, args...)
	if err != nil {
		return err
	}
	err = binding.Validate(v)
	return err
}

// Redirect to
func (c *Ctx) Redirect(url string) {
	c.writer.Header().Set("Location", url)
	c.writer.WriteHeader(http.StatusFound)
	c.String(200, "Redirecting to: "+url)
}

func (c *Ctx) Abort() {
	c.index = abortIndex
}

// AbortWithStatusJSON calls `Abort()` and then `JSON` internally.
// This method stops the chain, writes the status code and return a JSON body.
// It also sets the Content-Type as "application/json".
func (c *Ctx) AbortWithStatusJSON(code int, jsonObj interface{}) {
	c.Abort()
	c.JSON(code, jsonObj)
}

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
func (c *Ctx) ClientIP() (clientIP string) {
	var values []string

	if values, _ = c.request.Header[XRealIP]; len(values) > 0 {

		clientIP = strings.TrimSpace(values[0])
		if clientIP != blank {
			return
		}
	}

	if values, _ = c.request.Header[XForwardedFor]; len(values) > 0 {
		clientIP = values[0]

		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}

		clientIP = strings.TrimSpace(clientIP)
		if clientIP != blank {
			return
		}
	}

	clientIP, _, _ = net.SplitHostPort(strings.TrimSpace(c.request.RemoteAddr))

	return
}

// Attachment is a helper method for returning an attachement file
// to be downloaded, if you with to open inline see function
func (c *Ctx) Attachment(r io.Reader, filename string) (err error) {
	c.writer.Header().Set(ContentDisposition, "attachment;filename="+filename)
	c.writer.Header().Set(ContentType, detectContentType(filename))
	c.writer.WriteHeader(http.StatusOK)

	_, err = io.Copy(c.writer, r)

	return
}

// Inline is a helper method for returning a file inline to
// be rendered/opened by the browser
func (c *Ctx) Inline(r io.Reader, filename string) (err error) {
	c.writer.Header().Set(ContentDisposition, "inline;filename="+filename)
	c.writer.Header().Set(ContentType, detectContentType(filename))
	c.writer.WriteHeader(http.StatusOK)

	_, err = io.Copy(c.writer, r)

	return
}

// form params

/////////////////////////

func (c *Ctx) MustPostInt(key string, d int) int {
	val := c.Request().PostFormValue(key)
	if val == "" {
		return d
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err.Error())
	}

	return i
}

func (c *Ctx) MustPostString(key, d string) string {
	val := c.Request().PostFormValue(key)
	if val == "" {
		return d
	}

	return val
}

func (c *Ctx) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.request.FormFile(name)
	return fh, err
}

func (c *Ctx) MustParamInt(name string, d int) int {
	val := c.Param(name).(string)
	if val == "" {
		return d
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err.Error())
	}

	return i
}

func (c *Ctx) MustQueryInt(name string, d int) int {
	val, bool := c.GetQuery(name)
	if !bool {
		return d
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err.Error())
	}

	return i
}

func (c *Ctx) MustQueryString(name string, d string) string {
	val, bool := c.GetQuery(name)
	if !bool {
		return d
	}
	return val
}
