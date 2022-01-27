package jvmao

import (
	"bufio"
	"net"
	"net/http"
)

func NewResponse(jm *Jvmao, w http.ResponseWriter) *Response {
	return &Response{jm: jm, writer: w}
}

// Response implements http.ResponseWriter/http.Flusher/http.Hijacker
// to be used by an HTTP handler to construct an HTTP response .
// more: [http.ResponseWriter](https://golang.org/pkg/net/http/#ResponseWriter)
type Response struct {
	jm     *Jvmao
	writer http.ResponseWriter

	Status      int
	Size        int64
	wroteHeader bool // reply header has been (logically) written
}

// Header returns the header map that will be sent by
// WriteHeader. The Header map also is the mechanism with which
// Handlers can set HTTP trailers.
func (r *Response) Header() http.Header {
	return r.writer.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *Response) Write(buf []byte) (n int, err error) {

	if !r.wroteHeader {
		if r.Status == 0 {
			r.Status = http.StatusOK
		}
		r.WriteHeader(r.Status)
	}
	n, err = r.writer.Write(buf)
	r.Size += int64(n)
	return
}

// WriteHeader sends an HTTP response header with the provided
// status code.
func (r *Response) WriteHeader(statusCode int) {
	if r.wroteHeader {
		r.jm.Logger.Warn(" superfluous Response.WriteHeader call.")
		return
	}
	r.Status = statusCode
	r.writer.WriteHeader(statusCode)
	r.wroteHeader = true
}

// Flush sends any buffered data to the client.
// more [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *Response) Flush() {
	r.writer.(http.Flusher).Flush()
}

// Hijack allow an HTTP handler to take over the connection.
// more [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.writer.(http.Hijacker).Hijack()
}

func (r *Response) reset(w http.ResponseWriter) {
	r.writer = w
	r.Size = 0
	r.Status = http.StatusOK
	r.wroteHeader = false
}
