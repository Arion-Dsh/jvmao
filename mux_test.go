package jvmao

import (
	"fmt"
	"testing"
)

//            /
//            a
//     c                 b
//     /             cd    c
//    c d              /     /

func TestMux(t *testing.T) {
	n := &entry{}
	// n.addPattern("/ab cd", entity{pat: "/abcd"})
	// n.addPattern("/ab ", entity{pat: "/ab"})
	// n.addPattern("/ab/dc", entity{pat: "/ab/dc"})
	n.addPattern("/abcsfg", &entity{pat: "/abccsfg"})
	n.addPattern("/abc", &entity{pat: "/abc"})
	n.addPattern("/abc/dd", &entity{pat: "/abc/d"})

	// n.addattern("/abc/cc/ds", entity{pat: "/abc/c"})
	// n.addPattern("/abc/:id", entity{pat: "/abc/:id"})
	n.addPattern("/abc/:id/:ui", &entity{pat: "/abc/:id/:ui"})
	// n.addPattern("/abcdef", entity{})
	e, err := n.matchPath("/abc/123/sdg", []string{})
	if err == nil {
		fmt.Println(e)
	}
	// fmt.Println(n.subs)

}

func BenchmarkMux(b *testing.B) {
	n := &entry{}
	n.addPattern("/abcd", &entity{pat: "/abcd"})
	n.addPattern("/ab", &entity{pat: "/ab"})
	n.addPattern("/abc", &entity{pat: "/abc"})
	// n.addPattern("/abcdef", entity{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = n.matchPath("/abcd", []string{})

	}
}
