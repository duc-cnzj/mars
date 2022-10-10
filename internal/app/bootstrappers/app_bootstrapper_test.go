package bootstrappers

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/utils"

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

type runHooksEqual struct{}

func (r *runHooksEqual) Matches(x any) bool {
	_, ok := x.(contracts.Callback)
	return ok
}

func (r *runHooksEqual) String() string {
	return ""
}

func TestAppBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	h := &runHooksEqual{}
	app.EXPECT().BeforeServerRunHooks(h).Times(3)

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
	pubsub.EXPECT().Publish(int64(nsModel.ID), gomock.Any()).Times(1).Return(errors.New("xxx"))
	podFanOutObj.ch <- contracts.NewObj(nil, &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "devtest-ns",
			Name:      "p2",
		},
	}, contracts.Delete)
	pubsub.EXPECT().Publish(int64(nsModel.ID), gomock.Any()).Times(1).Return(errors.New("xxx"))
	podFanOutObj.ch <- contracts.NewObj(&corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "devtest-ns",
			Name:      "p3",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}, &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "devtest-ns",
			Name:      "p4",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}, contracts.Update)

	podFanOutObj.ch <- contracts.NewObj[*corev1.Pod](nil, nil, contracts.FanOutType(999))
	time.Sleep(1 * time.Second)
	close(ch)
	wg.Wait()

	podFanOutObj2 := &fanOut[*corev1.Pod]{
		name:      "pod-2",
		ch:        make(chan contracts.Obj[*corev1.Pod], 100),
		listeners: make(map[string]chan<- contracts.Obj[*corev1.Pod]),
	}

	close(podFanOutObj2.ch)
	podFanOutObj2.Distribute(nil)
	assert.True(t, true)
}

func TestUpdateImagePullSecrets(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client: fake.NewSimpleClientset(),
	}).AnyTimes()
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	SyncImagePullSecrets(app)

	// changed
	app.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "name",
				Password: "new",
				Email:    "mars@q.c",
				Server:   "mars",
			},
		},
	}).Times(1)
	assert.Nil(t, db.AutoMigrate(&models.Namespace{}))

	secret, err := utils.CreateDockerSecrets("test", config.DockerAuths{
		{
			Username: "1",
			Password: "1",
			Email:    "1@q.c",
			Server:   "mars",
		},
	})
	assert.Nil(t, err)
	assert.Nil(t, db.Create(&models.Namespace{
		Name:             "test",
		ImagePullSecrets: secret.Name,
	}).Error)
	SyncImagePullSecrets(app)
	get, _ := app.K8sClient().Client.CoreV1().Secrets("test").Get(context.TODO(), secret.Name, v1.GetOptions{})
	dockerConfig, _ := utils.DecodeDockerConfigJSON(get.Data[corev1.DockerConfigJsonKey])
	entry := dockerConfig.Auths["mars"]
	assert.Equal(t, "name", entry.Username)
	assert.Equal(t, "new", entry.Password)
	assert.Equal(t, "mars@q.c", entry.Email)

	// add
	app.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "1",
				Password: "1",
				Email:    "1@q.c",
				Server:   "mars",
			},
			{
				Username: "name2",
				Password: "new2",
				Email:    "mars2@q.c",
				Server:   "mars2",
			},
		},
	}).Times(1)
	SyncImagePullSecrets(app)
	var newNs models.Namespace
	assert.Nil(t, app.DB().Model(&models.Namespace{}).First(&newNs).Error)
	assert.Len(t, newNs.ImagePullSecretsArray(), 2)

	list, err := app.K8sClient().Client.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 2)

	// deleted
	app.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "1",
				Password: "1",
				Email:    "1@q.c",
				Server:   "mars",
			},
		},
	}).Times(1)
	SyncImagePullSecrets(app)

	var newNs2 models.Namespace
	assert.Nil(t, app.DB().Model(&models.Namespace{}).First(&newNs2).Error)
	assert.Len(t, newNs2.ImagePullSecretsArray(), 1)

	list, err = app.K8sClient().Client.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 1)
	assert.Equal(t, newNs2.ImagePullSecretsArray()[0], list.Items[0].Name)

	// delete k8s secret
	app.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "1",
				Password: "1",
				Email:    "1@q.c",
				Server:   "mars",
			},
		},
	}).Times(1)
	assert.Nil(t, app.K8sClient().Client.CoreV1().Secrets("test").Delete(context.TODO(), list.Items[0].Name, v1.DeleteOptions{}))

	list, err = app.K8sClient().Client.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 0)
	SyncImagePullSecrets(app)

	var newNs3 models.Namespace
	assert.Nil(t, app.DB().Model(&models.Namespace{}).First(&newNs3).Error)
	assert.Len(t, newNs3.ImagePullSecretsArray(), 1)

	// delete db secret
	// 不会自动删除之前的 secret
	app.DB().Model(&newNs3).Updates(map[string]any{
		"image_pull_secrets": "",
	})
	app.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "1",
				Password: "1",
				Email:    "1@q.c",
				Server:   "mars",
			},
		},
	}).Times(1)

	list, err = app.K8sClient().Client.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 1)
	SyncImagePullSecrets(app)

	list, err = app.K8sClient().Client.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 2)

	var newNs4 models.Namespace
	assert.Nil(t, app.DB().Model(&models.Namespace{}).First(&newNs4).Error)
	assert.Len(t, newNs4.ImagePullSecretsArray(), 1)
}

func TestUpdateCertTls(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	app.EXPECT().Config().Return(&config.Config{DomainManagerPlugin: config.Plugin{Name: "test_domain"}}).AnyTimes()

	d := mock.NewMockDomainManager(m)
	app.EXPECT().GetPluginByName("test_domain").Return(d).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	d.EXPECT().Initialize(gomock.All()).AnyTimes()

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
	UpdateCertTls(app)
	s, _ := app.K8sClient().Client.CoreV1().Secrets("ns").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s.StringData["tls.key"])
	assert.Equal(t, "crt", s.StringData["tls.crt"])
	s2, _ := app.K8sClient().Client.CoreV1().Secrets("ns-2").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s2.StringData["tls.key"])
	assert.Equal(t, "crt", s2.StringData["tls.crt"])
}
