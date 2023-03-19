package cmd

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/duc-cnzj/mars/v4/internal/adapter"
	"github.com/duc-cnzj/mars/v4/internal/app"
	"github.com/duc-cnzj/mars/v4/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/plugins"

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

		for i, boot := range ServerBootstrappers {
			s := strings.Split(reflect.TypeOf(boot).String(), ".")
			name := s[len(s)-1]
			tags := strings.Join(boot.Tags(), ",")
			table.Append([]string{fmt.Sprintf("%d", i+1), name, tags})
		}
		table.Render()
	},
}

type loggerBootstrapper struct{}

func (l *loggerBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	mlog.SetLogger(adapter.NewEmptyLogger())
	return nil
}

func (l *loggerBootstrapper) Tags() []string {
	return []string{}
}

var showCronJobsCmd = &cobra.Command{
	Use:   "cronjobs",
	Short: "app cron jobs.",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(config.Init(cfgFile), app.WithMustBootedBootstrappers(&loggerBootstrapper{}))
		cm := cron.NewManager(nil, app)
		for _, callback := range cron.RegisteredCronJobs() {
			callback(cm, app)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		table.SetHeader([]string{"ID", "Name", "Expression"})
		for i, command := range cm.List() {
			table.Append([]string{fmt.Sprintf("%d", i+1), command.Name(), command.Expression()})
		}

		table.Render()
	},
}

var showEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "app events.",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(config.Init(cfgFile),
			app.WithMustBootedBootstrappers(&loggerBootstrapper{}),
			app.WithBootstrappers(&bootstrappers.EventBootstrapper{}))
		app.Bootstrap()
		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		table.SetHeader([]string{"ID", "Event Name", "Listener Names", "Listener Count"})
		i := 0
		for event, listeners := range events.RegisteredEvents() {
			i++
			var listenerNames []string
			for _, listener := range listeners {
				s := strings.Split(GetFunctionName(listener), ".")
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

		var others [][]string
		i := 0
		for name := range plugins.GetPlugins() {
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
		table := tablewriter.NewWriter(os.Stdout)
		table.SetColWidth(200)
		table.SetRowLine(true)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"ID", "Key", "Value"})
		rv := reflect.ValueOf(cfg).Elem()
		rt := reflect.TypeOf(cfg).Elem()
		for i := 0; i < rv.NumField(); i++ {
			fieldName := rt.Field(i).Name
			filedValue := fmt.Sprintf("%v", rv.Field(i).Interface())
			if fieldName == "PrivateKey" {
				filedValue = "MASKED"
			}
			table.Append([]string{fmt.Sprintf("%d", i+1), fieldName, filedValue})
		}

		table.Render()
	},
}

func GetFunctionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
