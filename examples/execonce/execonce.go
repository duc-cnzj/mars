package main

import (
	"context"
	"fmt"
	"log"

	"github.com/duc-cnzj/mars/api/v5"
	"github.com/duc-cnzj/mars/api/v5/container"
)

func main() {
	client, _ := api.NewClient("localhost:50000", api.WithAuth("admin", "123456"))
	defer client.Close()
	ns := "devops-duc"
	pod := "nginx-54bff68475-k69gh"
	exec, err := client.Container().ExecOnce(context.TODO(), &container.ExecOnceRequest{
		Namespace: ns,
		Pod:       pod,
		Command:   []string{"bash", "-c", "echo 'hello world'"},
	})
	if err != nil {
		log.Println(err)
		return
	}
	for {
		recv, err := exec.Recv()
		if err != nil {
			return
		}
		fmt.Print(string(recv.Message))
	}

	exec.CloseSend()
}
