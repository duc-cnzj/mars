package utils

import (
	"errors"
	"fmt"
	"testing"

	v12 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetSlugName(t *testing.T) {
	assert.Equal(t, Md5(fmt.Sprintf("%d-%s", 1, "aa")), GetSlugName(1, "aa"))
}

func TestNewCloser(t *testing.T) {
	called := 0
	closer := NewCloser(func() error {
		called++
		return nil
	})
	closer.Close()
	assert.Equal(t, 1, called)
}

func TestPackageChart(t *testing.T) {}

func TestReleaseList_Add(t *testing.T) {
	rl := ReleaseList{}
	rl.Add(&release.Release{Name: "rl1", Namespace: "dev", Info: &release.Info{Status: "deployed"}})
	rl.Add(&release.Release{Name: "rl2", Namespace: "dev", Info: &release.Info{Status: "pending-upgrade"}})
	rl.Add(&release.Release{Name: "rl3", Namespace: "dev", Info: &release.Info{Status: "pending-rollback"}})
	rl.Add(&release.Release{Name: "rl4", Namespace: "dev", Info: &release.Info{Status: "pending-install"}})
	rl.Add(&release.Release{Name: "rl5", Namespace: "dev", Info: &release.Info{Status: "uninstalling"}})
	rl.Add(&release.Release{Name: "rl6", Namespace: "dev", Info: &release.Info{Status: "failed"}})
	rl.Add(&release.Release{Name: "rl7", Namespace: "dev", Info: &release.Info{Status: "superseded"}})
	rl.Add(&release.Release{Name: "rl8", Namespace: "dev", Info: &release.Info{Status: "unknown"}})
	assert.Len(t, rl, 8)
	_, ok := rl["dev-rl1"]
	assert.True(t, ok)
	assert.Equal(t, "deployed", rl["dev-rl1"].Status)
	assert.Equal(t, "pending", rl["dev-rl2"].Status)
	assert.Equal(t, "pending", rl["dev-rl3"].Status)
	assert.Equal(t, "pending", rl["dev-rl4"].Status)
	assert.Equal(t, "unknown", rl["dev-rl5"].Status)
	assert.Equal(t, "failed", rl["dev-rl6"].Status)
	assert.Equal(t, "unknown", rl["dev-rl7"].Status)
	assert.Equal(t, "unknown", rl["dev-rl8"].Status)
}

func TestReleaseList_GetStatus(t *testing.T) {
	rl := ReleaseList{}
	rl.Add(&release.Release{Name: "rl1", Namespace: "dev", Info: &release.Info{Status: "deployed"}})
	assert.Equal(t, "deployed", rl.GetStatus("dev", "rl1"))
	assert.Equal(t, "unknown", rl.GetStatus("dev", "xxx"))
}

func TestReleaseStatus(t *testing.T) {}

func TestRollback(t *testing.T) {}

func TestUninstallRelease(t *testing.T) {}

func TestUpgradeOrInstall(t *testing.T) {}

func TestWriteConfigYamlToTmpFile(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	info := mock.NewMockFileInfo(m)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(info, nil).Times(1)
	info.EXPECT().Path().Return("/aa.txt").Times(1)
	file, closer, err := WriteConfigYamlToTmpFile([]byte("xx"))
	assert.Nil(t, err)
	assert.Equal(t, "/aa.txt", file)
	up.EXPECT().Delete("/aa.txt").Times(1).Return(nil)
	assert.Nil(t, closer.Close())

	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil, errors.New("xx")).Times(1)
	_, _, err = WriteConfigYamlToTmpFile([]byte("xx"))
	assert.Equal(t, "xx", err.Error())

	up.EXPECT().Delete("/aa.txt").Times(1).Return(errors.New("xx"))
	assert.Equal(t, "xx", closer.Close().Error())
}

func Test_checkIfInstallable(t *testing.T) {}

func Test_getActionConfigAndSettings(t *testing.T) {}

func Test_runInstall(t *testing.T) {}

func Test_send(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "po",
			Namespace: "ns",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "app",
			},
		},
		Spec:   v1.PodSpec{},
		Status: v1.PodStatus{},
	}
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	idxer.Add(pod)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fake.NewSimpleClientset(pod), PodLister: v12.NewPodLister(idxer)})
	called := 0
	var str string
	var obj any = &eventv1.Event{
		Note: "aaa",
		Regarding: v1.ObjectReference{
			Name:            "po",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
	}
	event := obj.(*eventv1.Event)
	p := event.Regarding
	get, _ := app.K8sClient().PodLister.Pods(p.Namespace).Get(p.Name)

	for _, value := range get.Labels {
		if value == ("app") {
			func(cs []*types.Container, format string, v ...any) {
				called++
				str = format
			}(nil, event.Note)
			break
		}
	}
	assert.Equal(t, 1, called)
	assert.Equal(t, "aaa", str)
}
