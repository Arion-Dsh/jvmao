package middleware

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
)

var IsTTY bool
var useColor bool

type tColor []byte

func (c tColor) WriteIn(w io.Writer, s string, args ...interface{}) {
	if IsTTY && useColor {
		w.Write(c)
	}
	fmt.Fprintf(w, s, args...)

	// reset ...
	reset := []byte{'\033', '[', '0', 'm'}
	if IsTTY && useColor {
		w.Write(reset)
	}
}

var (
	// Normal colors
	nBlack   = []byte{'\033', '[', '3', '0', 'm'}
	nRed     = []byte{'\033', '[', '3', '1', 'm'}
	nGreen   = []byte{'\033', '[', '3', '2', 'm'}
	nYellow  = []byte{'\033', '[', '3', '3', 'm'}
	nBlue    = []byte{'\033', '[', '3', '4', 'm'}
	nMagenta = []byte{'\033', '[', '3', '5', 'm'}
	nCyan    = []byte{'\033', '[', '3', '6', 'm'}
	nWhite   = []byte{'\033', '[', '3', '7', 'm'}
	// Bright colors
	bBlack   = []byte{'\033', '[', '3', '0', ';', '1', 'm'}
	bRed     = []byte{'\033', '[', '3', '1', ';', '1', 'm'}
	bGreen   = []byte{'\033', '[', '3', '2', ';', '1', 'm'}
	bYellow  = []byte{'\033', '[', '3', '3', ';', '1', 'm'}
	bBlue    = []byte{'\033', '[', '3', '4', ';', '1', 'm'}
	bMagenta = []byte{'\033', '[', '3', '5', ';', '1', 'm'}
	bCyan    = []byte{'\033', '[', '3', '6', ';', '1', 'm'}
	bWhite   = []byte{'\033', '[', '3', '7', ';', '1', 'm'}
)

func init() {
	fi, err := os.Stdout.Stat()
	if err == nil {
		m := os.ModeDevice | os.ModeCharDevice
		IsTTY = fi.Mode()&m == m
	}
	useColor = true
	if runtime.GOOS == "windows" {
		useColor = false
	}
}
func colorWrite(c tColor, s string, args ...interface{}) string {
	buf := bytes.NewBuffer([]byte{})
	c.WriteIn(buf, s, args...)
	return buf.String()
}
