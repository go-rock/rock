package rock

import (
	"fmt"
	"math"
	"net/http"
)

const (
	contentType    = "Content-Type"
	acceptLanguage = "Accept-Language"
	abortIndex     = math.MaxInt8 / 2
)

type Ctx2 struct {
	writercache responseWriter
	Request     *http.Request
	Writer      ResponseWriter
	index       int8
	handlers    []HandlerFunc
	data        map[string]interface{}
	render      Render
	// sessions    *sessions.Sessions
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
