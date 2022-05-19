package testutil

import (
	"testing"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSetGormDB(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	db, f := SetGormDB(m, app)
	defer f()
	assert.NotNil(t, db)
	assert.Equal(t, db, app.DBManager().DB())
}

func TestMockApp(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	a := MockApp(m)
	assert.Same(t, app.App(), a)
}
