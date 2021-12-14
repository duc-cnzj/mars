package utils

import (
	"fmt"
	"runtime"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
)

func HandlePanic(title string) {
	pc, _, _, _ := runtime.Caller(2)
	file, line := runtime.FuncForPC(pc).FileLine(pc)
	called := fmt.Sprintf("%s:%d.", file, line)

	err := recover()
	if err != nil {
		mlog.Errorf("[Panic]: title: %v, err: %v --- [%s]", title, err, called)
		if app.App() != nil && app.App().IsDebug() {
			panic(err)
		}
	}
}
