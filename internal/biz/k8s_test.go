package biz

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// TestDockerConfigJSONSerializationContract 锁定 docker config.json 的序列化契约：
// 字段必须小写（username/password/email/auth），kubelet 解析依赖此格式。
// 曾因缺 json tag 序列化出大写键导致镜像拉取认证失效（P0）。
func TestDockerConfigJSONSerializationContract(t *testing.T) {
	cfg := DockerConfigJSON{
		Auths: DockerConfig{"reg.io": {
			Username: "u",
			Password: "p",
			Email:    "e",
			Auth:     base64.StdEncoding.EncodeToString([]byte("u:p")),
		}},
	}
	data, err := json.Marshal(cfg)
	assert.NoError(t, err)

	raw := map[string]map[string]map[string]string{}
	assert.NoError(t, json.Unmarshal(data, &raw))
	entry := raw["auths"]["reg.io"]
	assert.Equal(t, "u", entry["username"])
	assert.Equal(t, "p", entry["password"])
	assert.Equal(t, "e", entry["email"])
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("u:p")), entry["auth"])
	assert.NotContains(t, entry, "Username")
}

// TestDecodeDockerConfigJSON 覆盖 config.json 字节流解析为领域类型的往返。
func TestDecodeDockerConfigJSON(t *testing.T) {
	input := []byte(`{"auths": {"https://index.docker.io/v1/": {"username": "tu", "password": "tp", "email": "te", "auth": "dXU6cA=="}}}`)
	res, err := DecodeDockerConfigJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, "tu", res.Auths["https://index.docker.io/v1/"].Username)
	assert.Equal(t, "tp", res.Auths["https://index.docker.io/v1/"].Password)
	assert.Equal(t, "te", res.Auths["https://index.docker.io/v1/"].Email)

	// 空 Auths 保持可解码
	empty, err := DecodeDockerConfigJSON([]byte(`{"auths": {}}`))
	assert.NoError(t, err)
	assert.Empty(t, empty.Auths)
}

// TestProjectCpuMemory 验证聚合规则把"项目全部 pod metrics"喂给聚合器并原样返回结果。
func TestProjectCpuMemory(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	k8sBiz := NewMockK8sBiz(m)
	proj := &Project{ID: 1, Name: "proj"}
	me := []v1beta1.PodMetrics{{}}

	k8sBiz.EXPECT().GetAllPodMetrics(gomock.Any(), proj).Return(me)
	k8sBiz.EXPECT().GetCpuAndMemory(gomock.Any(), me).Return("100m", "1Gi")

	cpu, memory := ProjectCpuMemory(context.Background(), k8sBiz, proj)
	assert.Equal(t, "100m", cpu)
	assert.Equal(t, "1Gi", memory)
}

func TestGetPreOccupiedLenByValuesYaml(t *testing.T) {
	t.Run("returns zero when values is empty", func(t *testing.T) {
		values := ""
		got := GetPreOccupiedLenByValuesYaml(values)
		assert.Equal(t, 0, got)
	})

	t.Run("returns correct length when values contains host", func(t *testing.T) {
		values := "  testHost< .Host1 >"
		got := GetPreOccupiedLenByValuesYaml(values)
		assert.Equal(t, len("testHost"), got)
	})

	t.Run("returns max length when values contains multiple hosts", func(t *testing.T) {
		values := "  testHost< .Host1 >  longerTestHost< .Host2 >"
		got := GetPreOccupiedLenByValuesYaml(values)
		assert.Equal(t, len("longerTestHost"), got)
	})

	t.Run("ignores non-host values", func(t *testing.T) {
		values := "  testHost< .Host1 >  nonHostValue"
		got := GetPreOccupiedLenByValuesYaml(values)
		assert.Equal(t, len("testHost"), got)
	})
}

// fakeK8sRepoForK8sBiz 记录各写操作是否被调用，输入校验测试中 repo 不被调用（调用即 panic）。
type fakeK8sRepoForK8sBiz struct {
	K8sRepo
	addTlsCalled, createNsCalled, deleteNsCalled, deleteSecretCalled, deletePodCalled bool
	deletePodOpts                                                                     metav1.DeleteOptions
}

func (f *fakeK8sRepoForK8sBiz) AddTlsSecret(ns string, name string, key string, crt string) (*corev1.Secret, error) {
	f.addTlsCalled = true
	return &corev1.Secret{}, nil
}

func (f *fakeK8sRepoForK8sBiz) CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	f.createNsCalled = true
	return &corev1.Namespace{}, nil
}

func (f *fakeK8sRepoForK8sBiz) DeleteSecret(ctx context.Context, namespace, secret string) error {
	f.deleteSecretCalled = true
	return nil
}

