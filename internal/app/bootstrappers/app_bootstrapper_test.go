package bootstrappers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type runHooksEqual struct {
	hook contracts.Callback
}

func (r *runHooksEqual) Matches(x any) bool {
	r.hook = x.(contracts.Callback)
	return true
}

func (r *runHooksEqual) String() string {
	return ""
}

func TestAppBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	h := &runHooksEqual{}
	app.EXPECT().BeforeServerRunHooks(h).Times(2)

	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin:     config.Plugin{Name: "test_git_server"},
		PicturePlugin:       config.Plugin{Name: "test_picture"},
		DomainManagerPlugin: config.Plugin{Name: "test_domain"},
		WsSenderPlugin:      config.Plugin{Name: "test_wssender"},
	}).AnyTimes()

	gits := mock.NewMockGitServer(m)
	app.EXPECT().GetPluginByName("test_git_server").Return(gits).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gits.EXPECT().Initialize(gomock.All()).AnyTimes()

	pictrure := mock.NewMockPictureInterface(m)
	app.EXPECT().GetPluginByName("test_picture").Return(pictrure).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	pictrure.EXPECT().Initialize(gomock.All()).AnyTimes()

	d := mock.NewMockDomainManager(m)
	app.EXPECT().GetPluginByName("test_domain").Return(d).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	d.EXPECT().Initialize(gomock.All()).AnyTimes()

	ws := mock.NewMockWsSender(m)
	app.EXPECT().GetPluginByName("test_wssender").Return(ws).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	ws.EXPECT().Initialize(gomock.All()).AnyTimes()

	assert.Nil(t, (&AppBootstrapper{}).Bootstrap(app))
	assert.NotNil(t, h.hook)
	d.EXPECT().GetCerts().Return("cert", "key", "crt")
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	assert.Nil(t, db.AutoMigrate(&models.Namespace{}))

	assert.Nil(t, db.Create(&models.Namespace{Name: "ns"}).Error)
	assert.Nil(t, db.Create(&models.Namespace{Name: "ns-2"}).Error)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client: fake.NewSimpleClientset(
			&corev1.Secret{
				TypeMeta: v1.TypeMeta{
					Kind: "Secret",
				},
				ObjectMeta: v1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "cert",
				},
				StringData: map[string]string{
					"tls.key": "key-2",
					"tls.crt": "crt-2",
				},
			},
		),
	}).AnyTimes()
	h.hook(app)
	s, _ := app.K8sClient().Client.CoreV1().Secrets("ns").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s.StringData["tls.key"])
	assert.Equal(t, "crt", s.StringData["tls.crt"])
	s2, _ := app.K8sClient().Client.CoreV1().Secrets("ns-2").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s2.StringData["tls.key"])
	assert.Equal(t, "crt", s2.StringData["tls.crt"])
}

func TestAppBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&AppBootstrapper{}).Tags())
}

func TestProjectPodEventListener(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()

	podFanOutObj := &fanOut[*corev1.Pod]{
		name:      "pod",
		ch:        make(chan contracts.Obj[*corev1.Pod], 100),
		listeners: make(map[string]chan<- contracts.Obj[*corev1.Pod]),
	}

	client := &contracts.K8sClient{
		PodFanOut: podFanOutObj,
	}
	ch := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		podFanOutObj.Distribute(ch)
	}()

	app.EXPECT().K8sClient().Return(client).AnyTimes()
	app.EXPECT().Done().Return(ch).AnyTimes()

	app.EXPECT().Config().Return(&config.Config{
		NsPrefix:       "devtest-",
		WsSenderPlugin: config.Plugin{Name: "test_wssender"},
	}).AnyTimes()

	ws := mock.NewMockWsSender(m)
	app.EXPECT().GetPluginByName("test_wssender").Return(ws).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	ws.EXPECT().Initialize(gomock.All()).AnyTimes()

	nsModel := &models.Namespace{Name: "devtest-ns"}
	db.AutoMigrate(&models.Namespace{})
	db.Create(nsModel)

	pubsub := mock.NewMockPubSub(m)
	ws.EXPECT().New("", "").Return(pubsub).Times(1)

	ProjectPodEventListener(app)

	pubsub.EXPECT().Publish(int64(nsModel.ID), gomock.Any()).Times(1)
	podFanOutObj.ch <- contracts.NewObj(nil, &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "devtest-ns",
			Name:      "p1",
		},
	}, contracts.Add)
	pubsub.EXPECT().Publish(int64(nsModel.ID), gomock.Any()).Times(1)
	podFanOutObj.ch <- contracts.NewObj(nil, &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "devtest-ns",
			Name:      "p2",
		},
	}, contracts.Delete)
	pubsub.EXPECT().Publish(int64(nsModel.ID), gomock.Any()).Times(0)
	podFanOutObj.ch <- contracts.NewObj(nil, &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "devtest-ns",
			Name:      "p3",
		},
	}, contracts.Update)
	time.Sleep(1 * time.Second)
	close(ch)
	wg.Wait()
}
