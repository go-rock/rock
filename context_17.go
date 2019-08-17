// +build go1.7

package rock

import "net/http"

// Context is the context interface type
type Context interface {
	Request() *http.Request
	Response() ResponseWriter
	Next()
	String(int, string, ...interface{})
	// RequestStart(w http.ResponseWriter, r *http.Request)
	// RequestEnd()
}
