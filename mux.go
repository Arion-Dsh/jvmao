package jvmao

import (
	"net/http"
	"path"
	"sort"
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
}

func newMux(jm *Jvmao) *mux {
	return &mux{
		jm:       jm,
		root:     &entry{},
		patCache: map[string]*entity{},
		routes:   map[string]string{},
	}
}

func (mux *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := mux.jm.pool.Get().(*context)
	defer mux.jm.pool.Put(c)
	c.reset(w, r)

	var err error
	var es *entity

	h := mux.jm.NotFoundHandler

	err = NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound))

	if es, err = mux.root.matchPath(r.URL.Path, []string{}); err == nil {
		if hf, ok := es.hf[r.Method]; ok {
			for i, n := range es.pName {
				c.params.Add(n, es.pValue[i])
			}
			h = hf
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
	h = applyMiddleware(h, mux.jm.middleware...)

	if err != nil {
		c.Error(err)
	}

	if err = h(c); err != nil {
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

	mux.root.addPattern(pattern, es)

	mux.patCache[pattern] = es
	mux.routes[name] = pattern
}

type entryTyp uint8

const (
	typStatic entryTyp = iota // /home
	typParam                  // /:id
	typAll
)

type entry struct {
	pattern string
	prefix  byte
	suffix  byte
	subs    [typAll + 1]entries
	es      *entity
}

type entries []*entry

type entity struct {
	pat     string
	hf      map[string]HandlerFunc // map[method]HandlerFunc
	methods []string
	pName   []string
	pValue  []string
}

type entities []*entity

func (e *entry) matchPath(path string, pValue []string) (es *entity, err error) {

	err = NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	if len(path) == 0 {
		return
	}
	var param string
	for typ, ss := range e.subs {
		if len(ss) == 0 {
			continue
		}
		var s *entry
		t := entryTyp(typ)
		switch t {
		case typStatic, typAll:
			s = ss.findSub(path[0])
			if s == nil || !strings.HasPrefix(path, s.pattern) {
				continue
			}
			path = path[len(s.pattern):]
		case typParam:
			s = ss.findSub('*')
			if s == nil {
				continue
			}

			if idx := strings.Index(path, "/"); idx > -1 {
				param, path = path[:idx], path[idx:]
			} else {
				param = path
				path = ""
			}
			pValue = append(pValue, param)

		}

		if s == nil {
			continue
		}
		if len(path) == 0 && s.es != nil {
			es = s.es
			err = nil
			es.pValue = append(es.pValue, pValue...)
			return
		}
		return s.matchPath(path, pValue)

	}
	return
}

func (es entries) findSub(prefix byte) *entry {
	num := len(es)

	idx := 0
	i, j := 0, num-1
	for i <= j {
		idx = i + (j-i)/2
		if prefix > es[idx].prefix {
			i = idx + 1
		} else if prefix < es[idx].prefix {
			j = idx - 1
		} else {
			i = num
		}
	}
	if es[idx].prefix != prefix {
		return nil
	}
	return es[idx]
}

func (e *entry) addPattern(pattern string, es *entity) {

	var typ entryTyp = typStatic
	var pp, param string

	var parent *entry = e
	var xp *entry
	var prefix byte
	pat := pattern
	var ses *entity

	typ, pp, pat, param = splitPattern(pat)
	if len(pp) == 0 {
		return
	}

	if len(param) > 0 {
		es.pName = append(es.pName, param)
	}

	switch pp {
	case "*", "/":
		if len(pat) == 0 {
			ses = es
		}
		parent = e.addSub(typ, pp, ses)
		parent.addPattern(pat, es)

	default:
		prefix = pp[0]
		xp = e.findSub(typ, prefix)

		if xp == nil {
			if len(pat) == 0 {
				ses = es
			}
			e.addSub(typ, pp, ses)
			e.addPattern(pat, es)
			return
		}

		common := longestCommon(pp, xp.pattern)

		sp := xp.pattern[:common]
		parent = &entry{pattern: sp, prefix: sp[0]}
		if len(xp.pattern) > common {
			e.delSub(typ, xp)
			// split xp
			xp.pattern = xp.pattern[common:]
			xp.prefix = xp.pattern[0]
			parent.addSubByType(typ, xp)

			if sp == "/" {
				e.addSubByType(typAll, parent)
			} else {
				e.addSubByType(typ, parent)
			}

		} else {
			parent = xp
		}
		if len(pp) > common {
			parent.addPattern(pp[common:]+pat, es)
		} else {
			parent.es = es
		}
	}
}

func (e *entry) addSubByType(typ entryTyp, s *entry) *entry {

	switch typ {
	case typStatic:
		e.subs[typ] = append(e.subs[typ], s)
		e.subs[typ].Sort()
	default:
		if len(e.subs[typ]) > 0 {
			e.subs[typ][0] = s
		} else {
			e.subs[typ] = append(e.subs[typ], s)
		}
	}
	return s

}
func (e *entry) addSub(typ entryTyp, pattern string, es *entity) *entry {
	sub := &entry{pattern: pattern, prefix: pattern[0], es: es}
	e.addSubByType(typ, sub)
	return sub
}

func (e *entry) delSub(typ entryTyp, s *entry) {
	for i, ss := range e.subs[typ] {
		if ss == s {
			e.subs[typ] = append(e.subs[typ][:i], e.subs[typ][i+1:]...)
			return
		}
	}
}

func (e *entry) findSub(typ entryTyp, prefix byte) *entry {

	es := e.subs[typ]
	num := len(es)
	if num == 0 {
		return nil
	}
	idx := 0

	switch typ {
	case typStatic:

		i, j := 0, num-1
		for i <= j {
			idx = i + (j-i)/2
			if prefix > es[idx].prefix {
				i = idx + 1
			} else if prefix < es[idx].prefix {
				j = idx - 1
			} else {
				i = num // breaks cond
			}
		}
		if es[idx].prefix != prefix {
			return nil
		}

		return es[idx]
	default:
		return es[idx]

	}

}

func splitPattern(p string) (t entryTyp, pp, next, param string) {
	t = typStatic
	i := strings.Index(p, ":")
	si := strings.Index(p, "/")

	// :ad
	// :ad/pwr/sg
	if strings.HasPrefix(p, ":") {

		t = typParam
		pp = "*"
		param = p[1:]
		if si != -1 {
			param = p[i+1 : si]
			if si > 0 {
				next = p[si:]
			}
		}
		return
	}
	if i == -1 {
		pp = p
		return
	}

	// /saf/:afg/sdf
	pp, next = p[:i], p[i:]
	if pp == "/" {
		t = typAll
	}
	return

}

// Sort the list of entries by prefix
func (es entries) Sort()              { sort.Sort(es) }
func (es entries) Len() int           { return len(es) }
func (es entries) Swap(i, j int)      { es[i], es[j] = es[j], es[i] }
func (es entries) Less(i, j int) bool { return es[i].prefix < es[j].prefix }

// longestCommon finds the length of the shared of two strings
func longestCommon(k1, k2 string) int {
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}
	var i int
	for i = 0; i < max; i++ {
		if k1[i] != k2[i] {
			break
		}
	}
	return i
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
