package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/cluster"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Test_clusterSvc_DeployTrend_BucketsZeroFill 成功路径：repo 回两条位于不同天的 created_at，
// 服务端按本地天界分桶：非零落在对应天、中间无部署的天补 0、长度恒等于 days 且升序。
func Test_clusterSvc_DeployTrend_BucketsZeroFill(t *testing.T) {
	svc, mocks := newClusterSvcWithMocks(t)

	loc := time.Local
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	// 两个采样点：两天前 + 今天（均落在窗口 [today-2, today+1) 内），用来验证分桶落位
	t0 := todayStart.AddDate(0, 0, -2).Add(6 * time.Hour) // 第 0 天
	t1 := todayStart.Add(13 * time.Hour)                  // 第 2 天（今天）
	mocks.clRepo.EXPECT().SelectCreatedAtBetween(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]time.Time{t0, t1}, nil)

	resp, err := svc.DeployTrend(context.TODO(), &cluster.DeployTrendRequest{Days: 3})
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, int32(3), resp.Days)
		assert.Len(t, resp.Items, 3)
		assert.Equal(t, t0.In(loc).Format("2006-01-02"), resp.Items[0].Date)
		assert.Equal(t, int32(1), resp.Items[0].Count)
		assert.Equal(t, int32(0), resp.Items[1].Count) // 中间天无部署 → 补 0
		assert.Equal(t, t1.In(loc).Format("2006-01-02"), resp.Items[2].Date)
		assert.Equal(t, int32(1), resp.Items[2].Count)
		assert.Equal(t, resp.Items[0].Date, resp.StartDate)
		assert.Equal(t, resp.Items[2].Date, resp.EndDate)
	}
}

// Test_clusterSvc_DeployTrend_DefaultAndClamp 窗口收敛：days 缺省取默认 30；超过上限收敛到 90。
func Test_clusterSvc_DeployTrend_DefaultAndClamp(t *testing.T) {
	svc, mocks := newClusterSvcWithMocks(t)

	// days=0/缺省 → 默认 30，空仓库 → 30 天全 0
	mocks.clRepo.EXPECT().SelectCreatedAtBetween(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil)
	resp, err := svc.DeployTrend(context.TODO(), nil)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, int32(biz.DeployTrendDefaultDays), resp.Days)
		assert.Len(t, resp.Items, biz.DeployTrendDefaultDays)
		for _, it := range resp.Items {
			assert.Zero(t, it.Count)
		}
	}

	// days=999 → 收敛到上限 90
	mocks.clRepo.EXPECT().SelectCreatedAtBetween(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil)
	resp2, err := svc.DeployTrend(context.TODO(), &cluster.DeployTrendRequest{Days: 999})
	assert.NoError(t, err)
	if assert.NotNil(t, resp2) {
		assert.Equal(t, int32(biz.DeployTrendMaxDays), resp2.Days)
		assert.Len(t, resp2.Items, biz.DeployTrendMaxDays)
	}
}

// Test_clusterSvc_DeployTrend_Error 失败路径：repo 查询失败时错误上抛（不吞错返回空表）。
func Test_clusterSvc_DeployTrend_Error(t *testing.T) {
	svc, mocks := newClusterSvcWithMocks(t)

	mocks.clRepo.EXPECT().SelectCreatedAtBetween(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db down"))
	resp, err := svc.DeployTrend(context.TODO(), &cluster.DeployTrendRequest{Days: 7})
	assert.Nil(t, resp)
	assert.Error(t, err)
}
