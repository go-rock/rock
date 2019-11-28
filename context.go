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
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

const (
	contentType    = "Content-Type"
	acceptLanguage = "Accept-Language"
	abortIndex     = math.MaxInt8 / 2
)

type Ctx struct {
	writercache         Response
	request             *http.Request
	response            *Response
	queryParams         url.Values
	parent              Context
	index               int
	handlers            []HandlerFunc
	data                map[string]interface{}
	render              HTMLRender
	formParsed          bool
	multipartFormParsed bool
}

// BaseContext returns the underlying context object App uses internally.
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

func (a *App) createContext(w http.ResponseWriter, r *http.Request) *Ctx {
	c := a.pool.Get().(*Ctx)
	c.writercache.reset(w)
	c.request = r
	c.index = -1
	c.data = nil
	c.render = a.render

	return c
}

//String response with text/html; charset=UTF-8 Content type
func (c *Ctx) String(status int, format string, val ...interface{}) {
	c.response.Header().Set(contentType, "text/html; charset=UTF-8")
	c.response.WriteHeader(status)

	buf := bufPool.Get()
	defer bufPool.Put(buf)

	if len(val) == 0 {
		buf.WriteString(format)
	} else {
		buf.WriteString(fmt.Sprintf(format, val...))
	}

	c.response.Write(buf.Bytes())
}

//HTML render template engine
// func (c *Ctx) HTML(name string, data interface{}) {
// 	if c.render == nil {
// 		c.String(200, "Not implement render")
// 	} else {
// 		c.render.Render(c.Writer, name, data)
// 	}
// }

func (c *Ctx) SetHandlers(handlers []HandlerFunc) {
	c.handlers = handlers
}

func (c *Ctx) Redirect(url string) {
	c.response.Header().Set("Location", url)
	c.response.WriteHeader(http.StatusFound)
	c.response.Write([]byte("Redirecting to: " + url))
}

func (c *Ctx) Abort() {
	c.index = abortIndex
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *Ctx) AbortWithStatus(code int) {
	c.response.WriteHeader(code)
	c.Abort()
}

// AbortWithStatusJSON calls `Abort()` and then `JSON` internally.
// This method stops the chain, writes the status code and return a JSON body.
// It also sets the Content-Type as "application/json".
func (c *Ctx) AbortWithStatusJSON(code int, jsonObj interface{}) {
	c.Abort()
	c.JSON(code, jsonObj)
}

func (c *Ctx) GetQuery(name string) (string, bool) {
	if val := c.QueryParams().Get(name); val != "" {
		return val, true
	} else {
		return "", false
	}
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

// QueryParams returns the http.Request.URL.Query() values
// this function is not for convenience, but rather performance
// URL.Query() reparses the RawQuery every time it's called, but this
// function will cache the initial parsing so it doesn't have to reparse;
// which is useful if when accessing these Params from multiple middleware.
func (c *Ctx) QueryParams() url.Values {

	if c.queryParams != nil {
		return c.queryParams
	}

	c.queryParams = c.request.URL.Query()

	return c.queryParams
}

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

	for _, entry := range c.params().Keys {
		c.request.Form.Add(entry, c.Param(entry))
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

	for _, entry := range c.params().Keys {
		c.request.Form.Add(entry, c.Param(entry))
	}

	c.multipartFormParsed = true

	return nil
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
func (c *Ctx) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c.parent)
		c.index++
	}
	// s := int(len(c.handlers))
	// if c.index < s {
	// 	c.handlers[c.index](c.parent)
	// }
}

// Params returns the url param by name.
func (c *Ctx) Param(name string) string {
	return chi.URLParam(c.request, name)
}

func (c *Ctx) params() chi.RouteParams {
	params := chi.RouteContext(c.request.Context()).URLParams
	return params
}

func (c *Ctx) MustParamInt(name string, d int) int {
	val := c.Param(name)
	if val == "" {
		return d
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err.Error())
	}

	return i
}

// http response helpers

// JSON marshals provided interface + returns JSON + status code
func (c *Ctx) JSON(code int, i interface{}) (err error) {

	b, err := json.Marshal(i)
	if err != nil {
		return err
	}

	return c.JSONBytes(code, b)
}

