package utils

import (
	"strings"

	app "github.com/duc-cnzj/mars/internal/app/helper"
)

func GetMarsNamespace(ns string) string {
	prefix := app.Config().NsPrefix
	if strings.HasPrefix(ns, prefix) {
		return ns
	}

	return prefix + ns
}
