package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

// fakeProjectBiz 只覆盖 EndpointBiz 用到的 Show/GetProjectEndpointsInNamespace，
// 其余接口方法由嵌入的 ProjectBiz 接口兜底（测试不会调用到）。
type fakeProjectBiz struct {
	ProjectBiz
	show         func(ctx context.Context, id int) (*Project, error)
	getEndpoints func(ctx context.Context, namespace string, projectIDs ...int) ([]*types.ServiceEndpoint, error)
}

func (f *fakeProjectBiz) Show(ctx context.Context, id int) (*Project, error) {
	return f.show(ctx, id)
}

func (f *fakeProjectBiz) GetProjectEndpointsInNamespace(ctx context.Context, namespace string, projectIDs ...int) ([]*types.ServiceEndpoint, error) {
	return f.getEndpoints(ctx, namespace, projectIDs...)
}

// fakeNamespaceRepo 只覆盖 EndpointBiz 用到的 Show 方法。
type fakeNamespaceRepo struct {
	NamespaceRepo
	show func(ctx context.Context, id int) (*Namespace, error)
}

func (f *fakeNamespaceRepo) Show(ctx context.Context, id int) (*Namespace, error) {
	return f.show(ctx, id)
}

func TestNewEndpointBiz(t *testing.T) {
	b := NewEndpointBiz(mlog.NewForConfig(nil), &fakeProjectBiz{}, &fakeNamespaceRepo{})
	assert.NotNil(t, b)
}

func TestEndpointBiz_InNamespace_HappyPath(t *testing.T) {
	proj := &fakeProjectBiz{}
	ns := &fakeNamespaceRepo{
		show: func(ctx context.Context, id int) (*Namespace, error) {
			return &Namespace{Name: "ns", Projects: []*Project{{ID: 1}}}, nil
		},
	}
	proj.getEndpoints = func(ctx context.Context, namespace string, projectIDs ...int) ([]*types.ServiceEndpoint, error) {
		assert.Equal(t, "ns", namespace)
		assert.Equal(t, []int{1}, projectIDs)
		return []*types.ServiceEndpoint{
			{Name: "a1", Url: "b1", PortName: "c1"},
			{Name: "a2", Url: "b2", PortName: "c2"},
			{Name: "a3", Url: "b3", PortName: "c4"},
		}, nil
	}
	b := NewEndpointBiz(mlog.NewForConfig(nil), proj, ns)

	res, err := b.InNamespace(context.TODO(), 1)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 3)
}

func TestEndpointBiz_InNamespace_NonExistentNamespace(t *testing.T) {
	ns := &fakeNamespaceRepo{
		show: func(ctx context.Context, id int) (*Namespace, error) {
			return nil, errors.New("x")
		},
	}
	b := NewEndpointBiz(mlog.NewForConfig(nil), &fakeProjectBiz{}, ns)

	res, err := b.InNamespace(context.TODO(), 999)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestEndpointBiz_InProject_HappyPath(t *testing.T) {
	proj := &fakeProjectBiz{
		show: func(ctx context.Context, id int) (*Project, error) {
			return &Project{Namespace: &Namespace{Name: "ns"}, ID: 1}, nil
		},
	}
	proj.getEndpoints = func(ctx context.Context, namespace string, projectIDs ...int) ([]*types.ServiceEndpoint, error) {
		assert.Equal(t, "ns", namespace)
		assert.Equal(t, []int{1}, projectIDs)
		return []*types.ServiceEndpoint{
			{Name: "ra1", Url: "rb1", PortName: "rc1"},
			{Name: "a1", Url: "b1", PortName: "c1"},
			{Name: "a2", Url: "b2", PortName: "c2"},
			{Name: "a3", Url: "b3", PortName: "c4"},
		}, nil
	}
	b := NewEndpointBiz(mlog.NewForConfig(nil), proj, &fakeNamespaceRepo{})

	res, err := b.InProject(context.TODO(), 1)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 4)
}

func TestEndpointBiz_InProject_NonExistentProject(t *testing.T) {
	proj := &fakeProjectBiz{
		show: func(ctx context.Context, id int) (*Project, error) {
			return nil, errors.New("x")
		},
	}
	b := NewEndpointBiz(mlog.NewForConfig(nil), proj, &fakeNamespaceRepo{})

	res, err := b.InProject(context.TODO(), 999)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
