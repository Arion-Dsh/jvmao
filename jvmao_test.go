package jvmao

import (
	"testing"
)

func TestJm(t *testing.T) {

	jm := New()

	jm.GET("home", "/:id/:name", func(c Context) error { return nil })
	jm.GET("home1", "/:id/name", func(c Context) error { return nil })

	// assert.Equal()
	// home := jm.Reverse("home", "123", "arion")
	// home1 := jm.Reverse("home1", "123")
	// fmt.Println(home, home1)

}
