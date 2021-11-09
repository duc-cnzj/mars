package services

import (
	"bufio"
	"context"
	"errors"
	"net/http"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/project"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type Project struct {
	project.UnimplementedProjectServer
}

func (p *Project) StreamPodContainerLog(request *project.PodContainerLogRequest, server project.Project_StreamPodContainerLogServer) error {
	var projectModal models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModal).Error; err != nil {
		return err
	}

	if running, reason := utils.IsPodRunning(projectModal.Namespace.Name, request.Pod); !running {
		return status.Errorf(codes.NotFound, reason)
	}

	var limit int64 = 2000
	logs := app.K8sClientSet().CoreV1().Pods(projectModal.Namespace.Name).GetLogs(request.Pod, &v1.PodLogOptions{
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

			if err := server.Send(&project.PodContainerLogResponse{
				Data: &project.PodLog{
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

func (p *Project) IsPodRunning(_ context.Context, request *project.IsPodRunningRequest) (*project.IsPodRunningResponse, error) {
	running, reason := utils.IsPodRunning(request.GetNamespace(), request.GetPod())

	return &project.IsPodRunningResponse{Running: running, Reason: reason}, nil
}

func (p *Project) Destroy(ctx context.Context, request *project.ProjectDestroyRequest) (*emptypb.Empty, error) {
	var projectModal models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModal).Error; err != nil {
		return nil, err
	}
	if err := utils.UninstallRelease(projectModal.Name, projectModal.Namespace.Name, mlog.Debugf); err != nil {
		mlog.Error(err)
	}
	app.DB().Delete(&projectModal)
	app.Event().Dispatch(events.EventProjectedDeleted, map[string]interface{}{"data": &projectModal})

	return &emptypb.Empty{}, nil
}

func (p *Project) Show(ctx context.Context, request *project.ProjectShowRequest) (*project.ProjectShowResponse, error) {
	var projectModal models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModal).Error; err != nil {
		return nil, err
	}
	cpu, memory := utils.GetCpuAndMemory(projectModal.GetAllPodMetrics())
	commit, _, err := app.GitlabClient().Commits.GetCommit(projectModal.GitlabProjectId, projectModal.GitlabCommit)
	if err != nil {
		mlog.Error(err)
		return nil, err
	}

	nodePortMapping := utils.GetNodePortMappingByNamespace(projectModal.Namespace.Name)
	ingMapping := utils.GetIngressMappingByNamespace(projectModal.Namespace.Name)

	var urls = make([]string, 0)
	for key, values := range ingMapping {
		if projectModal.Name == key {
			urls = append(urls, values...)
		}
	}
	for key, values := range nodePortMapping {
		if projectModal.Name == key {
			urls = append(urls, values...)
		}
	}

	return &project.ProjectShowResponse{
		Id:                 int64(projectModal.ID),
		Name:               projectModal.Name,
		GitlabProjectId:    int64(projectModal.GitlabProjectId),
		GitlabBranch:       projectModal.GitlabBranch,
		GitlabCommit:       projectModal.GitlabCommit,
		Config:             projectModal.Config,
		DockerImage:        projectModal.DockerImage,
		Atomic:             projectModal.Atomic,
		GitlabCommitWebUrl: commit.WebURL,
		GitlabCommitTitle:  commit.Title,
		GitlabCommitAuthor: commit.AuthorName,
		GitlabCommitDate:   utils.ToHumanizeDatetimeString(commit.CreatedAt),
		Urls:               urls,
		Namespace: &project.ProjectShowResponse_Namespace{
			Id:   int64(projectModal.NamespaceId),
			Name: projectModal.Namespace.Name,
		},
		Cpu:            cpu,
		Memory:         memory,
		OverrideValues: projectModal.OverrideValues,
		CreatedAt:      utils.ToHumanizeDatetimeString(&projectModal.CreatedAt),
		UpdatedAt:      utils.ToHumanizeDatetimeString(&projectModal.UpdatedAt),
	}, nil
}

func (p *Project) AllPodContainers(ctx context.Context, request *project.AllPodContainersRequest) (*project.AllPodContainersResponse, error) {
	var projectModal models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModal).Error; err != nil {
		return nil, err
	}

	var list = projectModal.GetAllPods()

	var containerList []*project.PodLog
	for _, item := range list {
		for _, c := range item.Spec.Containers {
			containerList = append(containerList,
				&project.PodLog{
					PodName:       item.Name,
					ContainerName: c.Name,
					Log:           "",
				},
			)
		}
	}

	return &project.AllPodContainersResponse{Data: containerList}, nil
}

func (p *Project) PodContainerLog(ctx context.Context, request *project.PodContainerLogRequest) (*project.PodContainerLogResponse, error) {
	var projectModal models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModal).Error; err != nil {
		return nil, err
	}

	if running, reason := utils.IsPodRunning(projectModal.Namespace.Name, request.Pod); !running {
		return nil, status.Errorf(codes.NotFound, reason)
	}

	var limit int64 = 2000
	logs := app.K8sClientSet().CoreV1().Pods(projectModal.Namespace.Name).GetLogs(request.Pod, &v1.PodLogOptions{
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
				for _, selector := range projectModal.GetPodSelectors() {
					utils.CleanEvictedPods(projectModal.Namespace.Name, selector)
				}
			}
		}
		return nil, err
	}

	return &project.PodContainerLogResponse{
		Data: &project.PodLog{
			PodName:       request.Pod,
			ContainerName: request.Container,
			Log:           string(raw),
		},
	}, nil
}
