package jvmao

import (
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogPriority int

const (
	// LOG__EMERG LogPriority = iota
	LOG_PANIC = iota + 1
	LOG_FATAL
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
	LOG_PRINT
)

type Logger interface {
	Output() io.Writer
	SetOutput(w io.Writer)
	Priorty() LogPriority
	SetPriority(p LogPriority)
	Print(s string)
	Debug(s string)
	Info(s string)
	Warn(s string)
	Error(s string)
	Fatal(s string)
	Panic(s string)
}

type logger struct {
	p LogPriority
	w io.Writer

	mu sync.Mutex
}

func DefaultLogger() Logger {
	return &logger{p: LOG_ERR, w: os.Stdout}
}

func (l *logger) Output() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w
}

func (l *logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.w = w
}

func (l *logger) Priorty() LogPriority {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.p
}

func (l *logger) SetPriority(p LogPriority) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.p = p
}
func (l *logger) Print(s string) {
	l.log(LOG_PRINT, NBlack, "PRINT", s)
}
func (l *logger) Debug(s string) {
	l.log(LOG_DEBUG, NBlack, "DEBUG", s)
}
func (l *logger) Info(s string) {
	l.log(LOG_INFO, BGreen, "INFO", s)
}
func (l *logger) Warn(s string) {
	l.log(LOG_WARNING, BCyan, "WARNING", s)
}

func (l *logger) Error(s string) {
	l.log(LOG_ERR, BRed, "ERROR", s)
}
func (l *logger) Fatal(s string) {
	l.log(LOG_FATAL, BRed, "FATAL", s)
}
func (l *logger) Panic(s string) {
	l.log(LOG_PANIC, BRed, "PANIC", s)
	panic(s)
}

func (l *logger) log(p LogPriority, c TColor, tag, msg string) {
	if p > l.p {
		return
	}
	useColor := true
	if p > LOG_INFO {
		useColor = false
	}
	nl := ""
	if !strings.HasSuffix(msg, "\n") {
		nl = "\n"
	}
	timestamp := time.Now().Format(time.Stamp)
	s := "[%s]%s %s line %d : %s%s"
	_, file, line, _ := runtime.Caller(2)
	all := TColorWrite(c, useColor, s, tag, timestamp, path.Base(file), line, msg, nl)
	l.w.Write([]byte(all))
}
