package cmd

import (
	"context"
	"github.com/DuC-cnZj/mars/pkg/app"
	"github.com/DuC-cnZj/mars/pkg/config"
	"github.com/DuC-cnZj/mars/pkg/mlog"
	"github.com/DuC-cnZj/mars/pkg/models"
	"github.com/DuC-cnZj/mars/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.Init(cfgFile)
		c.Debug = true
		app := app.NewApplication(c)
		if err := app.Bootstrap(); err != nil {
			mlog.Fatal(err)
		}
		var project models.Project
		utils.DB().Preload("Namespace").Where("`id` = ?", "19").First(&project)

		list, _ := utils.K8sClientSet().CoreV1().Pods(project.Namespace.Name).List(context.Background(), v1.ListOptions{
			LabelSelector: "app.kubernetes.io/instance=" + project.Name,
		})

		type PodContainer struct {
			Pod       v12.Pod
			Container v12.Container
		}
		var containerList []PodContainer
		for _, item := range list.Items {
			for _, container := range item.Spec.Containers {
				containerList = append(containerList, PodContainer{
					Pod:       item,
					Container: container,
				})
			}
		}

		if len(containerList) > 0 {
			//var l int64 = 10000
			first := containerList[0]
			var ss int64 = 5
			logs := utils.K8sClientSet().CoreV1().Pods(project.Namespace.Name).GetLogs(first.Pod.Name, &v12.PodLogOptions{
				Container: first.Container.Name,
				SinceSeconds: &ss,
				//TailLines: &l,
			})
			s, err := logs.Stream(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			b := make([]byte, 200)
			for {
				_, err := s.Read(b)
				if err == io.EOF {
					mlog.Fatal("EOF")
				}
				mlog.Warning(string(b))
				time.Sleep(1*time.Second)
			}
		}

		app.Shutdown()
	},
}
