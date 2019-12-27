// +build go1.7

package rock

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
)

// Context is the context interface type
type Context interface {
	Request() *http.Request
	Response() *Response
	GetQuery(name string) (string, bool)
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
	// Decode(includeFormQueryParams bool, maxMemory int64, v interface{}) (err error)
	Decode(v interface{}, args ...interface{}) (err error)
	ShouldBindJSON(v interface{}, args ...interface{}) (err error)
	BaseContext() *Ctx
	RequestStart(w http.ResponseWriter, r *http.Request)
	String(int, string, ...interface{})

	// Custom
	SetHandlers([]HandlerFunc)
	Data() M
	SetData(M)
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)

	Abort()
	AbortWithStatus(code int)
	AbortWithStatusJSON(code int, jsonObj interface{})
	Redirect(url string)

	// Render
	HTMLRender(HTMLRender)
	Render(code int, r Render)
	Status(code int)
	HTML(string, ...int)

	// Post body

	MustPostInt(key string, d int) int
	MustPostString(key string, d string) string

	MustParamInt(name string, d int) int

	MustQueryInt(name string, d int) int

	FormFile(name string) (*multipart.FileHeader, error)
}
