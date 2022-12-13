package testutil

import (
	"testing"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetGormDB(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	db, f := SetGormDB(m, app)
	defer f()
	assert.NotNil(t, db)
	assert.Equal(t, db, app.DB())
}

func TestMockApp(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	a := MockApp(m)
	assert.Same(t, app.App(), a)
}

func TestAssertAuditLogFired(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	a := MockApp(m)
	AssertAuditLogFired(m, a)
	a.EventDispatcher().Dispatch(events.EventAuditLog, nil)
}

func TestNewPodLister(t *testing.T) {
	lister := NewPodLister(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app1",
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app2",
			},
		},
	)
	_, err := lister.Pods("ns").Get("app1")
	assert.Nil(t, err)
	_, err = lister.Pods("ns").Get("app2")
	assert.Nil(t, err)
	_, err = lister.Pods("ns").Get("app3")
	assert.Error(t, err)
}

func TestNewRsLister(t *testing.T) {
	lister := NewRsLister(&appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "rs1",
		},
	}, &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "rs2",
		},
	})
	_, err := lister.ReplicaSets("ns").Get("rs1")
	assert.Nil(t, err)
	_, err = lister.ReplicaSets("ns").Get("rs2")
	assert.Nil(t, err)
	_, err = lister.ReplicaSets("ns").Get("rs3")
	assert.Error(t, err)
}

func TestNewServiceLister(t *testing.T) {
	lister := NewServiceLister(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "svc1",
		},
	}, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "svc2",
		},
	})
	_, err := lister.Services("ns").Get("svc1")
	assert.Nil(t, err)
	_, err = lister.Services("ns").Get("svc2")
	assert.Nil(t, err)
	_, err = lister.Services("ns").Get("svc3")
	assert.Error(t, err)
}
