package rock

// Map 是 map[string]interface{} 的别名，用于路径参数与通用键值数据。
type Map map[string]interface{}

// H 是 map[string]interface{} 的别名，常用于 JSON 响应数据。
type H Map

// M 是 map[string]interface{} 的别名，常用于视图数据与通用 KV。
type M Map

// HandlerFunc 定义路由处理函数与中间件的签名。
type HandlerFunc func(Context)

// MiddlewareFunc 是 HandlerFunc 的别名，语义上表示中间件。
type MiddlewareFunc = HandlerFunc

// PreMiddlewareFunc 是 HandlerFunc 的别名，语义上表示前置中间件。
type PreMiddlewareFunc = HandlerFunc

// Handler 是 HandlerFunc 的别名。
type Handler = HandlerFunc
