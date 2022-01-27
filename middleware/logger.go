package middleware

import (
	"bytes"
	"io"
	"jvmao"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

func Logger() jvmao.MiddlewareFunc {
	return LoggerWithConfig(defaultLoggerConfig)
}

type LoggerConfig struct {
	Format     string
	Output     io.Writer
	TimeFormat string
}

var defaultLoggerConfig = LoggerConfig{
	Format:     `{{real-ip}} - {{id}} [{{time}}] {{uri}} {{status}} {{bytes-out}} {{referer}} {{user-agent}}`,
	TimeFormat: "2006-01-02 15:04:05.00000",
	Output:     os.Stdout,
}

func LoggerWithConfig(config LoggerConfig) jvmao.MiddlewareFunc {

	pool := &sync.Pool{
		New: func() interface{} {
			return &loggerWR{conf: config, buf: bytes.NewBuffer(make([]byte, 256))}
		},
	}

	tmlp := template.New("Logger Template")

	f := formatf(config.Format)

	return func(next jvmao.HandlerFunc) jvmao.HandlerFunc {
		return func(c *jvmao.Context) error {
			req := c.Request()
			resp := c.Response()
			wr := pool.Get().(*loggerWR)
			wr.reset(req, resp)

			defer func(wr *loggerWR) {
				wr.stop = time.Now()

				t := template.Must(tmlp.Funcs(wr.mf()).Parse(f))
				t.Execute(wr.buf, nil)

				config.Output.Write(wr.buf.Bytes())

				pool.Put(wr)

			}(wr)

			return next(c)
		}
	}
}

func formatf(s string) string {

	s = regexp.MustCompile(`(\{[^\}]+)+`).ReplaceAllStringFunc(s, func(sub string) string {
		return strings.ReplaceAll(sub, " ", "")
	})

	ch := func(re *regexp.Regexp, s string) string {
		return re.ReplaceAllStringFunc(s, func(sub string) string {
			sub = strings.ReplaceAll(sub, ":", " \"")
			sub += "\""
			return sub
		})
	}
	var re *regexp.Regexp

	re = regexp.MustCompile(`((header|query|form):\s*([^\}]+))+`)
	s = ch(re, s)

	re = regexp.MustCompile(`(\{real\-ip\}|\{user-agent\}|\{bytes-in\}|\{bytes-out\}|\{ latency-human\})+`)

	s = re.ReplaceAllStringFunc(s, func(sub string) string {
		return strings.ReplaceAll(sub, "-", "")
	})

	if !strings.HasSuffix(s, "\n") {
		s = s + "\n"
	}

	return s
}

type loggerWR struct {
	r     *http.Request
	w     *jvmao.Response
	conf  LoggerConfig
	start time.Time
	stop  time.Time
	buf   *bytes.Buffer

	tmpl *template.Template
}

func (wr *loggerWR) mf() template.FuncMap {
	return template.FuncMap{
		"time":         wr.time,
		"id":           wr.id,
		"realip":       wr.remoteIP,
		"host":         wr.host,
		"uri":          wr.uri,
		"path":         wr.path,
		"method":       wr.method,
		"protocol":     wr.protocol,
		"referer":      wr.referer,
		"useragent":    wr.userAgent,
		"status":       wr.status,
		"header":       wr.header,
		"query":        wr.query,
		"form":         wr.form,
		"bytesin":      wr.bytesIn,
		"bytesout":     wr.bytesOut,
		"latency":      wr.latency,
		"latencyhuman": wr.latencyHuman,
	}
}
func (wr *loggerWR) latencyHuman() string {
	return wr.latency().String()
}

func (wr *loggerWR) latency() time.Duration {
	return wr.stop.Sub(wr.start)
}

func (wr *loggerWR) bytesOut() int64 {
	return wr.w.Size
}
func (wr *loggerWR) bytesIn() string {

	cl := wr.header("Content-Length")
	if cl == "" {
		cl = "0"
	}
	return cl
}

func (wr *loggerWR) form(n string) string {
	return wr.r.FormValue(n)
}
func (wr loggerWR) query(n string) string {
	return wr.r.URL.Query().Get(n)
}

func (wr *loggerWR) header(n string) string {
	return wr.r.Header.Get(n)
}
func (wr *loggerWR) status() string {
	uc := true
	var s string
	code := wr.w.Status
	switch {
	case code < 200:
		s = jvmao.TColorWrite(jvmao.BBlack, uc, "%d", code)
	case code < 300:
		s = jvmao.TColorWrite(jvmao.BGreen, uc, "%d", code)
		s = colorWrite(jvmao.BGreen, "%d", code)
	case code < 400:
		s = jvmao.TColorWrite(jvmao.BCyan, uc, "%d", code)
		s = colorWrite(jvmao.BCyan, "%d", code)
	case code < 500:
		s = jvmao.TColorWrite(jvmao.BRed, uc, "%d", code)
	default:
		s = jvmao.TColorWrite(jvmao.BRed, uc, "%d", code)
	}
	return s
}

func (wr *loggerWR) userAgent() string {
	return wr.r.UserAgent()
}

func (wr *loggerWR) referer() string {
	return wr.r.Referer()
}

func (wr *loggerWR) protocol() string {
	return wr.r.Proto
}

func (wr *loggerWR) method() string {
	return wr.r.Method
}

func (wr *loggerWR) path() string {
	p := wr.r.URL.Path
	if p == "" {
		p = "/"
	}
	return p
}

func (wr *loggerWR) uri() string {
	return wr.r.RequestURI
}

func (wr *loggerWR) host() string {
	return wr.r.Host
}

func (wr *loggerWR) remoteIP() string {
	if ip := wr.r.Header.Get("X-Forwarded-For"); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			return strings.TrimSpace(ip[:i])
		}
		return ip
	}
	if ip := wr.r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(wr.r.RemoteAddr)
	return ra

}

func (wr *loggerWR) id() string {
	id := wr.r.Header.Get("X-Request-ID")
	if id == "" {
		id = wr.w.Header().Get("X-Request-ID")
	}
	return id
}

func (wr *loggerWR) time() string {
	return wr.start.Format(wr.conf.TimeFormat)
}

func (wr *loggerWR) reset(r *http.Request, w *jvmao.Response) {
	wr.r = r
	wr.w = w
	wr.start = time.Now()
	wr.buf.Reset()
}
