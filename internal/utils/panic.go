package utils

import (
	"runtime"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
)

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
