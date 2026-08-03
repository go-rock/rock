package rock

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/go-rock/rock/binding"
)

// defaultMaxMemory 是请求体解码的默认上限（10MB）。
const defaultMaxMemory = 10 << 20

type (
	// Context 封装单个 HTTP 请求的请求/响应能力，
	// 由框架注入到每个处理函数与中间件中。
	Context interface {
		Application() *App
		ResetRequest(r *http.Request)
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
		GetMethod() string
		GetPath() string
		// render
		// HTML(code int, name string, data interface{})
		ViewEngine(engine ViewEngine)
		HTML(name string, viewData ...interface{})
		Data() M
		SetData(M)
		Set(key string, value interface{})
		Get(key string) (value interface{}, exists bool)

		Values() *Store
		ViewData(key string, value interface{})
		GetViewData() map[string]interface{}

		GetView() View

		// binding
		Decode(v interface{}, args ...interface{}) (err error)
		ShouldBind(v interface{}, args ...interface{}) (err error)
		ShouldBindJSON(v interface{}) error
		ShouldBindQuery(v interface{}) error

		// Methods
		Abort()
		AbortWithStatusJSON(code int, jsonObj interface{})
		Redirect(url string)
		Attachment(r io.Reader, filename string) (err error)
		Inline(r io.Reader, filename string) (err error)
		File(filepath string)

		// Post body

		MustPostInt(key string, d int) int
		MustPostString(key string, d string) string

		MustParamInt(name string, d int) int

		MustQueryInt(name string, d int) int
		MustQueryString(name string, d string) string

		FormFile(name string) (*multipart.FileHeader, error)

		// 文件上传功能
		SaveSingleFile(name string, config *FileUploadConfig) (*FileInfo, error)
		SaveMultipleFiles(name string, config *FileUploadConfig) ([]*FileInfo, error)
		UploadSingleImage(name string) (*FileInfo, error)
		UploadSingleDocument(name string) (*FileInfo, error)
		UploadMultipleImages(name string) ([]*FileInfo, error)

		// 日志功能
		App() *App
		Logger() *RockLogger
		LogDebug(msg string, args ...interface{})
		LogInfo(msg string, args ...interface{})
		LogWarn(msg string, args ...interface{})
		LogError(msg string, args ...interface{})
	}

	// Ctx 是 Context 接口的默认实现，通过 sync.Pool 复用。
	Ctx struct {
		app     *App
		request *http.Request
		writer  http.ResponseWriter
		params  Map
		// request info
		Path   string
		Method string
		// response info
		statusCode    int
		headerWritten bool // 响应头是否已写入，避免重复调用 WriteHeader
		// middleware
		handlers []HandlerFunc
		index    int
		aborted  bool // Abort() 标志，终止中间件链（不依赖固定 index 哨兵，支持任意长链）
		// render
		// render HTMLRender
		data   M
		values Store
		// form
		formParsed          bool
		multipartFormParsed bool
	}
)

func (c *Ctx) Application() *App {
	return c.app
}

func (c *Ctx) ResetRequest(r *http.Request) {
	c.request = r
	c.Path = r.URL.Path
	c.Method = r.Method
	c.statusCode = http.StatusOK
	c.headerWritten = false
	c.index = -1
	c.aborted = false
	c.formParsed = false
	c.multipartFormParsed = false
	// 重置状态数据以防止污染
	c.params = nil
	c.data = nil
	c.handlers = nil // 清掉上个请求的处理器链，避免复用 Ctx 时误执行旧 handler
	// 每个请求都应使用全新的 Store，
	// 否则对象池复用时会泄漏上个请求通过 Values/ViewData 写入的数据。
	// 同一请求内的中间件通过共享的 c 访问 values，不需要跨请求保留。
	c.values = Store{}
}

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
	c.headerWritten = false
	c.index = -1
	c.aborted = false
	c.formParsed = false
	c.multipartFormParsed = false
	// 重置状态数据以防止污染
	c.params = nil
	c.data = nil
	// 为每个请求分配全新的 Store，
	// 避免对象池复用时带上个请求通过 Values/ViewData 写入的数据。
	c.values = Store{}
	return c
}

func (c *Ctx) Request() *http.Request {
	return c.request
}

func (c *Ctx) Writer() http.ResponseWriter {
	return c.writer
}

func (c *Ctx) Next() {
	// 安全检查：确保handlers不为空且未被abort
	if c.handlers == nil || c.aborted {
		return
	}

	c.index++
	s := len(c.handlers)

	// 执行下一个中间件或处理器
	for c.index < s {
		// 检查是否已经被abort
		if c.aborted {
			return
		}

		handler := c.handlers[c.index]
		if handler == nil {
			c.index++
			continue
		}

		// 执行中间件/处理器
		handler(c)

		// 检查是否在执行过程中被abort
		if c.aborted {
			return
		}

		c.index++
	}
}

