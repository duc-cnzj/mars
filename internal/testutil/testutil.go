package testutil

import (
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	appsv1lister "k8s.io/client-go/listers/apps/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	networkingv1lister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
)

func SetGormDB(m *gomock.Controller, app *mock.MockApplicationInterface) (*gorm.DB, func()) {
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.Exec("PRAGMA foreign_keys = ON", nil)
	s, _ := db.DB()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	return db, func() {
		s.Close()
	}
}

func MockApp(m *gomock.Controller) *mock.MockApplicationInterface {
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	return app
}

func AssertAuditLogFired(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockDispatcherInterface {
	e := mock.NewMockDispatcherInterface(m)
	e.EXPECT().Dispatch(contracts.Event("audit_log"), gomock.Any()).Times(1)
	app.EXPECT().EventDispatcher().Return(e).AnyTimes()

	return e
}

func NewPodLister(pods ...*corev1.Pod) corev1lister.PodLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range pods {
		idxer.Add(po)
	}
	return corev1lister.NewPodLister(idxer)
}

func NewRsLister(rs ...*appsv1.ReplicaSet) appsv1lister.ReplicaSetLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range rs {
		idxer.Add(po)
	}
	return appsv1lister.NewReplicaSetLister(idxer)
}

func NewSecretLister(rs ...*corev1.Secret) corev1lister.SecretLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range rs {
		idxer.Add(po)
	}
	return corev1lister.NewSecretLister(idxer)
}

func NewServiceLister(svcs ...*corev1.Service) corev1lister.ServiceLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range svcs {
		idxer.Add(po)
	}
	return corev1lister.NewServiceLister(idxer)
}

func NewIngressLister(svcs ...*networkingv1.Ingress) networkingv1lister.IngressLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range svcs {
		idxer.Add(po)
	}
	return networkingv1lister.NewIngressLister(idxer)
}

func MockGitServer(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockGitServer {
	gits := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{GitServerPlugin: config.Plugin{Name: "gits"}}).AnyTimes()
	app.EXPECT().GetPluginByName("gits").Return(gits).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gits.EXPECT().Initialize(gomock.Any()).AnyTimes()
	return gits
}

func MockWsServer(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockWsSender {
	wssender := mock.NewMockWsSender(m)
	app.EXPECT().Config().Return(&config.Config{WsSenderPlugin: config.Plugin{Name: "wssender"}}).AnyTimes()
	app.EXPECT().GetPluginByName("wssender").Return(wssender).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	wssender.EXPECT().Initialize(gomock.Any()).AnyTimes()
	return wssender
}

type ValueMatcher struct {
	Value any
}

func (v *ValueMatcher) Matches(x any) bool {
	v.Value = x
	return true
}

func (v *ValueMatcher) String() string {
	return ""
}
