package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	v13 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func TestDecodeDockerConfigJSON(t *testing.T) {
	a := DockerConfigJSON{
		Auths: map[string]DockerConfigEntry{
			"duc": {
				Username: "duc",
				Password: "pwd",
				Email:    "em",
				Auth:     "au:xx",
			},
		},
		HttpHeaders: map[string]string{"a": "a"},
	}
	marshal, err := json.Marshal(&a)
	assert.Nil(t, err)
	configJSON, err := DecodeDockerConfigJSON(marshal)
	assert.Nil(t, err)
	assert.Equal(t, a, configJSON)

	_, err = DecodeDockerConfigJSON(nil)
	assert.Error(t, err)
	_, err = DecodeDockerConfigJSON([]byte{})
	assert.Error(t, err)
}

func TestCreateDockerSecrets(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fk := fake.NewSimpleClientset()
	secret, _ := CreateDockerSecrets(fk, "default", config.DockerAuths{
		&config.DockerAuth{
			Username: "name",
			Password: "pwd",
			Email:    "1@q.c",
			Server:   "https://index.docker.io/v1/",
		},
		&config.DockerAuth{
			Username: "mars",
			Password: "pwd-mars",
			Email:    "mars@q.c",
			Server:   "mars.io",
		},
	})

	dockercfgAuthOne := DockerConfigEntry{
		Username: "name",
		Password: "pwd",
		Email:    "1@q.c",
		Auth:     base64.StdEncoding.EncodeToString([]byte("name:pwd")),
	}
	dockercfgAuthTwo := DockerConfigEntry{
		Username: "mars",
		Password: "pwd-mars",
		Email:    "mars@q.c",
		Auth:     base64.StdEncoding.EncodeToString([]byte("mars:pwd-mars")),
	}

	dockerCfgJSON := DockerConfigJSON{
		Auths: map[string]DockerConfigEntry{"https://index.docker.io/v1/": dockercfgAuthOne, "mars.io": dockercfgAuthTwo},
	}

	marshal, _ := json.Marshal(dockerCfgJSON)
	secret.Name = ""
	assert.Equal(t, &corev1.Secret{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
		},
		Data: map[string][]byte{
			".dockerconfigjson": marshal,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}, secret)
}

func NewPodLister(pods ...*corev1.Pod) v13.PodLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	for _, po := range pods {
		idxer.Add(po)
	}
	return v13.NewPodLister(idxer)
}

func TestIsPodRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	pod1 := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	pod2 := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod2",
			Namespace: "duc",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}
	pod3 := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod3",
			Namespace: "duc",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "Reason",
							Message: "Message",
						},
					},
				},
			},
		},
	}
	pod4 := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod4",
			Namespace: "duc",
		},
		Status: corev1.PodStatus{
			Phase:  corev1.PodFailed,
			Reason: "Evicted",
		},
	}
	fk := fake.NewSimpleClientset(
		pod1,
		pod2,
		pod3,
		pod4,
	)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: NewPodLister(pod1, pod2, pod3, pod4),
	})
	_, e := IsPodRunning("duc", "pod_not_exists")
	assert.Equal(t, "pod \"pod_not_exists\" not found", e)

	running, _ := IsPodRunning("duc", "pod1")
	assert.True(t, running)
	running, r := IsPodRunning("duc", "pod2")
	assert.False(t, running)
	assert.Equal(t, "pod not running.", r)
	running, r = IsPodRunning("duc", "pod3")
	assert.False(t, running)
	assert.Equal(t, "Reason Message", r)
	running, r = IsPodRunning("duc", "pod4")
	assert.False(t, running)
	assert.Equal(t, "po pod4 already evicted in namespace duc!", r)
	po, err := fk.CoreV1().Pods("duc").Get(context.TODO(), "pod4", v1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, corev1.PodFailed, po.Status.Phase)
}

func TestGetIngressMappingByNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	ing1 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "ing1",
		},
		Spec: v12.IngressSpec{
			TLS: []v12.IngressTLS{
				{
					Hosts:      []string{"app1.com", "app1.io"},
					SecretName: "sec1",
				},
				{
					Hosts:      []string{"app1.org"},
					SecretName: "sec2",
				},
			},
		},
	}
	ing2 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "ing2",
		},
		Spec: v12.IngressSpec{
			TLS: []v12.IngressTLS{
				{
					Hosts:      []string{"app2.org"},
					SecretName: "sec2",
				},
			},
			Rules: []v12.IngressRule{
				{
					Host: "http.com",
				},
				{
					Host: "app2.org",
				},
			},
		},
	}
	ing3 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "xxx",
		},
		Spec: v12.IngressSpec{
			TLS: []v12.IngressTLS{
				{
					Hosts:      []string{"xxx.org"},
					SecretName: "sec2",
				},
			},
		},
	}
	ing4 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "yyy",
		},
		Spec: v12.IngressSpec{
			Rules: []v12.IngressRule{
				{
					Host: "yyy.com",
				},
				{
					Host: "zzz.com",
				},
			},
		},
	}
	fk := fake.NewSimpleClientset(
		&v12.IngressList{
			Items: []v12.Ingress{
				ing1,
				ing2,
				ing3,
				ing4,
			},
		},
	)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:        fk,
		IngressLister: testutil.NewIngressLister(&ing1, &ing2, &ing3, &ing4),
	})
	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := models.Namespace{
		Name: "duc",
	}
	db.Create(&ns)
	p1 := &models.Project{
		Name:        "app1",
		Manifest:    strings.Join(encodeToYaml(&ing1), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p1)
	p2 := &models.Project{
		Name:        "app2",
		Manifest:    strings.Join(encodeToYaml(&ing2), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p2)
	p3 := &models.Project{
		Name:        "xxx",
		NamespaceId: ns.ID,
	}
	db.Create(p3)
	p4 := &models.Project{
		Name:        "yyy",
		Manifest:    strings.Join(encodeToYaml(&ing4), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p4)
	db.Preload("Projects").First(&ns)
	mapping := GetIngressMappingByProjects(ns.Name, ns.Projects...)

	assert.Len(t, mapping["app1"], 3)
	assert.Len(t, mapping["app2"], 2)
	assert.Equal(t, "https://app2.org", mapping["app2"][0].Url)
	assert.Equal(t, "http://http.com", mapping["app2"][1].Url)
	assert.Len(t, mapping["xxx"], 0)
	assert.Len(t, mapping["yyy"], 2)
	for _, endpoint := range mapping["yyy"] {
		assert.True(t, strings.HasPrefix(endpoint.Url, "http://"))
	}
}

func encodeToYaml(objs ...runtime.Object) []string {
	var results []string
	for _, obj := range objs {
		bf := bytes.Buffer{}
		info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
		info.Serializer.Encode(obj, &bf)
		results = append(results, bf.String())
	}
	return results
}

func TestGetNodePortMappingByProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	svc1 := corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "svc1",
		},
		Spec: corev1.ServiceSpec{
			Type: "NodePort",
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30000,
				},
				{
					Name:     "ui",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30001,
				},
				{
					Name:     "web",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30002,
				},
				{
					Name:     "api",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30003,
				},
				{
					Name:     "grpc",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30004,
				},
				{
					Name:     "xxxx",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30005,
				},
			},
		},
	}
	lister := testutil.NewServiceLister(&svc1)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		ServiceLister: lister,
	})
	app.EXPECT().Config().Return(&config.Config{ExternalIp: "127.0.0.1"})
	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := models.Namespace{
		Name: "duc",
	}
	db.Create(&ns)
	p1 := &models.Project{
		Name:        "svc1",
		Manifest:    strings.Join(encodeToYaml(&svc1), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p1)
	db.Preload("Projects").First(&ns)
	mapping := GetNodePortMappingByProjects(ns.Name, ns.Projects...)
	httpCount := 0
	total := 0
	for _, endpoints := range mapping {
		for _, endpoint := range endpoints {
			total++
			if strings.HasPrefix(endpoint.Url, "http") {
				httpCount++
			}
		}
	}
	assert.Equal(t, 4, httpCount)
	assert.Equal(t, 6, total)
}

func TestGetLoadBalancerMappingByProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	svc1 := corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "svc1",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30000,
				},
				{
					Name:     "https",
					Protocol: "tcp",
					Port:     443,
					NodePort: 30001,
				},
				{
					Name:     "xxxx",
					Protocol: "tcp",
					Port:     8080,
					NodePort: 30005,
				},
				{
					Name:     "httpx",
					Protocol: "tcp",
					Port:     8080,
					NodePort: 30006,
				},
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{
						IP: "111.111.111.111",
					},
				},
			},
		},
	}
	lister := testutil.NewServiceLister(&svc1)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		ServiceLister: lister,
	})
	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := models.Namespace{
		Name: "duc",
	}
	db.Create(&ns)
	p1 := &models.Project{
		Name:        "svc1",
		Manifest:    strings.Join(encodeToYaml(&svc1), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p1)
	db.Preload("Projects").First(&ns)
	mapping := GetLoadBalancerMappingByProjects(ns.Name, ns.Projects...)
	for _, endpoints := range mapping {
		for _, endpoint := range endpoints {
			if endpoint.Name == "http" {
				assert.Equal(t, "http://111.111.111.111", endpoint.Url)
			}
			if endpoint.Name == "https" {
				assert.Equal(t, "https://111.111.111.111", endpoint.Url)
			}
			if endpoint.Name == "xxxx" {
				assert.Equal(t, "111.111.111.111:8080", endpoint.Url)
			}
			if endpoint.Name == "httpx" {
				assert.Equal(t, "http://111.111.111.111:8080", endpoint.Url)
			}
		}
	}
}

