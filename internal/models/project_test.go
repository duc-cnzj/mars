package models

import (
	"sort"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/utils/date"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fake2 "k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/kubectl/pkg/util/deployment"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func TestProject_GetAllPodMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	fm := &fake.Clientset{}
	ex := []v1beta1.PodMetrics{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test", ResourceVersion: "10", Labels: map[string]string{"a": "a"}},
			Window:     metav1.Duration{Duration: time.Minute},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1-2",
					Usage: v1.ResourceList{
						v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory:  *resource.NewQuantity(5*(1024*1024), resource.DecimalSI),
						v1.ResourceStorage: *resource.NewQuantity(6*(1024*1024), resource.DecimalSI),
					},
				},
			},
		},
	}

	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		MetricsClient: fm,
	}).AnyTimes()
	sqlDB, _, _ := sqlmock.New()
	defer sqlDB.Close()
	dbManager := mock.NewMockDBManager(ctrl)
	gormDB, _ := gorm.Open(mysql.New(mysql.Config{SkipInitializeWithVersion: true, Conn: sqlDB}), &gorm.Config{})
	app.EXPECT().DBManager().AnyTimes().Return(dbManager)
	dbManager.EXPECT().DB().Return(gormDB).AnyTimes()
	m := Project{
		ID:          1,
		Namespace:   Namespace{Name: "test"},
		NamespaceId: 1,
	}
	assert.Nil(t, m.GetAllPodMetrics())
	m.PodSelectors = "a=a"
	fm.AddReactor("list", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		res := &v1beta1.PodMetricsList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "2",
			},
			Items: ex,
		}
		return true, res, nil
	})
	assert.Len(t, m.GetAllPodMetrics(), 1)

}

func TestProject_GetAllPods(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	rs := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				deployment.RevisionAnnotation: "1",
			},
			UID:       "aaaa",
			Namespace: "test",
			Name:      "rs",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "deploy-1",
				},
			},
		},
	}
	rs2 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				deployment.RevisionAnnotation: "5",
			},
			UID:       "bbbb",
			Namespace: "test",
			Name:      "rs2",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "deploy-1",
				},
			},
		},
	}
	rs3 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				deployment.RevisionAnnotation: "2",
			},
			UID:       "cccc",
			Namespace: "test",
			Name:      "rs3",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "deploy-1",
				},
			},
		},
	}
	rs4 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				deployment.RevisionAnnotation: "4",
			},
			UID:       "dddd",
			Namespace: "test",
			Name:      "rs4",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "deploy-1",
				},
			},
		},
	}
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "test",
			Labels: map[string]string{
				"a": "a",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs",
					UID:        "aaaa",
				},
			},
		},
	}
	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod4",
			Namespace: "test",
			Labels: map[string]string{
				"b": "b",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs3",
					UID:        "cccc",
				},
			},
		},
	}
	pod3 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "test",
			Labels: map[string]string{
				"c": "c",
			},
		},
	}
	pod4 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "test",
			Labels: map[string]string{
				"b": "b",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs2",
					UID:        "bbbb",
				},
			},
		},
	}
	pod5 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod5",
			Namespace: "test",
			Labels: map[string]string{
				"b": "b",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs4",
					UID:        "dddd",
				},
			},
		},
	}

	podWithErrorRsName := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-x-error",
			Namespace: "test",
			Labels: map[string]string{
				"b": "b",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs-x",
					UID:        "uid-not-exist",
				},
			},
		},
	}
	fk := fake2.NewSimpleClientset(rs, rs2, rs3, rs4, pod1, pod2, pod3, pod4, pod5, podWithErrorRsName)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:           fk,
		PodLister:        testutil.NewPodLister(pod1, pod2, pod3, pod4, pod5, podWithErrorRsName),
		ReplicaSetLister: testutil.NewRsLister(rs, rs2, rs3, rs4),
	}).AnyTimes()
	m := Project{
		ID:           1,
		Namespace:    Namespace{Name: "test"},
		NamespaceId:  1,
		PodSelectors: "",
	}
	assert.Nil(t, m.GetAllPods())
	m.PodSelectors = "a=a|b=b"
	assert.Len(t, m.GetAllPods(), 5)
	var oldCount int
	for _, po := range m.GetAllPods() {
		if po.IsOld {
			oldCount++
			continue
		}
		assert.True(t, po.Pod.Name == "pod2" || po.Pod.Name == "pod-x-error" || po.Pod.Name == "pod5")
	}
	assert.Equal(t, 3, oldCount)
}

func TestProject_GetEnvValues(t *testing.T) {
	m := Project{
		ID:        1,
		EnvValues: `{"name": "duc", "age": 18}`,
	}
	// json.Marshal 是 float64 的
	assert.Equal(t, map[string]any{"age": float64(18), "name": "duc"}, m.GetEnvValues())
}

func TestProject_GetExtraValues(t *testing.T) {
	m := Project{
		ID:          1,
		ExtraValues: `[{"path": "/bbb", "value": "bbb"}, {"path": "/aaa", "value": "aaa"}]`,
	}
	assert.Len(t, m.GetExtraValues(), 2)
	assert.Equal(t, []*types.ExtraValue{
		{
			Path:  "/bbb",
			Value: "bbb",
		},
		{
			Path:  "/aaa",
			Value: "aaa",
		},
	}, m.GetExtraValues())
}

func TestProject_GetPodSelectors(t *testing.T) {
	m := Project{
		ID:           1,
		PodSelectors: "a=a|b=b|c=c",
	}
	assert.Equal(t, []string{"a=a", "b=b", "c=c"}, m.GetPodSelectors())
}

