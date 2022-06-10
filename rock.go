package rock

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-playground/form/v4"
	log "github.com/kataras/golog"
)

var (
	formDecoder     *form.Decoder
	formDecoderInit sync.Once
)

type App struct {
	*RouterGroup
	router *Router
	pool   sync.Pool
	groups []*RouterGroup

	// template
	config *Configuration
	view   View
}

func (app *App) GetView() View {
	return app.view
}

func New() *App {
	config := DefaultConfiguration()
	app := &App{
		config: &config,
		router: NewRouter(),
	}
	app.RouterGroup = &RouterGroup{app: app}
	app.groups = []*RouterGroup{app.RouterGroup}
	app.pool.New = func() interface{} {
		return app.allocateContext()
	}
	return app
}

func (app *App) allocateContext() *Ctx {
	return &Ctx{app: app}
}

func (app *App) createContext(w http.ResponseWriter, r *http.Request) *Ctx {
	c := app.pool.Get().(*Ctx)
	c.newContext(w, r)
	return c
}

// Run defines the method to start a http server
func (app *App) Run(args ...string) (err error) {
	addr := ":8989"
	if len(args) > 0 {
		addr = args[0]
	}
	debugPrint("Rock running on http://localhost%s", addr)
	return http.ListenAndServe(addr, app)
}

func (app *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range app.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := app.createContext(w, req)

	c.handlers = middlewares

	app.router.handle(c)
	app.pool.Put(c)
}

// ConfigurationReadOnly returns an object which doesn't allow field writing.
func (app *App) ConfigurationReadOnly() *Configuration {
	return app.config
}

func (app *App) View(writer io.Writer, filename string, bindingData interface{}) error {
	if !app.view.Registered() {
		err := errors.New("view engine is missing, use `RegisterView`")
		// app.logger.Error(err)
		log.Error(err)
		return err
	}

	return app.view.ExecuteWriter(writer, filename, bindingData)
}

func (app *App) RegisterView(viewEngine ViewEngine) {
	log.Info("register view from app")
	app.view.Register(viewEngine)
}

// Bidning

func initFormDecoder() {
	formDecoderInit.Do(func() {
		formDecoder = form.NewDecoder()
	})
}
