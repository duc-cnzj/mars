package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
)

type DockerConfig map[string]DockerConfigEntry

type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

type DockerConfigJSON struct {
	Auths       DockerConfig      `json:"auths"`
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`
}

func CreateDockerSecret(namespace, username, password, email, server string) (*v1.Secret, error) {
	if server == "" {
		server = "https://index.docker.io/v1/"
	}
	dockercfgAuth := DockerConfigEntry{
		Username: username,
		Password: password,
		Email:    email,
		Auth:     base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
	}

	dockerCfgJSON := DockerConfigJSON{
		Auths: map[string]DockerConfigEntry{server: dockercfgAuth},
	}

	marshal, _ := json.Marshal(dockerCfgJSON)

	return app.K8sClientSet().CoreV1().Secrets(namespace).Create(context.Background(), &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: "mars-",
		},
		Data: map[string][]byte{
			v1.DockerConfigJsonKey: marshal,
		},
		Type: v1.SecretTypeDockerConfigJson,
	}, metav1.CreateOptions{})
}

type Endpoint = types.ServiceEndpoint

type RuntimeObjectList []runtime.Object

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

func FilterRuntimeObjectFromManifests[T runtime.Object](manifests []string) RuntimeObjectList {
	var m = make(RuntimeObjectList, 0)
	info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	for _, f := range manifests {
		obj, _, err := info.Serializer.Decode([]byte(f), nil, nil)
		if err != nil {
			mlog.Error(err)
			continue
		}
		switch obj.(type) {
		case T:
			m = append(m, obj)
		}
	}

	return m
}

type projectObjectMap map[string]RuntimeObjectList

func (m projectObjectMap) GetProject(svc runtime.Object) (string, bool) {
	for projectName, set := range m {
		if set.Has(svc) {
			return projectName, true
		}
	}
	return "", false
}

func GetNodePortMappingByProjects(namespace string, projects ...models.Project) map[string][]*Endpoint {
	cfg := app.Config()
	var projectMap = make(projectObjectMap)
	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*v1.Service](Filter[string](strings.Split(project.Manifest, "---"), func(item string, index int) bool { return len(item) > 0 }))
	}

	list, _ := app.K8sClientSet().CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	var m = map[string][]*Endpoint{}

	isHttp := func(name string) bool {
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

	for _, item := range list.Items {
		if projectName, ok := projectMap.GetProject(&item); ok && item.Spec.Type == v1.ServiceTypeNodePort {
			for _, port := range item.Spec.Ports {
				data := m[projectName]

				switch {
				case strings.Contains(port.Name, "rpc"):
					fallthrough
				case strings.Contains(port.Name, "tcp"):
					m[projectName] = append(data, &Endpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("%s://%s:%d", port.Name, cfg.ExternalIp, port.NodePort),
					})
				case isHttp(port.Name):
					m[projectName] = append(data, &Endpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("http://%s:%d", cfg.ExternalIp, port.NodePort),
					})
				default:
					m[projectName] = append(data, &Endpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("%s:%d", cfg.ExternalIp, port.NodePort),
					})
				}
			}
		}
	}
	return m
}

func GetIngressMappingByProjects(namespace string, projects ...models.Project) map[string][]*Endpoint {
	var projectMap = make(projectObjectMap)
	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*networkingv1.Ingress](Filter[string](strings.Split(project.Manifest, "---"), func(item string, index int) bool { return len(item) > 0 }))
	}

	var m = map[string][]*Endpoint{}

	list, _ := app.K8sClientSet().NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	for _, item := range list.Items {
		if len(item.Spec.TLS) > 0 {
			for _, tls := range item.Spec.TLS {
				if projectName, ok := projectMap.GetProject(&item); ok {
					data := m[projectName]
					var hosts []*Endpoint
					for _, host := range tls.Hosts {
						hosts = append(hosts, &Endpoint{Name: projectName, Url: fmt.Sprintf("https://%s", host)})
					}
					m[projectName] = append(data, hosts...)
				}
			}
		} else {
			for _, rules := range item.Spec.Rules {
				if projectName, ok := projectMap.GetProject(&item); ok {
					data := m[projectName]
					var hosts []*Endpoint
					hosts = append(hosts, &Endpoint{Name: projectName, Url: fmt.Sprintf("http://%s", rules.Host)})
					m[projectName] = append(data, hosts...)
				}
			}
		}
	}

	return m
}

func IsPodRunning(namespace, podName string) (running bool, notRunningReason string) {
	podInfo, err := app.K8sClientSet().CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return false, err.Error()
	}

	if podInfo.Status.Phase == v1.PodRunning {
		return true, ""
	}

	if podInfo.Status.Phase == v1.PodFailed && podInfo.Status.Reason == "Evicted" {
		return false, fmt.Sprintf("po %s already evicted in namespace %s!", podName, namespace)
	}

	for _, status := range podInfo.Status.ContainerStatuses {
		return false, fmt.Sprintf("%s %s", status.State.Waiting.Reason, status.State.Waiting.Message)
	}

	return false, "pod not running."
}

func AddTlsSecret(ns string, name string, key string, crt string) error {
	_, err := app.K8sClientSet().CoreV1().Secrets(ns).Create(context.TODO(), &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				"created-by": "mars",
			},
		},
		StringData: map[string]string{
			"tls.key": key,
			"tls.crt": crt,
		},
		Type: v1.SecretTypeTLS,
	}, metav1.CreateOptions{})
	return err
}
