package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/DuC-cnZj/mars/pkg/app"
	"github.com/DuC-cnZj/mars/pkg/config"
	"github.com/DuC-cnZj/mars/pkg/mlog"
	"github.com/DuC-cnZj/mars/pkg/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(config.Init(cfgFile))
		if err := app.Bootstrap(); err != nil {
			mlog.Fatal(err)
		}
		defer app.Shutdown()
		list, err := utils.K8sMetrics().
			MetricsV1beta1().
			PodMetricses("default").
			List(context.Background(), v1.ListOptions{
				LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s", "xuanji", "xuanji"),
			})
		if err != nil {
			log.Fatal(err)
			return
		}

		// helm 部署必带的 label
		//app.kubernetes.io/name: {{ include "charts.name" . }}
		//app.kubernetes.io/instance: {{ .Release.Name }}

		for _, item := range list.Items {
			log.Println(item.Name)
		}
	},
}
