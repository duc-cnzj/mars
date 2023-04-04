package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFunctionName(t *testing.T) {
	name := GetFunctionName(GetFunctionName)
	assert.Equal(t, "github.com/duc-cnzj/mars/v4/internal/utils/runtime.GetFunctionName", name)
}
