package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/util/hash"
	"github.com/duc-cnzj/mars/v4/internal/util/rand"
	"helm.sh/helm/v3/pkg/releaseutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
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

func DecodeDockerConfigJSON(data []byte) (res DockerConfigJSON, err error) {
	err = json.Unmarshal(data, &res)
	return
}

func CreateDockerSecrets(client kubernetes.Interface, namespace string, auths config.DockerAuths) (*v1.Secret, error) {
	var entries = make(map[string]DockerConfigEntry)
	for _, auth := range auths {
		entries[auth.Server] = DockerConfigEntry{
			Username: auth.Username,
			Password: auth.Password,
			Email:    auth.Email,
			Auth:     base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password)),
		}
	}

	dockerCfgJSON := DockerConfigJSON{
		Auths: entries,
	}

	marshal, _ := json.Marshal(dockerCfgJSON)

	return client.CoreV1().Secrets(namespace).Create(context.Background(), &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "mars-" + strings.ToLower(rand.String(10)),
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

// SplitManifests
// 因为有些 secret 自带 --- 的值，导致 spilt "---" 解析异常
func SplitManifests(manifest string) []string {
	mapManifests := releaseutil.SplitManifests(manifest)
	var manifests []string
	for _, s := range mapManifests {
		manifests = append(manifests, s)
	}
	return manifests
}

type sortEndpoint []*Endpoint

func (s sortEndpoint) Len() int {
	return len(s)
}

func (s sortEndpoint) Less(i, j int) bool {
	return strings.HasPrefix(s[i].Url, "https") && !strings.HasPrefix(s[j].Url, "https")
}

func (s sortEndpoint) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type EndpointMapping map[string][]*Endpoint

func (e EndpointMapping) Sort() {
	for _, endpoints := range e {
		sort.Sort(sortEndpoint(endpoints))
	}
}

func (e EndpointMapping) Get(projName string) []*Endpoint {
	return e[projName]
}

func (e EndpointMapping) AllEndpoints() []*Endpoint {
	var res = make([]*Endpoint, 0)
	for _, endpoints := range e {
		res = append(res, endpoints...)
	}
	return res
}

type internalCloser struct {
	closeFn func() error
}

func (i *internalCloser) Close() error {
	return i.closeFn()
}

func NewCloser(fn func() error) io.Closer {
	return &internalCloser{closeFn: fn}
}

func GetSlugName[T int64 | int](namespaceId T, name string) string {
	return hash.Hash(fmt.Sprintf("%d-%s", namespaceId, name))
}
