package rock

import (
	"net/http"
)

type M map[string]interface{}

//Renderer html render interface
type Renderer interface {
	Render(w http.ResponseWriter, name string, data interface{})
}

//HtmlTemplate use html template middleware
func (app *App) HtmlTemplate(render Renderer) {
	app.render = render
}