func (c *Ctx) StatusCode() int {
	return c.statusCode
}

// Status 设置响应状态码，但不会立即写入响应头。
// 状态码会在首次写入响应体时（见 writeHeader）才真正发送，
// 因此允许在写出 body 之前多次修改状态码，也不会触发重复的 WriteHeader。
func (c *Ctx) Status(code int) {
	c.statusCode = code
}

// writeHeader 在首次写入响应体前，把当前状态码写入响应头。
// 每个请求只执行一次，避免重复调用 WriteHeader 产生的 "superfluous" 告警。
func (c *Ctx) writeHeader() {
	if c.headerWritten {
		return
	}
	c.writer.WriteHeader(c.statusCode)
	c.headerWritten = true
}

func (c *Ctx) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.aborted = true
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
		WriteError(c, 500, NewAppError(ErrInternalServer, "Failed to write response"))
	}
}

func (c *Ctx) JSON(code int, obj interface{}) {
	// 确保Content-Type设置为application/json
	c.SetHeader("Content-Type", "application/json")

	// code 参数是响应状态码的最终来源，先写头再写 body
	c.Status(code)
	c.writeHeader()

	// 使用统一的JSON响应写入方法
	if err := writeJSONResponse(c.writer, obj); err != nil {
		WriteError(c, 500, NewAppError(ErrInternalServer, "Failed to encode JSON response"))
	}
}

// XML marshals provided interface + returns XML + status code
func (c *Ctx) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}

	c.writer.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.Status(code)
	c.writeHeader()

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
	values, exists := c.request.URL.Query()[name]
	if !exists || len(values) == 0 {
		return "", false
	}
	return values[0], true
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
	// c.values.Set(key, value)
}

func (c *Ctx) Get(key string) (value interface{}, exists bool) {
	value, exists = c.data[key]
	return
}

