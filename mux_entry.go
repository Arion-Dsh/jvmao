package jvmao

import (
	"sort"
	"strings"
)

type entryTyp uint8

const (
	typStatic entryTyp = iota // /home
	typParam                  // /:id
	typAll
)

type entry struct {
	typ entryTyp

	pattern string
	prefix  byte
	sub     [typAll + 1]entries
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

func (e *entry) match(ctx *muxCtx, path string) *entity {

	for typ, ss := range e.sub {
		if len(ss) == 0 {
			continue
		}

		var sub *entry

		switch entryTyp(typ) {
		case typParam:
			sub = ss.findSub('*')
			ps := strings.SplitN(path, "/", 2)
			pattern := "*"
			if len(ps) == 2 && len(ps[1]) > 0 {
				pattern = "*/" + ps[1]
			}
			if !strings.HasPrefix(pattern, sub.pattern) {
				continue
			}
			path = strings.TrimPrefix(pattern, sub.pattern)
			ctx.param(ps[0])
		default:
			sub = ss.findSub(path[0])
			if sub == nil || !strings.HasPrefix(path, sub.pattern) {
				continue
			}
			path = strings.TrimPrefix(path, sub.pattern)
		}

		if sub == nil {
			continue
		}

		if len(path) == 0 {
			if sub.es != nil {
				return sub.es
			} else {
				return nil
			}
		}

		return sub.match(ctx, path)
	}
	return nil
}

func (e *entry) addPat(pat string, es *entity) {

	var prefix byte

	typ, pp, next := splitPat(pat)

	if len(pp) == 0 {
		return
	}
	prefix = pp[0]

	sub := e.sub[typ].findSub(prefix)
	if sub == nil {
		et := &entry{pattern: pp, prefix: prefix, typ: typ}
		e.sub[typ] = append(e.sub[typ], et)
		e.sub[typ].Sort()
		if len(next) == 0 {
			et.es = es
		} else {
			et.addPat(next, es)
		}
		return
	}

	//remove sub
	e.subDel(sub)

	common := longestCommon(pp, sub.pattern)
	et := &entry{pattern: pp[:common], prefix: prefix, typ: typ}
	// split sub entry
	if len(sub.pattern) > common {
		sub.pattern = sub.pattern[common:]
		sub.prefix = sub.pattern[0]
		et.subAdd(sub)
	} else {
		et = sub
	}

	if len(pp) > common {
		et.addPat(pp[common:]+":"+next, es)
	} else {
		if len(next) == 0 {
			et.es = es
		} else {
			et.addPat(next, es)

		}
	}
	e.sub[typ] = append(e.sub[typ], et)
	e.sub[typ].Sort()
}

func (e *entry) subFind(typ entryTyp, prefix byte) *entry {
	for _, s := range e.sub[typ] {
		if s.prefix == prefix {
			return s
		}
	}
	return nil
}

func (e *entry) subDel(sub *entry) {

	subs := e.sub[sub.typ]

	for i, s := range subs {
		if sub == s {
			e.sub[sub.typ] = append(subs[:i], subs[i+1:]...)
			return
		}
	}

}

func (e *entry) subAdd(sub *entry) {
	subs := e.sub[sub.typ]

	e.sub[sub.typ] = append(subs, sub)
	e.sub[sub.typ].Sort()
}

func (es entries) findSub(prefix byte) *entry {
	num := len(es)
	if num == 0 {
		return nil
	}

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

// Sort the list of entries by prefix
func (es entries) Sort()              { sort.Sort(es) }
func (es entries) Len() int           { return len(es) }
func (es entries) Swap(i, j int)      { es[i], es[j] = es[j], es[i] }
func (es entries) Less(i, j int) bool { return es[i].prefix < es[j].prefix }

func splitPat(p string) (t entryTyp, pp, next string) {
	// /bac/f :d/c
	// /bac/f :d/c/ :e/f
	if p == "" {
		return
	}
	if strings.HasPrefix(p, ":") {
		p = p[1:]
	}

	t = typStatic

	// /saf/:afg/sdf
	ps := strings.SplitN(p, ":", 2)
	pp = ps[0]
	if len(ps) == 2 {
		next = ps[1]
	}
	if strings.HasPrefix(pp, "*") {
		t = typParam

	}
	return

}

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
