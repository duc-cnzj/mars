package utils

import (
	"strings"

	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
)

// GetMarsNamespace
// prefix `devops-`
// namespace    output
// dev          devops-dev
// devops-dev   devops-dev
func GetMarsNamespace(ns string) string {
	prefix := app.Config().NsPrefix
	if strings.HasPrefix(ns, prefix) {
		return ns
	}

	return prefix + ns
}
