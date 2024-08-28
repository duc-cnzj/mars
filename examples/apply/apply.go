package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/duc-cnzj/mars/api/v5"
	"github.com/duc-cnzj/mars/api/v5/project"
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/samber/lo"
)

func main() {
	client, _ := api.NewClient("localhost:50000", api.WithAuth("admin", "123456"))
	defer client.Close()
	// 创建一个 app
	model, err := create(client, "app1")
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(15 * time.Second)
	log.Println("####### 开始更新 app 的副本数量 1 -> 2 #######")
	// 更新 app 的副本数量 1 -> 2
	update(client, int(model.Id))
}

func create(client api.Interface, name string) (*types.ProjectModel, error) {
	input := &project.ApplyRequest{
		// 项目空间 id, 鼠标移动到项目空间名称上会显示 id
		NamespaceId: 20,
		// 仓库 id, 后台仓库管理页面可见
		RepoId: 107,
		// helm 的 atomic 模式, 如果为 true, 会等到部署成功再结束
		Atomic: false,
		// 是否 websocket 通知其他用户刷新页面
		WebsocketSync: true,
		SendPercent:   false,
		// optional，不传使用 repo 的 name，此能力只有接口有，页面没开放
		Name: name,
		// 覆盖自定义值
		ExtraValues: []*websocket.ExtraValue{
			{
				Path:  "replicaCount",
				Value: "1",
			},
		},
	}
	apply, err := client.Project().Apply(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	for {
		recv, err := apply.Recv()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		if recv.Metadata.End {
			fmt.Println(recv.Metadata.Message, recv.Project)
			return recv.Project, nil
		}
		if input.SendPercent {
			fmt.Println(recv.Metadata.Percent)
		}
		fmt.Println(recv.Metadata.Message)
	}
}

func update(client api.Interface, id int) {
	show, err := client.Project().Show(context.TODO(), &project.ShowRequest{Id: int32(id)})
	if err != nil {
		log.Fatal(err)
	}
	input := &project.ApplyRequest{
		// 项目空间 id, 鼠标移动到项目空间名称上会显示 id
		NamespaceId: show.Item.NamespaceId,
		// 仓库 id, 后台仓库管理页面可见
		RepoId: show.Item.RepoId,
		// helm 的 atomic 模式, 如果为 true, 会等到部署成功再结束
		Atomic: false,
		// 是否 websocket 通知其他用户刷新页面
		WebsocketSync: true,
		SendPercent:   false,
		// optional，不传使用 repo 的 name，此能力只有接口有，页面没开放
		Name: show.Item.Name,
		// 覆盖自定义值
		ExtraValues: []*websocket.ExtraValue{
			{
				Path:  "replicaCount",
				Value: "2",
			},
		},
		Version: lo.ToPtr(int32(show.Item.Version)),
	}
	apply, err := client.Project().Apply(context.TODO(), input)
	if err != nil {
		return
	}

	for {
		recv, err := apply.Recv()
		if err != nil {
			fmt.Println(err)
			return
		}
		if recv.Metadata.End {
			fmt.Println(recv.Metadata.Message, recv.Project)
			return
		}
		if input.SendPercent {
			fmt.Println(recv.Metadata.Percent)
		}
		fmt.Println(recv.Metadata.Message)
	}
}
