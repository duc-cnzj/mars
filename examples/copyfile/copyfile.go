package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/duc-cnzj/mars/api/v5"
	"github.com/duc-cnzj/mars/api/v5/container"
)

func main() {
	c, _ := api.NewClient("localhost:50000", api.WithAuth("admin", "123456"))
	defer c.Close()
	cp, _ := c.Container().StreamCopyToPod(context.TODO())
	open, _ := os.Open("/Users/duc/Downloads/ducc.xlsx")
	defer open.Close()
	bf := bufio.NewReaderSize(open, 1024*1024*5)
	var (
		filename = open.Name()
		pod      = "nginx-54bff68475-k69gh"
		// 在有多个容器的情况下，不传会拿默认的容器
		//containerName = "demo"
		namespace = "devops-duc"
	)
	for {
		bts := make([]byte, 1024*1024)
		n, err := bf.Read(bts)
		if err != nil {
			if err == io.EOF {
				cp.Send(&container.StreamCopyToPodRequest{
					FileName:  filename,
					Data:      bts[0:n],
					Namespace: namespace,
					Pod:       pod,
					//Container: containerName,
				})
				recv, err := cp.CloseAndRecv()
				if err != nil {
					log.Fatal(err)
				}
				indent, _ := json.MarshalIndent(recv, "", "  ")
				fmt.Printf("上传成功\n%s", string(indent))
			}
			return
		}
		cp.Send(&container.StreamCopyToPodRequest{
			FileName:  filename,
			Data:      bts[0:n],
			Namespace: namespace,
			Pod:       pod,
		})
	}
}