func TestFilterK8sTypeFromManifest(t *testing.T) {
	data := []string{`apiVersion: v1
kind: Service
metadata:
 name: devops-misc-consul-server
 namespace: devops-aa
 labels:
   app: consul
   chart: consul-helm
   heritage: Helm
   release: devops-misc
   component: server
 annotations:
   service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
spec:
 publishNotReadyAddresses: true
 ports:
 - name: http
   port: 8500
   targetPort: 8500
`, `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: busybox
  name: busybox-56c8cc5468-fd59w
  namespace: default
spec:
  containers:
  - command:
    - sh
    - -c
    - sleep 3600;
    image: busybox:latest
    name: busybox
    resources:
      limits:
        cpu: 10m
        memory: 10Mi
      requests:
        cpu: 10m
        memory: 10Mi
`}
	res := FilterRuntimeObjectFromManifests[*corev1.Service](data)
	assert.Len(t, res, 1)
	res1 := FilterRuntimeObjectFromManifests[*corev1.Pod](data)
	assert.Len(t, res1, 1)
	res2 := FilterRuntimeObjectFromManifests[*corev1.Namespace](data)
	assert.Len(t, res2, 0)
}

func TestRuntimeObjectList_Has(t *testing.T) {
	l := RuntimeObjectList{
		&corev1.Service{
			TypeMeta: v1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "svc1",
			},
		},
		&corev1.Service{
			TypeMeta: v1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "svc2",
			},
		},
	}
	assert.True(t, l.Has(&corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "svc2",
		},
	}))
	assert.False(t, l.Has(&corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "svc3",
		},
	}))
	assert.True(t, l.Has(&corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "svc1",
		},
	}))
	assert.False(t, l.Has(&corev1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "svc2",
		},
	}))
}

const f = `
---
# Source: mars/templates/rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: user-mars
  namespace: devops-test
  labels:
    helm.sh/chart: mars-1.3.1
    app.kubernetes.io/name: mars
    app.kubernetes.io/instance: mars-charts
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
---
# Source: mars/templates/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mars-charts-secret
  labels:
    helm.sh/chart: mars-1.3.1
    app.kubernetes.io/name: mars
    app.kubernetes.io/instance: mars-charts
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
stringData:
  config.yaml: |
    private_key: "-----BEGIN RSA PRIVATE KEY-----\n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\n-----END RSA PRIVATE KEY-----"
---
# Source: mars/templates/rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: user-mars-devops-test-ClusterRoleBinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: user-mars
    namespace: devops-test
`

func TestSplitManifests(t *testing.T) {
	manifests := SplitManifests(f)
	assert.Len(t, manifests, 3)
	assert.NotEqual(t, 3, len(strings.Split(f, "---")))
}

func Test_isHttpPortName(t *testing.T) {
	var tests = []struct {
		name  string
		wants bool
	}{
		{
			name:  "webx",
			wants: true,
		},
		{
			name:  "http",
			wants: true,
		},
		{
			name:  "ui",
			wants: true,
		},
		{
			name:  "api",
			wants: true,
		},
		{
			name:  "xapix",
			wants: true,
		},
		{
			name:  "xxxx",
			wants: false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, isHttpPortName(tt.name))
		})
	}
}

func TestEndpointMapping_AllEndpoints(t *testing.T) {
	em := EndpointMapping{
		"a": []*Endpoint{
			{
				Name: "a",
			},
			{
				Name: "b",
			},
		},
		"b": []*Endpoint{
			{
				Name: "c",
			},
			{
				Name: "d",
			},
		},
	}
	assert.Len(t, em.AllEndpoints(), 4)
}

func TestEndpointMapping_Get(t *testing.T) {
	em := EndpointMapping{
		"a": []*Endpoint{
			{
				Name: "a",
			},
			{
				Name: "b",
			},
		},
	}
	assert.Nil(t, em.Get("xxxx"))
	assert.Len(t, em.Get("a"), 2)
}

func TestEndpointMapping_Sort(t *testing.T) {
	e := EndpointMapping{
		"a": []*Endpoint{
			{
				Name: "a1",
				Url:  "https://xxx",
			},
			{
				Name: "a2",
				Url:  "http://xxx",
			},
		},
		"b": []*Endpoint{
			{
				Name: "b1",
				Url:  "http://xxx",
			},
			{
				Name: "b2",
				Url:  "https://xxx",
			},
		},
	}
	e.Sort()
	assert.Equal(t, []*Endpoint{
		{
			Name: "a1",
			Url:  "https://xxx",
		},
		{
			Name: "a2",
			Url:  "http://xxx",
		},
	}, e["a"])
	assert.Equal(t, []*Endpoint{
		{
			Name: "b2",
			Url:  "https://xxx",
		},
		{
			Name: "b1",
			Url:  "http://xxx",
		},
	}, e["b"])
}

