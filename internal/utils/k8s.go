package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars-client/v4/endpoint"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

type Endpoint = endpoint.ServiceEndpoint

func GetNodePortMappingByNamespace(namespace string) map[string][]*Endpoint {
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
		if item.Spec.Type == v1.ServiceTypeNodePort {
			for _, port := range item.Spec.Ports {
				if projectName, ok := item.Spec.Selector["app.kubernetes.io/instance"]; ok {
					data := m[projectName]

					switch {
					case strings.Contains(port.Name, "rpc"):
						fallthrough
					case strings.Contains(port.Name, "tcp"):
						m[projectName] = append(data, &Endpoint{
							Name:     projectName,
							PortName: port.Name,
							Url:      fmt.Sprintf("%s://%s:%d", port.Name, app.Config().ExternalIp, port.NodePort),
						})
					case isHttp(port.Name):
						m[projectName] = append(data, &Endpoint{
							Name:     projectName,
							PortName: port.Name,
							Url:      fmt.Sprintf("http://%s:%d", app.Config().ExternalIp, port.NodePort),
						})
					default:
						m[projectName] = append(data, &Endpoint{
							Name:     projectName,
							PortName: port.Name,
							Url:      fmt.Sprintf("%s:%d", app.Config().ExternalIp, port.NodePort),
						})
					}
				}
			}
		}
	}
	return m
}

func GetIngressMappingByNamespace(namespace string) map[string][]*Endpoint {
	var m = map[string][]*Endpoint{}

	list, _ := app.K8sClientSet().NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	for _, item := range list.Items {
		for _, tls := range item.Spec.TLS {
			if projectName, ok := item.Labels["app.kubernetes.io/instance"]; ok {
				data := m[projectName]
				var hosts []*Endpoint
				for _, host := range tls.Hosts {
					hosts = append(hosts, &Endpoint{Name: projectName, Url: fmt.Sprintf("https://%s", host)})
				}
				m[projectName] = append(data, hosts...)
			}
		}
	}

	return m
}

func CleanEvictedPods(namespace string, selectors string) {
	mlog.Warningf("[K8s]: delete Evicted pods namespace: %s", namespace)
	list, err := app.K8sClientSet().CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selectors,
		FieldSelector: labels.Set{"status.phase": string(v1.PodFailed)}.String(),
	})
	if err != nil {
		mlog.Error(err)
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(list.Items))
	for _, item := range list.Items {
		go func(item v1.Pod) {
			defer wg.Done()
			if item.Status.Reason == "Evicted" {
				err := app.K8sClientSet().CoreV1().Pods(namespace).Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
				if err != nil {
					mlog.Error(err)
				}
			}
		}(item)
	}
	wg.Wait()
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
		go CleanEvictedPods(namespace, labels.Everything().String())
		return false, fmt.Sprintf("delete po %s when evicted in namespace %s!", podName, namespace)
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
