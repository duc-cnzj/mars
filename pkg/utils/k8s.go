package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/generate/versioned"
)

func CreateDockerSecret(namespace, username, password, email string) (*v1.Secret, error) {
	dockercfgAuth := versioned.DockerConfigEntry{
		Username: username,
		Password: password,
		Email:    email,
		Auth:     base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
	}

	dockerCfgJSON := versioned.DockerConfigJSON{
		Auths: map[string]versioned.DockerConfigEntry{"server": dockercfgAuth},
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
