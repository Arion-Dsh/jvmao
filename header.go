package jvmao

const (
	charsetUTF8 = "charset=UTF-8"
)

// https://www.iana.org/assignments/media-types/media-types.xhtml
const (
	MIMETextPlain                 = "text/plain"
	MIMETextPlainUTF8             = "text/plain; " + charsetUTF8
	MIMETextHTML                  = "text/html"
	MIMETextHTMLUTF8              = "text/html; " + charsetUTF8
	MIMEApplicationJavaScript     = "application/javascript"
	MIMEApplicationJavaScriptUTF8 = "application/javascript; " + charsetUTF8
	MIMEApplicationJSON           = "application/json"
	MIMEApplicationJSONUTF8       = "application/json; " + charsetUTF8
	MIMEApplicationForm           = "application/x-www-form-urlencoded"
	MIMEMultipartForm             = "multipart/form-data"
	MIMEApplicationGrpc           = "application/grpc"
	MIMEApplicationGrpcWeb        = "application/grpc-web"
	MIMEApplicationGrpcWebText    = "application/grpc-web-text"
)

// Headers
const (
	HeaderContentEncoding    = "Content-Encoding"
	HeaderContentLength      = "Content-Length"
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"

	HeaderAuthorization   = "Authorization"
	HeaderCookie          = "Cookie"
	HeaderSetCookie       = "Set-Cookie"
	HeaderWWWAuthenticate = "WWW-Authenticate"

	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderXContentTypeOptions = "X-Content-Type-Options"

	HeaderLocation = "Location"
)
