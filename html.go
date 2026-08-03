package rock

import (
	"io"
	"log"
)

// ViewEngine 是模板引擎需要实现的接口，
// 由外部引擎（如 rock-pongo2）提供实现并通过 RegisterView 注册。
type ViewEngine interface {
	Name() string
	Ext() string
	ExecuteWriter(writer io.Writer, filename string, bindingData interface{}) error
	SetViewDir(viewDir string)
	GetViewDir() string
}

// View 持有当前注册的模板引擎。
type View struct {
	Engine ViewEngine
}

// Engine 是 ViewEngine 的别名。
type Engine = ViewEngine

// Register registers a view engine.
func (v *View) Register(e Engine) {
	if v.Engine != nil {
		log.Printf("Engine already exists, replacing the old %q with the new one %q", v.Engine.Name(), e.Name())
	}

	v.Engine = e
}

// Registered reports whether an engine was registered.
func (v *View) Registered() bool {
	return v.Engine != nil
}

// func (v *View) ensureTemplateName(s string) string {
// 	log.Printf("name %s %s", s, v.Engine.Ext())
// 	if s == "" {
// 		return s
// 	}

// 	s = strings.TrimPrefix(s, "/")

// 	if ext := v.Engine.Ext(); ext != "" {
// 		if !strings.HasSuffix(s, ext) {
// 			return s + ext
// 		}
// 	}

// 	return s
// }

// ExecuteWriter calls the correct view Engine's ExecuteWriter func
func (v *View) ExecuteWriter(w io.Writer, filename string, bindingData interface{}) error {
	return v.Engine.ExecuteWriter(w, filename, bindingData)
}

// BlockEngine 是预留的空模板引擎类型。
type BlockEngine struct{}

// HTML Engine
// type HtmlEngine struct {
// 	name string
// }

// func NewHtmlEngine(name string) *HtmlEngine {
// 	return &HtmlEngine{name}
// }

// func (e *HtmlEngine) Name() string {
// 	return e.name
// }

// func (e *HtmlEngine) Render(w io.Writer, tmplName, data interface{}) error {
// 	w.Write([]byte("html engine"))
// 	return nil
// }

// // ExecuteWriter renders a template on "w".
// func (s *HtmlEngine) ExecuteWriter(w io.Writer, tmplName, data interface{}) error {
// 	return s.Render(w, tmplName, data)
// }
