package gus

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

var ErrorLogger = log.New(os.Stderr, "GUS ERR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
var DebugLogger = log.New(os.Stdout, "GUS: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

func Debug(in ...interface{}) {
	if DebugLogger == nil {
		return
	}
	DebugLogger.Output(2, fmt.Sprintf("%v", in))
}

func LogErr(err error) {
	if ErrorLogger == nil {
		return
	}
	ErrorLogger.Println(err)
	ErrorLogger.Output(2, string(debug.Stack()))
}
