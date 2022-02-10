package rock

import (
	"io"
	"log"
)

type ViewEngine interface {
	Name() string
	Ext() string
	ExecuteWriter(writer io.Writer, filename string, bindingData interface{}) error
	SetViewDir(viewDir string)
	GetViewDir() string
}

type View struct {
	Engine ViewEngine
}

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
	filename = EnsureTemplateName(filename, v.Engine)

	return v.Engine.ExecuteWriter(w, filename, bindingData)
}

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
