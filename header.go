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
	HeaderContentExposeHeaders = "access-control-expose-headers"
	HeaderContentEncoding      = "content-encoding"
	HeaderContentLength        = "content-length"
	HeaderContentType          = "content-type"
	HeaderContentDisposition   = "content-disposition"

	HeaderAuthorization   = "authorization"
	HeaderCookie          = "cookie"
	HeaderSetCookie       = "set-cookie"
	HeaderWWWAuthenticate = "www-authenticate"

	HeaderXRealIP             = "x-real-ip"
	HeaderXRequestID          = "x-request-id"
	HeaderXContentTypeOptions = "x-content-type-options"

	HeaderLocation = "location"

	//Grpc Header
	HeaderTe                 = "te"
	HeaderGrpcAcceptEncoding = "grpc-accept-encoding"

	//Header value
	HeaderValueTrailers = "trailers"
	HeaderValueIdentity = "identity"
)
