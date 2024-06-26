package instance

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
