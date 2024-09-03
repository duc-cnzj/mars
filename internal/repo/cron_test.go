package repo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/util/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corev1lister "k8s.io/client-go/listers/core/v1"
	cache2 "k8s.io/client-go/tools/cache"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/cache"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/cron"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"

	"github.com/duc-cnzj/mars/v5/internal/mlog"

	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/stretchr/testify/assert"
)

func Test_cronRepo_allNamespaces(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	cr := &cronRepo{
		logger: mlog.NewForConfig(nil),
		data:   data.NewDataImpl(&data.NewDataParams{DB: db}),
	}
	for i := 0; i < 10; i++ {
		db.Namespace.Create().SetCreatorEmail("a").SetName(fmt.Sprintf("test%d", i)).SaveX(context.TODO())
	}
	namespaces, err := cr.allNamespaces()
	assert.Nil(t, err)
	assert.Len(t, namespaces, 10)
	for i, namespace := range namespaces {
		assert.Equal(t, fmt.Sprintf("test%d", i), namespace.Name)
	}
}

func TestNewCronRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fRepo := NewMockFileRepo(m)
	mockRepoRepo := NewMockRepoRepo(m)
	nsRepo := NewMockNamespaceRepo(m)
	mockCache := cache.NewMockCache(m)
	mockK8sRepo := NewMockK8sRepo(m)
	mockEventRepo := NewMockEventRepo(m)
	pl := application.NewMockPluginManger(m)
	mockUploader := uploader.NewMockUploader(m)
	mockData := data.NewMockData(m)
	mockHelm := NewMockHelmerRepo(m)
	mockGit := NewMockGitRepo(m)
	mockcron := cron.NewMockManager(m)

	mockData.EXPECT().Config().Return(&config.Config{GitServerCached: true, KubeConfig: "xx"})

	command := cron.NewMockCommand(m)
	mockcron.EXPECT().NewCommand(gomock.Any(), gomock.Any()).Times(8).Return(command)
	command.EXPECT().DailyAt(gomock.Any())
	command.EXPECT().EveryTenMinutes()
	command.EXPECT().EveryTwoMinutes()
	command.EXPECT().EveryMinute()
	command.EXPECT().EveryTwoMinutes()
	command.EXPECT().EveryFiveMinutes()
	command.EXPECT().EveryFiveMinutes()
	command.EXPECT().EveryFiveSeconds()

	cronRepo := NewCronRepo(
		mlog.NewForConfig(nil),
		fRepo,
		mockCache,
		mockRepoRepo,
		nsRepo,
		mockK8sRepo,
		pl,
		mockEventRepo,
		mockData,
		mockUploader,
		mockHelm,
		mockGit,
		mockcron,
	).(*cronRepo)

	assert.NotNil(t, cronRepo)
	assert.NotNil(t, cronRepo.logger)
	assert.NotNil(t, cronRepo.fileRepo)
	assert.NotNil(t, cronRepo.cache)
	assert.NotNil(t, cronRepo.repoRepo)
	assert.NotNil(t, cronRepo.nsRepo)
	assert.NotNil(t, cronRepo.k8sRepo)
	assert.NotNil(t, cronRepo.pluginMgr)
	assert.NotNil(t, cronRepo.event)
	assert.NotNil(t, cronRepo.data)
	assert.NotNil(t, cronRepo.up)
	assert.NotNil(t, cronRepo.helm)
	assert.NotNil(t, cronRepo.gitRepo)
	assert.NotNil(t, cronRepo.cronManager)
}

func Test_cronRepo_CacheAllBranches(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockrepo := NewMockRepoRepo(m)
	gRepo := NewMockGitRepo(m)
	cr := &cronRepo{
		logger:   mlog.NewForConfig(nil),
		repoRepo: mockrepo,
		gitRepo:  gRepo,
	}
	gRepo.EXPECT().AllBranches(context.TODO(), 1, true)
	gRepo.EXPECT().AllBranches(context.TODO(), 2, true)
	gRepo.EXPECT().AllBranches(context.TODO(), 3, true)
	mockrepo.EXPECT().All(gomock.Any(), &AllRepoRequest{Enabled: lo.ToPtr(true), NeedGitRepo: lo.ToPtr(true)}).
		Return([]*Repo{
			{GitProjectID: 1},
			{GitProjectID: 1},
			{GitProjectID: 2},
			{GitProjectID: 3},
		}, nil)
	cr.CacheAllBranches()
}

