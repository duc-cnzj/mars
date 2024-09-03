package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewClusterSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewClusterSvc(repo.NewMockK8sRepo(m), mlog.NewForConfig(nil))
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*clusterSvc).repo)
	assert.NotNil(t, svc.(*clusterSvc).logger)
}

func Test_clusterSvc_ClusterInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewClusterSvc(k8sRepo, mlog.NewForConfig(nil))

	k8sRepo.EXPECT().ClusterInfo().Return(nil)

	resp, err := svc.ClusterInfo(context.TODO(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}
