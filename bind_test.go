package jvmao

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type Post struct {
	Name string
}

type NewPost struct {
	*Post
	Name    string    `query:"d" form:"d"`
	Time    time.Time `query:"t" form:"t"`
	Content []string  `query:"c" form:"c"`
}

func TestBindQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?d=name&c=c1&c=c2", nil)
	rec := httptest.NewRecorder()
	ctx := &context{r: req, w: &Response{writer: rec}}

	p := new(NewPost)
	err := ctx.BindQuery(p)
	if err != nil {
		t.Fatal("BindQuery:", err)
	}
	if p.Name != "name" {
		fmt.Println(p.Name)
		t.Fatal("BindQuery error")
	}

	fmt.Println(p)
}

func TestBindForm(t *testing.T) {
	p := new(NewPost)
	ti := time.Date(2014, 5, 3, 13, 9, 7, 64, time.UTC)
	encoding, _ := ti.MarshalText()

	form := url.Values{}
	form.Add("d", "name")
	form.Add("t", string(encoding))
	form.Add("c", "c1")
	form.Add("c", "c2")

	req := httptest.NewRequest(http.MethodPost, "/?d=123", strings.NewReader(form.Encode()))
	req.Header.Add(HeaderContentType, MIMEApplicationForm)
	rec := httptest.NewRecorder()
	ctx := &context{r: req, w: &Response{writer: rec}}

	err := ctx.BindForm(p)
	if err != nil {
		t.Fatal("BindForm:", err)
	}

	fmt.Println(p)

}
