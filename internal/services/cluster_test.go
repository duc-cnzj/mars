package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewClusterSvc(t *testing.T) {
	svc, _ := newClusterSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.k8sBiz)
	assert.NotNil(t, svc.logger)
}

func Test_clusterSvc_ClusterInfo(t *testing.T) {
	svc, mocks := newClusterSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo

	k8sRepo.EXPECT().ClusterInfo().Return(&biz.ClusterInfo{
		Status:            "ok",
		FreeMemory:        "100Mi",
		FreeCpu:           "2",
		FreeRequestMemory: "50Mi",
		FreeRequestCpu:    "1",
		TotalMemory:       "200Mi",
		TotalCpu:          "4",
		UsageMemoryRate:   "0.5",
		UsageCpuRate:      "0.25",
		RequestMemoryRate: "0.3",
		RequestCpuRate:    "0.4",
	})

	resp, err := svc.ClusterInfo(context.TODO(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	if assert.NotNil(t, resp.Item) {
		assert.Equal(t, "ok", resp.Item.Status)
		assert.Equal(t, "100Mi", resp.Item.FreeMemory)
		assert.Equal(t, "2", resp.Item.FreeCpu)
		assert.Equal(t, "50Mi", resp.Item.FreeRequestMemory)
		assert.Equal(t, "1", resp.Item.FreeRequestCpu)
		assert.Equal(t, "200Mi", resp.Item.TotalMemory)
		assert.Equal(t, "4", resp.Item.TotalCpu)
		assert.Equal(t, "0.5", resp.Item.UsageMemoryRate)
		assert.Equal(t, "0.25", resp.Item.UsageCpuRate)
		assert.Equal(t, "0.3", resp.Item.RequestMemoryRate)
		assert.Equal(t, "0.4", resp.Item.RequestCpuRate)
	}
}

type clusterSvcMocks struct {
	ctrl    *gomock.Controller
	k8sRepo *data.MockK8sRepo
}

func newClusterSvcWithMocks(t *testing.T) (*clusterSvc, *clusterSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &clusterSvcMocks{
		ctrl:    ctrl,
		k8sRepo: data.NewMockK8sRepo(ctrl),
	}
	s, ok := NewClusterSvc(ClusterSvcDeps{
		K8sBiz: biz.NewK8sBiz(mocks.k8sRepo),
		Logger: mlog.NewForConfig(nil),
	}).(*clusterSvc)
	if !ok {
		panic("NewClusterSvc returned unexpected type")
	}
	return s, mocks
}
