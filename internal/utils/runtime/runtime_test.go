package runtime

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFunctionName(t *testing.T) {
	name := GetFunctionName(GetFunctionName)
	assert.Equal(t, "github.com/duc-cnzj/mars/v4/internal/utils/runtime.GetFunctionName", name)
}
