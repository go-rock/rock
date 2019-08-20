// +build go1.7

package rock

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
)

// Context is the context interface type
type Context interface {
	Request() *http.Request
	Response() *Response
	params() chi.RouteParams
	Param(name string) string
	QueryParams() url.Values
	ParseForm() error
	ParseMultipartForm(maxMemory int64) error
	Next()
	ClientIP() (clientIP string)
	AcceptedLanguages(lowercase bool) []string
	Stream(step func(w io.Writer) bool)
	JSON(int, interface{}) error
	JSONBytes(int, []byte) error
	JSONP(int, interface{}, string) error
	XML(int, interface{}) error
	XMLBytes(int, []byte) error
	Text(int, string) error
	TextBytes(int, []byte) error
	Attachment(r io.Reader, filename string) (err error)
	Inline(r io.Reader, filename string) (err error)
	Decode(includeFormQueryParams bool, maxMemory int64, v interface{}) (err error)
	BaseContext() *Ctx
	RequestStart(w http.ResponseWriter, r *http.Request)
	SetHandlers([]HandlerFunc)
	String(int, string, ...interface{})

	// RequestEnd()
}
