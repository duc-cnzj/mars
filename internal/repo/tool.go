package repo

import "regexp"

type ToolRepo interface {
	GetPreOccupiedLenByValuesYaml(values string) int
}

var _ ToolRepo = (*toolRepo)(nil)

type toolRepo struct{}

func NewToolRepo() ToolRepo {
	return &toolRepo{}
}

var hostMatch = regexp.MustCompile(`\s+([\w-_]*)<\s*.Host\d+\s*>`)

func (*toolRepo) GetPreOccupiedLenByValuesYaml(values string) int {
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