func Test_getPodSelectorsInDeploymentAndStatefulSetByManifest(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: Deployment
				metadata:
				  annotations:
				    meta.helm.sh/release-name: mars
				  generation: 56
				  labels:
				    app.kubernetes.io/name: mars
				  name: mars
				  namespace: default
				spec:
				  selector:
				    matchLabels:
				      app.kubernetes.io/instance: mars
				      app.kubernetes.io/name: mars
				`),
			out: "app.kubernetes.io/instance=mars,app.kubernetes.io/name=mars",
		},
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: Deployment
				metadata:
				  annotations:
				    meta.helm.sh/release-name: mars
				  generation: 56
				  labels:
				    app.kubernetes.io/name: mars
				  name: mars
				  namespace: default
				spec:
				  selector:
				    matchLabels:
				      app.kubernetes.io/instance: abc
				      app.kubernetes.io/name: abc
				`),
			out: "app.kubernetes.io/instance=abc,app.kubernetes.io/name=abc",
		},
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: StatefulSet
				metadata:
				  labels:
				    app.kubernetes.io/component: primary
				    app.kubernetes.io/instance: mars-db
				  name: mars-db-mysql-primary
				  namespace: default
				spec:
				  selector:
				    matchLabels:
				      app.kubernetes.io/component: primary
				      app.kubernetes.io/instance: mars-db
				      app.kubernetes.io/name: mysql
				`),
			out: "app.kubernetes.io/component=primary,app.kubernetes.io/instance=mars-db,app.kubernetes.io/name=mysql",
		},
		{
			in: dedent.Dedent(`
				W0509 17:36:48.835823   98185 helpers.go:555] --dry-run is deprecated and can be replaced with --dry-run=client.
				apiVersion: v1
				kind: Pod
				metadata:
				  creationTimestamp: null
				  labels:
				    run: nginx
				  name: nginx
				spec:
				  containers:
				  - image: nginx
				    name: nginx
				    resources: {}
				  dnsPolicy: ClusterFirst
				  restartPolicy: Always
				status: {}
				`),
			out: "",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: Job
				metadata:
				  name: pi
				spec:
				  template:
				    spec:
				      containers:
				      - name: pi
				        image: perl:5.34.0
				        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
				      restartPolicy: Never
				  backoffLimit: 4
				`),
			out: "",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: Job
				metadata:
				  name: pi
				spec:
				  template:
				    metadata:
				      labels:
				        app: jobRunner-one
				    spec:
				      containers:
				      - name: pi
				        image: perl:5.34.0
				        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
				      restartPolicy: Never
				  backoffLimit: 4
				`),
			out: "app=jobRunner-one",
		},
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: DaemonSet
				metadata:
				  name: fluentd-elasticsearch
				spec:
				  selector:
				    matchLabels:
				      name: fluentd-elasticsearch
				  template:
				    metadata:
				      labels:
				        name: fluentd-elasticsearch
				    spec:
				      containers:
				      - name: fluentd-elasticsearch
				        image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
				        volumeMounts:
				        - name: varlog
				          mountPath: /var/log
				`),
			out: "name=fluentd-elasticsearch",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: CronJob
				metadata:
				  name: hello
				spec:
				  schedule: "* * * * *"
				  jobTemplate:
				    spec:
				      template:
				        metadata:
				          labels:
				            app: cronjob
				        spec:
				          containers:
				          - name: hello
				            image: busybox:1.28
				            imagePullPolicy: IfNotPresent
				            command:
				            - /bin/sh
				            - -c
				            - date; echo Hello from the Kubernetes cluster
				          restartPolicy: OnFailure
				`),
			out: "app=cronjob",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1beta1
				kind: CronJob
				metadata:
				  name: hello
				spec:
				  schedule: "* * * * *"
				  jobTemplate:
				    spec:
				      template:
				        metadata:
				          labels:
				            app: cronjob-v1beta1
				        spec:
				          containers:
				          - name: hello
				            image: busybox:1.28
				            imagePullPolicy: IfNotPresent
				            command:
				            - /bin/sh
				            - -c
				            - date; echo Hello from the Kubernetes cluster
				          restartPolicy: OnFailure
				`),
			out: "app=cronjob-v1beta1",
		},
	}

	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			labels := GetPodSelectorsByManifest([]string{tt.in})
			if len(labels) > 0 {
				assert.Equal(t, tt.out, labels[0])
			} else {
				assert.Equal(t, tt.out, "")
			}
		})
	}
}

func TestGetSlugName(t *testing.T) {
	assert.Equal(t, Hash(fmt.Sprintf("%d-%s", 1, "aa")), GetSlugName(1, "aa"))
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