func TestProject_ProtoTransform(t *testing.T) {
	tt := time.Now().Add(13 * time.Minute)
	m := Project{
		ID:               1,
		Name:             "project",
		GitProjectId:     100,
		GitBranch:        "dev",
		GitCommit:        "commit",
		Config:           "xxx",
		OverrideValues:   "",
		DockerImage:      "duccnzj/mars:v2,duccnzj/mars:v1",
		PodSelectors:     "a=a|b=b",
		NamespaceId:      1,
		Atomic:           true,
		DeployStatus:     1,
		EnvValues:        "",
		ExtraValues:      "",
		FinalExtraValues: "",
		ConfigType:       "",
		GitCommitWebUrl:  "",
		GitCommitTitle:   "",
		GitCommitAuthor:  "",
		GitCommitDate:    &tt,
		CreatedAt:        time.Now().Add(15 * time.Minute),
		UpdatedAt:        time.Now().Add(30 * time.Minute),
		DeletedAt: gorm.DeletedAt{
			Time:  time.Now().Add(-10 * time.Second),
			Valid: true,
		},
		Namespace: Namespace{},
	}
	assert.Equal(t, &types.ProjectModel{
		Id:                int64(m.ID),
		Name:              m.Name,
		GitProjectId:      int64(m.GitProjectId),
		GitBranch:         m.GitBranch,
		GitCommit:         m.GitCommit,
		Config:            m.Config,
		OverrideValues:    m.OverrideValues,
		DockerImage:       m.DockerImage,
		PodSelectors:      m.PodSelectors,
		NamespaceId:       int64(m.NamespaceId),
		Atomic:            m.Atomic,
		EnvValues:         m.EnvValues,
		ExtraValues:       m.GetExtraValues(),
		FinalExtraValues:  m.FinalExtraValues,
		DeployStatus:      types.Deploy(m.DeployStatus),
		HumanizeCreatedAt: date.ToHumanizeDatetimeString(&m.CreatedAt),
		HumanizeUpdatedAt: date.ToHumanizeDatetimeString(&m.UpdatedAt),
		ConfigType:        m.ConfigType,
		GitCommitWebUrl:   m.GitCommitWebUrl,
		GitCommitTitle:    m.GitCommitTitle,
		GitCommitAuthor:   m.GitCommitAuthor,
		GitCommitDate:     date.ToHumanizeDatetimeString(m.GitCommitDate),
		Namespace:         m.Namespace.ProtoTransform(),
		CreatedAt:         date.ToRFC3339DatetimeString(&m.CreatedAt),
		UpdatedAt:         date.ToRFC3339DatetimeString(&m.UpdatedAt),
		DeletedAt:         date.ToRFC3339DatetimeString(&m.DeletedAt.Time),
	}, m.ProtoTransform())
}

func TestSortStatePod(t *testing.T) {
	s := SortStatePod{
		{
			IsOld:       false,
			Terminating: true,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "Terminating-PodRunning"},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
		},
		{
			IsOld:       false,
			Terminating: false,
			Pending:     true,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "PodPending-1", CreationTimestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)}},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
		},
		{
			IsOld:       false,
			Terminating: false,
			Pending:     true,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "PodPending-2", CreationTimestamp: metav1.Time{Time: time.Now()}},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
		},
		{
			IsOld:       false,
			Terminating: false,
			Pending:     false,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "PodFailed-1"},
				Status: v1.PodStatus{
					Phase: v1.PodFailed,
				},
			},
		},
		{
			IsOld:       false,
			Terminating: false,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "PodRunning-a", CreationTimestamp: metav1.Time{Time: time.Now()}},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
		},
		{
			IsOld:       false,
			Terminating: false,
			Pending:     true,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "PodPending-3", CreationTimestamp: metav1.Time{Time: time.Now().Add(1 * time.Hour)}},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
		},
		{
			IsOld:       false,
			Terminating: false,
			Pending:     false,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "PodRunning-3"},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
		},
		{
			IsOld:       true,
			Terminating: false,
			Pending:     true,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "old-PodPending-a", CreationTimestamp: metav1.Time{Time: time.Now().Add(2 * time.Hour)}},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
		},
		{
			IsOld:       false,
			Terminating: true,
			Pending:     true,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "Terminating-PodPending"},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
		},
		{
			IsOld:       true,
			Terminating: false,
			Pending:     true,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "old-PodPending-b", CreationTimestamp: metav1.Time{Time: time.Now()}},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
		},
		{
			IsOld:       false,
			Terminating: false,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "PodRunning-b", CreationTimestamp: metav1.Time{Time: time.Now().Add(2 * time.Hour)}},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
		},
		{
			IsOld:       true,
			Terminating: false,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "old-PodRunning"},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
		},
		{
			IsOld:       true,
			Terminating: true,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "old-Terminating-PodRunning"},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
		},
	}
	sort.Sort(s)
	var names []string
	for _, pod := range s {
		names = append(names, pod.Pod.Name)
	}

	assert.Equal(t, []string{
		"PodRunning-3",
		"PodRunning-a",
		"PodRunning-b",
		"PodFailed-1",
		"PodPending-1",
		"PodPending-2",
		"PodPending-3",
		"Terminating-PodRunning",
		"Terminating-PodPending",
		"old-PodRunning",
		"old-Terminating-PodRunning",
		"old-PodPending-b",
		"old-PodPending-a",
	}, names)
}

func TestProject_SetPodSelectors(t *testing.T) {
	p := &Project{}
	p.SetPodSelectors([]string{"app=a", "tag=1.0"})
	assert.Equal(t, p.PodSelectors, "app=a|tag=1.0")
}
