package jvmao

import (
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
)

type mux struct {
	jm   *Jvmao
	mu   sync.RWMutex
	root *entry

	patCache map[string]*entity //map[pattern]entity
	routes   map[string]string  // map[route name] pattern

	sPrefix string // static prefix
	sRoot   http.Dir

	ctx *muxCtx
}

type muxCtx struct {
	handler HandlerFunc
	pValue  []string
}

func (c *muxCtx) param(p string) {
	c.pValue = append(c.pValue, p)
}

func (c *muxCtx) reset() {
	c.pValue = []string{}
}

func newMux(jm *Jvmao) *mux {
	return &mux{
		jm:       jm,
		root:     &entry{},
		patCache: map[string]*entity{},
		routes:   map[string]string{},
		ctx:      new(muxCtx),
	}
}

func (mux *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := mux.jm.pool.Get().(*context)
	defer mux.jm.pool.Put(c)
	c.reset(w, r)
	mux.ctx.reset()

	mux.ctx.handler = mux.jm.NotFoundHandler

	err := NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound))

	if es := mux.root.match(mux.ctx, r.URL.Path); es != nil {

		if hf, ok := es.hf[r.Method]; ok {
			for i, n := range es.pName {
				c.params.Add(n, mux.ctx.pValue[i])
			}
			mux.ctx.handler = hf
			err = nil
		}
	}

	if mux.sPrefix != "" && strings.HasPrefix(r.URL.Path, mux.sPrefix) {
		if containsDotDot(r.URL.Path) {
			err = NewHTTPError(http.StatusBadRequest, "invalid URL path")
		} else {
			p := strings.TrimPrefix(r.URL.Path, mux.sPrefix)
			err = c.File(mux.sRoot, p)
			if err != nil {
				err = NewHTTPErrorWithError(err)
			}
		}
	}
	// apply middleware
	mux.ctx.handler = applyMiddleware(mux.ctx.handler, mux.jm.middleware...)

	if err != nil {
		c.Error(err)
	}

	if err = mux.ctx.handler(c); err != nil {
		err = NewHTTPErrorWithError(err)
		mux.jm.HTTPErrHandler(err, c)
	}

}

func (mux *mux) setStatic(root, prefix string) {
	mux.sRoot = http.Dir(root)
	mux.sPrefix = prefix
}

func (mux *mux) handle(name, method, pattern string, hf HandlerFunc) {

	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	if _, ok := mux.routes[name]; ok {
		panic("route name " + name + " already have it.")
	}

	if e, ok := mux.patCache[pattern]; ok {
		e.hf[method] = hf
		e.methods = append(e.methods, method)
		return
	}

	es := &entity{
		pat:     pattern,
		hf:      map[string]HandlerFunc{method: hf},
		methods: []string{http.MethodOptions, method},
	}
	mux.patCache[pattern] = es
	mux.routes[name] = pattern

	re := regexp.MustCompile(`(:[^\/]+)`)
	pattern = re.ReplaceAllStringFunc(pattern, func(s string) string {
		es.pName = append(es.pName, s[1:])
		return ":*"
	})

	mux.root.addPat(pattern, es)
}

// cleanPath returns the canonical path for p, eliminating . and .. elements.
// copy from go std net/http package
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

// isSlashRune
// copy from go std net/http package
func isSlashRune(r rune) bool { return r == '/' || r == '\\' }

// containsDotDot
// copy from go std net/http package
func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}
