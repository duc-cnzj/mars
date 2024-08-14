package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v4/changelog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewChangelogSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewChangelogSvc(repo.NewMockChangelogRepo(m))
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*changelogSvc).repo)
}

func Test_changelogSvc_FindLastChangelogsByProjectID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	changelogRepo := repo.NewMockChangelogRepo(m)
	svc := NewChangelogSvc(changelogRepo)

	changelogRepo.EXPECT().FindLastChangelogsByProjectID(gomock.Any(), &repo.FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        true,
		ProjectID:          1,
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              5,
	}).Return(nil, errors.New("x"))

	_, err := svc.FindLastChangelogsByProjectID(context.TODO(), &changelog.FindLastChangelogsByProjectIDRequest{
		ProjectId:   1,
		OnlyChanged: true,
	})
	assert.Error(t, err)
}

func Test_changelogSvc_FindLastChangelogsByProjectID_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	changelogRepo := repo.NewMockChangelogRepo(m)
	svc := NewChangelogSvc(changelogRepo)

	changelogRepo.EXPECT().FindLastChangelogsByProjectID(gomock.Any(), &repo.FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        true,
		ProjectID:          1,
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              5,
	}).Return([]*repo.Changelog{}, nil)

	resp, err := svc.FindLastChangelogsByProjectID(context.TODO(), &changelog.FindLastChangelogsByProjectIDRequest{
		ProjectId:   1,
		OnlyChanged: true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}
