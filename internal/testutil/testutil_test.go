package testutil

import (
	"testing"

	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/duc-cnzj/mars/v4/internal/plugins"

	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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

func TestNewIngressLister(t *testing.T) {
	lister := NewIngressLister(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "ing1",
		},
	}, &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "ing2",
		},
	})
	_, err := lister.Ingresses("ns").Get("ing1")
	assert.Nil(t, err)
	_, err = lister.Ingresses("ns").Get("ing2")
	assert.Nil(t, err)
	_, err = lister.Ingresses("ns").Get("ing3")
	assert.Error(t, err)
}

func TestMockGitServer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := MockApp(m)
	server := MockGitServer(m, app)
	assert.IsType(t, (*mock.MockGitServer)(nil), server)
	assert.Same(t, server, app.GetPluginByName("gits"))
	assert.Same(t, server, plugins.GetGitServer())
}

func TestValueMatcher(t *testing.T) {
	assert.Implements(t, (*gomock.Matcher)(nil), &ValueMatcher{})
	vm := &ValueMatcher{}
	vm.Matches("aa")
	assert.Equal(t, "aa", vm.Value)
}

func TestMockWsServer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := MockApp(m)
	server := MockWsServer(m, app)
	assert.IsType(t, (*mock.MockWsSender)(nil), server)
	assert.Same(t, server, app.GetPluginByName("wssender"))
	assert.Same(t, server, plugins.GetWsSender())
}

func TestMockDomainManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := MockApp(m)
	dm := MockDomainManager(m, app)
	assert.IsType(t, (*mock.MockDomainManager)(nil), dm)
	assert.Same(t, dm, app.GetPluginByName("domain_manager"))
	assert.Same(t, dm, plugins.GetDomainManager())
}

func TestNewSecretLister(t *testing.T) {
	lister := NewSecretLister(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "sec1",
		},
	}, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "sec2",
		},
	})
	_, err := lister.Secrets("ns").Get("sec1")
	assert.Nil(t, err)
	_, err = lister.Secrets("ns").Get("sec2")
	assert.Nil(t, err)
	_, err = lister.Secrets("ns").Get("sec3")
	assert.Error(t, err)
}

func TestAssertAuditLogFiredWithMsg(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	a := MockApp(m)
	AssertAuditLogFiredWithMsg(m, a, "hello")
	a.EventDispatcher().Dispatch(events.EventAuditLog, events.NewEventAuditLog("duc", 1, "hello"))
}

func Test_auditLogEqMatcher_Matches(t *testing.T) {
	m := auditLogEqMatcher{msg: "aaa"}
	assert.False(t, m.Matches(events.NewEventAuditLog("", 1, "11")))
	assert.True(t, m.Matches(events.NewEventAuditLog("", 1, "aaa")))
}

func TestAssertAuditLogFiredWithLog(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	a := MockApp(m)
	AssertAuditLogFiredWithLog(m, a, events.NewEventAuditLog("duc", 1, "hello"))
	a.EventDispatcher().Dispatch(events.EventAuditLog, events.NewEventAuditLog("duc", 1, "hello"))
}

func TestNewEventLister(t *testing.T) {
	lister := NewEventLister(
		&eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ev1",
				Namespace: "duc",
			},
		},
		&eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ev2",
				Namespace: "duc",
			},
		},
	)

	_, err := lister.Events("duc").Get("ev1")
	assert.Nil(t, err)
	_, err = lister.Events("duc").Get("ev2")
	assert.Nil(t, err)
	ret, _ := lister.Events("duc").List(labels.Everything())
	assert.Len(t, ret, 2)
}
