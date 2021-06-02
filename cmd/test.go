package cmd

import (
	"context"
	"fmt"

	"github.com/DuC-cnZj/mars/pkg/app"
	"github.com/DuC-cnZj/mars/pkg/config"
	"github.com/DuC-cnZj/mars/pkg/mlog"
	"github.com/DuC-cnZj/mars/pkg/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		var m = map[string][]string{}

		list, _ := utils.K8sClientSet().NetworkingV1().Ingresses("devops-aaa").List(context.Background(), metav1.ListOptions{})
		for _, item := range list.Items {
			for _, tls := range item.Spec.TLS {
				if projectName, ok := item.Labels["app.kubernetes.io/instance"]; ok {
					data := m[projectName]
					for _, host := range tls.Hosts {
						m[projectName] = append(data, fmt.Sprintf("https://%s", host))
					}
				}
			}
		}

		mlog.Warningf("%#v", m)

		app.Shutdown()
	},
}
