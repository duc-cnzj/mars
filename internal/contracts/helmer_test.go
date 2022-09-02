package contracts

import (
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWrapLogFn_UnWrap(t *testing.T) {
	called := 0
	var fn WrapLogFn = func(container []*types.Container, format string, v ...any) {
		called++
	}
	fn.UnWrap()("")
	assert.Equal(t, 1, called)
}
