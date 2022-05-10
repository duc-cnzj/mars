package instance

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	contracts.ApplicationInterface
}

func TestApp(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	assert.Nil(t, App())
	ta := &testApp{}
	SetInstance(ta)
	assert.Same(t, ta, App())
}
