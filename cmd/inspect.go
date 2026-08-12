package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	inspect.AddCommand(inspectBootTagsCmd)
	inspect.AddCommand(inspectAllCmd)
	inspect.AddCommand(inspectCronJobsCmd)
	inspect.AddCommand(inspectEventsCmd)
	inspect.AddCommand(inspectPluginsCmd)
	inspect.AddCommand(inspectConfigCmd)
}

var inspect = &cobra.Command{
	Use:   "inspect",
	Short: "inspect app info.",
}

var inspectAllCmd = &cobra.Command{
	Use:   "all",
	Short: "all app info.",
	Run: func(cmd *cobra.Command, args []string) {
		for _, command := range inspect.Commands() {
			if command.Use != "all" {
				fmt.Println(command.Short)
				command.Run(cmd, args)
			}
		}
	},
}

var inspectBootTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "app boot tags.",
	Run: func(cmd *cobra.Command, args []string) {
		table := newInspectTable([]string{"ID", "Name", "Tags"})

		for i, boot := range serverBootstrappers {
			s := strings.Split(reflect.TypeOf(boot).String(), ".")
			name := s[len(s)-1]
			tags := strings.Join(boot.Tags(), ",")
			_ = table.Append([]string{fmt.Sprintf("%d", i+1), name, tags})
		}
		_ = table.Render()
	},
}

var inspectCronJobsCmd = &cobra.Command{
	Use:     "cronjobs",
	Aliases: []string{"cronjob", "cron", "job", "jobs", "cj"},
	Short:   "app cron jobs.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(viper.GetString("config"))
		cfg.LogChannel = ""
		logger := mlog.NewForConfig(cfg)
		app, err := InitializeApp(cfg, logger, nil)
		if err != nil {
			logger.Fatal(err)
		}
		// Flush 注册于 Shutdown 之前：defer LIFO 先执行 Shutdown，再冲刷日志，关闭期日志不丢。
		defer logger.Flush()
		defer app.Shutdown()

		table := newInspectTable([]string{"ID", "Name", "Expression"})
		for i, command := range app.CronManager().List() {
			_ = table.Append([]string{fmt.Sprintf("%d", i+1), command.Name(), command.Expression()})
		}

		_ = table.Render()
	},
}

var inspectEventsCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"event", "ev"},
	Short:   "app events.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(viper.GetString("config"))
		cfg.LogChannel = ""
		logger := mlog.NewForConfig(cfg)
		app, err := InitializeApp(cfg, logger, []app.Bootstrapper{})
		if err != nil {
			logger.Fatal(err)
		}
		// Flush 注册于 Shutdown 之前：defer LIFO 先执行 Shutdown，再冲刷日志，关闭期日志不丢。
		defer logger.Flush()
		defer app.Shutdown()

		table := newInspectTable([]string{"ID", "Event Name", "Listener Names", "Listener Count"})
		i := 0
		for event, listeners := range app.Dispatcher().List() {
			i++
			var listenerNames []string
			for _, listener := range listeners {
				s := strings.Split(GetFunctionName(listener), ".")
				listenerNames = append(listenerNames, s[len(s)-1])
			}
			_ = table.Append([]string{fmt.Sprintf("%d", i), event.String(), strings.Join(listenerNames, " "), fmt.Sprintf("%d", len(listeners))})
		}

		_ = table.Render()
	},
}

var inspectPluginsCmd = &cobra.Command{
	Use:     "plugins",
	Aliases: []string{"plugin"},
	Short:   "app plugins.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(viper.GetString("config"))
		table := newInspectTable([]string{"ID", "Plugin", "Current"})

		usedPlugins := []string{
			cfg.PicturePlugin.Name,
			cfg.WsSenderPlugin.Name,
			cfg.DomainManagerPlugin.Name,
			cfg.GitServerPlugin.Name,
		}

		cfg.LogChannel = ""
		logger := mlog.NewForConfig(cfg)
		app, err := InitializeApp(cfg, logger, []app.Bootstrapper{})
		if err != nil {
			logger.Fatal(err)
		}
		// Flush 注册于 Shutdown 之前：defer LIFO 先执行 Shutdown，再冲刷日志，关闭期日志不丢。
		defer logger.Flush()
		defer app.Shutdown()

		var others [][]string
		i := 0
		for name := range app.PluginManager().GetPlugins() {
			i++
			used := false
			for _, plugin := range usedPlugins {
				if name == plugin {
					used = true
					break
				}
			}
			if used {
				_ = table.Append([]string{fmt.Sprintf("%d", i), name, "⭐︎"})
			} else {
				others = append(others, []string{fmt.Sprintf("%d", i), name, ""})
			}
		}
		for _, other := range others {
			_ = table.Append(other)
		}

		_ = table.Render()
	},
}

var inspectConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg", "conf"},
	Short:   "app config.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(viper.GetString("config"))
		var c = struct {
			*config.Config
			InstallTimeout string
		}{
			Config:         cfg,
			InstallTimeout: cfg.InstallTimeout.String(),
		}
		indent, _ := json.MarshalIndent(c, "", "  ")
		fmt.Println(string(indent))
	},
}

// GetFunctionName 返回函数指针 i 对应的函数名。
func GetFunctionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// newInspectTable 创建带表头与行分隔线的 CLI 输出表格。
// tablewriter v1.x 用 builder 配置取代了 v0.0.5 的 SetHeader/SetRowLine，
// 此处收敛为单一出口，4 个 inspect 子命令共用。
func newInspectTable(header []string) *tablewriter.Table {
	return tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader(header),
		tablewriter.WithRendition(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.Success},
			},
		}),
	)
}
