package utils

import "regexp"

var hostMatch = regexp.MustCompile(`\s+([\w-_]*)<\s*.Host\d+\s*>`)

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func GetPreOccupiedLenByValuesYaml(values string) int {
	var sub = 0
	if len(values) > 0 {
		submatch := hostMatch.FindAllStringSubmatch(values, -1)
		for _, i := range submatch {
			if len(i) == 2 {
				sub = max(sub, len(i[1]))
			}
		}
	}
	return sub
}
