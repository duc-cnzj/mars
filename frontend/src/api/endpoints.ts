/** 后端 API 端点常量表：全库 API 路径收敛于此，消灭散落字面量。
 *  typed client（openapi-fetch）路径沿用占位符语法（如 /api/projects/{id}），
 *  与 src/api/schema.d.ts 的 paths key 一一对应——引用时若拼写错误 tsc 直接报错（天然护栏）。
 *  裸 fetch / SSE 的模板路径收敛为导出函数，保留各自编码/拼接语义。 */
export const API = {
  // --- access tokens ---
  accessTokens: '/api/access_tokens',
  accessTokenDetail: '/api/access_tokens/{token}',

  // --- auth ---
  authExchange: '/api/auth/exchange',
  authInfo: '/api/auth/info',
  authLogin: '/api/auth/login',
  authSettings: '/api/auth/settings',
  pictureBackground: '/api/picture/background',

  // --- admin（后台） ---
  adminBoard: '/api/admin/cluster/board',
  adminResources: '/api/admin/cluster/resources',
  adminNamespaces: '/api/admin/namespaces',
  adminProjectsLiveness: '/api/admin/projects/liveness',
  adminSettings: '/api/admin/settings',
  adminUsers: '/api/admin/users',
  adminUserRole: '/api/admin/users/{email}/role',

  // --- changelogs（配置历史） ---
  changelogsFindLast: '/api/changelogs/find_last_changelogs_by_project_id',

  // --- cluster ---
  clusterInfo: '/api/cluster_info',

  // --- containers ---
  containerCopyToPod: '/api/containers/copy_to_pod',
  containerForceDelete: '/api/containers/namespaces/{namespace}/pods/{pod}/force_delete',

  // --- endpoints（空间/项目访问地址） ---
  endpointsNamespace: '/api/endpoints/namespaces/{namespaceId}',
  endpointsProject: '/api/endpoints/projects/{projectId}',

  // --- events / files / record_files ---
  events: '/api/events',
  eventsDetail: '/api/events/{id}',
  recordFileDetail: '/api/record_files/{id}',
  fileDiskInfo: '/api/files/disk_info',
  fileMaxUploadSize: '/api/files/max_upload_size',
  fileDetail: '/api/files/{id}',
  /** 文件上传（裸 fetch 用，schema 有该端点） */
  filesUpload: '/api/files',
  /** 从 pod 复制文件（裸 fetch 用） */
  copyFromPod: '/api/copy_from_pod',

  // --- git ---
  gitAllRepos: '/api/git/all_repos',
  gitChartValuesYaml: '/api/git/get_chart_values_yaml',
  gitProjectOptions: '/api/git/project_options',
  gitBranchOptions: '/api/git/projects/{gitProjectId}/branch_options',
  gitCommitOptions: '/api/git/projects/{gitProjectId}/branches/{branch}/commit_options',
  gitPipelineJobOptions: '/api/git/projects/{gitProjectId}/pipeline_job_options',
  gitPipelineInfo: '/api/git/repos/{repoId}/branches/{branch}/commits/{commit}/pipeline_info',

  // --- metrics ---
  metricsNamespaceCpuMemory: '/api/metrics/namespace/{namespaceId}/cpu_memory',
  metricsProjectCpuMemory: '/api/metrics/projects/{projectId}/cpu_memory',

  // --- namespaces ---
  namespaces: '/api/namespaces',
  namespacesFavorite: '/api/namespaces/favorite',
  namespacesFavoriteSort: '/api/namespaces/favorite/sort',
  namespacesUpdateConfig: '/api/namespaces/update_config',
  namespacesDetail: '/api/namespaces/{id}',
  namespacesUpdateDesc: '/api/namespaces/{id}/update_desc',

  // --- projects ---
  projectsDetail: '/api/projects/{id}',
  projectsContainers: '/api/projects/{id}/containers',
  projectsMemoryCpuEndpoints: '/api/projects/{id}/memory_cpu_and_endpoints',
  projectsResourceTree: '/api/projects/{id}/resource_tree',

  // --- repos ---
  repos: '/api/repos',
  reposClone: '/api/repos/clone',
  reposExport: '/api/repos/export',
  reposImport: '/api/repos/import',
  reposToggleEnabled: '/api/repos/toggle_enabled',
  reposDetail: '/api/repos/{id}',
  reposDetailExport: '/api/repos/{id}/export',

  // --- misc ---
  version: '/api/version',
} as const

/** 事件文件下载（裸 fetch，无 typed 端点） */
export const downloadFileUrl = (fileId: number | string) => `/api/download_file/${fileId}`

/** Pod 实时指标 SSE 流（query 由调用方追加，如 ?time=） */
export const podMetricsStreamUrl = (namespace: string, pod: string) =>
  `/api/metrics/namespace/${namespace}/pods/${pod}/stream`

/** 容器日志流（SSE）：路径段统一 encodeURIComponent，showEvents 控制是否带事件日志 */
export const containerStreamLogsUrl = (ns: string, pod: string, container: string, showEvents: boolean) =>
  `/api/containers/namespaces/${encodeURIComponent(ns)}/pods/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/stream_logs?showEvents=${showEvents}`
