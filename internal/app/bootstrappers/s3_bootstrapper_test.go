package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestS3Bootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := app.NewMockApp(m)
	mockData := data.NewMockData(m)
	app.EXPECT().Data().Return(mockData)
	mockData.EXPECT().InitS3()
	assert.Nil(t, (&S3Bootstrapper{}).Bootstrap(app))
}

func TestS3Bootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&S3Bootstrapper{}).Tags())
}