func Test_cronRepo_FixDeployStatus(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	helmrepo := NewMockHelmerRepo(m)
	cr := &cronRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
		helm:   helmrepo,
	}
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db)
	namespace := createNamespace(db)
	p1 := createProject(db, namespace.ID)
	p2 := createProject(db, namespace.ID)
	p3 := createProject(db, namespace.ID)
	p1.Update().SetDeployStatus(types.Deploy_StatusFailed).Save(context.TODO())
	p2.Update().SetDeployStatus(types.Deploy_StatusUnknown).Save(context.TODO())
	p3.Update().SetDeployStatus(types.Deploy_StatusDeploying).Save(context.TODO())
	helmrepo.EXPECT().ReleaseStatus(p1.Name, namespace.Name).Return(types.Deploy_StatusDeployed)
	helmrepo.EXPECT().ReleaseStatus(p2.Name, namespace.Name).Return(types.Deploy_StatusFailed)

	cr.FixDeployStatus()
	get, _ := db.Project.Get(context.TODO(), p1.ID)
	assert.Equal(t, types.Deploy_StatusDeployed, get.DeployStatus)
	get2, _ := db.Project.Get(context.TODO(), p2.ID)
	assert.Equal(t, types.Deploy_StatusUnknown, get2.DeployStatus)
}

func Test_cronRepo_CacheAllProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gRepo := NewMockGitRepo(m)
	cr := &cronRepo{
		logger:  mlog.NewForConfig(nil),
		gitRepo: gRepo,
	}
	gRepo.EXPECT().AllProjects(context.TODO(), true)
	cr.CacheAllProjects()
}

func TestContainerStatusChanged(t *testing.T) {
	logger := mlog.NewForConfig(nil)

	t.Run("returns true when container statuses length differ", func(t *testing.T) {
		oldPod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{{Name: "container1", Ready: true}},
			},
		}
		currentPod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{Name: "container1", Ready: true},
					{Name: "container2", Ready: true},
				},
			},
		}
		assert.True(t, containerStatusChanged(logger, oldPod, currentPod))
	})

	t.Run("returns true when container readiness status changes", func(t *testing.T) {
		oldPod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{{Name: "container1", Ready: true}},
			},
		}
		currentPod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{{Name: "container1", Ready: false}},
			},
		}
		assert.True(t, containerStatusChanged(logger, oldPod, currentPod))
	})

	t.Run("returns false when container statuses are identical", func(t *testing.T) {
		oldPod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{{Name: "container1", Ready: true}},
			},
		}
		currentPod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{{Name: "container1", Ready: true}},
			},
		}
		assert.False(t, containerStatusChanged(logger, oldPod, currentPod))
	})
}

func TestListFiles_PrettyYaml(t *testing.T) {
	files := listFiles{
		&File{Path: "path1", HumanizeSize: "1MB"},
		&File{Path: "path2", HumanizeSize: "2MB"},
	}

	expected := `- name: path1
  size: 1MB
- name: path2
  size: 2MB
`

	assert.Equal(t, expected, files.PrettyYaml())
}

func TestListFiles_PrettyYaml_Empty(t *testing.T) {
	files := listFiles{}

	expected := `[]
`

	assert.Equal(t, expected, files.PrettyYaml())
}

func Test_listFiles_PrettyYaml(t *testing.T) {
	lf := listFiles{
		{
			Path:         "/tmp/2.txt",
			HumanizeSize: "10 MB",
		},
		{
			Path:         "/tmp/1.txt",
			HumanizeSize: "1 B",
		},
	}
	assert.Equal(t, `- name: /tmp/2.txt
  size: 10 MB
- name: /tmp/1.txt
  size: 1 B
`, lf.PrettyYaml())
}

