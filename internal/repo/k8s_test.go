package repo

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	eventsv1lister "k8s.io/client-go/listers/events/v1"
	"k8s.io/client-go/tools/cache"
)

func TestNewK8sRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	fileRepo := NewMockFileRepo(m)
	mockUploader := uploader.NewMockUploader(m)
	mockData.EXPECT().Config().Return(&config.Config{})
	repo := NewK8sRepo(
		mlog.NewLogger(nil),
		timer.NewRealTimer(),
		mockData,
		fileRepo,
		mockUploader,
		NewDefaultArchiver(),
		NewExecutorManager(mockData),
	).(*k8sRepo)
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.logger)
	assert.NotNil(t, repo.timer)
	assert.NotNil(t, repo.data)
	assert.NotNil(t, repo.fileRepo)
	assert.NotNil(t, repo.uploader)
	assert.NotNil(t, repo.archiver)
	assert.NotNil(t, repo.executor)
}

func TestSplitManifests(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	fileRepo := NewMockFileRepo(m)
	mockUploader := uploader.NewMockUploader(m)
	mockData.EXPECT().Config().Return(&config.Config{})
	repo := NewK8sRepo(
		mlog.NewLogger(nil),
		timer.NewRealTimer(),
		mockData,
		fileRepo,
		mockUploader,
		NewDefaultArchiver(),
		NewExecutorManager(mockData),
	).(*k8sRepo)

	t.Run("should split manifest string correctly", func(t *testing.T) {
		manifest := "manifest1\n---\nmanifest2\n---\nmanifest3"
		expected := []string{"manifest1", "manifest2", "manifest3"}

		result := repo.SplitManifests(manifest)

		assert.Equal(t, expected, result)
	})

	t.Run("should return single manifest when no delimiters", func(t *testing.T) {
		manifest := "manifest1"
		expected := []string{"manifest1"}

		result := repo.SplitManifests(manifest)

		assert.Equal(t, expected, result)
	})

	t.Run("should handle empty manifest string", func(t *testing.T) {
		manifest := ""
		expected := []string{}

		result := repo.SplitManifests(manifest)

		assert.Equal(t, expected, result)
	})
}

func Test_k8sRepo_CreateDockerSecrets(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockData.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "a",
				Password: "b",
				Email:    "cc",
				Server:   "d",
			},
		},
	})
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: clientset})
	kr := &k8sRepo{
		logger: mlog.NewLogger(nil),
		data:   mockData,
	}
	secret, err := kr.CreateDockerSecret(context.TODO(), "a")
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(secret.Name, "mars-"))
	assert.Equal(t, corev1.SecretTypeDockerConfigJson, secret.Type)
	d := DockerConfigJSON{}
	json.Unmarshal(secret.Data[corev1.DockerConfigJsonKey], &d)
	assert.Equal(t, "a", d.Auths["d"].Username)
	assert.Equal(t, "b", d.Auths["d"].Password)
	assert.Equal(t, "cc", d.Auths["d"].Email)
}

func Test_k8sRepo_GetNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: clientset}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewLogger(nil),
		data:   mockData,
	}
	namespace, err := kr.GetNamespace(context.TODO(), "a")
	assert.Error(t, err)
	assert.Nil(t, namespace)
	_, err = kr.CreateNamespace(context.TODO(), "a")
	assert.Nil(t, err)
	namespace, err = kr.GetNamespace(context.TODO(), "a")
	assert.Nil(t, err)
	assert.Equal(t, "a", namespace.Name)
}

func NewEventLister(events ...*eventsv1.Event) eventsv1lister.EventLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range events {
		idxer.Add(po)
	}
	return eventsv1lister.NewEventLister(idxer)
}

func Test_k8sRepo_ListEvents(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{EventLister: NewEventLister(&eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ev",
			Namespace: "a",
		},
	})}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewLogger(nil),
		data:   mockData,
	}
	events, err := kr.ListEvents("a")
	assert.Nil(t, err)
	assert.Len(t, events, 1)
}
func TestGetPod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{PodLister: NewPodLister(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "po",
			Namespace: "a",
		},
	})}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewLogger(nil),
		data:   mockData,
	}
	pod, err := kr.GetPod("a", "po")
	assert.Nil(t, err)
	assert.Equal(t, "po", pod.Name)
}

func TestFindDefaultContainer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	kr := &k8sRepo{
		logger: mlog.NewLogger(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{PodLister: NewPodLister(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				defaultContainerAnnotationName: "second-container",
			},
			Name:      "pod",
			Namespace: "a",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "first-container",
				},
				{
					Name: "second-container",
				},
			},
		},
	})}).AnyTimes()

	_, err := kr.FindDefaultContainer(context.Background(), "a", "c")
	assert.Error(t, err)

	container, err := kr.FindDefaultContainer(context.Background(), "a", "pod")
	assert.Nil(t, err)
	assert.Equal(t, "second-container", container)
}
