package biz

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestIsHttpPortName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Should return true when input contains 'web'",
			input:    "web",
			expected: true,
		},
		{
			name:     "Should return true when input contains 'ui'",
			input:    "ui",
			expected: true,
		},
		{
			name:     "Should return true when input contains 'api'",
			input:    "api",
			expected: true,
		},
		{
			name:     "Should return true when input contains 'http'",
			input:    "http",
			expected: true,
		},
		{
			name:     "Should return false when input does not contain 'web', 'ui', 'api', or 'http'",
			input:    "test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHttpPortName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortEndpoint_Len(t *testing.T) {
	endpoints := sortEndpoint{
		{Name: "Endpoint1"},
		{Name: "Endpoint2"},
		{Name: "Endpoint3"},
	}
	assert.Equal(t, 3, endpoints.Len())
}

func TestSortEndpoint_Swap(t *testing.T) {
	endpoints := sortEndpoint{
		{Name: "Endpoint1"},
		{Name: "Endpoint2"},
	}
	endpoints.Swap(0, 1)
	assert.Equal(t, "Endpoint2", endpoints[0].Name)
	assert.Equal(t, "Endpoint1", endpoints[1].Name)
}

func TestSortEndpoint_Less(t *testing.T) {
	endpoints := sortEndpoint{
		{Name: "Endpoint1", Url: "http://example.com"},
		{Name: "Endpoint2", Url: "https://example.com"},
	}
	assert.False(t, endpoints.Less(0, 1))
}

func TestRuntimeObjectList_Has(t *testing.T) {
	list := RuntimeObjectList{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}},
	}

	t.Run("returns true when object is in list", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}}
		assert.True(t, list.Has(obj))
	})

	t.Run("returns false when object is not in list", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod3"}}
		assert.False(t, list.Has(obj))
	})
}

func TestProjectObjectMap_GetProject(t *testing.T) {
	mapObj := projectObjectMap{
		"Project1": RuntimeObjectList{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}},
		},
		"Project2": RuntimeObjectList{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}},
		},
	}

	t.Run("returns project name and true when object is in map", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}}
		projectName, ok := mapObj.GetProject(obj)
		assert.True(t, ok)
		assert.Equal(t, "Project1", projectName)
	})

	t.Run("returns empty string and false when object is not in map", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod3"}}
		projectName, ok := mapObj.GetProject(obj)
		assert.False(t, ok)
		assert.Equal(t, "", projectName)
	})
}

func TestEndpointMapping_AllEndpoints(t *testing.T) {
	mapping := EndpointMapping{
		"Project1": []*types.ServiceEndpoint{
			{Name: "Endpoint1", Url: "http://example.com"},
		},
		"Project2": []*types.ServiceEndpoint{
			{Name: "Endpoint2", Url: "https://example.com"},
		},
	}

	t.Run("returns all endpoints", func(t *testing.T) {
		endpoints := mapping.AllEndpoints()
		assert.Equal(t, 2, len(endpoints))
		// AllEndpoints 遍历 map，顺序不确定，断言改为无序集合比较。
		names := []string{endpoints[0].Name, endpoints[1].Name}
		assert.ElementsMatch(t, []string{"Endpoint1", "Endpoint2"}, names)
	})
}

func TestEndpointMapping_Sort(t *testing.T) {
	mapping := EndpointMapping{
		"Project1": []*types.ServiceEndpoint{
			{Name: "Endpoint1", Url: "http://example.com"},
			{Name: "Endpoint2", Url: "https://example.com"},
		},
		"Project2": []*types.ServiceEndpoint{
			{Name: "Endpoint3", Url: "http://example.com"},
			{Name: "Endpoint4", Url: "https://example.com"},
		},
	}

	mapping.Sort()

	t.Run("Endpoints should be sorted by Url", func(t *testing.T) {
		for _, endpoints := range mapping {
			for i := 0; i < len(endpoints)-1; i++ {
				if strings.HasPrefix(endpoints[i].Url, "http") && strings.HasPrefix(endpoints[i+1].Url, "https") {
					t.Errorf("Endpoints are not sorted correctly")
				}
			}
		}
	})
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
	res := FilterRuntimeObjectFromManifests[*corev1.Service](mlog.NewForConfig(nil), data)
	assert.Len(t, res, 1)
	res1 := FilterRuntimeObjectFromManifests[*corev1.Pod](mlog.NewForConfig(nil), data)
	assert.Len(t, res1, 1)
	res2 := FilterRuntimeObjectFromManifests[*corev1.Namespace](mlog.NewForConfig(nil), data)
	assert.Len(t, res2, 0)
}

