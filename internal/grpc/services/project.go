package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/duc-cnzj/mars/client/model"

	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/client/websocket"
	"github.com/duc-cnzj/mars/internal/socket"

	"github.com/duc-cnzj/mars/client/event"
	"github.com/duc-cnzj/mars/client/project"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type Project struct {
	project.UnimplementedProjectServer
}

func (p *Project) Apply(input *project.ProjectApplyRequest, server project.Project_ApplyServer) error {
	user := MustGetUser(server.Context())
	ch := make(chan struct{}, 1)
	t := websocket.Type_ApplyProject
	job := socket.NewJober(&websocket.ProjectInput{
		Type:            t,
		NamespaceId:     input.NamespaceId,
		Name:            input.Name,
		GitlabProjectId: input.GitlabProjectId,
		GitlabBranch:    input.GitlabBranch,
		GitlabCommit:    input.GitlabCommit,
		Config:          input.Config,
		Atomic:          input.Atomic,
	}, *user, "", &messager{
		slugName: utils.GetSlugName(input.NamespaceId, input.Name),
		t:        t,
		server:   server,
	}, &plugins.EmptyPubSub{})

	go func() {
		select {
		case <-server.Context().Done():
			job.Stop(server.Context().Err())
		case <-ch:
			return
		}
	}()
	socket.InstallProject(job)
	ch <- struct{}{}

	return nil
}

func (p *Project) StreamPodContainerLog(request *project.ProjectPodContainerLogRequest, server project.Project_StreamPodContainerLogServer) error {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return err
	}

	if running, reason := utils.IsPodRunning(projectModel.Namespace.Name, request.Pod); !running {
		return status.Errorf(codes.NotFound, reason)
	}

	var limit int64 = 2000
	logs := app.K8sClientSet().CoreV1().Pods(projectModel.Namespace.Name).GetLogs(request.Pod, &v1.PodLogOptions{
		Follow:    true,
		Container: request.Container,
		TailLines: &limit,
	})
	stream, _ := logs.Stream(context.TODO())
	bf := bufio.NewReader(stream)

	ch := make(chan []byte)
	go func() {
		defer mlog.Debug("[Stream]:  read exit!")
		for {
			bytes, err := bf.ReadBytes('\n')
			if err != nil {
				mlog.Debugf("[Stream]: %v", err)
				close(ch)
				return
			}
			ch <- bytes
		}
	}()

	for {
		select {
		case <-app.App().Done():
			stream.Close()
			err := errors.New("server shutdown")
			mlog.Debug("[Stream]: client exit with: ", err)
			return err
		case <-server.Context().Done():
			stream.Close()
			mlog.Debug("[Stream]: client exit with: ", server.Context().Err())
			return server.Context().Err()
		case msg, ok := <-ch:
			if !ok {
				stream.Close()
				return errors.New("[Stream]: channel close")
			}

			if err := server.Send(&project.ProjectPodContainerLogResponse{
				Data: &project.ProjectPodLog{
					Namespace:     projectModel.Namespace.Name,
					PodName:       request.Pod,
					ContainerName: request.Container,
					Log:           string(msg),
				},
			}); err != nil {
				stream.Close()
				return err
			}
		}
	}
}

func (p *Project) IsPodRunning(_ context.Context, request *project.ProjectIsPodRunningRequest) (*project.ProjectIsPodRunningResponse, error) {
	running, reason := utils.IsPodRunning(request.GetNamespace(), request.GetPod())

	return &project.ProjectIsPodRunningResponse{Running: running, Reason: reason}, nil
}

func (p *Project) Delete(ctx context.Context, request *project.ProjectDeleteRequest) (*project.ProjectDeleteResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}
	if err := utils.UninstallRelease(projectModel.Name, projectModel.Namespace.Name, mlog.Debugf); err != nil {
		mlog.Error(err)
	}
	app.DB().Delete(&projectModel)
	app.Event().Dispatch(events.EventProjectDeleted, map[string]interface{}{"data": &projectModel})

	AuditLog(MustGetUser(ctx).Name, event.ActionType_Delete,
		fmt.Sprintf("删除项目: %d: %s/%s ", projectModel.ID, projectModel.Namespace.Name, projectModel.Name))

	return &project.ProjectDeleteResponse{}, nil
}

func (p *Project) Show(ctx context.Context, request *project.ProjectShowRequest) (*project.ProjectShowResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	cpu, memory := utils.GetCpuAndMemory(projectModel.GetAllPodMetrics())
	commit, err := plugins.GetGitServer().GetCommit(fmt.Sprintf("%d", projectModel.GitlabProjectId), projectModel.GitlabCommit)
	if err != nil {
		mlog.Error(err)
		return nil, err
	}

	nodePortMapping := utils.GetNodePortMappingByNamespace(projectModel.Namespace.Name)
	ingMapping := utils.GetIngressMappingByNamespace(projectModel.Namespace.Name)

	var urls = make([]string, 0)
	for key, values := range ingMapping {
		if projectModel.Name == key {
			urls = append(urls, values...)
		}
	}
	for key, values := range nodePortMapping {
		if projectModel.Name == key {
			urls = append(urls, values...)
		}
	}

	return &project.ProjectShowResponse{
		Id:                 int64(projectModel.ID),
		Name:               projectModel.Name,
		GitlabProjectId:    int64(projectModel.GitlabProjectId),
		GitlabBranch:       projectModel.GitlabBranch,
		GitlabCommit:       projectModel.GitlabCommit,
		Config:             projectModel.Config,
		DockerImage:        projectModel.DockerImage,
		Atomic:             projectModel.Atomic,
		GitlabCommitWebUrl: commit.GetWebURL(),
		GitlabCommitTitle:  commit.GetTitle(),
		GitlabCommitAuthor: commit.GetAuthorName(),
		GitlabCommitDate:   utils.ToHumanizeDatetimeString(commit.GetCreatedAt()),
		Urls:               urls,
		Namespace: &project.ProjectShowResponse_Namespace{
			Id:   int64(projectModel.NamespaceId),
			Name: projectModel.Namespace.Name,
		},
		Cpu:               cpu,
		Memory:            memory,
		OverrideValues:    projectModel.OverrideValues,
		CreatedAt:         utils.ToRFC3339DatetimeString(&projectModel.CreatedAt),
		UpdatedAt:         utils.ToRFC3339DatetimeString(&projectModel.UpdatedAt),
		HumanizeCreatedAt: utils.ToHumanizeDatetimeString(&projectModel.CreatedAt),
		HumanizeUpdatedAt: utils.ToHumanizeDatetimeString(&projectModel.CreatedAt),
	}, nil
}

