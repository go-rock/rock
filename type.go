package rock

type Map map[string]interface{}
type H Map
type M Map
type HandlerFunc func(Context)
type MiddlewareFunc = HandlerFunc
type PreMiddlewareFunc = HandlerFunc
type Handler = HandlerFunc
