package rock

import "net/http"

// HTTP Constant Terms and Variables
const (
	// CONNECT HTTP method
	CONNECT = http.MethodConnect
	// DELETE HTTP method
	DELETE = http.MethodDelete
	// GET HTTP method
	GET = http.MethodGet
	// HEAD HTTP method
	HEAD = http.MethodHead
	// OPTIONS HTTP method
	OPTIONS = http.MethodOptions
	// PATCH HTTP method
	PATCH = http.MethodPatch
	// POST HTTP method
	POST = http.MethodPost
	// PUT HTTP method
	PUT = http.MethodPut
	// TRACE HTTP method
	TRACE = http.MethodTrace

	//-------------
	// Media types
	//-------------

	ApplicationJSON                  = "application/json"
	ApplicationJSONCharsetUTF8       = ApplicationJSON + "; " + CharsetUTF8
	ApplicationJavaScript            = "application/javascript"
	ApplicationJavaScriptCharsetUTF8 = ApplicationJavaScript + "; " + CharsetUTF8
	ApplicationXML                   = "application/xml"
	ApplicationXMLCharsetUTF8        = ApplicationXML + "; " + CharsetUTF8
	ApplicationForm                  = "application/x-www-form-urlencoded"
	ApplicationProtobuf              = "application/protobuf"
	ApplicationMsgpack               = "application/msgpack"
	TextHTML                         = "text/html"
	TextHTMLCharsetUTF8              = TextHTML + "; " + CharsetUTF8
	TextPlain                        = "text/plain"
	TextPlainCharsetUTF8             = TextPlain + "; " + CharsetUTF8
	MultipartForm                    = "multipart/form-data"
	OctetStream                      = "application/octet-stream"

	//---------
	// Charset
	//---------

	CharsetUTF8 = "charset=utf-8"

	//---------
	// Headers
	//---------

	AcceptedLanguage   = "Accept-Language"
	AcceptEncoding     = "Accept-Encoding"
	Authorization      = "Authorization"
	ContentDisposition = "Content-Disposition"
	ContentEncoding    = "Content-Encoding"
	ContentLength      = "Content-Length"
	ContentType        = "Content-Type"
	Location           = "Location"
	Upgrade            = "Upgrade"
	Vary               = "Vary"
	WWWAuthenticate    = "WWW-Authenticate"
	XForwardedFor      = "X-Forwarded-For"
	XRealIP            = "X-Real-Ip"
	Allow              = "Allow"
	Origin             = "Origin"

	Gzip = "gzip"

	blank = ""
)
