package data

import (
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_dataImpl(t *testing.T) {
	d := &dataImpl{minioCli: &minio.Client{}, oidc: OidcConfig{}}
	assert.NotNil(t, d.MinioCli())
	assert.NotNil(t, d.OidcConfig())
}

func Test_filterEvent(t *testing.T) {
	b := filterEvent("aaa")(&eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "aaa-bbb-ccc",
		},
		Reason: "x",
		Regarding: corev1.ObjectReference{
			Kind: "Pod",
		},
	})
	assert.True(t, b)
	b = filterEvent("aaa")(eventsv1.Event{
		Reason: "x",
		Regarding: corev1.ObjectReference{
			Kind: "Pod",
		},
	})
	assert.False(t, b)
}

func Test_filterPod(t *testing.T) {
	b := filterPod("aaa")(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "aaa-bbb-ccc",
		},
	})
	assert.True(t, b)
	b = filterPod("aaa")(corev1.Pod{})
	assert.False(t, b)
}
