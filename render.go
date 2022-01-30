package jvmao

import (
	"io"
	"text/template"
)

type Renderer interface {
	Render(w io.Writer, name string, data interface{}, c Context) error
}

type DefaultRenderer struct{}

func (dr *DefaultRenderer) Render(w io.Writer, name string, data interface{}, c Context) error {

	t, err := template.New(name).ParseFiles(name)
	if err != nil {
		return err
	}
	return t.Execute(w, data)

}
