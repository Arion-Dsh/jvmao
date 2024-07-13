package jvmao

import (
	"encoding"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type BindUnmarshaler interface {
	UnmarshalBind(i interface{}) error
}

func (c *context) BindQuery(dest interface{}) error {
	return bind("query", dest, c)
}

func (c *context) BindForm(dest interface{}) error {
	ct := c.r.Header.Get(HeaderContentType)
	switch {
	case strings.HasPrefix(ct, MIMEApplicationForm), strings.HasPrefix(ct, MIMEMultipartForm):
		return bind("form", dest, c)
	case strings.HasPrefix(ct, MIMEApplicationJSON):
		return json.NewDecoder(c.r.Body).Decode(dest)
	default:
		return errors.New("BindForm unsupports header content type " + ct)
	}

}

func (c *context) BindParam(dest interface{}) error {
	return bind("param", dest, c)
}

func bind(tag string, dest interface{}, c Context) error {

	typ := reflect.TypeOf(dest).Elem()

	if typ.Kind() != reflect.Struct {
		return errors.New("bind elem must be struct.")
	}

	val := reflect.ValueOf(dest).Elem()

	for i := 0; i < typ.NumField(); i++ {

		tf := typ.Field(i)

		fName, hast := tf.Tag.Lookup(tag)
		if !hast || fName == "-" {
			continue
		}
		fName = strings.Split(fName, ",")[0]

		vf := val.Field(i)
		if tf.Anonymous {
			// when struct field is struct just igrone.
			if vf.Kind() == reflect.Struct {
				continue
			}
			vf = vf.Elem()
		}

		if !vf.CanSet() {
			continue
		}

		if vf.Kind() == reflect.Slice {
			d := getSliceData(tag, fName, c)
			if len(d) > 0 {
				_ = bindSliceField(vf, d)
			}
		} else {
			d := getData(tag, fName, c)
			_ = bindField(vf, d)
		}
	}

	return nil
}

func bindSliceField(v reflect.Value, data []string) error {

	l := len(data)

	slice := reflect.MakeSlice(v.Type(), l, l)

	for i := 0; i < l; i++ {
		if err := bindField(slice.Index(i), data[i]); err != nil {
			return err
		}
	}
	v.Set(slice)
	return nil
}

func bindField(v reflect.Value, data string) error {

	inter := v.Addr().Interface()
	if um, ok := inter.(BindUnmarshaler); ok {
		return um.UnmarshalBind(data)
	}
	if um, ok := inter.(encoding.TextUnmarshaler); ok {
		return um.UnmarshalText([]byte(data))
	}

	switch v.Kind() {

	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		_ = bindField(v.Elem(), data)
	case reflect.String:
		v.SetString(data)
	case reflect.Bool:
		if data == "" {
			data = "false"
		}
		b, err := strconv.ParseBool(data)

		if err != nil {
			return err
		}
		v.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convSetInt(data, v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return convSetUint(data, v)

	case reflect.Float32, reflect.Float64:
		return convSetFloat(data, v)

	default:
		return errors.New("type unsupported.")

	}

	return nil

}

func convSetFloat(s string, v reflect.Value) error {
	if s == "" {
		s = "0.0"
	}
	bitSize := 32
	if v.Kind() == reflect.Float64 {
		bitSize = 64
	}

	f, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		return err
	}
	v.SetFloat(f)
	return nil
}

func convSetUint(s string, v reflect.Value) error {
	if s == "" {
		s = "0"
	}
	bitSize := 0

	switch v.Kind() {
	case reflect.Int8:
		bitSize = 8
	case reflect.Int16:
		bitSize = 16
	case reflect.Int32:
		bitSize = 32
	case reflect.Int64:
		bitSize = 32
	}

	i, err := strconv.ParseUint(s, 10, bitSize)
	if err != nil {
		return err
	}
	v.SetUint(i)

	return nil

}

func convSetInt(s string, v reflect.Value) error {
	if s == "" {
		s = "0"
	}
	bitSize := 0

	switch v.Kind() {
	case reflect.Int8:
		bitSize = 8
	case reflect.Int16:
		bitSize = 16
	case reflect.Int32:
		bitSize = 32
	case reflect.Int64:
		bitSize = 32
	}

	i, err := strconv.ParseInt(s, 10, bitSize)
	if err != nil {
		return err
	}
	v.SetInt(i)
	return nil
}

func getSliceData(tag, key string, c Context) (d []string) {

	switch tag {
	case "query":
		d = c.QueryValues(key)
	case "form":
		d = c.FormValues(key)
	default:
		d = []string{}
	}
	return

}
func getData(tag, key string, c Context) (d string) {
	switch tag {
	case "query":
		d = c.QueryValue(key)
	case "form":
		d = c.FormValue(key)
	case "param":
		d = c.ParamValue(key)
	default:
		d = ""
	}
	return
}
