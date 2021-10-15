package render

import (
	"net/http"

	"github.com/doabit/rock"

	"github.com/flosch/pongo2"
)

type ViewRender struct {
	engine *ViewEngine
	name   string
	data   interface{}
}
type ViewEngine struct {
	Config ViewConfig
}

type ViewConfig struct {
	ViewDir   string
	Extension string
	Box       *pongo2.TemplateSet
}

func Default() *ViewEngine {
	config := ViewConfig{
		ViewDir:   "./views/",
		Extension: ".html",
	}
	return New(config)
}

func New(config ViewConfig) *ViewEngine {
	return &ViewEngine{
		Config: config,
	}
}
func (e *ViewEngine) SetViewDir(viewDir string) {
	e.Config.ViewDir = viewDir
}

func (e *ViewEngine) GetViewDir() string {
	return e.Config.ViewDir
}

func (e *ViewEngine) Instance(name string, data interface{}) rock.Render {
	return ViewRender{
		engine: e,
		name:   name,
		data:   data,
	}
}

func (r ViewRender) Render(w http.ResponseWriter, statusCode int) {
	// viewDir := r.engine.Config.ViewDir
	ext := r.engine.Config.Extension
	file := r.name + ext
	data := r.data
	template, err := r.loadFile(file)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	content := convertContext(data)
	err = template.ExecuteWriter(content, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func convertContext(templateData interface{}) pongo2.Context {
	if templateData == nil {
		return nil
	}

	if contextData, isPongoContext := templateData.(pongo2.Context); isPongoContext {
		return contextData
	}

	if contextData, isContextViewData := templateData.(rock.M); isContextViewData {
		return pongo2.Context(contextData)
	}

	return templateData.(map[string]interface{})
}

func (r ViewRender) loadFile(name string) (*pongo2.Template, error) {
	// return nil, nil
	box := r.engine.Config.Box
	// viewDir := r.engine.Config.ViewDir
	// var html string
	var err error
	var template *pongo2.Template
	if box != nil {
		template, err = box.FromCache(name)
		if err != nil {
			return nil, err
		}
	} else {
		template, err = pongo2.FromFile(r.engine.Config.ViewDir + name)
		if err != nil {
			return nil, err
		}
		// html, err = ioutil.ReadFile(viewDir + name + ext)
		// pp.Println("from file")
	}

	// if err != nil {
	// 	http.Error(w, "can't find file "+viewDir+file, http.StatusInternalServerError)
	// 	return
	// }
	return template, nil
}
