package rock

import (
	"net/http"
)

type M map[string]interface{}

//Renderer html render interface
// type Renderer interface {
// 	Render(w http.ResponseWriter, name string, data interface{})
// }

// //HtmlTemplate use html template middleware
// func (app *App) HtmlTemplate(render Renderer) {
// 	app.render = render
// }

type HTMLRender interface {
	GetViewDir() string
	SetViewDir(string)
	Instance(string, interface{}) Render
}
type Render interface {
	Render(w http.ResponseWriter, statusCode int)
}

//HtmlTemplate use html template middleware
func (r *App) HTMLRender(render HTMLRender) {
	r.render = render
}

func (r *App) GetHTMLRender() HTMLRender {
	return r.render
}

// import (
// 	"net/http"
// )
// type HTMLRender interface {
// 	// Instance returns an HTML instance.
// 	Instance(string, interface{}) Render
// }
// type Render interface {
// 	// Render writes data with custom ContentType.
// 	Render(http.ResponseWriter) error
// 	// WriteContentType writes custom ContentType.
// 	WriteContentType(w http.ResponseWriter)
// }

// func writeContentType(w http.ResponseWriter, value []string) {
// 	header := w.Header()
// 	if val := header["Content-Type"]; len(val) == 0 {
// 		header["Content-Type"] = value
// 	}
// }

func (ctx *Ctx) Status(status int) {
	ctx.response.WriteHeader(status)
}

// Render writes the response headers and calls render.Render to render data.
func (c *Ctx) Render(statusCode int, r Render) {
	// c.Status(code)
	r.Render(c.response, statusCode)

	// if !bodyAllowedForStatus(code) {
	// 	r.WriteContentType(c.Writer)
	// 	c.Writer.WriteHeaderNow()
	// 	return
	// }

	// if err := r.Render(c.Writer); err != nil {
	// 	panic(err)
	// }
}
func (c *Ctx) HTML(name string, status ...int) {
	if c.render == nil {
		c.String(200, "Please set render")
		return
	}
	defaultStatus := http.StatusOK
	if len(status) > 0 {
		defaultStatus = status[0]
	}

	instance := c.render.Instance(name, c.Data())
	c.Render(defaultStatus, instance)
	// if err := c.App.Render.HTML(name, c, defaultStatus); err != nil {
	// 	panic(err)
	// }
}