func (f *fakeK8sRepoForK8sBiz) DeleteNamespace(ctx context.Context, name string) error {
	f.deleteNsCalled = true
	return nil
}

func (f *fakeK8sRepoForK8sBiz) DeletePod(ctx context.Context, namespace, pod string, opts metav1.DeleteOptions) error {
	f.deletePodCalled = true
	f.deletePodOpts = opts
	return nil
}

func TestK8sBiz_AddTlsSecret_EmptyNSOrName(t *testing.T) {
	k := NewK8sBiz(&fakeK8sRepoForK8sBiz{})
	got, err := k.AddTlsSecret("", "", "key", "crt")
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace 或 secret 名称不能为空", status.Convert(err).Message())
}

func TestK8sBiz_AddTlsSecret_Valid(t *testing.T) {
	f := &fakeK8sRepoForK8sBiz{}
	b := NewK8sBiz(f)
	got, err := b.AddTlsSecret("ns", "name", "key", "crt")
	assert.NoError(t, err)
	assert.True(t, f.addTlsCalled)
	assert.NotNil(t, got)
}

func TestK8sBiz_CreateNamespace_EmptyName(t *testing.T) {
	k := NewK8sBiz(&fakeK8sRepoForK8sBiz{})
	got, err := k.CreateNamespace(context.TODO(), "")
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace 名称不能为空", status.Convert(err).Message())
}

func TestK8sBiz_CreateNamespace_Valid(t *testing.T) {
	f := &fakeK8sRepoForK8sBiz{}
	b := NewK8sBiz(f)
	got, err := b.CreateNamespace(context.TODO(), "ns")
	assert.NoError(t, err)
	assert.True(t, f.createNsCalled)
	assert.NotNil(t, got)
}

func TestK8sBiz_DeleteNamespace_EmptyName(t *testing.T) {
	k := NewK8sBiz(&fakeK8sRepoForK8sBiz{})
	err := k.DeleteNamespace(context.TODO(), "")
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace 名称不能为空", status.Convert(err).Message())
}

func TestK8sBiz_DeleteNamespace_Valid(t *testing.T) {
	f := &fakeK8sRepoForK8sBiz{}
	b := NewK8sBiz(f)
	assert.NoError(t, b.DeleteNamespace(context.TODO(), "ns"))
	assert.True(t, f.deleteNsCalled)
}

func TestK8sBiz_DeleteSecret_EmptyNamespaceOrSecret(t *testing.T) {
	k := NewK8sBiz(&fakeK8sRepoForK8sBiz{})
	err := k.DeleteSecret(context.TODO(), "", "")
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace 或 secret 名称不能为空", status.Convert(err).Message())
}

func TestK8sBiz_DeleteSecret_Valid(t *testing.T) {
	f := &fakeK8sRepoForK8sBiz{}
	b := NewK8sBiz(f)
	assert.NoError(t, b.DeleteSecret(context.TODO(), "ns", "sec"))
	assert.True(t, f.deleteSecretCalled)
}

func TestK8sBiz_ForceDeletePod_InvalidArgs(t *testing.T) {
	k := NewK8sBiz(&fakeK8sRepoForK8sBiz{})
	err := k.ForceDeletePod(context.TODO(), "", "pod", 0)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace 或 pod 名称不能为空", status.Convert(err).Message())
	err = k.ForceDeletePod(context.TODO(), "ns", "", 0)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	err = k.ForceDeletePod(context.TODO(), "ns", "pod", -1)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "grace period seconds 不能为负数", status.Convert(err).Message())
}

func TestK8sBiz_ForceDeletePod_Valid(t *testing.T) {
	f := &fakeK8sRepoForK8sBiz{}
	b := NewK8sBiz(f)
	assert.NoError(t, b.ForceDeletePod(context.TODO(), "ns", "pod", 0))
	assert.True(t, f.deletePodCalled)
	// 透传强制删除策略：grace-period=0 + 后台传播
	assert.Equal(t, int64(0), *f.deletePodOpts.GracePeriodSeconds)
	assert.Equal(t, metav1.DeletePropagationBackground, *f.deletePodOpts.PropagationPolicy)
}

func TestK8sBiz_ForceDeletePod_ValidNonZeroGrace(t *testing.T) {
	f := &fakeK8sRepoForK8sBiz{}
	b := NewK8sBiz(f)
	assert.NoError(t, b.ForceDeletePod(context.TODO(), "ns", "pod", 30))
	assert.True(t, f.deletePodCalled)
	assert.Equal(t, int64(30), *f.deletePodOpts.GracePeriodSeconds)
}

