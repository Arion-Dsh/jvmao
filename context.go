package jvmao

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const maxMemory = 32 << 20 // 32 MB

type Context interface {
	Request() *http.Request

	Response() *Response

	//Reverse the path with name.
	Reverse(name string, params ...string) string

	HanderValue(key string) string

	Set(key string, value interface{})

	Get(key string) interface{}

	Del(key string)

	WriteHeader(statusCode int)

	// Cookie returns the named cookie provided in the request or ErrNoCookie
	// if not found. If multiple cookies match the given name, only one cookie
	// will be returned.
	Cookie(name string) (*http.Cookie, error)

	// SetCookie add a cookie in response.
	SetCookie(cookie *http.Cookie)

	// Query return all query values in url.
	Query() url.Values

	// QueryValue returns first value with the given key.
	QueryValue(key string) string

	// QueryValues returns all values with the given key
	QueryValues(key string) []string

	// Param is the parameters in route pattern.
	// such as "id" in /post/:id .

	// ParamValue get the parameter.
	ParamValue(key string) string

	// FormValue returns firs value with the given name.
	// it calls ParseMultipartForm and ParseForm
	// if necessary and ignores any errors returned by these functions.
	FormValue(name string) string

	// FormValues returns all values with the given name.
	// it calls ParseMultipartForm and ParseForm
	// if necessary and ignores any errors returned by these functions.
	FormValues(name string) []string

	FormFile(name string) (*multipart.FileHeader, error)

	BindForm(i interface{}) error
	BindQuery(i interface{}) error
	BindParam(i interface{}) error

	//Render render a template then send a HTML response with status code
	// it'll use the DefalultRenderer when the jumao's Renderer was not set
	Render(statusCode int, tmpl string, data interface{}) (err error)

	Error(statusCode int, err error) error

	NoContent(statusCode int) error

	// String send a text response with status code.
	String(statusCode int, s string) error

	// HTML send a HTML response with status code.
	HTML(statusCode int, html string) error

	// Blob send a blob response with status code and content type.
	Blob(statusCode int, contentType string, b []byte) (err error)

	//Json send a json response with status code.
	Json(statusCode int, i interface{}) error

	//File send a file response with status code
	File(file string, dir http.Dir) error

	//FileFS send a file response with status code
	FileFS(file string, fsys fs.FS) error

	//Redirect to provided URL
	Redirect(statusCode int, url string) error

	// Logger() Logger
}

type context struct {
	r *http.Request
	w *Response

	params url.Values
	data   map[string]interface{}
	err    *HTTPError

	route *routeChache
}

func (c *context) Request() *http.Request {
	return c.r
}

func (c *context) Response() *Response {
	return c.w
}

func (c *context) Reverse(name string, params ...string) string {
	// return c.jm.Reverse(name, params...)
	return ""
}

func (c *context) HanderValue(key string) string {
	return c.r.Header.Get(key)
}

func (c *context) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *context) Get(key string) interface{} {

	if v, ok := c.data[key]; ok {
		return v
	}
	return nil
}

func (c *context) Del(key string) {
	delete(c.data, key)
}

func (c *context) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *context) Cookie(name string) (*http.Cookie, error) {
	return c.r.Cookie(name)
}

func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

func (c *context) Query() url.Values {
	return c.r.URL.Query()
}

func (c *context) QueryValue(key string) string {
	return c.Query().Get(key)
}

func (c *context) QueryValues(key string) []string {
	return c.Query()[key]
}

func (c *context) Param() url.Values {
	return c.params
}

func (c *context) ParamValue(key string) string {
	return c.r.PathValue(key)
}

func (c *context) ParamValues(key string) []string {
	return c.params[key]
}

func (c *context) FormValue(name string) string {
	_ = c.ParseForm()
	return c.r.PostFormValue(name)
}

func (c *context) FormValues(name string) []string {

	_ = c.ParseForm()

	return c.r.PostForm[name]
}

