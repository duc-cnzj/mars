package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
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
			table.Append([]string{fmt.Sprintf("%d", i+1), name, tags})
		}
		table.Render()
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
			table.Append([]string{fmt.Sprintf("%d", i+1), command.Name(), command.Expression()})
		}

		table.Render()
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
			table.Append([]string{fmt.Sprintf("%d", i), event.String(), strings.Join(listenerNames, " "), fmt.Sprintf("%d", len(listeners))})
		}

		table.Render()
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

// inspectTable 是 inspect 子命令共用的终端表格：表头居中大写、数据左对齐、
// 行与行之间以制表线分隔，整体用制表符（Unicode 方框字符）绘制。
type inspectTable struct {
	header []string
	rows   [][]string
	out    io.Writer
}

// newInspectTable 创建一张渲染到标准输出的表格。
func newInspectTable(header []string) *inspectTable {
	return &inspectTable{header: header, out: os.Stdout}
}

// Append 追加一行数据；列数少于表头时缺列按空串渲染。
func (t *inspectTable) Append(row []string) {
	t.rows = append(t.rows, row)
}

// Render 计算各列宽度并输出整张表格。
// 无数据行时只输出表头，不输出行分隔线。
func (t *inspectTable) Render() {
	headers := make([]string, len(t.header))
	widths := make([]int, len(t.header))
	for i, h := range t.header {
		headers[i] = strings.ToUpper(h)
		widths[i] = displayWidth(headers[i])
	}
	for _, row := range t.rows {
		for i := range headers {
			if i < len(row) {
				widths[i] = max(widths[i], displayWidth(row[i]))
			}
		}
	}

	var sb strings.Builder
	writeBorder(&sb, "┌", "┬", "┐", widths)
	writeCells(&sb, headers, widths, true)
	for _, row := range t.rows {
		writeBorder(&sb, "├", "┼", "┤", widths)
		writeCells(&sb, row, widths, false)
	}
	writeBorder(&sb, "└", "┴", "┘", widths)

	_, _ = io.WriteString(t.out, sb.String())
}

// writeBorder 渲染一条水平边框：每个单元格位置重复 width+2 个横线，
// 与单元格左右各一个空格的内边距对齐。
func writeBorder(sb *strings.Builder, left, join, right string, widths []int) {
	sb.WriteString(left)
	for i, w := range widths {
		if i > 0 {
			sb.WriteString(join)
		}
		sb.WriteString(strings.Repeat("─", w+2))
	}
	sb.WriteString(right)
	sb.WriteByte('\n')
}

// writeCells 渲染一行单元格：center 为真时表头居中（余数偏向右侧），否则数据左对齐。
func writeCells(sb *strings.Builder, cells []string, widths []int, center bool) {
	sb.WriteString("│")
	for i, w := range widths {
		var cell string
		if i < len(cells) {
			cell = cells[i]
		}
		pad := w - displayWidth(cell)
		sb.WriteByte(' ')
		if center {
			left := pad / 2
			sb.WriteString(strings.Repeat(" ", left))
			sb.WriteString(cell)
			sb.WriteString(strings.Repeat(" ", pad-left))
		} else {
			sb.WriteString(cell)
			sb.WriteString(strings.Repeat(" ", pad))
		}
		sb.WriteString(" │")
	}
	sb.WriteByte('\n')
}

// displayWidth 返回文本的终端显示宽度。
// inspect 表格的数据源都是 ASCII（数字 ID、Go 标识符、配置里的插件名、事件名）
// 以及固定两 rune 的 "⭐︎" 标记，其 rune 数与显示宽度一致，故直接按 rune 计数。
func displayWidth(s string) int {
	return utf8.RuneCountInString(s)
}
