// +build go1.7

package rock

import "net/http"

// Context is the context interface type
type Context interface {
	Request() *http.Request
	Response() *Response
	Next()
	String(int, string, ...interface{})
	BaseContext() *Ctx
	Param(name string) string
	RequestStart(w http.ResponseWriter, r *http.Request)
	SetHandlers([]HandlerFunc)
	Text(int, string) error

	// RequestStart(w http.ResponseWriter, r *http.Request)
	// RequestEnd()
}
