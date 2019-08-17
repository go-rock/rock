package rock

import "reflect"

// type internally.
type Handler interface{}

// HandlerFunc is the internal handler type used for middleware and handlers
type HandlerFunc func(Context)

// HandlersChain is an array of HanderFunc handlers to run
type HandlersChain []HandlerFunc

// ContextFunc is the function to run when creating a new context
type ContextFunc func(l *App) Context

// CustomHandlerFunc wraped by HandlerFunc and called where you can type cast both Context and Handler
// and call Handler
type CustomHandlerFunc func(Context, Handler)

// customHandlers is a map of your registered custom CustomHandlerFunc's
// used in determining how to wrap them.
type customHandlers map[reflect.Type]CustomHandlerFunc