// TestFilterRuntimeObjectFromManifests_InvalidYaml 覆盖 YAML 反序列化失败分支：无效 manifest
// 记录 Warning 日志后跳过，不影响其余对象解析。
func TestFilterRuntimeObjectFromManifests_InvalidYaml(t *testing.T) {
	res := FilterRuntimeObjectFromManifests[*corev1.Service](mlog.NewForConfig(nil), []string{"[invalid yaml"})
	assert.Len(t, res, 0)
}

// fakeEndpointK8sRepo 是 endpoint 编排 + 容器拓扑推导共用的 K8sRepo 替身，
// 覆盖全部被调用的低层读取原语，字段为零值时返回零值/空列表。
type fakeEndpointK8sRepo struct {
	K8sRepo
	httpRoutes         []*gatewayv1.HTTPRoute
	services           []*corev1.Service
	ingresses          []*networkingv1.Ingress
	pods               []*corev1.Pod
	replicaSets        map[string]*appsv1.ReplicaSet
	gatewayInstalled   bool
	externalIP         string
	listHTTPRoutesErr  error
	listServicesErr    error
	listIngressesErr   error
	listPodsErr        error
	servicesCall       int
	servicesFailOnCall int // 指定第几次 ListServices 调用返回错误；0 表示第一次调用即失败
}

func (f *fakeEndpointK8sRepo) GatewayApiInstalled() bool { return f.gatewayInstalled }

func (f *fakeEndpointK8sRepo) ExternalIp() string { return f.externalIP }

func (f *fakeEndpointK8sRepo) ListHTTPRoutes(namespace string) ([]*gatewayv1.HTTPRoute, error) {
	return f.httpRoutes, f.listHTTPRoutesErr
}

func (f *fakeEndpointK8sRepo) ListServices(namespace string) ([]*corev1.Service, error) {
	f.servicesCall++
	if f.listServicesErr != nil && (f.servicesFailOnCall == 0 || f.servicesCall == f.servicesFailOnCall) {
		return nil, f.listServicesErr
	}
	return f.services, nil
}

func (f *fakeEndpointK8sRepo) ListIngresses(namespace string) ([]*networkingv1.Ingress, error) {
	return f.ingresses, f.listIngressesErr
}

func (f *fakeEndpointK8sRepo) ListPodsBySelectors(namespace string, selectors []string) ([]*corev1.Pod, error) {
	return f.pods, f.listPodsErr
}

func (f *fakeEndpointK8sRepo) GetReplicaSet(namespace, name string) (*appsv1.ReplicaSet, error) {
	return f.replicaSets[name], nil
}

// 以下 YAML 是单个项目 Manifest 的快照，供四个 Build* 编排函数匹配集群对象使用。
// 对象名（web-svc / web-ing / web-route）必须与集群返回对象同名，GetProject 按名匹配。
const (
	svcManifest = `apiVersion: v1
kind: Service
metadata:
  name: web-svc
  namespace: ns
spec:
  ports:
  - name: web
    port: 80
`
	ingressManifest = `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-ing
  namespace: ns
spec:
  rules:
  - host: a.example.com
`
	httpRouteManifest = `apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: web-route
  namespace: ns
spec:
  hostnames:
  - x.example.com
`
)

func TestBuildGatewayHTTPRouteMappingByProjects_ListErr(t *testing.T) {
	k := &fakeEndpointK8sRepo{gatewayInstalled: true, listHTTPRoutesErr: errors.New("routes down")}
	_, err := BuildGatewayHTTPRouteMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{httpRouteManifest}})
	assert.ErrorContains(t, err, "routes down")
}

func TestBuildGatewayHTTPRouteMappingByProjects_GatewayNotInstalled(t *testing.T) {
	k := &fakeEndpointK8sRepo{gatewayInstalled: false}
	got, err := BuildGatewayHTTPRouteMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{httpRouteManifest}})
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestBuildGatewayHTTPRouteMappingByProjects_HappyPath(t *testing.T) {
	k := &fakeEndpointK8sRepo{
		gatewayInstalled: true,
		httpRoutes: []*gatewayv1.HTTPRoute{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-route", Namespace: "ns"},
			Spec:       gatewayv1.HTTPRouteSpec{Hostnames: []gatewayv1.Hostname{"x.example.com"}},
		}},
	}
	got, err := BuildGatewayHTTPRouteMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{httpRouteManifest}})
	assert.NoError(t, err)
	assert.Equal(t, "https://x.example.com", got["proj1"][0].Url)
}

func TestBuildNodePortMappingByProjects_ListErr(t *testing.T) {
	k := &fakeEndpointK8sRepo{listServicesErr: errors.New("svc down")}
	_, err := BuildNodePortMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{svcManifest}})
	assert.ErrorContains(t, err, "svc down")
}

