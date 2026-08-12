package biz

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"go.opentelemetry.io/otel"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// init 注册 gateway API v1 到 scheme，供 Manifest 反序列化 HTTPRoute 使用。
// Install 对内置类型恒返回 nil，panic 分支不可达已删。
func init() {
	_ = gatewayv1.Install(scheme.Scheme)
}

// RuntimeObjectList 是同一类型的一组 k8s 对象（从项目 Manifest 反序列化而来）。
type RuntimeObjectList []runtime.Object

// Has 判断列表中是否存在与 in 类型相同且同名的对象。
func (l RuntimeObjectList) Has(in runtime.Object) bool {
	inAccessor, _ := meta.Accessor(in)
	for _, set := range l {
		accessor, _ := meta.Accessor(set)
		if reflect.TypeOf(set) == reflect.TypeOf(in) && accessor.GetName() == inAccessor.GetName() {
			return true
		}
	}

	return false
}

// FilterRuntimeObjectFromManifests 从一组 YAML manifest 中挑出类型为 T 的对象。
func FilterRuntimeObjectFromManifests[T runtime.Object](logger mlog.Logger, manifests []string) RuntimeObjectList {
	var m = make(RuntimeObjectList, 0)
	info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	for _, f := range manifests {
		obj, _, err := info.Serializer.Decode([]byte(f), nil, nil)
		if err != nil {
			logger.Warning(err.Error())
			continue
		}
		switch obj.(type) {
		case T:
			m = append(m, obj)
		}
	}

	return m
}

// projectObjectMap 按项目名聚类项目 Manifest 反序列化出的对象列表。
type projectObjectMap map[string]RuntimeObjectList

// GetProject 返回对象所属的项目名；对象不属于任何已知项目时返回 false。
func (m projectObjectMap) GetProject(svc runtime.Object) (string, bool) {
	for projectName, set := range m {
		if set.Has(svc) {
			return projectName, true
		}
	}
	return "", false
}

// isHttpPortName 判断端口名是否暗示 HTTP 协议（web/ui/api/http 等）。
func isHttpPortName(name string) bool {
	switch {
	case strings.Contains(name, "web"):
		fallthrough
	case strings.Contains(name, "ui"):
		fallthrough
	case strings.Contains(name, "api"):
		fallthrough
	case strings.Contains(name, "http"):
		return true
	default:
		return false
	}
}

// EndpointMapping 是按项目名聚类的 endpoint 集合，用于多来源 endpoint 合并。
type EndpointMapping map[string][]*types.ServiceEndpoint

// sortEndpoint 把 HTTPS 端点排在 HTTP 之前。
type sortEndpoint []*types.ServiceEndpoint

// Len 返回端点个数。
func (s sortEndpoint) Len() int {
	return len(s)
}

// Less 判定第 i 个端点应排在第 j 个之前：HTTPS 优先于 HTTP。
func (s sortEndpoint) Less(i, j int) bool {
	return strings.HasPrefix(s[i].Url, "https") && !strings.HasPrefix(s[j].Url, "https")
}

// Swap 交换两个端点的位置。
func (s sortEndpoint) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Sort 对每个项目下的端点按 HTTPS 优先排序。
func (e EndpointMapping) Sort() {
	for _, endpoints := range e {
		sort.Sort(sortEndpoint(endpoints))
	}
}

// AllEndpoints 扁平化返回全部项目的端点列表。
func (e EndpointMapping) AllEndpoints() []*types.ServiceEndpoint {
	var res = make([]*types.ServiceEndpoint, 0)
	for _, endpoints := range e {
		res = append(res, endpoints...)
	}
	return res
}

// BuildGatewayHTTPRouteMappingByProjects 从集群 HTTPRoute 匹配项目 Manifest，产出 HTTPS 端点。
// ListHTTPRoutes 失败原样上抛（由最上层 services 统一打印），不静默返回空端点。
func BuildGatewayHTTPRouteMappingByProjects(ctx context.Context, logger mlog.Logger, k8sRepo K8sRepo, namespace string, projects ...*Project) (EndpointMapping, error) {
	_, span := otel.Tracer("").Start(ctx, "BuildGatewayHTTPRouteMappingByProjects")
	defer span.End()
	var (
		projectMap = make(projectObjectMap)
	)

	if !k8sRepo.GatewayApiInstalled() {
		return nil, nil
	}

	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*gatewayv1.HTTPRoute](logger, project.Manifest)
	}
	if len(projectMap) == 0 {
		return nil, nil
	}

	list, err := k8sRepo.ListHTTPRoutes(namespace)
	if err != nil {
		return nil, err
	}

	var m = map[string][]*types.ServiceEndpoint{}

	for idx := range list {
		item := list[idx]
		if projectName, ok := projectMap.GetProject(item); ok && len(item.Spec.Hostnames) > 0 {
			for _, hostname := range item.Spec.Hostnames {
				data := m[projectName]
				m[projectName] = append(data, &types.ServiceEndpoint{
					Name: projectName,
					Url:  "https://" + string(hostname),
				})
			}
		}
	}

	return m, nil
}