// fakeK8sRepoPassthrough 覆盖 k8sBiz 全部纯透传方法，记录调用并返回罐头数据。
type fakeK8sRepoPassthrough struct {
	K8sRepo
	calls map[string]bool
}

func (f *fakeK8sRepoPassthrough) mark(name string) {
	if f.calls == nil {
		f.calls = map[string]bool{}
	}
	f.calls[name] = true
}

func (f *fakeK8sRepoPassthrough) SplitManifests(manifest string) []string {
	f.mark("SplitManifests")
	return []string{"a", "b"}
}

func (f *fakeK8sRepoPassthrough) GetPodMetrics(ctx context.Context, namespace, podName string) (*v1beta1.PodMetrics, error) {
	f.mark("GetPodMetrics")
	return &v1beta1.PodMetrics{}, nil
}

func (f *fakeK8sRepoPassthrough) CreateDockerSecret(ctx context.Context, namespace string) (*corev1.Secret, error) {
	f.mark("CreateDockerSecret")
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "docker"}}, nil
}

func (f *fakeK8sRepoPassthrough) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	f.mark("GetNamespace")
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}, nil
}

func (f *fakeK8sRepoPassthrough) LogStream(ctx context.Context, namespace, pod, container string) (chan []byte, error) {
	f.mark("LogStream")
	return make(chan []byte), nil
}

func (f *fakeK8sRepoPassthrough) GetPodLogs(ctx context.Context, namespace, podName string, options *corev1.PodLogOptions) (string, error) {
	f.mark("GetPodLogs")
	return "logs", nil
}

func (f *fakeK8sRepoPassthrough) FindDefaultContainer(ctx context.Context, namespace string, pod string) (string, error) {
	f.mark("FindDefaultContainer")
	return "app", nil
}

func (f *fakeK8sRepoPassthrough) GetPod(namespace, podName string) (*corev1.Pod, error) {
	f.mark("GetPod")
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName}}, nil
}

func (f *fakeK8sRepoPassthrough) ListEvents(namespace string) ([]*eventv1.Event, error) {
	f.mark("ListEvents")
	return []*eventv1.Event{{}}, nil
}

func (f *fakeK8sRepoPassthrough) IsPodRunning(namespace, podName string) (running bool, notRunningReason string) {
	f.mark("IsPodRunning")
	return true, ""
}

func (f *fakeK8sRepoPassthrough) GetPodSelectorsByManifest(manifests []string) []string {
	f.mark("GetPodSelectorsByManifest")
	return []string{"app=nginx"}
}

func (f *fakeK8sRepoPassthrough) GetCpuAndMemoryInNamespace(ctx context.Context, namespace string) (string, string) {
	f.mark("GetCpuAndMemoryInNamespace")
	return "100m", "1Gi"
}

func (f *fakeK8sRepoPassthrough) GetCpuAndMemory(ctx context.Context, list []v1beta1.PodMetrics) (string, string) {
	f.mark("GetCpuAndMemory")
	return "200m", "2Gi"
}

func (f *fakeK8sRepoPassthrough) GetCpuAndMemoryQuantity(pod v1beta1.PodMetrics) (cpu *resource.Quantity, memory *resource.Quantity) {
	f.mark("GetCpuAndMemoryQuantity")
	return resource.NewMilliQuantity(100, resource.DecimalSI), resource.NewQuantity(1024, resource.BinarySI)
}

func (f *fakeK8sRepoPassthrough) ClusterInfo() *ClusterInfo {
	f.mark("ClusterInfo")
	return &ClusterInfo{Status: StatusHealth}
}

func (f *fakeK8sRepoPassthrough) Execute(ctx context.Context, c *Container, input *ExecuteInput) error {
	f.mark("Execute")
	return nil
}

func (f *fakeK8sRepoPassthrough) GetAllPodMetrics(ctx context.Context, proj *Project) []v1beta1.PodMetrics {
	f.mark("GetAllPodMetrics")
	return []v1beta1.PodMetrics{{}}
}

func (f *fakeK8sRepoPassthrough) CopyFileToPod(ctx context.Context, input *CopyFileToPodInput) (*File, error) {
	f.mark("CopyFileToPod")
	return &File{ID: 1}, nil
}

func (f *fakeK8sRepoPassthrough) CopyFromPod(ctx context.Context, input *CopyFromPodInput) (*File, error) {
	f.mark("CopyFromPod")
	return &File{ID: 2}, nil
}

