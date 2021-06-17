package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func CreateDockerSecret(namespace, username, password, email string) (*v1.Secret, error) {
	dockercfgAuth := DockerConfigEntry{
		Username: username,
		Password: password,
		Email:    email,
		Auth:     base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
	}

	dockerCfgJSON := DockerConfigJSON{
		Auths: map[string]DockerConfigEntry{"server": dockercfgAuth},
	}

	marshal, _ := json.Marshal(dockerCfgJSON)

	return K8sClientSet().CoreV1().Secrets(namespace).Create(context.Background(), &v1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: "mars-",
		},
		Data: map[string][]byte{
			v1.DockerConfigJsonKey: marshal,
		},
		Type: v1.DockerConfigJsonKey,
	}, metav1.CreateOptions{})
}

func GetNodePortMappingByNamespace(namespace string) map[string][]string {
	list, _ := K8sClientSet().CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	var m = map[string][]string{}

	for _, item := range list.Items {
		if item.Spec.Type == v1.ServiceTypeNodePort {
			for _, port := range item.Spec.Ports {
				if projectName, ok := item.Spec.Selector["app.kubernetes.io/instance"]; ok {
					data := m[projectName]
					switch {
					case strings.Contains(port.Name, "http"):
						m[projectName] = append(data, fmt.Sprintf("http://%s:%d", Config().ClusterIp, port.NodePort))
					case strings.Contains(port.Name, "rpc"):
						fallthrough
					case strings.Contains(port.Name, "tcp"):
						m[projectName] = append(data, fmt.Sprintf("%s://%s:%d", port.Name, Config().ClusterIp, port.NodePort))
					}
				}
			}
		}
	}

	return m
}

func GetIngressMappingByNamespace(namespace string) map[string][]string {
	var m = map[string][]string{}

	list, _ := K8sClientSet().NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	for _, item := range list.Items {
		for _, tls := range item.Spec.TLS {
			if projectName, ok := item.Labels["app.kubernetes.io/instance"]; ok {
				data := m[projectName]
				for _, host := range tls.Hosts {
					m[projectName] = append(data, fmt.Sprintf("https://%s", host))
				}
			}
		}
	}

	return m
}
