package recovery

import (
	"errors"
	"runtime"

	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

// HandlePanic handle panic.
func HandlePanic(title string) {
	bf := make([]byte, 1024*5)
	n := runtime.Stack(bf, false)
	bf = bf[:n]

	err := recover()
	if err != nil {
		mlog.Errorf("[Panic]: title: %v, err: %v --- [%s]", title, err, string(bf))
		if app.App() != nil && app.App().IsDebug() {
			panic(err)
		}
	}
}

// HandlePanicWithCallback handle panic with custom callback.
func HandlePanicWithCallback(title string, callback func(error)) {
	bf := make([]byte, 1024*5)
	n := runtime.Stack(bf, false)
	bf = bf[:n]

	err := recover()
	if err != nil {
		switch e := err.(type) {
		case error:
			callback(e)
		case string:
			callback(errors.New(e))
		}
		mlog.Errorf("[Panic]: title: %v, err: %v --- [%s]", title, err, string(bf))
		if app.App() != nil && app.App().IsDebug() {
			panic(err)
		}
	}
}