func TestBuildNodePortMappingByProjects_HappyPath(t *testing.T) {
	k := &fakeEndpointK8sRepo{
		externalIP: "10.0.0.1",
		services: []*corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{{Name: "web", Port: 80, NodePort: 30080}},
			},
		}},
	}
	got, err := BuildNodePortMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{svcManifest}})
	assert.NoError(t, err)
	assert.Equal(t, "http://10.0.0.1:30080", got["proj1"][0].Url)
}

func TestBuildIngressMappingByProjects_ListErr(t *testing.T) {
	k := &fakeEndpointK8sRepo{listIngressesErr: errors.New("ing down")}
	_, err := BuildIngressMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{ingressManifest}})
	assert.ErrorContains(t, err, "ing down")
}

func TestBuildIngressMappingByProjects_HappyPath(t *testing.T) {
	k := &fakeEndpointK8sRepo{
		ingresses: []*networkingv1.Ingress{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-ing", Namespace: "ns"},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{{Host: "a.example.com"}},
				TLS:   []networkingv1.IngressTLS{{Hosts: []string{"b.example.com"}}},
			},
		}},
	}
	got, err := BuildIngressMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{ingressManifest}})
	assert.NoError(t, err)
	// 排序后 https 在前：https://b.example.com 应排在 http://a.example.com 前。
	assert.ElementsMatch(t, []string{"http://a.example.com", "https://b.example.com"}, []string{got["proj1"][0].Url, got["proj1"][1].Url})
	assert.True(t, strings.HasPrefix(got["proj1"][0].Url, "https"))
}

func TestBuildLoadBalancerMappingByProjects_ListErr(t *testing.T) {
	k := &fakeEndpointK8sRepo{listServicesErr: errors.New("svc down")}
	_, err := BuildLoadBalancerMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{svcManifest}})
	assert.ErrorContains(t, err, "svc down")
}

func TestBuildLoadBalancerMappingByProjects_HappyPath(t *testing.T) {
	k := &fakeEndpointK8sRepo{
		services: []*corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{{Name: "web", Port: 443}},
			},
			Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: "1.2.3.4"}}}},
		}},
	}
	got, err := BuildLoadBalancerMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{svcManifest}})
	assert.NoError(t, err)
	// 端口 443 + http 端口名 → https 无端口号。
	assert.Equal(t, "https://1.2.3.4", got["proj1"][0].Url)
}

func TestBuildLoadBalancerMappingByProjects_Port80(t *testing.T) {
	k := &fakeEndpointK8sRepo{
		services: []*corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{{Name: "http", Port: 80}},
			},
			Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: "1.2.3.4"}}}},
		}},
	}
	got, err := BuildLoadBalancerMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{svcManifest}})
	assert.NoError(t, err)
	// 端口 80 + http 端口名 → http 无端口号。
	assert.Equal(t, "http://1.2.3.4", got["proj1"][0].Url)
}

func TestBuildLoadBalancerMappingByProjects_NonHttpPort(t *testing.T) {
	k := &fakeEndpointK8sRepo{
		services: []*corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{{Name: "mysql", Port: 3306}},
			},
			Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: "1.2.3.4"}}}},
		}},
	}
	got, err := BuildLoadBalancerMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{svcManifest}})
	assert.NoError(t, err)
	// 非 http 端口名 → 不带协议前缀的 host:port。
	assert.Equal(t, "1.2.3.4:3306", got["proj1"][0].Url)
}

// 以下四个测试覆盖各 Build* 在"无项目"时 projectMap 为空的提前返回分支。

func TestBuildGatewayHTTPRouteMappingByProjects_NoProjects(t *testing.T) {
	k := &fakeEndpointK8sRepo{gatewayInstalled: true}
	got, err := BuildGatewayHTTPRouteMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestBuildNodePortMappingByProjects_NoProjects(t *testing.T) {
	k := &fakeEndpointK8sRepo{}
	got, err := BuildNodePortMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestBuildIngressMappingByProjects_NoProjects(t *testing.T) {
	k := &fakeEndpointK8sRepo{}
	got, err := BuildIngressMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestBuildLoadBalancerMappingByProjects_NoProjects(t *testing.T) {
	k := &fakeEndpointK8sRepo{}
	got, err := BuildLoadBalancerMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

// TestBuildNodePortMappingByProjects_NonHttpPort 覆盖 NodePort 编排的非 http 端口名 default 分支。
func TestBuildNodePortMappingByProjects_NonHttpPort(t *testing.T) {
	k := &fakeEndpointK8sRepo{
		externalIP: "10.0.0.1",
		services: []*corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{{Name: "mysql", Port: 3306, NodePort: 33306}},
			},
		}},
	}
	got, err := BuildNodePortMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k, "ns", &Project{Name: "proj1", Manifest: []string{svcManifest}})
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1:33306", got["proj1"][0].Url)
}
