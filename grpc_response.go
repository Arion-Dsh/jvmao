package jvmao

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/http2"
)

type grpcResponse struct {
	req *http.Request //origin http request

	header http.Header

	writer http.ResponseWriter

	isText bool

	Status      int
	Size        int64
	wroteHeader bool // reply header has been (logically) written

	ct string
}

// Header returns the header map that will be sent by
// WriteHeader. The Header map also is the mechanism with which
// Handlers can set HTTP trailers.
func (r *grpcResponse) Header() http.Header {
	return r.header
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *grpcResponse) Write(buf []byte) (n int, err error) {

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
func (r *grpcResponse) WriteHeader(statusCode int) {

	if r.wroteHeader {
		return
	}
	r.encodeHeaders()
	r.Status = statusCode
	r.writer.WriteHeader(statusCode)
	r.wroteHeader = true
}

// Flush sends any buffered data to the client.
// more [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *grpcResponse) Flush() {
	if r.wroteHeader {
		f, _ := r.writer.(http.Flusher)
		f.Flush()
	}
}

func (r *grpcResponse) finish() {
	r.WriteHeader(http.StatusOK)
	r.Flush()
}

func (r *grpcResponse) decodeHeaders() {
	// Remove content-length header since it represents http1.1 payload size, not the sum of the h2
	// DATA frame payload lengths. https://http2.github.io/http2-spec/#malformed This effectively
	// switches to chunked encoding which is the default for h2
	r.req.Header.Del(HeaderContentLength)

	// Adds te:trailers to upstream HTTP2 request. It's required for gRPC.
	r.req.Header.Set(HeaderTe, HeaderValueTrailers)

	if r.req.Header.Get(HeaderGrpcAcceptEncoding) == "" {
		r.req.Header.Set(HeaderGrpcAcceptEncoding, HeaderValueIdentity)
	}

	ct := r.req.Header.Get(HeaderContentType)
	// replace content-type
	mime := MIMEApplicationGrpcWeb
	if strings.HasPrefix(ct, MIMEApplicationGrpcWebText) {
		mime = MIMEApplicationGrpcWebText
		r.isText = true
		r.ct = mime
	}
	r.req.Header.Set(HeaderContentType, strings.Replace(ct, mime, MIMEApplicationGrpc, 1))

}

func (r *grpcResponse) encodeHeaders() {

	wh := r.writer.Header()
	ks := []string{}
	for k, v := range r.header {
		k = strings.ToLower(k)

		if k == HeaderContentType {
			nv := []string{}
			for _, vv := range v {
				nv = append(
					nv,
					strings.Replace(vv, MIMEApplicationGrpc, r.ct, 1),
				)
			}
			v = nv

		}

		// if strings.Contains(k, http2.TrailerPrefix) {
		k = strings.Replace(k, http2.TrailerPrefix, "", 1)

		// }
		ks = append(ks, k)
		wh[http.CanonicalHeaderKey(k)] = v
	}

	ks = append(ks, "grpc-status", "grpc-message", "transfer-encoding")

	wh.Set(HeaderContentExposeHeaders, strings.Join(ks, ","))
}

func (r *grpcResponse) reset(w http.ResponseWriter, req *http.Request) {

	r.req = req
	r.writer = w
	r.decodeHeaders()
	r.header = make(http.Header)
	r.ct = MIMEApplicationGrpcWeb

	r.Size = 0
	r.Status = http.StatusOK
	r.wroteHeader = false

	if r.isText {
		tr := &grpcTextResponse{w: w}
		tr.newEncoder()
		r.writer = tr
	}
}

type grpcTextResponse struct {
	w       http.ResponseWriter
	encoder io.WriteCloser
}

func (r *grpcTextResponse) newEncoder() *grpcTextResponse {
	r.encoder = base64.NewEncoder(base64.StdEncoding, r.w)
	return r
}

func (r *grpcTextResponse) Header() http.Header {
	return r.w.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *grpcTextResponse) Write(buf []byte) (n int, err error) {
	return r.encoder.Write(buf)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
func (r *grpcTextResponse) WriteHeader(statusCode int) {
	r.w.WriteHeader(statusCode)
}

// Flush sends any buffered data to the client.
// more [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *grpcTextResponse) Flush() {
	// close base64 encoder and ignore error
	r.encoder.Close()
	r.newEncoder()
	r.w.(http.Flusher).Flush()
}
