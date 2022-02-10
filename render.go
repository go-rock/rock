package rock

// import "net/http"

// type HTMLRender interface {
// 	GetViewDir() string
// 	SetViewDir(string)
// 	Instance(string, interface{}) Render
// }
// type Render interface {
// 	Render(w http.ResponseWriter, statusCode int)
// }

// //HtmlTemplate use html template middleware
// func (r *RouterGroup) HTMLRender(render HTMLRender) {
// 	r.app.render = render
// }

// func (r *App) GetHTMLRender() HTMLRender {
// 	return r.render
// }

// // Render writes the response headers and calls render.Render to render data.
// func (c *Ctx) Render(statusCode int, r Render) {
// 	r.Render(c.writer, statusCode)
// }
// func (c *Ctx) HTML(name string, status ...int) {
// 	if c.render == nil {
// 		c.String(200, "Please set render")
// 		return
// 	}
// 	defaultStatus := http.StatusOK
// 	if len(status) > 0 {
// 		defaultStatus = status[0]
// 	}

// 	instance := c.render.Instance(name, c.Data())
// 	c.Render(defaultStatus, instance)
// }
