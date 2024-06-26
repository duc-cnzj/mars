package services

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

func TestRegisteredEndpoints(t *testing.T) {
	assert.Len(t, RegisteredEndpoints(), 15)
}

type testServiceRegistrar struct {
	m map[*grpc.ServiceDesc]any
}

func (t *testServiceRegistrar) RegisterService(desc *grpc.ServiceDesc, impl any) {
	t.m[desc] = impl
}

func TestRegisteredServers(t *testing.T) {
	assert.Len(t, RegisteredServers(), 15)
	sr := &testServiceRegistrar{m: map[*grpc.ServiceDesc]any{}}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Auth().Return(nil)
	app.EXPECT().Oidc().Return(nil)
	app.EXPECT().Config().Return(&config.Config{})

	for _, r := range RegisteredServers() {
		r(sr, app)
	}
	assert.Len(t, sr.m, 15)
}
