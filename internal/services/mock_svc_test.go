package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
)

// TestMockMetricsStreamTopPodServer_AllMethods 逐个调用生成 mock 的全部方法，
// 让 mockgen 生成的样板代码被完整覆盖。
func TestMockMetricsStreamTopPodServer_AllMethods(t *testing.T) {
	m := gomock.NewController(t)
	srv := NewMockMetrics_StreamTopPodServer(m)

	srv.EXPECT().Context().Return(context.TODO())
	srv.EXPECT().RecvMsg(gomock.Any()).Return(nil)
	srv.EXPECT().Send(gomock.Any()).Return(nil)
	srv.EXPECT().SendHeader(gomock.Any()).Return(nil)
	srv.EXPECT().SendMsg(gomock.Any()).Return(nil)
	srv.EXPECT().SetHeader(gomock.Any()).Return(nil)
	srv.EXPECT().SetTrailer(gomock.Any())

	assert.NotNil(t, srv.Context())
	assert.NoError(t, srv.RecvMsg(nil))
	assert.NoError(t, srv.Send(&metrics.TopPodResponse{}))
	assert.NoError(t, srv.SendHeader(metadata.MD{}))
	assert.NoError(t, srv.SendMsg(nil))
	assert.NoError(t, srv.SetHeader(metadata.MD{}))
	srv.SetTrailer(metadata.MD{})
}
