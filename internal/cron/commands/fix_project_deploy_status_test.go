package commands

import (
	"testing"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_fixDeployStatus(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mh := mock.NewMockHelmer(m)
	app := testutil.MockApp(m)
	app.EXPECT().Helmer().Return(mh).AnyTimes()
	db, f := testutil.SetGormDB(m, app)
	defer f()
	assert.Nil(t, db.AutoMigrate(&models.Project{}, &models.Namespace{}))
	ns := &models.Namespace{
		Name: "test",
	}
	db.Create(ns)
	p := &models.Project{
		Name:         "app",
		NamespaceId:  ns.ID,
		DeployStatus: uint8(types.Deploy_StatusFailed),
	}
	db.Create(p)

	mh.EXPECT().ReleaseStatus("app", "test").Return(types.Deploy_StatusUnknown)
	assert.Nil(t, fixDeployStatus())
	var found models.Project
	db.First(&found, p.ID)
	assert.Equal(t, types.Deploy_StatusFailed, types.Deploy(found.DeployStatus))

	mh.EXPECT().ReleaseStatus("app", "test").Return(types.Deploy_StatusDeployed)
	assert.Nil(t, fixDeployStatus())
	db.First(&found, p.ID)

	assert.Equal(t, types.Deploy_StatusDeployed, types.Deploy(found.DeployStatus))
}
