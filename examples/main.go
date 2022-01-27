package main

import (
	"jvmao"
	"jvmao/middleware"
)

func tM(next jvmao.HandlerFunc) jvmao.HandlerFunc {
	return func(c *jvmao.Context) error {
		return next(c)
	}
}

func main() {

	j := jvmao.New()

	h := func(c *jvmao.Context) error {
		return c.NoContent(200)
	}

	j.Use(middleware.Logger())
	j.Use(middleware.Recover())
	j.Use(tM)
	j.GET("home", "/:id/:name", h)
	j.Static("/home/arion/Develop/jvmao/examples/", "/static/")

	j.Start(":8000")

}
