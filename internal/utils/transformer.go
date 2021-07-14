package utils

import "strings"

var NsPrefix = "devops-"

func GetMarsNamespace(ns string) string {
	if strings.HasPrefix(ns, NsPrefix) {
		return ns
	}

	return NsPrefix + ns
}