func TestCleanUploadFiles(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	up := uploader.NewMockUploader(m)
	mockData := data.NewMockData(m)
	mockData.EXPECT().DB().Return(db).AnyTimes()
	mockEvent := NewMockEventRepo(m)
	localUp := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(localUp).AnyTimes()
	cr := &cronRepo{
		up:     up,
		logger: mlog.NewForConfig(nil),
		data:   mockData,
		event:  mockEvent,
	}

	mockEvent.EXPECT().AuditLogWithChange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	var files = []*File{
		{
			UploadType: schematype.Local,
			Path:       "/tmp/path1",
			CreatedAt:  time.Now().Add(-24 * time.Hour),
		},
		{
			UploadType: schematype.S3,
			Path:       "/tmp/path2",
			CreatedAt:  time.Now().Add(-24 * time.Hour),
		},
		{
			UploadType: schematype.S3,
			Path:       "/tmp/path3",
			CreatedAt:  time.Now(),
		},
		{
			UploadType: schematype.Local,
			Path:       "/tmp/path4",
			CreatedAt:  time.Now().Add(-48 * time.Hour),
		},
	}

	for _, file := range files {
		db.File.Create().SetPath(file.Path).SetUploadType(file.UploadType).SetCreatedAt(file.CreatedAt).SaveX(context.TODO())
	}

	up.EXPECT().LocalUploader().Return(localUp).AnyTimes()

	up.EXPECT().Type().Return(schematype.S3).AnyTimes()
	localUp.EXPECT().Type().Return(schematype.Local).AnyTimes()
	localUp.EXPECT().Exists("/tmp/path1").Return(true)
	up.EXPECT().Exists("/tmp/path2").Return(false)
	localUp.EXPECT().RemoveEmptyDir()

	db.File.Create().SetUploadType(schematype.Local).SetPath("/tmp/local/path4").
		SetCreatedAt(time.Now().Add(-24 * time.Hour)).Save(context.TODO())
	up.EXPECT().AllDirectoryFiles(gomock.Any()).Return([]uploader.FileInfo{
		uploader.NewFileInfo("/tmp/up/path1", 100, time.Now().Add(-24*time.Hour)),
		uploader.NewFileInfo("/tmp/up/path2", 100, time.Now()),
		uploader.NewFileInfo("/tmp/up/path3", 100, time.Now().Add(-48*time.Hour)),
		uploader.NewFileInfo("/tmp/up/path4", 100, time.Now().Add(-24*time.Hour)),
	}, nil)
	up.EXPECT().Delete("/tmp/up/path4").Times(1)
	up.EXPECT().Delete("/tmp/up/path1").Times(1).Return(errors.New("xxx"))
	up.EXPECT().Delete("/tmp/up/path2").Times(0)
	up.EXPECT().Delete("/tmp/up/path3").Times(0)

	localUp.EXPECT().Exists("/tmp/local/path4").Return(true).Times(1)
	localUp.EXPECT().AllDirectoryFiles(gomock.Any()).Return([]uploader.FileInfo{
		uploader.NewFileInfo("/tmp/local/path1", 100, time.Now().Add(-24*time.Hour)),
		uploader.NewFileInfo("/tmp/local/path2", 100, time.Now()),
		uploader.NewFileInfo("/tmp/local/path3", 100, time.Now().Add(-48*time.Hour)),
		uploader.NewFileInfo("/tmp/local/path4", 100, time.Now().Add(-24*time.Hour)),
	}, nil)
	localUp.EXPECT().Delete("/tmp/local/path1").Times(1)
	cr.CleanUploadFiles()
}

