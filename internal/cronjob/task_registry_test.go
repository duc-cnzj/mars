package cronjob

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/cron"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// taskNames 提取任务名列表，供断言。
func taskNames(tasks []CronTask) []string {
	names := make([]string, 0, len(tasks))
	for _, task := range tasks {
		names = append(names, task.Name)
	}
	return names
}

// findTask 按名字取任务（值副本），未命中返回零值。
func findTask(tasks []CronTask, name string) CronTask {
	for _, task := range tasks {
		if task.Name == name {
			return task
		}
	}
	return CronTask{}
}

// TestTasks_Registry_Base 覆盖基础 4 任务：无条件任务的枚举与调度/执行体非空。
func TestTasks_Registry_Base(t *testing.T) {
	tasks := Registry(&Tasks{}, &config.Config{})
	assert.ElementsMatch(t, []string{
		"clean_upload_files", "fix_project_deploy_status", "sync_domain_secret", "disk_info",
	}, taskNames(tasks))
	for _, task := range tasks {
		assert.NotNil(t, task.Schedule)
		assert.NotNil(t, task.Run)
	}
}

// TestTasks_Registry_Conditional 覆盖条件任务：GitServerCached 追加两个缓存任务，
// K8s 环境（KubeConfig 非空）追加镜像拉取 secret 同步与两个快照预热任务。
func TestTasks_Registry_Conditional(t *testing.T) {
	withCache := Registry(&Tasks{}, &config.Config{GitServerCached: true})
	assert.ElementsMatch(t, []string{
		"clean_upload_files", "fix_project_deploy_status", "sync_domain_secret", "disk_info",
		"all_branch_cache", "all_project_cache",
	}, taskNames(withCache))

	inK8s := Registry(&Tasks{}, &config.Config{KubeConfig: "/tmp/kube"})
	assert.ElementsMatch(t, []string{
		"clean_upload_files", "fix_project_deploy_status", "sync_domain_secret", "disk_info",
		"sync_image_pull_secrets", "cache_cluster_board", "cache_resource_snapshot",
	}, taskNames(inK8s))

	// 条件全开：6 + 3。
	both := Registry(&Tasks{}, &config.Config{GitServerCached: true, KubeConfig: "/tmp/kube"})
	assert.ElementsMatch(t, []string{
		"clean_upload_files", "fix_project_deploy_status", "sync_domain_secret", "disk_info",
		"all_branch_cache", "all_project_cache", "sync_image_pull_secrets", "cache_cluster_board", "cache_resource_snapshot",
	}, taskNames(both))
}

// TestTasks_Registry_DiskInfoAdapter 覆盖 DiskInfo 适配闭包：把 (int64, error)
// 降为 error 供 cron 执行体使用，size 被丢弃。
func TestTasks_Registry_DiskInfoAdapter(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)

	fr := data.NewMockFileRepo(m)
	fr.EXPECT().DiskInfo(true).Return(int64(1024), nil)

	tasks := Registry(&Tasks{fileRepo: fr}, &config.Config{})
	assert.NoError(t, findTask(tasks, "disk_info").Run())
}

// TestTasks_Registry_Schedules 覆盖 5 个调度构造器：经真实 cron.Manager 注册
// Registry 产出的全部任务（含条件任务），断言命令名与 cron 表达式。
func TestTasks_Registry_Schedules(t *testing.T) {
	cm := cron.NewManager(timer.NewReal(), nil, nil, mlog.NewForConfig(nil))
	tasks := Registry(&Tasks{}, &config.Config{GitServerCached: true, KubeConfig: "/tmp/kube"})
	for _, task := range tasks {
		task.Schedule(cm.NewCommand(task.Name, task.Run))
	}

	exprs := make(map[string]string)
	for _, c := range cm.List() {
		exprs[c.Name()] = c.Expression()
	}
	assert.Len(t, exprs, len(tasks))
	assert.Equal(t, "0 00 2 * * *", exprs["clean_upload_files"])
	assert.Equal(t, "0 */2 * * * *", exprs["fix_project_deploy_status"])
	assert.Equal(t, "0 * * * * *", exprs["sync_domain_secret"])
	assert.Equal(t, "0 */10 * * * *", exprs["disk_info"])
	assert.Equal(t, "0 */2 * * * *", exprs["all_branch_cache"])
	assert.Equal(t, "0 */5 * * * *", exprs["all_project_cache"])
	assert.Equal(t, "0 */5 * * * *", exprs["sync_image_pull_secrets"])
	assert.Equal(t, "0,30 * * * * *", exprs["cache_cluster_board"])
	assert.Equal(t, "0 */5 * * * *", exprs["cache_resource_snapshot"])
}
