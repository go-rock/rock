package rock

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-playground/form"
	"github.com/plimble/utils/pool"
)

var (
	bufPool = pool.NewBufferPool(100)

	formDecoder     *form.Decoder
	formDecoderInit sync.Once
)

type (
	App struct {
		*Router
		mux    *chi.Mux
		pool   sync.Pool
		render Renderer
		// panicFunc    PanicHandler
		notfoundFunc HandlerFunc

		// add
		contextFunc         ContextFunc
		customHandlersFuncs customHandlers
	}
)

func GetPool() *pool.BufferPool {
	return bufPool
}
func New() *App {
	app := &App{
		contextFunc: func(l *App) Context {
			return NewContext(l)
		},
	}
	app.Router = &Router{
		handlers: nil,
		prefix:   "/",
		app:      app,
	}
	app.mux = chi.NewMux()
	app.pool.New = func() interface{} {
		// c := &Ctx{}
		// c.index = -1
		// c.response = &c.writercache
		// return c
		// pp.Println(app, "this is app")
		c := app.contextFunc(app)
		// c.response = &c.writercache

		b := c.BaseContext()
		b.index = -1
		b.parent = c
		return b
	}

	return app
}

func Default() *App {
	a := New()
	a.Use(Logger())
	return a
}

func DefaultWithout() *App {
	a := New()
	a.Use(Logger())
	return a
}

//Run server with specific address and port
func (app *App) Run(addr string) {
	logo, _ := base64.StdEncoding.DecodeString("LC0tLS0tLS4gICwtLS0tLS4gICwtLS0tLS4sLS0uICwtLS4KfCAgLi0tLiAnJyAgLi0uICAnJyAgLi0tLi98ICAuJyAgIC8KfCAgJy0tJy4nfCAgfCB8ICB8fCAgfCAgICB8ICAuICAgJwp8ICB8XCAgXCAnICAnLScgICcnICAnLS0nXHwgIHxcICAgXApgLS0nICctLScgYC0tLS0tJyAgYC0tLS0tJ2AtLScgJy0tJw==")
	fmt.Println(string(logo))
	fmt.Printf("*Rock* -- Listen on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, app); err != nil {
		panic(err)
	}
}

//ServeHTTP implement http.Handler
// func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
// 	// a.mux.ServeHTTP(w, req)
// 	c := a.pool.Get().(*Ctx)
// 	// c := NewContext(a)
// 	c.RequestStart(w, req)
// 	pp.Println(c)
// 	// c.Next()
// 	// c.parent.RequestEnd()
// 	// a.mux.ServeHTTP(c.response, c.request)

// 	// a.pool.Put(c)
// }

func (app *App) GetMux() *chi.Mux {
	return app.mux
}

// RegisterCustomHandler registers a custom handler that gets wrapped by HandlerFunc
func (l *App) RegisterCustomHandler(customType interface{}, fn CustomHandlerFunc) {

	if l.customHandlersFuncs == nil {
		l.customHandlersFuncs = make(customHandlers)
	}

	t := reflect.TypeOf(customType)

	if _, ok := l.customHandlersFuncs[t]; ok {
		panic(fmt.Sprint("Custom Type + CustomHandlerFunc already declared: ", t))
	}

	l.customHandlersFuncs[t] = fn
}

// RegisterContext registers a custom Context function for creation
// and resetting of a global object passed per http request
func (l *App) RegisterContext(fn ContextFunc) {
	l.contextFunc = fn
}

func initFormDecoder() {
	formDecoderInit.Do(func() {
		formDecoder = form.NewDecoder()
	})
}

// BuiltInFormDecoder returns the built in form decoder github.com/go-playground/form
// in order for custom type to be registered.
func (l *App) BuiltInFormDecoder() *form.Decoder {

	initFormDecoder()

	return formDecoder
}