func (p *Project) AllPodContainers(ctx context.Context, request *project.ProjectAllPodContainersRequest) (*project.ProjectAllPodContainersResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}

	var list = projectModel.GetAllPods()

	var containerList []*project.ProjectPodLog
	for _, item := range list {
		for _, c := range item.Spec.Containers {
			containerList = append(containerList,
				&project.ProjectPodLog{
					Namespace:     projectModel.Namespace.Name,
					PodName:       item.Name,
					ContainerName: c.Name,
					Log:           "",
				},
			)
		}
	}

	return &project.ProjectAllPodContainersResponse{Data: containerList}, nil
}

func (p *Project) PodContainerLog(ctx context.Context, request *project.ProjectPodContainerLogRequest) (*project.ProjectPodContainerLogResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}

	if running, reason := utils.IsPodRunning(projectModel.Namespace.Name, request.Pod); !running {
		return nil, status.Errorf(codes.NotFound, reason)
	}

	var limit int64 = 2000
	logs := app.K8sClientSet().CoreV1().Pods(projectModel.Namespace.Name).GetLogs(request.Pod, &v1.PodLogOptions{
		Container: request.Container,
		TailLines: &limit,
	})
	var raw = []byte("未找到日志")
	do := logs.Do(context.Background())
	raw, err := do.Raw()
	if err != nil {
		if status, ok := err.(apierrors.APIStatus); ok {
			if status.Status().Code == http.StatusBadRequest {
				mlog.Warningf("CleanEvictedPods code: %d message: %s", status.Status().Code, status.Status().Reason)
				for _, selector := range projectModel.GetPodSelectors() {
					utils.CleanEvictedPods(projectModel.Namespace.Name, selector)
				}
			}
		}
		return nil, err
	}

	return &project.ProjectPodContainerLogResponse{
		Data: &project.ProjectPodLog{
			Namespace:     projectModel.Namespace.Name,
			PodName:       request.Pod,
			ContainerName: request.Container,
			Log:           string(raw),
		},
	}, nil
}

type messager struct {
	slugName string
	t        websocket.Type
	server   project.Project_ApplyServer
}

func (m *messager) SendDeployedResult(resultType websocket.ResultType, s string, p *models.Project) {
	var ns model.NamespaceModel
	if p.Namespace.ID != 0 {
		ns = model.NamespaceModel{
			Id:               int64(p.Namespace.ID),
			Name:             p.Namespace.Name,
			ImagePullSecrets: p.Namespace.ImagePullSecretsArray(),
			CreatedAt:        utils.ToRFC3339DatetimeString(&p.Namespace.CreatedAt),
			UpdatedAt:        utils.ToRFC3339DatetimeString(&p.Namespace.UpdatedAt),
		}
	}
	m.server.Send(&project.ProjectApplyResponse{
		Metadata: &websocket.Metadata{
			Slug:   m.slugName,
			Type:   m.t,
			Result: resultType,
			End:    true,
			Data:   s,
		},
		Project: &model.ProjectModel{
			Id:              int64(p.ID),
			Name:            p.Name,
			GitlabProjectId: int64(p.GitlabProjectId),
			GitlabBranch:    p.GitlabBranch,
			GitlabCommit:    p.GitlabCommit,
			Config:          p.Config,
			OverrideValues:  p.OverrideValues,
			DockerImage:     p.DockerImage,
			PodSelectors:    p.PodSelectors,
			NamespaceId:     int64(p.NamespaceId),
			Atomic:          p.Atomic,
			CreatedAt:       utils.ToRFC3339DatetimeString(&p.CreatedAt),
			UpdatedAt:       utils.ToRFC3339DatetimeString(&p.UpdatedAt),
			Namespace:       &ns,
		},
	})
}

func (m *messager) SendEndError(err error) {
	m.server.Send(&project.ProjectApplyResponse{Metadata: &websocket.Metadata{
		Slug:   m.slugName,
		Type:   m.t,
		Result: websocket.ResultType_Error,
		End:    true,
		Data:   err.Error(),
	}})
}

func (m *messager) SendError(err error) {
	m.server.Send(&project.ProjectApplyResponse{Metadata: &websocket.Metadata{
		Slug:   m.slugName,
		Type:   m.t,
		Result: websocket.ResultType_Error,
		End:    false,
		Data:   err.Error(),
	}})
}

func (m *messager) SendProcessPercent(s string) {
	// 不需要
}

func (m *messager) SendMsg(s string) {
	m.server.Send(&project.ProjectApplyResponse{Metadata: &websocket.Metadata{
		Slug:   m.slugName,
		Type:   m.t,
		Result: websocket.ResultType_Success,
		End:    false,
		Data:   s,
	}})
}

func (m *messager) SendProtoMsg(message plugins.WebsocketMessage) {
	m.server.Send(&project.ProjectApplyResponse{Metadata: message.GetMetadata()})
}
