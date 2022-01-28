package jvmao

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

var tPool sync.Pool

var isTTY bool

func init() {

	tPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer([]byte{})
		},
	}

	fi, err := os.Stdout.Stat()
	if err == nil {
		m := os.ModeDevice | os.ModeCharDevice
		isTTY = fi.Mode()&m == m
	}
	if runtime.GOOS == "windows" {
		isTTY = false
	}
}

// TColorWrite return string with terminal color
func TColorWrite(c TColor, useColor bool, s string, args ...interface{}) string {
	buf := tPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer tPool.Put(buf)
	c.WriteIn(buf, useColor, s, args...)
	return buf.String()
}

// TColor color in terminal, igonre in windows
type TColor []byte

//WriteIn write string in io.Writer with color
func (c TColor) WriteIn(w io.Writer, useColor bool, s string, args ...interface{}) {
	if isTTY && useColor {
		w.Write(c)
	}
	fmt.Fprintf(w, s, args...)

	// reset
	if isTTY && useColor {
		w.Write([]byte{'\033', '[', '0', 'm'})
	}
}

var (
	// Normal colors
	NBlack   = []byte{'\033', '[', '3', '0', 'm'}
	NRed     = []byte{'\033', '[', '3', '1', 'm'}
	NGreen   = []byte{'\033', '[', '3', '2', 'm'}
	NYellow  = []byte{'\033', '[', '3', '3', 'm'}
	NBlue    = []byte{'\033', '[', '3', '4', 'm'}
	NMagenta = []byte{'\033', '[', '3', '5', 'm'}
	NCyan    = []byte{'\033', '[', '3', '6', 'm'}
	NWhite   = []byte{'\033', '[', '3', '7', 'm'}
	// Bright colors
	BBlack   = []byte{'\033', '[', '3', '0', ';', '1', 'm'}
	BRed     = []byte{'\033', '[', '3', '1', ';', '1', 'm'}
	BGreen   = []byte{'\033', '[', '3', '2', ';', '1', 'm'}
	BYellow  = []byte{'\033', '[', '3', '3', ';', '1', 'm'}
	BBlue    = []byte{'\033', '[', '3', '4', ';', '1', 'm'}
	BMagenta = []byte{'\033', '[', '3', '5', ';', '1', 'm'}
	BCyan    = []byte{'\033', '[', '3', '6', ';', '1', 'm'}
	BWhite   = []byte{'\033', '[', '3', '7', ';', '1', 'm'}
)
