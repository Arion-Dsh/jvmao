package jvmao

import "testing"

func TestLogger(t *testing.T) {

	logger := DefaultLogger()

	logger.SetPriority(LOG_PRINT)

	logger.Info("info log ")
	logger.Error("error log")
	// logger.Panic("panic")

}
