package cronjob

import (
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/cron"
)

// CronTask 一条定时任务的完整声明：名字 + 调度设置器 + 执行体。新增任务只需在
// Registry 里加一行；新增无参调度方式直接复用 cron.Command 方法表达式
// （如 cron.Command.EveryMinute），无需新增任何辅助函数——这是注册表模式的扩展点。
type CronTask struct {
	Name string
	// Schedule 命令调度设置器：对 cmd 产出的命令做链式调度，返回命令自身。
	// 无参调度（EveryMinute/EveryTwoMinutes/...）用 cron.Command.Xxx 方法表达式直接引用；
	// 带参调度（如 DailyAt）用内联 lambda 固定参数。
	Schedule func(cron.Command) cron.Command
	Run      func() error
}

// Registry 数据驱动枚举全部定时任务；cfg 条件任务就地声明，不进 cmd。
// cmd 组合根调用 RegisterCronJobs 时只需对返回的 []CronTask 做机械注册。
func Registry(t *Tasks, cfg *config.Config) []CronTask {
	tasks := []CronTask{
		{Name: "clean_upload_files",
			Schedule: func(cmd cron.Command) cron.Command { return cmd.DailyAt("2:00") },
			Run:      t.CleanUploadFiles},
		{Name: "fix_project_deploy_status", Schedule: cron.Command.EveryTwoMinutes, Run: t.FixDeployStatus},
		{Name: "sync_domain_secret", Schedule: cron.Command.EveryMinute, Run: t.SyncDomainSecret},
		{Name: "disk_info", Schedule: cron.Command.EveryTenMinutes, Run: func() error {
			_, err := t.DiskInfo()
			return err
		}},
	}
	if cfg.GitServerCached {
		tasks = append(tasks,
			CronTask{Name: "all_branch_cache", Schedule: cron.Command.EveryTwoMinutes, Run: t.CacheAllBranches},
			CronTask{Name: "all_project_cache", Schedule: cron.Command.EveryFiveMinutes, Run: t.CacheAllProjects},
		)
	}
	if cfg.IsK8sEnv() {
		tasks = append(tasks,
			CronTask{Name: "sync_image_pull_secrets", Schedule: cron.Command.EveryFiveMinutes, Run: t.SyncImagePullSecrets},
			CronTask{Name: "cache_cluster_board", Schedule: cron.Command.EveryThirtySeconds, Run: t.CacheClusterBoard},
			CronTask{Name: "cache_cluster_info", Schedule: cron.Command.EveryThirtySeconds, Run: t.CacheClusterInfo},
			CronTask{Name: "cache_resource_snapshot", Schedule: cron.Command.EveryFiveMinutes, Run: t.CacheResourceSnapshot},
		)
	}
	return tasks
}
