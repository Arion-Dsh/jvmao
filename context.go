package jvmao

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type Context struct {
	jm *Jvmao

	r *http.Request
	w *Response

	param map[string]string
	data  map[string]interface{}
	err   *HTTPError
}

func (c *Context) Request() *http.Request {
	return c.r
}

func (c *Context) Response() *Response {
	return c.w
}

func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *Context) Get(key string) interface{} {

	if v, ok := c.data[key]; ok {
		return v
	}
	return nil
}

func (c *Context) Del(key string) {
	delete(c.data, key)
}

func (c *Context) Query(name string) string {
	return c.r.URL.Query().Get(name)
}

func (c *Context) Param(name string) string {
	if v, ok := c.param[name]; ok {
		return v
	}
	return ""
}

func (c *Context) FormValue(name string) string {
	return c.r.FormValue(name)
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	f, fh, err := c.r.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, nil
}

func (c *Context) Header() http.Header {
	return c.w.Header()
}
func (c *Context) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

func (c *Context) GetHeader(key string) string {
	return c.w.Header().Get(key)
}

func (c *Context) DelHeader(key string) {
	c.w.Header().Del(key)
}

func (c *Context) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

//Render render a template then send a HTML response with status code
// it'll use the DefalultRenderer when the jumao's Renderer was not set
func (c *Context) Render(statusCode int, tmpl string, data interface{}) (err error) {

	buf := new(bytes.Buffer)

	if err = c.jm.renderer.Render(buf, tmpl, data, c); err != nil {
		return c.Error(err)
	}

	return c.Blob(statusCode, MIMETextHTMLUTF8, buf.Bytes())

}

func (c *Context) Error(err error) error {
	err = NewHTTPErrorWithError(err)
	c.jm.HTTPErrHandler(err, c)
	return nil
}
func (c *Context) NoContent(statusCode int) error {
	c.WriteHeader(statusCode)
	return nil
}

// String send a text response with status code.
func (c *Context) String(statusCode int, s string) error {
	return c.Blob(statusCode, MIMETextPlainUTF8, []byte(s))
}

// HTML send a HTML response with status code.
func (c *Context) HTML(statusCode int, html string) error {
	return c.Blob(statusCode, MIMETextHTMLUTF8, []byte(html))
}

// Blob send a blob response with status code and content type.
func (c *Context) Blob(statusCode int, contentType string, b []byte) (err error) {
	c.setHct(contentType)
	c.WriteHeader(statusCode)
	_, err = c.w.Write(b)
	return
}

//Json send a json response with status code.
func (c *Context) Json(code int, i interface{}) error {

	b, err := json.Marshal(i)
	if err != nil {
		return c.Error(err)
	}
	return c.Blob(code, MIMEApplicationJSONUTF8, b)
}

//File send a file response with status code
func (c *Context) File(dir http.Dir, file string) error {
	const indexPage = "index.html"

	f, err := dir.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = dir.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		fi, err = f.Stat()
	}

	if err != nil {
		return err
	}

	http.ServeContent(c.w, c.r, file, fi.ModTime(), f)
	return nil
}

func (c *Context) setHct(t string) {
	c.SetHeader(HeaderContentType, t)
}

func (c *Context) Logger() Logger {
	return c.jm.Logger
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.w.reset(w)
	c.err = nil
	c.r = r
	c.param = map[string]string{}
	c.data = map[string]interface{}{}
}
