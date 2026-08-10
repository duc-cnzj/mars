package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewHttpHandler(t *testing.T) {
	handler, mocks := newHttpHandlerWithMocks(t)
	httpHandler := mocks.httpServer
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.logger)
	// 门面按关注点拆分：swagger 与文件处理器独立存在，不再共享同一依赖集。
	assert.NotNil(t, handler.swagger)
	assert.NotNil(t, handler.files)
	assert.NotNil(t, handler.files.logger)
	assert.NotNil(t, handler.files.authBiz)
	assert.NotNil(t, handler.files.uploader)
	assert.NotNil(t, handler.files.eventBiz)
	assert.NotNil(t, handler.files.fileBiz)
	assert.NotNil(t, handler.files.timer)
	assert.NotNil(t, handler.files.k8sBiz)
	assert.NotNil(t, handler.files.accessBiz)

	router := mux.NewRouter()
	handler.RegisterWsRoute(router)
	handler.RegisterSwaggerUIRoute(router)
	handler.RegisterFileRoute(runtime.NewServeMux())

	httpHandler.EXPECT().Shutdown(gomock.Not(nil)).Times(1)
	_ = handler.Shutdown(context.TODO())
}

func Test_httpHandlerImpl_Shutdown_Error(t *testing.T) {
	handler, mocks := newHttpHandlerWithMocks(t)
	ws := mocks.httpServer

	ws.EXPECT().Shutdown(gomock.Any()).Return(errors.New("boom"))

	err := handler.Shutdown(context.TODO())
	assert.Error(t, err)
}

type httpHandlerMocks struct {
	ctrl       *gomock.Controller
	httpServer *application.MockHttpHandler
	authBiz    *biz.MockAuthBiz
	uploader   *uploader.MockUploader
	fileRepo   *data.MockFileRepo
	eventRepo  *data.MockEventRepo
	k8sRepo    *data.MockK8sRepo
	nsRepo     *data.MockNamespaceRepo
}

func newHttpHandlerMocks(t *testing.T) *httpHandlerMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &httpHandlerMocks{
		ctrl:       ctrl,
		httpServer: application.NewMockHttpHandler(ctrl),
		authBiz:    biz.NewMockAuthBiz(ctrl),
		uploader:   uploader.NewMockUploader(ctrl),
		fileRepo:   data.NewMockFileRepo(ctrl),
		eventRepo:  data.NewMockEventRepo(ctrl),
		k8sRepo:    data.NewMockK8sRepo(ctrl),
		nsRepo:     data.NewMockNamespaceRepo(ctrl),
	}
}

// buildHttpHandlerDeps 组装 NewHttpHandler 的完整依赖面，供各子处理器测试复用。
func buildHttpHandlerDeps(t *testing.T, mocks *httpHandlerMocks) HttpHandlerDeps {
	t.Helper()
	logger := mlog.NewForConfig(nil)
	return HttpHandlerDeps{
		WsHttpServer: mocks.httpServer,
		Logger:       logger,
		Uploader:     mocks.uploader,
		AuthBiz:      mocks.authBiz,
		EventBiz:     biz.NewEventBiz(mocks.eventRepo),
		FileBiz:      biz.NewFileBiz(mocks.fileRepo),
		Timer:        timer.NewReal(),
		K8sBiz:       biz.NewK8sBiz(mocks.k8sRepo),
		// copyFromPod 的容器解析走真实 ContainerBiz：非空 container 直返，不触达 k8sRepo。
		ContainerBiz: biz.NewContainerBiz(logger, biz.NewK8sBiz(mocks.k8sRepo), biz.NewFileBiz(mocks.fileRepo), biz.NewEventBiz(mocks.eventRepo), timer.NewReal()),
		AccessBiz:    biz.NewAccessBiz(logger, biz.NewNsRepoBiz(mocks.nsRepo), nil),
	}
}

func newHttpHandlerWithMocks(t *testing.T) (*httpHandlerImpl, *httpHandlerMocks) {
	t.Helper()
	mocks := newHttpHandlerMocks(t)
	h, ok := NewHttpHandler(buildHttpHandlerDeps(t, mocks)).(*httpHandlerImpl)
	if !ok {
		panic("NewHttpHandler returned unexpected type")
	}
	return h, mocks
}

func newFileHandlerWithMocks(t *testing.T) (*fileHandler, *httpHandlerMocks) {
	t.Helper()
	h, mocks := newHttpHandlerWithMocks(t)
	return h.files, mocks
}

// newSwaggerHandlerWithMocks 构造 swagger 处理器：swagger 只依赖 logger，无需 mock。
func newSwaggerHandlerWithMocks(t *testing.T) *swaggerHandler {
	t.Helper()
	return newSwaggerHandler(mlog.NewForConfig(nil))
}