func TestK8sBiz_PassthroughMethods(t *testing.T) {
	f := &fakeK8sRepoPassthrough{}
	b := NewK8sBiz(f)

	t.Run("SplitManifests", func(t *testing.T) {
		assert.Equal(t, []string{"a", "b"}, b.SplitManifests("---"))
	})
	t.Run("GetPodMetrics", func(t *testing.T) {
		m, err := b.GetPodMetrics(context.TODO(), "ns", "pod")
		assert.NoError(t, err)
		assert.NotNil(t, m)
	})
	t.Run("CreateDockerSecret", func(t *testing.T) {
		s, err := b.CreateDockerSecret(context.TODO(), "ns")
		assert.NoError(t, err)
		assert.Equal(t, "docker", s.Name)
	})
	t.Run("GetNamespace", func(t *testing.T) {
		ns, err := b.GetNamespace(context.TODO(), "ns")
		assert.NoError(t, err)
		assert.Equal(t, "ns", ns.Name)
	})
	t.Run("LogStream", func(t *testing.T) {
		ch, err := b.LogStream(context.TODO(), "ns", "pod", "c")
		assert.NoError(t, err)
		assert.NotNil(t, ch)
	})
	t.Run("GetPodLogs", func(t *testing.T) {
		logs, err := b.GetPodLogs(context.TODO(), "ns", "pod", &corev1.PodLogOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "logs", logs)
	})
	t.Run("FindDefaultContainer", func(t *testing.T) {
		c, err := b.FindDefaultContainer(context.TODO(), "ns", "pod")
		assert.NoError(t, err)
		assert.Equal(t, "app", c)
	})
	t.Run("GetPod", func(t *testing.T) {
		p, err := b.GetPod("ns", "pod")
		assert.NoError(t, err)
		assert.Equal(t, "pod", p.Name)
	})
	t.Run("ListEvents", func(t *testing.T) {
		ev, err := b.ListEvents("ns")
		assert.NoError(t, err)
		assert.Len(t, ev, 1)
	})
	t.Run("IsPodRunning", func(t *testing.T) {
		running, reason := b.IsPodRunning("ns", "pod")
		assert.True(t, running)
		assert.Equal(t, "", reason)
	})
	t.Run("GetPodSelectorsByManifest", func(t *testing.T) {
		assert.Equal(t, []string{"app=nginx"}, b.GetPodSelectorsByManifest([]string{"x"}))
	})
	t.Run("GetCpuAndMemoryInNamespace", func(t *testing.T) {
		cpu, mem := b.GetCpuAndMemoryInNamespace(context.TODO(), "ns")
		assert.Equal(t, "100m", cpu)
		assert.Equal(t, "1Gi", mem)
	})
	t.Run("GetCpuAndMemory", func(t *testing.T) {
		cpu, mem := b.GetCpuAndMemory(context.TODO(), nil)
		assert.Equal(t, "200m", cpu)
		assert.Equal(t, "2Gi", mem)
	})
	t.Run("GetCpuAndMemoryQuantity", func(t *testing.T) {
		cpu, mem := b.GetCpuAndMemoryQuantity(v1beta1.PodMetrics{})
		assert.Equal(t, "100m", cpu.String())
		assert.Equal(t, "1Ki", mem.String())
	})
	t.Run("ClusterInfo", func(t *testing.T) {
		ci := b.ClusterInfo()
		assert.Equal(t, StatusHealth, ci.Status)
	})
	t.Run("Execute", func(t *testing.T) {
		assert.NoError(t, b.Execute(context.TODO(), &Container{}, &ExecuteInput{}))
	})
	t.Run("GetAllPodMetrics", func(t *testing.T) {
		assert.Len(t, b.GetAllPodMetrics(context.TODO(), &Project{}), 1)
	})
	t.Run("CopyFileToPod", func(t *testing.T) {
		got, err := b.CopyFileToPod(context.TODO(), &CopyFileToPodInput{})
		assert.NoError(t, err)
		assert.Equal(t, 1, got.ID)
	})
	t.Run("CopyFromPod", func(t *testing.T) {
		got, err := b.CopyFromPod(context.TODO(), &CopyFromPodInput{})
		assert.NoError(t, err)
		assert.Equal(t, 2, got.ID)
	})

	// 全部透传方法都实际调用了 repo。
	for _, name := range []string{
		"SplitManifests", "GetPodMetrics", "CreateDockerSecret", "GetNamespace",
		"LogStream", "GetPodLogs", "FindDefaultContainer", "GetPod", "ListEvents",
		"IsPodRunning", "GetPodSelectorsByManifest", "GetCpuAndMemoryInNamespace",
		"GetCpuAndMemory", "GetCpuAndMemoryQuantity", "ClusterInfo", "Execute",
		"GetAllPodMetrics", "CopyFileToPod", "CopyFromPod",
	} {
		assert.True(t, f.calls[name], "method %s 未调用", name)
	}
}