func TestSyncImagePullSecretsWithBadSecret(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	fk := fake.NewSimpleClientset()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData := data.NewMockData(m)
	mockData.EXPECT().DB().Return(db).AnyTimes()
	cr := &cronRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}

	mockData.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "name",
				Password: "new",
				Email:    "mars@q.c",
				Server:   "mars",
			},
		},
	}).Times(1)

	secret, err := k8s.CreateDockerSecrets(fk, "test", config.DockerAuths{
		{
			Username: "1",
			Password: "1",
			Email:    "1@q.c",
			Server:   "mars",
		},
	})
	assert.Nil(t, err)

	mockData.EXPECT().K8sClient().Return(&data.K8sClient{
		Client:       fk,
		SecretLister: getLister(fk),
	}).Times(1)

	db.Namespace.Create().SetCreatorEmail("a").SetName("test").SetImagePullSecrets([]string{secret.Name}).SaveX(context.Background())
	cr.SyncImagePullSecrets()
	get, _ := fk.CoreV1().Secrets("test").Get(context.TODO(), secret.Name, v1.GetOptions{})
	dockerConfig, _ := k8s.DecodeDockerConfigJSON(get.Data[corev1.DockerConfigJsonKey])
	entry := dockerConfig.Auths["mars"]
	assert.Equal(t, "name", entry.Username)
	assert.Equal(t, "new", entry.Password)
	assert.Equal(t, "mars@q.c", entry.Email)

	// add
	mockData.EXPECT().Config().Return(&config.Config{
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
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{
		Client:       fk,
		SecretLister: getLister(fk),
	}).Times(1)
	cr.SyncImagePullSecrets()
	newNs, _ := db.Namespace.Query().First(context.Background())
	assert.Len(t, newNs.ImagePullSecrets, 2)
	list, err := fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 2)

	// deleted
	mockData.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "1",
				Password: "1",
				Email:    "1@q.c",
				Server:   "mars",
			},
		},
	}).Times(1)

	mockData.EXPECT().K8sClient().Return(&data.K8sClient{
		Client:       fk,
		SecretLister: getLister(fk),
	}).Times(1)
	cr.SyncImagePullSecrets()
	newNs2, _ := db.Namespace.Query().First(context.Background())
	assert.Len(t, newNs2.ImagePullSecrets, 1)
	list, err = fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 1)
	assert.Equal(t, newNs2.ImagePullSecrets[0], list.Items[0].Name)

	//delete k8s secret
	mockData.EXPECT().Config().Return(&config.Config{
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
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{
		Client:       fk,
		SecretLister: getLister(fk),
	}).Times(1)
	cr.SyncImagePullSecrets()
	newNs3, _ := db.Namespace.Query().First(context.Background())
	assert.Len(t, newNs3.ImagePullSecrets, 1)
}

func getLister(fk kubernetes.Interface) corev1lister.SecretLister {
	var ss []*corev1.Secret
	list, _ := fk.CoreV1().Secrets("test").List(context.TODO(), v1.ListOptions{})
	for idx := range list.Items {
		ss = append(ss, &list.Items[idx])
	}
	return NewSecretLister(ss...)
}

func NewSecretLister(rs ...*corev1.Secret) corev1lister.SecretLister {
	idxer := cache2.NewIndexer(cache2.MetaNamespaceKeyFunc, cache2.Indexers{cache2.NamespaceIndex: cache2.MetaNamespaceIndexFunc})
	for _, po := range rs {
		idxer.Add(po)
	}
	return corev1lister.NewSecretLister(idxer)
}

func TestUpdateCertTls(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()
	db.Namespace.Create().SetCreatorEmail("a").SetName("ns").SaveX(context.Background())
	db.Namespace.Create().SetCreatorEmail("a").SetName("ns-2").SaveX(context.Background())
	db.Namespace.Create().SetCreatorEmail("a").SetName("ns-3").SaveX(context.Background())

	sec := &corev1.Secret{
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
	}
	mockData := data.NewMockData(m)
	fk := fake.NewSimpleClientset(sec)
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{
		SecretLister: NewSecretLister(sec),
		Client:       fk,
	}).AnyTimes()
	mockData.EXPECT().DB().Return(db).AnyTimes()
	pl := application.NewMockPluginManger(m)
	domainManager := application.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domainManager)
	domainManager.EXPECT().GetCerts().Return("cert", "key", "crt")
	(&cronRepo{
		logger:    mlog.NewForConfig(nil),
		data:      mockData,
		pluginMgr: pl,
		nsRepo: &namespaceRepo{
			logger: mlog.NewForConfig(nil),
			data:   data.NewDataImpl(&data.NewDataParams{DB: db}),
		},
		k8sRepo: &k8sRepo{
			data: mockData,
		},
	}).SyncDomainSecret()
	s, _ := fk.CoreV1().Secrets("ns").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s.StringData["tls.key"])
	assert.Equal(t, "crt", s.StringData["tls.crt"])
	s2, _ := fk.CoreV1().Secrets("ns-2").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s2.StringData["tls.key"])
	assert.Equal(t, "crt", s2.StringData["tls.crt"])
	s3, _ := fk.CoreV1().Secrets("ns-3").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s3.StringData["tls.key"])
	assert.Equal(t, "crt", s3.StringData["tls.crt"])
}