// JSONBytes returns provided JSON response with status code
func (c *Ctx) JSONBytes(code int, b []byte) (err error) {

	c.response.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

// JSONP sends a JSONP response with status code and uses `callback` to construct
// the JSONP payload.
func (c *Ctx) JSONP(code int, i interface{}, callback string) (err error) {

	b, e := json.Marshal(i)
	if e != nil {
		err = e
		return
	}

	c.response.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)

	if _, err = c.response.Write([]byte(callback + "(")); err == nil {

		if _, err = c.response.Write(b); err == nil {
			_, err = c.response.Write([]byte(");"))
		}
	}

	return
}

// XML marshals provided interface + returns XML + status code
func (c *Ctx) XML(code int, i interface{}) error {

	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}

	return c.XMLBytes(code, b)
}

// XMLBytes returns provided XML response with status code
func (c *Ctx) XMLBytes(code int, b []byte) (err error) {

	c.response.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)

	if _, err = c.response.Write([]byte(xml.Header)); err == nil {
		_, err = c.response.Write(b)
	}

	return
}

// Text returns the provided string with status code
func (c *Ctx) Text(code int, s string) error {
	return c.TextBytes(code, []byte(s))
}

// TextBytes returns the provided response with status code
func (c *Ctx) TextBytes(code int, b []byte) (err error) {

	c.response.Header().Set(ContentType, TextPlainCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

// http request helpers

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

// AcceptedLanguages returns an array of accepted languages denoted by
// the Accept-Language header sent by the browser
// NOTE: some stupid browsers send in locales lowercase when all the rest send it properly
func (c *Ctx) AcceptedLanguages(lowercase bool) []string {

	var accepted string

	if accepted = c.request.Header.Get(AcceptedLanguage); accepted == blank {
		return []string{}
	}

	options := strings.Split(accepted, ",")
	l := len(options)

	language := make([]string, l)

	if lowercase {

		for i := 0; i < l; i++ {
			locale := strings.SplitN(options[i], ";", 2)
			language[i] = strings.ToLower(strings.Trim(locale[0], " "))
		}
	} else {

		for i := 0; i < l; i++ {
			locale := strings.SplitN(options[i], ";", 2)
			language[i] = strings.Trim(locale[0], " ")
		}
	}

	return language
}

// Stream provides HTTP Streaming
func (c *Ctx) Stream(step func(w io.Writer) bool) {
	w := c.response
	clientGone := w.CloseNotify()

	for {
		select {
		case <-clientGone:
			return
		default:
			keepOpen := step(w)
			w.Flush()
			if !keepOpen {
				return
			}
		}
	}
}

// Attachment is a helper method for returning an attachement file
// to be downloaded, if you with to open inline see function
func (c *Ctx) Attachment(r io.Reader, filename string) (err error) {

	c.response.Header().Set(ContentDisposition, "attachment;filename="+filename)
	c.response.Header().Set(ContentType, detectContentType(filename))
	c.response.WriteHeader(http.StatusOK)

	_, err = io.Copy(c.response, r)

	return
}

// Inline is a helper method for returning a file inline to
// be rendered/opened by the browser
func (c *Ctx) Inline(r io.Reader, filename string) (err error) {

	c.response.Header().Set(ContentDisposition, "inline;filename="+filename)
	c.response.Header().Set(ContentType, detectContentType(filename))
	c.response.WriteHeader(http.StatusOK)

	_, err = io.Copy(c.response, r)

	return
}

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

// Render

func (c *Ctx) HTMLRender(render HTMLRender) {
	c.render = render
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

// func (c *Ctx) MustPostFloat64(key string, d float64) float64 {
// 	val := c.Request().PostFormValue(key)
// 	if val == "" {
// 		return d
// 	}
// 	f, err := strconv.ParseFloat(c.Request().URL.Query().Get(key), 64)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return f
// }

func (c *Ctx) MustPostString(key, d string) string {
	val := c.Request().PostFormValue(key)
	if val == "" {
		return d
	}

	return val
}

// func (c *Ctx) MustPostStrings(key string, d []string) []string {
// 	if c.Request().PostForm == nil {
// 		c.Request().ParseForm()
// 	}

// 	val := c.Request().PostForm[key]
// 	if len(val) == 0 {
// 		return d
// 	}

// 	return val
// }

// func (c *Ctx) MustPostTime(key string, layout string, d time.Time) time.Time {
// 	val := c.Request().PostFormValue(key)
// 	if val == "" {
// 		return d
// 	}
// 	t, err := time.Parse(layout, c.Request().URL.Query().Get(key))
// 	if err != nil {
// 		panic(err)
// 	}

// 	return t
// }

func (c *Ctx) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.request.FormFile(name)
	return fh, err
}
