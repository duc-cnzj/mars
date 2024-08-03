package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/utils/runtime"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	showCmd.AddCommand(showBootTagsCmd)
	showCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $DIR/config.yaml)")
	viper.BindPFlag("config", showCmd.PersistentFlags().Lookup("config"))

	showCmd.AddCommand(showAllCmd)
	showCmd.AddCommand(showCronJobsCmd)
	showCmd.AddCommand(showEventsCmd)
	showCmd.AddCommand(showPluginsCmd)
	showCmd.AddCommand(showConfigCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show app info.",
}

var showAllCmd = &cobra.Command{
	Use:   "all",
	Short: "all app info.",
	Run: func(cmd *cobra.Command, args []string) {
		for _, command := range showCmd.Commands() {
			if command.Use != "all" {
				fmt.Println(command.Short)
				command.Run(cmd, args)
			}
		}
	},
}

var showBootTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "app boot tags.",
	Run: func(cmd *cobra.Command, args []string) {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Tags"})
		table.SetRowLine(true)

		for i, boot := range serverBootstrappers {
			s := strings.Split(reflect.TypeOf(boot).String(), ".")
			name := s[len(s)-1]
			tags := strings.Join(boot.Tags(), ",")
			table.Append([]string{fmt.Sprintf("%d", i+1), name, tags})
		}
		table.Render()
	},
}

type loggerBootstrapper struct{}

// Bootstrap boot logger.
func (l *loggerBootstrapper) Bootstrap(app application.App) error {
	return nil
}

// Tags boot tags.
func (l *loggerBootstrapper) Tags() []string {
	return []string{}
}

var showCronJobsCmd = &cobra.Command{
	Use:   "cronjobs",
	Short: "app cron jobs.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(cfgFile)
		cfg.LogChannel = ""
		logger := mlog.NewLogger(cfg)
		app, clean, err := InitializeApp(cfg, logger, nil)
		if err != nil {
			logger.Fatal(err)
		}
		app.RegisterAfterShutdownFunc(func(application.App) { clean() })
		defer app.Shutdown()

		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		table.SetHeader([]string{"ID", "Name", "Expression"})
		for i, command := range app.CronManager().List() {
			table.Append([]string{fmt.Sprintf("%d", i+1), command.Name(), command.Expression()})
		}

		table.Render()
	},
}

var showEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "app events.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(cfgFile)
		cfg.LogChannel = ""
		logger := mlog.NewLogger(cfg)
		app, clean, err := InitializeApp(cfg, logger, []application.Bootstrapper{})
		if err != nil {
			logger.Fatal(err)
		}
		app.RegisterAfterShutdownFunc(func(application.App) { clean() })
		defer app.Shutdown()

		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		table.SetHeader([]string{"ID", "Event Name", "Listener Names", "Listener Count"})
		i := 0
		for event, listeners := range app.Dispatcher().List() {
			i++
			var listenerNames []string
			for _, listener := range listeners {
				s := strings.Split(runtime.GetFunctionName(listener), ".")
				listenerNames = append(listenerNames, s[len(s)-1])
			}
			table.Append([]string{fmt.Sprintf("%d", i), event.String(), strings.Join(listenerNames, " "), fmt.Sprintf("%d", len(listeners))})
		}

		table.Render()
	},
}

var showPluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "app plugins.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(cfgFile)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		table.SetHeader([]string{"ID", "Plugin", "Current"})

		usedPlugins := []string{
			cfg.PicturePlugin.Name,
			cfg.WsSenderPlugin.Name,
			cfg.DomainManagerPlugin.Name,
			cfg.GitServerPlugin.Name,
		}

		cfg.LogChannel = ""
		logger := mlog.NewLogger(cfg)
		app, clean, err := InitializeApp(cfg, logger, []application.Bootstrapper{})
		if err != nil {
			logger.Fatal(err)
		}
		app.RegisterAfterShutdownFunc(func(application.App) { clean() })
		defer app.Shutdown()

		var others [][]string
		i := 0
		for name := range app.PluginMgr().GetPlugins() {
			i++
			used := false
			for _, plugin := range usedPlugins {
				if name == plugin {
					used = true
					break
				}
			}
			if used {
				table.Append([]string{fmt.Sprintf("%d", i), name, "⭐︎"})
			} else {
				others = append(others, []string{fmt.Sprintf("%d", i), name, ""})
			}
		}
		for _, other := range others {
			table.Append(other)
		}

		table.Render()
	},
}

var showConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "app config.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(cfgFile)
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