// Body (raw) Writers
func (ctx *Ctx) Write(rawBody []byte) (int, error) {
	ctx.writeHeader()
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
	// 默认 10MB，避免过小的 body 上限导致绑定失败
	var maxMemory int64 = defaultMaxMemory
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
		err = c.decodeBody(func(b []byte) error {
			return json.NewDecoder(bytes.NewReader(b)).Decode(v)
		}, maxMemory)

	case ApplicationXML:
		err = c.decodeBody(func(b []byte) error {
			return xml.NewDecoder(bytes.NewReader(b)).Decode(v)
		}, maxMemory)

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

// decodeBody 读取请求体（带上限）并调用 decode 解析。
// 读取后把 body 缓存回 request.Body，使同一请求内可以多次 Decode/ShouldBind。
func (c *Ctx) decodeBody(decode func([]byte) error, maxMemory int64) error {
	b, err := io.ReadAll(io.LimitReader(c.request.Body, maxMemory))
	if err != nil {
		return err
	}
	// 缓存 body，支持重复读取
	c.request.Body = io.NopCloser(bytes.NewReader(b))
	return decode(b)
}

func (c *Ctx) ShouldBind(v interface{}, args ...interface{}) (err error) {
	err = c.Decode(v, args...)
	if err != nil {
		return err
	}
	err = binding.Validate(v)
	return err
}

// ShouldBindJSON 强制按 JSON 绑定请求体并校验，不依赖 Content-Type。
func (c *Ctx) ShouldBindJSON(v interface{}) error {
	err := c.decodeBody(func(b []byte) error {
		return json.NewDecoder(bytes.NewReader(b)).Decode(v)
	}, defaultMaxMemory)
	if err != nil {
		return err
	}
	return binding.Validate(v)
}

// ShouldBindQuery 将 URL 查询参数绑定到结构体并校验。
func (c *Ctx) ShouldBindQuery(v interface{}) error {
	initFormDecoder()
	if err := formDecoder.Decode(v, c.request.URL.Query()); err != nil {
		return err
	}
	return binding.Validate(v)
}

// Redirect to
func (c *Ctx) Redirect(url string) {
	c.writer.Header().Set("Location", url)
	c.SetHeader("Content-Type", "text/plain")
	c.Status(http.StatusFound)
	c.Write([]byte("Redirecting to: " + url))
}

func (c *Ctx) Abort() {
	c.aborted = true
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
// 注意：只有当配置了 TrustProxyHeaders 时才信任这些头，
// 否则客户端可以伪造它们来绕过基于 IP 的限流/封禁/审计。
func (c *Ctx) ClientIP() (clientIP string) {
	trustProxy := c.app != nil && c.app.config != nil && c.app.config.TrustProxyHeaders

	if trustProxy {
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
	}

	clientIP, _, _ = net.SplitHostPort(strings.TrimSpace(c.request.RemoteAddr))

	return
}

// Attachment is a helper method for returning an attachement file
// to be downloaded, if you with to open inline see function
func (c *Ctx) Attachment(r io.Reader, filename string) (err error) {
	c.writer.Header().Set(ContentDisposition, "attachment;filename="+filename)
	c.writer.Header().Set(ContentType, detectContentType(filename))
	c.Status(http.StatusOK)
	c.writeHeader()

	_, err = io.Copy(c.writer, r)

	return
}

// Inline is a helper method for returning a file inline to
// be rendered/opened by the browser
func (c *Ctx) Inline(r io.Reader, filename string) (err error) {
	c.writer.Header().Set(ContentDisposition, "inline;filename="+filename)
	c.writer.Header().Set(ContentType, detectContentType(filename))
	c.Status(http.StatusOK)
	c.writeHeader()

	_, err = io.Copy(c.writer, r)

	return
}

// File 直接返回磁盘上的单个文件，
// 由 http.ServeFile 处理 Content-Type、Range 请求与 404。
func (c *Ctx) File(filepath string) {
	http.ServeFile(c.Writer(), c.Request(), filepath)
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
		return d
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

// 文件上传功能实现

func (c *Ctx) SaveSingleFile(name string, config *FileUploadConfig) (*FileInfo, error) {
	return SaveSingleFile(c, name, config)
}

func (c *Ctx) SaveMultipleFiles(name string, config *FileUploadConfig) ([]*FileInfo, error) {
	return SaveMultipleFiles(c, name, config)
}

func (c *Ctx) UploadSingleImage(name string) (*FileInfo, error) {
	return UploadSingleImage(c, name)
}

func (c *Ctx) UploadSingleDocument(name string) (*FileInfo, error) {
	return UploadSingleDocument(c, name)
}

func (c *Ctx) UploadMultipleImages(name string) ([]*FileInfo, error) {
	return UploadMultipleImages(c, name)
}

func (c *Ctx) MustParamInt(name string, d int) int {
	val, ok := c.Param(name).(string)
	if !ok || val == "" {
		return d
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return d
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
		return d
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

// Values returns the current "user" storage.
// Named path parameters and any optional data can be saved here.
// This storage, as the whole context, is per-request lifetime.
//
// You can use this function to Set and Get local values
// that can be used to share information between handlers and middleware.
func (ctx *Ctx) Values() *Store {
	return &ctx.values
}

// Set view data by key and value
func (ctx *Ctx) ViewData(key string, value interface{}) {
	viewDataContextKey := ctx.app.ConfigurationReadOnly().GetViewDataContextKey()
	if key == "" {
		ctx.values.Set(viewDataContextKey, value)
		return
	}

	v := ctx.values.Get(viewDataContextKey)
	if v == nil {
		ctx.values.Set(viewDataContextKey, M{key: value})
		return
	}

	if data, ok := v.(M); ok {
		data[key] = value
	}
}

// GetViewData returns the values registered by `context#ViewData`.
// The return value is `map[string]interface{}`, this means that
// if a custom struct registered to ViewData then this function
// will try to parse it to map, if failed then the return value is nil
// A check for nil is always a good practise if different
// kind of values or no data are registered via `ViewData`.
//
// Similarly to `viewData := ctx.Values().Get("rock.view.data")` or
// `viewData := ctx.Values().Get(ctx.Application().ConfigurationReadOnly().GetViewDataContextKey())`.
func (ctx *Ctx) GetViewData() map[string]interface{} {
	if v := ctx.values.Get(ctx.app.ConfigurationReadOnly().GetViewDataContextKey()); v != nil {
		// if pure map[string]interface{}
		if viewData, ok := v.(M); ok {
			return viewData
		}

		// if struct, convert it to map[string]interface{}
		if structs.IsStruct(v) {
			return structs.Map(v)
		}
	}

	// if no values found, then return nil
	return nil
}

// App 返回应用实例
func (c *Ctx) App() *App {
	return c.app
}

// Logger 返回应用日志器
func (c *Ctx) Logger() *RockLogger {
	if c.app != nil {
		return c.app.logger
	}
	return nil
}

// LogDebug 记录调试日志
func (c *Ctx) LogDebug(msg string, args ...interface{}) {
	if logger := c.Logger(); logger != nil {
		logger.Debugf(msg, args...)
	}
}

// LogInfo 记录信息日志
func (c *Ctx) LogInfo(msg string, args ...interface{}) {
	if logger := c.Logger(); logger != nil {
		logger.Infof(msg, args...)
	}
}

// LogWarn 记录警告日志
func (c *Ctx) LogWarn(msg string, args ...interface{}) {
	if logger := c.Logger(); logger != nil {
		logger.Warnf(msg, args...)
	}
}

// LogError 记录错误日志
func (c *Ctx) LogError(msg string, args ...interface{}) {
	if logger := c.Logger(); logger != nil {
		logger.Errorf(msg, args...)
	}
}

// GetMethod 返回请求方法
func (c *Ctx) GetMethod() string {
	return c.Method
}

// GetPath 返回请求路径
func (c *Ctx) GetPath() string {
	return c.Path
}
