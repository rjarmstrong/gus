package gus

import (
	"fmt"
	"log"
	"os"
	"io"
	"runtime/debug"
	"io/ioutil"
)

var errr = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
var dbg = log.New(ioutil.Discard, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

func SetDebugOutput(out io.Writer) {
	dbg.SetOutput(out)
}

func Debug(in ...interface{}) {
	dbg.Output(2, fmt.Sprintf("%v", in))
}

func LogErr(err error) {
	errr.Println(err)
	errr.Output(2, string(debug.Stack()))
}
