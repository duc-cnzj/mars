package bootstrappers

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/duc-cnzj/mars/v4/internal/utils"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corev1lister "k8s.io/client-go/listers/core/v1"
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

	projectPodEventListener(app)

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
func getLister(fk kubernetes.Interface) corev1lister.SecretLister {
	var ss []*corev1.Secret
	list, _ := fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	for idx := range list.Items {
		ss = append(ss, &list.Items[idx])
	}
	return testutil.NewSecretLister(ss...)
}

func TestSyncImagePullSecretsWithBadSecret(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	fk := fake.NewSimpleClientset()
	db, fn := testutil.SetGormDB(m, app)
	defer fn()

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
	}).AnyTimes()
	assert.Nil(t, db.AutoMigrate(&models.Namespace{}))

	secret, err := utils.CreateDockerSecrets(fk, "test", config.DockerAuths{
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

	secret.Data[corev1.DockerConfigJsonKey] = nil
	fk.CoreV1().Secrets("test").Update(context.TODO(), secret, v1.UpdateOptions{})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, SecretLister: getLister(fk)}).Times(1)
	syncImagePullSecrets(app)
	get, _ := fk.CoreV1().Secrets("test").Get(context.TODO(), secret.Name, v1.GetOptions{})
	assert.Nil(t, get.Data[corev1.DockerConfigJsonKey])

	secret.Data = map[string][]byte{}
	fk.CoreV1().Secrets("test").Update(context.TODO(), secret, v1.UpdateOptions{})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, SecretLister: getLister(fk)}).Times(1)
	syncImagePullSecrets(app)
	get, _ = fk.CoreV1().Secrets("test").Get(context.TODO(), secret.Name, v1.GetOptions{})
	assert.Len(t, get.Data, 0)
}

func TestSyncImagePullSecrets(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	fk := fake.NewSimpleClientset()
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	app.EXPECT().K8sClient().Times(1)
	syncImagePullSecrets(app)

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

	secret, err := utils.CreateDockerSecrets(fk, "test", config.DockerAuths{
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
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, SecretLister: getLister(fk)}).Times(1)
	syncImagePullSecrets(app)
	get, _ := fk.CoreV1().Secrets("test").Get(context.TODO(), secret.Name, v1.GetOptions{})
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
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, SecretLister: getLister(fk)}).Times(1)
	syncImagePullSecrets(app)
	var newNs models.Namespace
	assert.Nil(t, app.DB().Model(&models.Namespace{}).First(&newNs).Error)
	assert.Len(t, newNs.ImagePullSecretsArray(), 2)

	list, err := fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
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
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, SecretLister: getLister(fk)}).Times(1)
	syncImagePullSecrets(app)

	var newNs2 models.Namespace
	assert.Nil(t, app.DB().Model(&models.Namespace{}).First(&newNs2).Error)
	assert.Len(t, newNs2.ImagePullSecretsArray(), 1)

	list, err = fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
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
	assert.Nil(t, fk.CoreV1().Secrets("test").Delete(context.TODO(), list.Items[0].Name, v1.DeleteOptions{}))

	list, err = fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 0)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, SecretLister: getLister(fk)}).Times(1)
	syncImagePullSecrets(app)

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

	list, err = fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 1)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, SecretLister: getLister(fk)}).Times(1)
	syncImagePullSecrets(app)

	list, err = fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 2)

	var newNs4 models.Namespace
	assert.Nil(t, app.DB().Model(&models.Namespace{}).First(&newNs4).Error)
	assert.Len(t, newNs4.ImagePullSecretsArray(), 1)
}

func Test_updateCerts(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Namespace{})
	manager := testutil.MockDomainManager(m, app)
	manager.EXPECT().GetCerts().Return("name", "key", "crt").Times(1)
	updateCerts(app)
}

func Test_containerStatusChanged(t *testing.T) {
	tests := []struct {
		old     *corev1.Pod
		current *corev1.Pod
		want    bool
	}{
		{
			old: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							Ready: false,
						},
					},
				},
			},
			current: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							Ready: true,
						},
					},
				},
			},
			want: true,
		},
		{
			old: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			current: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							Ready: true,
						},
					},
				},
			},
			want: true,
		},
		{
			old: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							Ready: true,
						},
					},
				},
			},
			current: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							Ready: true,
						},
					},
				},
			},
			want: false,
		},
		{
			old: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			current: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, containerStatusChanged(tt.old, tt.current))
		})
	}
}
