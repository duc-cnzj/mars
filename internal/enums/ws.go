package enums

const (
	WsHandleExecShell    string = "handle_exec_shell"
	WsHandleExecShellMsg string = "handle_exec_shell_msg"
	WsHandleCloseShell   string = "handle_close_shell"

	WsSetUid string = "set_uid"
	// WsReloadProjects
	// TODO 最好是直接把数据返回，而不是让前端再去获取一次
	WsReloadProjects string = "reload_projects"
	WsCancel         string = "cancel_project"
	WsCreateProject  string = "create_project"
	WsUpdateProject  string = "update_project"

	WsProcessPercent  string = "process_percent"
	WsClusterInfoSync string = "cluster_info_sync"

	WsInternalError string = "internal_error"
)
