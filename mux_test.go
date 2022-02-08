package jvmao

import (
	"fmt"
	"testing"
)

func TestMux(t *testing.T) {

	mux := newMux(nil)
	mux.handle("test", "1", "/abc/agafsf", nil)
	mux.handle("test0", "2", "/abc/:a", nil)
	mux.handle("test1", "3", "/abc/:a/d", nil)
	mux.handle("test2", "4", "/abc/abc/d", nil)
	mux.handle("test3", "GET", "/abc/:a/b", nil)
	if e := mux.root.match(mux.ctx, "/abc/123"); e != nil {
		fmt.Println("ok:", e)
	}
}

func TestEntry(t *testing.T) {

	et := new(entry)
	et.addPat("/a/:*/c", &entity{pat: "0"})
	et.addPat("/abc/:*/d/:*", &entity{pat: "1"})
	et.addPat("/a/:*", &entity{pat: "2"})
	et.addPat("/abcd", &entity{pat: "3"})
	et.addPat("/ab", &entity{pat: "4"})
	et.addPat("/abc", &entity{pat: "5"})

	paths := map[string]string{
		"1": "/abc/sf/d/123",
		"2": "/a/sf",
		"3": "/abcd",
		"4": "/ab",
		"5": "/abc",
	}
	ctx := new(muxCtx)
	for pat, url := range paths {
		ctx.reset()

		if e := et.match(ctx, url); e != nil {
			if e.pat != pat {
				t.Error("error: ", e.pat)
			}
		} else {
			t.Fatal("error:", pat)
		}

	}
}

func BenchmarkMux(b *testing.B) {

	et := new(entry)
	et.addPat("/a/:*/c", &entity{pat: "0"})
	et.addPat("/abc/:*/d/:*", &entity{pat: "1"})
	et.addPat("/a/:*", &entity{pat: "2"})
	et.addPat("/abcd", &entity{pat: "3"})
	et.addPat("/ab", &entity{pat: "4"})
	et.addPat("/abc", &entity{pat: "5"})

	// n.addPattern("/abcdef", entity{})
	b.ReportAllocs()
	b.ResetTimer()
	ctx := new(muxCtx)
	for i := 0; i < b.N; i++ {
		ctx.reset()
		_ = et.match(ctx, "/abcd")

	}
}
