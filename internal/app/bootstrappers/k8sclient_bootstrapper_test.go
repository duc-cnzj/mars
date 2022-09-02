package bootstrappers

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestK8sClientBootstrapper_Bootstrap(t *testing.T) {}

func TestK8sClientBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&K8sClientBootstrapper{}).Tags())
}

func Test_filterEvent(t *testing.T) {
	var tests = []struct {
		nsPrefix string

		reason    string
		kind      string
		namespace string
		name      string
		wants     bool
	}{
		{
			nsPrefix:  "ns",
			reason:    "",
			kind:      "Pod",
			namespace: "ns-xxx",
			name:      "app",
			wants:     true,
		},
		{
			nsPrefix:  "ns",
			reason:    "Unhealthy",
			kind:      "Pod",
			namespace: "ns-xxx",
			name:      "app-unhealthy",
			wants:     false,
		},
		{
			reason:    "",
			nsPrefix:  "ns",
			kind:      "Deployment",
			namespace: "xx-xxx",
			name:      "app",
			wants:     false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.namespace, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, filterEvent(tt.nsPrefix)(&eventsv1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tt.namespace,
				},
				Reason: tt.reason,
				Regarding: v1.ObjectReference{
					Kind:      tt.kind,
					Namespace: tt.namespace,
					Name:      tt.name,
				},
			}))
		})
	}
}

func Test_filterPod(t *testing.T) {
	var tests = []struct {
		nsPrefix string

		namespace string
		name      string
		wants     bool
	}{
		{
			nsPrefix:  "ns",
			namespace: "ns-xxx",
			name:      "app",
			wants:     true,
		},
		{
			nsPrefix:  "ns",
			namespace: "xx-xxx",
			name:      "app",
			wants:     false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.namespace, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, filterPod(tt.nsPrefix)(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tt.namespace,
					Name:      tt.name,
				},
			}))
		})
	}

}
