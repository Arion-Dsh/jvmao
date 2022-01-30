package main

import (
	"net/http"

	"github.com/arion-dsh/jvmao/middleware"

	"github.com/arion-dsh/jvmao"
)

func tM(next jvmao.HandlerFunc) jvmao.HandlerFunc {
	return func(c jvmao.Context) error {
		return next(c)
	}
}

func main() {

	j := jvmao.New()

	h := func(c jvmao.Context) error {
		return c.String(http.StatusOK, "123")
	}

	j.Use(middleware.Logger())
	j.Use(middleware.Recover())
	j.Use(tM)

	j.GET("home", "", h)

	g := j.Group("/group")
	g.GET("g-home", "", h)

	j.Static("/home/arion/Develop/jvmao/examples/", "/static/")

	// j.StartTLS(":8000", "server.crt", "server.key")
	j.Start(":8000")
	// j.StartAutoTLS(":8000")

}
