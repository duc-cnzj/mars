package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/duc-cnzj/mars/api/v5"
	"github.com/duc-cnzj/mars/api/v5/container"
)

func main() {
	client, _ := api.NewClient("localhost:50000", api.WithAuth("admin", "123456"))
	defer client.Close()
	ns := "duc-abc"
	pod := "ng-nginx-594b65865-g975j"
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	exec, err := client.Container().ExecOnce(timeout, &container.ExecOnceRequest{
		Namespace: ns,
		Pod:       pod,
		Command:   []string{"bash", "-c", "echo 'hello world';sleep 5;"},
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer exec.CloseSend()
	for {
		recv, err := exec.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%q", string(recv.Message))
	}
}
