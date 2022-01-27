package jvmao

import (
	"fmt"
	"testing"
)

func TestJm(t *testing.T) {

	jm := New()

	jm.GET("home", "/:id/:name", func(c *Context) error { return nil })

	// assert.Equal()
	home := jm.Reverse("home", 123, "arion")
	fmt.Println(home)

}