// BuildNodePortMappingByProjects 从集群 NodePort Service 匹配项目 Manifest，产出节点 IP 端点。
// ListServices 失败原样上抛（由最上层 services 统一打印），不静默返回空端点。
func BuildNodePortMappingByProjects(ctx context.Context, logger mlog.Logger, k8sRepo K8sRepo, namespace string, projects ...*Project) (EndpointMapping, error) {
	_, span := otel.Tracer("").Start(ctx, "BuildNodePortMappingByProjects")
	defer span.End()
	var (
		projectMap = make(projectObjectMap)
	)
	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*corev1.Service](logger, project.Manifest)
	}
	if len(projectMap) == 0 {
		return nil, nil
	}

	list, err := k8sRepo.ListServices(namespace)
	if err != nil {
		return nil, err
	}
	var m = map[string][]*types.ServiceEndpoint{}

	externalIp := k8sRepo.ExternalIp()
	for _, item := range list {
		if projectName, ok := projectMap.GetProject(item); ok && item.Spec.Type == corev1.ServiceTypeNodePort {
			for _, port := range item.Spec.Ports {
				data := m[projectName]

				switch {
				case isHttpPortName(port.Name):
					m[projectName] = append(data, &types.ServiceEndpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("http://%s:%d", externalIp, port.NodePort),
					})
				default:
					m[projectName] = append(data, &types.ServiceEndpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("%s:%d", externalIp, port.NodePort),
					})
				}
			}
		}
	}

	return m, nil
}

// BuildIngressMappingByProjects 从集群 Ingress 匹配项目 Manifest，按 TLS 决定 http/https。
// ListIngresses 失败原样上抛（由最上层 services 统一打印），不静默返回空端点。
func BuildIngressMappingByProjects(ctx context.Context, logger mlog.Logger, k8sRepo K8sRepo, namespace string, projects ...*Project) (EndpointMapping, error) {
	_, span := otel.Tracer("").Start(ctx, "BuildIngressMappingByProjects")
	defer span.End()
	var projectMap = make(projectObjectMap)
	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*networkingv1.Ingress](logger, project.Manifest)
	}
	if len(projectMap) == 0 {
		return nil, nil
	}

	var m = EndpointMapping{}
	list, err := k8sRepo.ListIngresses(namespace)
	if err != nil {
		return nil, err
	}
	type Host = string
	var allHosts = make(map[Host]struct {
		projectName string
		tls         bool
	})
	for _, item := range list {
		for _, rules := range item.Spec.Rules {
			if projectName, ok := projectMap.GetProject(item); ok {
				allHosts[rules.Host] = struct {
					projectName string
					tls         bool
				}{projectName: projectName, tls: false}
			}
		}
		for _, tls := range item.Spec.TLS {
			if projectName, ok := projectMap.GetProject(item); ok {
				for _, host := range tls.Hosts {
					allHosts[host] = struct {
						projectName string
						tls         bool
					}{projectName: projectName, tls: true}
				}
			}
		}
	}
	for host, data := range allHosts {
		urlScheme := "http"
		if data.tls {
			urlScheme = "https"
		}
		m[data.projectName] = append(m[data.projectName], &types.ServiceEndpoint{
			Name: data.projectName,
			Url:  fmt.Sprintf("%s://%s", urlScheme, host),
		})
	}
	m.Sort()

	return m, nil
}

// BuildLoadBalancerMappingByProjects 从集群 LoadBalancer Service 匹配项目 Manifest，产出 LB IP 端点。
// ListServices 失败原样上抛（由最上层 services 统一打印），不静默返回空端点。
func BuildLoadBalancerMappingByProjects(ctx context.Context, logger mlog.Logger, k8sRepo K8sRepo, namespace string, projects ...*Project) (EndpointMapping, error) {
	_, span := otel.Tracer("").Start(ctx, "BuildLoadBalancerMappingByProjects")
	defer span.End()
	var projectMap = make(projectObjectMap)
	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*corev1.Service](logger, project.Manifest)
	}
	if len(projectMap) == 0 {
		return nil, nil
	}
	list, err := k8sRepo.ListServices(namespace)
	if err != nil {
		return nil, err
	}
	var m = EndpointMapping{}

	for _, item := range list {
		if projectName, ok := projectMap.GetProject(item); ok && item.Spec.Type == corev1.ServiceTypeLoadBalancer && len(item.Status.LoadBalancer.Ingress) > 0 {
			lbIP := item.Status.LoadBalancer.Ingress[0].IP
			for _, port := range item.Spec.Ports {
				data := m[projectName]

				switch {
				case isHttpPortName(port.Name):
					var url = fmt.Sprintf("http://%s:%d", lbIP, port.Port)
					if port.Port == 80 {
						url = fmt.Sprintf("http://%s", lbIP)
					}
					if port.Port == 443 {
						url = fmt.Sprintf("https://%s", lbIP)
					}
					m[projectName] = append(data, &types.ServiceEndpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      url,
					})
				default:
					m[projectName] = append(data, &types.ServiceEndpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("%s:%d", lbIP, port.Port),
					})
				}
			}
		}
	}
	m.Sort()

	return m, nil
}
