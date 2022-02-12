package main

import (
	"fmt"
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

	j.Static("static/", "/static/")

	// err := j.StartTLS(":8000", "server.crt", "server.key")
	err := j.Start(":8000")
	// j.AutoTLSManager.HostPolicy = autocert.HostWhitelist("example.org", "localhost")
	// err := j.StartAutoTLS(":4430")
	fmt.Println(err)

}
