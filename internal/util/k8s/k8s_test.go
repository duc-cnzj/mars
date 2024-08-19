package k8s_test

import (
	"encoding/base64"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/util/k8s"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDecodeDockerConfigJSON(t *testing.T) {
	data := []byte(`{"auths": {"https://index.docker.io/v1/": {"username": "testuser", "password": "testpass", "email": "test@example.com", "auth": "dGVzdHVzZXI6dGVzdHBhc3M="}}}`)
	expected := k8s.DockerConfigJSON{
		Auths: map[string]k8s.DockerConfigEntry{
			"https://index.docker.io/v1/": {
				Username: "testuser",
				Password: "testpass",
				Email:    "test@example.com",
				Auth:     "dGVzdHVzZXI6dGVzdHBhc3M=",
			},
		},
	}

	result, err := k8s.DecodeDockerConfigJSON(data)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCreateDockerSecrets(t *testing.T) {
	client := fake.NewSimpleClientset()
	namespace := "test-namespace"
	auths := config.DockerAuths{
		{
			Server:   "https://index.docker.io/v1/",
			Username: "testuser",
			Password: "testpass",
			Email:    "test@example.com",
		},
	}

	secret, err := k8s.CreateDockerSecrets(client, namespace, auths)

	assert.NoError(t, err)
	assert.Equal(t, namespace, secret.Namespace)
	assert.Equal(t, v1.SecretTypeDockerConfigJson, secret.Type)

	dockerConfigJSON, err := k8s.DecodeDockerConfigJSON(secret.Data[v1.DockerConfigJsonKey])

	assert.NoError(t, err)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(auths[0].Username+":"+auths[0].Password)), dockerConfigJSON.Auths[auths[0].Server].Auth)
}

func TestCreateDockerSecretsNoAuths(t *testing.T) {
	client := fake.NewSimpleClientset()
	namespace := "test-namespace"
	auths := config.DockerAuths{}

	secret, err := k8s.CreateDockerSecrets(client, namespace, auths)

	assert.NoError(t, err)
	assert.Equal(t, namespace, secret.Namespace)
	assert.Equal(t, v1.SecretTypeDockerConfigJson, secret.Type)

	dockerConfigJSON, err := k8s.DecodeDockerConfigJSON(secret.Data[v1.DockerConfigJsonKey])

	assert.NoError(t, err)
	assert.Empty(t, dockerConfigJSON.Auths)
}