func Test_cronRepo_DiskInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo := NewMockFileRepo(m)
	repo.EXPECT().DiskInfo(true)
	info, err := (&cronRepo{fileRepo: repo}).DiskInfo()
	assert.Nil(t, err)
	assert.NotNil(t, info)
}

func TestProjectPodEventListener(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	pch1 := make(chan data.Obj[*corev1.Pod], 100)
	podFanOutObj := data.NewFanOut[*corev1.Pod](
		mlog.NewForConfig(nil),
		"pod",
		pch1,
		make(map[string]chan<- data.Obj[*corev1.Pod]),
	)

	client := &data.K8sClient{
		PodFanOut: podFanOutObj,
	}
	ch := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	mockData := data.NewMockData(m)
	mockData.EXPECT().K8sClient().Return(client).AnyTimes()

	mockData.EXPECT().Config().Return(&config.Config{
		NsPrefix:       "devtest-",
		WsSenderPlugin: config.Plugin{Name: "test_wssender"},
	}).AnyTimes()

	ws := application.NewMockWsSender(m)

	pl := application.NewMockPluginManger(m)
	pl.EXPECT().Ws().Return(ws)

	db, _ := data.NewSqliteDB()
	defer db.Close()

	nsModel := db.Namespace.Create().SetCreatorEmail("a").SetName("devtest-ns").SaveX(context.TODO())

	pubsub := application.NewMockPubSub(m)
	ws.EXPECT().New("", "").Return(pubsub).Times(1)

	go (&cronRepo{
		pluginMgr: pl,
		logger:    mlog.NewForConfig(nil),
		data:      mockData,
		nsRepo: &namespaceRepo{
			logger: mlog.NewForConfig(nil),
			data:   data.NewDataImpl(&data.NewDataParams{DB: db}),
		},
		k8sRepo: &k8sRepo{
			data: mockData,
		},
	}).ProjectPodEventListener()
	time.Sleep(1 * time.Second)
	go func() {
		defer wg.Done()
		podFanOutObj.Distribute(ch)
	}()
	pubsub.EXPECT().Publish(int64(nsModel.ID), gomock.Any()).Times(1)
	pch1 <- data.NewObj(nil, &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "devtest-ns",
			Name:      "p1",
		},
	}, data.Add)
	pubsub.EXPECT().Publish(int64(nsModel.ID), gomock.Any()).Times(1).Return(errors.New("xxx"))
	pch1 <- data.NewObj(nil, &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "devtest-ns",
			Name:      "p2",
		},
	}, data.Delete)
	pubsub.EXPECT().Publish(int64(nsModel.ID), gomock.Any()).Times(1).Return(errors.New("xxx"))
	pch1 <- data.NewObj(&corev1.Pod{
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
	}, data.Update)

	pch1 <- data.NewObj[*corev1.Pod](nil, nil, data.FanOutType(999))
	time.Sleep(1 * time.Second)
	close(ch)
	wg.Wait()

	pch := make(chan data.Obj[*corev1.Pod], 100)
	podFanOutObj2 := data.NewFanOut[*corev1.Pod](
		mlog.NewForConfig(nil),
		"pod-2",
		pch,
		make(map[string]chan<- data.Obj[*corev1.Pod]),
	)

	close(pch)
	podFanOutObj2.Distribute(nil)
	assert.True(t, true)
}
