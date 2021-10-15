package rock

type Map map[string]interface{}
type H Map
type M Map
type HandlerFunc func(Context)
type MiddlewareFunc func(Context)
type PreMiddlewareFunc func(Context)
type Handler HandlerFunc
