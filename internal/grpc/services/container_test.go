package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars-client/v4/container"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func mockApp(m *gomock.Controller) *mock.MockApplicationInterface {
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	return app
}

func TestContainer_ContainerLog(t *testing.T) {
}

func TestContainer_CopyToPod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mockApp(m)
	fk := fake.NewSimpleClientset(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "po1",
			Namespace: "dev",
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: "Running",
		},
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk}).AnyTimes()
	db, f := SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.File{})
	file := &models.File{}
	db.Create(file)
	assertAuditLogFired(m, app)
	res, err := (&Container{
		CopyFileToPodFunc: func(namespace, pod, container, fpath, targetContainerDir string) (*utils.CopyFileToPodResult, error) {
			return &utils.CopyFileToPodResult{
				TargetDir:     "/tmp",
				ContainerPath: "/tmp/aa.txt",
				FileName:      "aa.txt",
			}, nil
		},
	}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    int64(file.ID),
		Namespace: "dev",
		Pod:       "po1",
		Container: "app",
	})
	assert.Nil(t, err)
	assert.Equal(t, "/tmp", res.PodFilePath)
	assert.Equal(t, "aa.txt", res.FileName)

	_, err = (&Container{}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    int64(file.ID),
		Namespace: "xxx",
		Pod:       "xxx",
		Container: "xxx",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
	_, err = (&Container{}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    1111111,
		Namespace: "dev",
		Pod:       "po1",
		Container: "app",
	})
	assert.Error(t, err)

	_, err = (&Container{
		CopyFileToPodFunc: func(namespace, pod, container, fpath, targetContainerDir string) (*utils.CopyFileToPodResult, error) {
			return nil, errors.New("xxx")
		},
	}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    int64(file.ID),
		Namespace: "dev",
		Pod:       "po1",
		Container: "app",
	})
	fromError, _ = status.FromError(err)

	assert.Equal(t, "xxx", fromError.Message())
	assert.Equal(t, codes.Internal, fromError.Code())
}

func TestContainer_Exec(t *testing.T) {
}

func TestContainer_IsPodExists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mockApp(m)

	app.EXPECT().K8sClient().Times(2).Return(&contracts.K8sClient{Client: fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns",
			},
		},
	)})
	exist, _ := new(Container).IsPodExists(context.TODO(), &container.IsPodExistsRequest{
		Namespace: "nsxx",
		Pod:       "podxxx",
	})
	assert.Equal(t, false, exist.Exists)
	exist, _ = new(Container).IsPodExists(context.TODO(), &container.IsPodExistsRequest{
		Namespace: "ns",
		Pod:       "pod",
	})
	assert.Equal(t, true, exist.Exists)
}

func TestContainer_IsPodRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mockApp(m)

	app.EXPECT().K8sClient().Times(2).Return(&contracts.K8sClient{Client: fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns",
			},
			Spec: v1.PodSpec{},
			Status: v1.PodStatus{
				Phase: "Running",
			},
		},
	)})
	running, _ := new(Container).IsPodRunning(context.TODO(), &container.IsPodRunningRequest{
		Namespace: "nsxx",
		Pod:       "podxxx",
	})
	assert.Equal(t, false, running.Running)
	running, _ = new(Container).IsPodRunning(context.TODO(), &container.IsPodRunningRequest{
		Namespace: "ns",
		Pod:       "pod",
	})
	assert.Equal(t, true, running.Running)
}

func TestContainer_StreamContainerLog(t *testing.T) {

}

func TestContainer_StreamCopyToPod(t *testing.T) {

}

func TestFindDefaultContainer(t *testing.T) {
	defaultContainer := FindDefaultContainer(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "xx",
			Namespace:   "ns",
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "app",
				},
				{
					Name: "app2",
				},
			},
		},
	})
	assert.Equal(t, "app", defaultContainer)
	defaultContainer = FindDefaultContainer(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "xx",
			Namespace: "ns",
			Annotations: map[string]string{
				DefaultContainerAnnotationName: "app2",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "app",
				},
				{
					Name: "app2",
				},
			},
		},
	})
	assert.Equal(t, "app2", defaultContainer)
	defaultContainer = FindDefaultContainer(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "xx",
			Namespace: "ns",
			Annotations: map[string]string{
				DefaultContainerAnnotationName: "app3",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "app",
				},
				{
					Name: "app2",
				},
			},
		},
	})
	assert.Equal(t, "app", defaultContainer)
}

func Test_closeable_IsClosed(t *testing.T) {
	c := &closeable{}
	assert.False(t, c.IsClosed())
	c.Close()
	assert.True(t, c.IsClosed())
}

func Test_execWriter_IsClosed(t *testing.T) {
	w := newExecWriter()
	assert.False(t, w.IsClosed())
	w.Close()
	w.Close()
	assert.True(t, w.IsClosed())
	_, ok := <-w.ch
	assert.False(t, ok)
}

func Test_execWriter_Write(t *testing.T) {
	w := newExecWriter()
	_, err2 := w.Write([]byte("aaa"))
	assert.Nil(t, err2)
	data := <-w.ch
	assert.Equal(t, "aaa", data)
	w.Close()
	_, err := w.Write([]byte("bbb"))
	assert.Error(t, err)
}

func Test_newExecWriter(t *testing.T) {
	assert.NotNil(t, newExecWriter())
}