func (c *context) FormFile(name string) (*multipart.FileHeader, error) {

	f, fh, err := c.r.FormFile(name)
	if err != nil {
		return nil, http.ErrMissingFile
	}
	f.Close()
	return fh, nil
}

func (c *context) ParseForm() error {
	if strings.HasPrefix(c.r.Header.Get(HeaderContentType), MIMEMultipartForm) {
		return c.r.ParseMultipartForm(maxMemory)
	} else {
		return c.r.ParseForm()
	}
}

// Render render a template then send a HTML response with status code
// it'll use the DefalultRenderer when the jumao's Renderer was not set
func (c *context) Render(statusCode int, tmpl string, data interface{}) (err error) {

	buf := new(bytes.Buffer)

	// if err = c.jm.renderer.Render(buf, tmpl, data, c); err != nil {
	// 	panic(err)
	// }

	return c.Blob(statusCode, MIMETextHTMLUTF8, buf.Bytes())

}

func (c *context) Error(statusCode int, err error) error {
	if _, ok := err.(*HTTPError); ok {
		return err
	}
	err = NewHTTPError(statusCode, err.Error())
	return err
}

func (c *context) NoContent(statusCode int) error {
	c.WriteHeader(statusCode)
	return nil
}

// String send a text response with status code.
func (c *context) String(statusCode int, s string) error {
	return c.Blob(statusCode, MIMETextPlainUTF8, []byte(s))
}

// HTML send a HTML response with status code.
func (c *context) HTML(statusCode int, html string) error {
	return c.Blob(statusCode, MIMETextHTMLUTF8, []byte(html))
}

// Blob send a blob response with status code and content type.
func (c *context) Blob(statusCode int, contentType string, b []byte) (err error) {
	c.setHct(contentType)
	c.WriteHeader(statusCode)
	_, err = c.w.Write(b)
	return
}

// Json send a json response with status code.
func (c *context) Json(statusCode int, i interface{}) error {

	b, err := json.Marshal(i)
	if err != nil {
		return c.Error(500, err)
	}
	return c.Blob(statusCode, MIMEApplicationJSONUTF8, b)
}

func (c *context) FileFS(file string, fsys fs.FS) error {
	return c.openFile(file, http.FS(fsys))
}

// File send a file response with status code
func (c *context) File(file string, dir http.Dir) error {
	fsys := newCtxFS(dir)
	return c.openFile(file, http.FS(fsys))
}

func (c *context) openFile(file string, dir http.FileSystem) error {
	const indexPage = "index.html"

	f, err := dir.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if fi.IsDir() {
		http.NotFound(c.w, c.r)
		return nil
	}

	if err != nil {
		return err
	}

	ff, ok := f.(io.ReadSeeker)
	if !ok {
		return errors.New("file is not io.ReadSeeker")
	}
	http.ServeContent(c.w, c.r, file, fi.ModTime(), ff)

	return nil
}

/* func (c *context) Logger() Logger { */
/* return c.jm.Logger */
/* } */

func (c *context) Redirect(statusCode int, url string) error {
	if statusCode < 300 || statusCode > 308 {
		return errors.New("invalid redirect status code.")
	}

	c.w.Header().Add(HeaderLocation, url)
	c.WriteHeader(statusCode)
	return nil
}

func (c *context) setHct(t string) {
	c.w.Header().Add(HeaderContentType, t)
}

func (c *context) reset(w http.ResponseWriter, r *http.Request) {
	c.w.reset(w)
	c.err = nil
	c.r = r
	c.params = url.Values{}
	c.data = map[string]interface{}{}
}

func newCtxFS(dir http.Dir) fs.FS {
	return &ctxFS{dir}
}

type ctxFS struct {
	http.Dir
}

func (f *ctxFS) Open(name string) (fs.File, error) {
	return f.Dir.Open(name)
}
