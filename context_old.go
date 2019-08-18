package rock

import (
	"fmt"
	"math"
	"net/http"

	"github.com/plimble/sessions"
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
	render      Renderer
	sessions    *sessions.Sessions
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
