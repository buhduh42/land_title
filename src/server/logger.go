package server

import (
	logger "github.com/buhduh42/go-logger"
)

// logging was sort of an after thought and didn't want to keep passing a logger
// around with each struct to log, this is package global for ease
// might increase granularity if necessary
var myLogger logger.Logger

func init() {
	myLogger = logger.NopLogger()
}

func addLogger(newLogger logger.Logger) {
	if newLogger == nil {
		return
	}
	myLogger = logger.MultiLogger(myLogger, newLogger)
}
