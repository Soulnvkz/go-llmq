package log

import (
	"log"
	"os"
)

var logger_info *log.Logger = log.New(os.Stdout, "INFO:", log.LstdFlags|log.Lshortfile)
var logger_error *log.Logger = log.New(os.Stderr, "ERROR:", log.LstdFlags|log.Lshortfile)

func Info() *log.Logger {
	return logger_info
}

func Error() *log.Logger {
	return logger_error
}
