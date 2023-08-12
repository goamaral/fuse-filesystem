package fusefs

import (
	"fmt"
	"log"
	"os"
)

func Info(msg string) {
	fmt.Println(msg)
}

func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

var stderr = log.New(os.Stderr, "error: ", 0)

func Error(err error) {
	stderr.Println(err)
}

func Errorf(format string, args ...interface{}) {
	Error(fmt.Errorf(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
	os.Exit(1)
}
