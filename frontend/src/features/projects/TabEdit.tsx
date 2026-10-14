import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { getHibernateWindows } from '@/api/hibernateWindowsCache'
import { nextZIndex } from '@/lib/zIndex'
import { Icon } from '@/components/Icons'
import { CodeEditor, FILE_TYPES } from '@/components/CodeEditor'
import { DiffViewer } from '@/components/DiffViewer'
import { useConfetti } from '@/hooks/useConfetti'
import { Button } from '@/components/ui/shadcn/button'
import { Switch } from '@/components/ui/shadcn/switch'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/shadcn/popover'
import { SearchableSelect } from '@/components/SearchableSelect'
import { BottomTabButton, handleTablistKeyDown } from './configTabs'
import { ConfigHistory } from './ConfigHistory'
import { DeployLog } from './DeployLog'
import { Elements } from './Elements'
import { HibernateReplicaWarning } from './HibernateReplicaWarning'
import { HibernateWindowSelect } from './HibernateWindowSelect'
import { hibernateIconActive, iconTriggerCls } from './iconTrigger'
import {
  hibernateWindowTimeRange,
  replicaFieldLabels,
  shouldWarnReplicaInHibernate,
} from './hibernate'
import { PipelineInfo } from './PipelineInfo'
import { SegmentedSkeleton } from '@/components/ui'
import { useDeployStream } from './useDeployStream'

type ProjectModel = components['schemas']['types.ProjectModel']
type GitOption = components['schemas']['git.Option']
type Element = components['schemas']['mars.Element']
type GroupSetting = components['schemas']['mars.GroupSetting']
type ExtraValue = components['schemas']['websocket.ExtraValue']
type HibernateWindowModel = components['schemas']['hibernate.HibernateWindow']

/**
 * 配置更新 Tab（布局对齐旧版 DeployProjectForm）：
 * - 吸顶头部：pipeline 状态 + 项目/分支/commit 级联选择 + 操作按钮行（部署/重置/取消/查看日志/历史）
 * - 部署参数（Elements）3 列紧凑排布（TextArea 长文本块下移底部 tab，不占网格）
 * - 底部「配置文件 / 各 TextArea」可切换 tab（仅在存在 TextArea 字段时出现）：每个 TextArea 独立一个
 *   tab、标题取各自 description||path。TextArea 面板与配置文件面板共用同一套 grid 定高布局
 *   （min-h-[500px] flex-1 + grid-rows-[minmax(0,1fr)] + 编辑器 h-full），两面板视觉一致——编辑器都
 *   占满整列高度；相对部署值有改动时右侧并排 diff（统一视图、仅变更、无工具栏）。无 TextArea 时无 tab、
 *   配置文件编辑器直接展示
 * - 部署时日志替换表单（查看/隐藏日志切换），实时进度 + 日志行
 * 正式部署走 WS 实时部署流（WebApply）。
 */
export function TabEdit({
  detail,
  active,
  onChanged,
  onRefresh,
  onDeployed,
}: {
  detail: ProjectModel
  /** 当前是否为激活 Tab：失活（keep-alive 隐藏）时通过 key 重建三个 SearchableSelect，
   *  强制关闭可能开着、portal 挂在 body 上的弹层——否则切走瞬间残留成幽灵下拉（闪一帧淡出） */
  active?: boolean
  onChanged: () => void
  /** 刷新配置：重拉项目最新详情（父级同时更新其 detail，其它 Tab/列表一并刷新）。
   *  返回最新详情，失败返回 null（错误提示已由父级 reload 弹出，此处不重复提示） */
  onRefresh: () => Promise<ProjectModel | null>
  /** 部署成功回调（弹窗据此从配置 Tab 切到拓扑 Tab 看最终资源树） */
  onDeployed?: () => void
}) {
  const { t } = useTranslation()
  const fireConfetti = useConfetti()

  const gitProjectId = detail.repo?.gitProjectId ?? 0
  const needGit = detail.repo?.needGitRepo

  const [branch, setBranch] = useState<string>(detail.gitBranch || '')
  const [commit, setCommit] = useState<string>(detail.gitCommit || '')
  const [config, setConfig] = useState<string>(detail.config || '')
  const [extraValues, setExtraValues] = useState<ExtraValue[]>(
    detail.finalExtraValues ?? [],
  )
  const [elements, setElements] = useState<Element[]>([])
  const [groupSettings, setGroupSettings] = useState<GroupSetting[]>([])
  const [configFileType, setConfigFileType] = useState('yaml')
  const [branchOptions, setBranchOptions] = useState<GitOption[]>([])
  const [commitOptions, setCommitOptions] = useState<GitOption[]>([])
  const [loadingRepo, setLoadingRepo] = useState(false)
  const [loadingBranch, setLoadingBranch] = useState(false)
  const [loadingCommit, setLoadingCommit] = useState(false)
  // 部署日志占位开关：部署中/后展示日志（替换表单），对齐旧版「查看/隐藏日志」切换
  const [showLog, setShowLog] = useState(false)
  // 刷新按钮进行中：禁用按钮 + 转圈，避免重复点击
  const [refreshing, setRefreshing] = useState(false)
  // 仓库级休眠配置里的副本字段 path（来自 marsConfig.hibernate）：用于判定「窗口内被改成 >0」的提示
  const [hibernateReplicaPaths, setHibernateReplicaPaths] = useState<string[]>([])
  // 定时休眠开关显隐：仓库级能力（marsConfig.hibernate.enabled）开启才显示（同 TabInfo 门控）
  const [hasHibernate, setHasHibernate] = useState(false)
  // 本地开关态：初始同步 detail，切换时乐观更新（失败回滚），父级刷新后按最新 detail 重新对齐
  const [hibEnabled, setHibEnabled] = useState(detail.hibernateEnabled)
  const [hibSaving, setHibSaving] = useState(false)
  // 仓库声明的休眠时间窗列表（项目多选的选择范围）+ 项目本地选定的窗子集（空 = 未选）。
  // 展示用的时段由「生效窗」拼出（走缓存模块去重的全局时间窗列表）。
  const [repoWindowNames, setRepoWindowNames] = useState<string[]>([])
  const [hibWindows, setHibWindows] = useState<string[]>(detail.hibernateWindowNames ?? [])
  const [windows, setWindows] = useState<HibernateWindowModel[]>([])
  // 月亮弹层（编辑时间窗）开合 + z-index：**受控**开合，对齐 SearchableSelect / useOverlayZ 的
  // 既定写法——Radix 触发器打开的浮层在点击事件里同步 onOpenChange，同一批次就能定好 z；
  // 初始 0 是有意的（不打开就不占用共享计数器）。portal 到 body，默认 z-50 会被宿主项目详情
  // 弹窗（useOverlayZ 分配的 52+）盖住，故打开时现取 nextZIndex() 压到所有已分配浮层之上。
  const [hibOpen, setHibOpen] = useState(false)
  const [hibZ, setHibZ] = useState(0)
  useEffect(() => {
    setHibEnabled(detail.hibernateEnabled)
    setHibWindows(detail.hibernateWindowNames ?? [])
  }, [detail.id, detail.hibernateEnabled, detail.hibernateWindowNames])
  useEffect(() => {
    if (!hasHibernate) return
    let alive = true
    void getHibernateWindows().then((w) => {
      if (alive) setWindows(w)
    })
    return () => {
      alive = false
    }
  }, [hasHibernate])

  // 拉取 repo 的 mars 配置：elements 动态表单 + 编辑器语言。
  // loadingRepo 期间项目 select 骨架占位（本页「项目」API）
  useEffect(() => {
    if (detail.repoId <= 0) return
    let alive = true
    setLoadingRepo(true)
    api
      .GET(API.reposDetail, { params: { path: { id: detail.repoId } } })
      .then(({ data }) => {
        if (!alive || !data?.item?.marsConfig) return
        const mc = data.item.marsConfig
        setElements(mc.elements ?? [])
        setGroupSettings(mc.groupSettings ?? [])
        setConfigFileType(mc.configFileType || 'yaml')
        setHibernateReplicaPaths(mc.hibernate?.replicaPaths ?? [])
        setHasHibernate(Boolean(mc.hibernate?.enabled))
        setRepoWindowNames(mc.hibernate?.windowNames ?? [])
      })
      .finally(() => {
        if (alive) setLoadingRepo(false)
      })
    return () => {
      alive = false
    }
  }, [detail.repoId])

  /** 切换项目级定时休眠开关：仅改开关不触发重部署（副本映射/时间窗来自仓库级配置）；成功后通知
   *  父级刷新详情与工作台列表，失败回滚本地态（同 TabInfo 交互）。开启前必须已选至少一个时间窗
   *  （空选 = 未选，不继承仓库全量）——未选则拒绝开启并提示，不发请求、不翻本地开关 */
  const toggleHibernate = async (next: boolean) => {
    if (next && hibWindows.length === 0) {
      toast.error(t('project.hibernateWindowRequired'))
      return
    }
    setHibEnabled(next)
    setHibSaving(true)
    try {
      const { error } = await api.POST(API.projectsHibernate, {
        params: { path: { projectId: detail.id } },
        body: { projectId: detail.id, enabled: next, windowNames: hibWindows },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(next ? t('project.hibernateToggleOn') : t('project.hibernateToggleOff'))
      onChanged()
    } catch (e) {
      setHibEnabled(!next)
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setHibSaving(false)
    }
  }

  /** 保存项目选定的休眠时间窗子集（月亮弹层内多选）：复用 ToggleHibernate 端点，只改 windowNames。
   *  清空（不选任何窗）= 关闭该项目休眠：项目级休眠必须有生效窗，故提交 enabled=false 并清空选择，
   *  避免留下「开着开关却没有窗」的空转态；仍在选窗时 enabled 原样回传，仅换子集。
   *  成功后通知父级刷新；失败回滚本地选择/开关并提示（不让失败的改动误回显）。
   *  保存中守卫 + 弹层内多选禁用（同 ProjectGovernance/RepoProjectsDialog 范式）：避免并发写竞态，
   *  也避免受控组件在守卫下吞 onChange 造成勾选回弹。 */
  const saveWindows = async (names: string[]) => {
    if (hibSaving) return
    const prevEnabled = hibEnabled
    const prevWindows = hibWindows
    const nextEnabled = names.length > 0 ? hibEnabled : false
    setHibWindows(names)
    setHibEnabled(nextEnabled)
    setHibSaving(true)
    try {
      const { error } = await api.POST(API.projectsHibernate, {
        params: { path: { projectId: detail.id } },
        body: { projectId: detail.id, enabled: nextEnabled, windowNames: names },
      })
      if (error) throw new Error(error.message ?? String(error))
      onChanged()
    } catch (e) {
      setHibWindows(prevWindows)
      setHibEnabled(prevEnabled)
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setHibSaving(false)
    }
  }

  // 生效窗 = 项目选定子集本身（空 = 未选，不继承仓库全量；与后端对账同名同义）
  // 生效时间段（只显示 HH:MM-HH:MM、不掺窗名）；TabEdit 操作行只读展示，改窗在月亮弹层
  const hibWindow = hibernateWindowTimeRange(hibWindows, windows)
  // 休眠期间会被置 0 的副本字段展示名（描述优先、回退 path），供横幅告知用户具体是哪些字段
  const replicaLabels = replicaFieldLabels(hibernateReplicaPaths, elements)

  // 实时部署流：正式部署走 WS 通道（Create/UpdateProject 帧流式回投进度与日志）
  const stream = useDeployStream(detail.namespaceId, detail.name)

  // 部署终态：成功 → 彩蛋 + 成功 toast + 隐藏日志回到表单 + 刷新列表；
  // 失败/取消 → 终态 toast（失败保留日志面板供排查，取消无需隐藏）
  useEffect(() => {
    if (stream.status === 'deployed') {
      fireConfetti()
      toast.success(t('project.deploySuccess', { name: detail.name }))
      // 部署成功后才隐藏日志、回到配置表单：部署中日志替换表单，成功后展示配置结果（失败保留日志排查）
      setShowLog(false)
      onChanged()
      // 成功后切到拓扑 Tab：弹窗内看最终资源树 + 后续 pod 事件驱动实时刷新
      onDeployed?.()
    } else if (stream.status === 'failed') {
      toast.error(t('project.deployFailed', { name: detail.name }))
    } else if (stream.status === 'canceled') {
      toast.info(t('project.deployCanceled', { name: detail.name }))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stream.status])

  // 分支选项拉取：挂载时初次拉（loading 骨架），每次点开分支下拉再拉一次最新。
  // 点开时的刷新不置 loading —— 骨架会把打开的下拉换掉导致闪关，故静默更新选项
  const fetchBranchOptions = useCallback(() => {
    if (!needGit || gitProjectId <= 0) return Promise.resolve()
    return api
      .GET(API.gitBranchOptions, {
        params: { path: { gitProjectId }, query: { repoId: detail.repoId } },
      })
      .then(({ data, error }) => {
        if (!error && data) setBranchOptions(data.items)
      })
      .catch(() => {})
  }, [needGit, gitProjectId, detail.repoId])

  useEffect(() => {
    if (!needGit || gitProjectId <= 0) return
    setLoadingBranch(true)
    void fetchBranchOptions().finally(() => setLoadingBranch(false))
  }, [fetchBranchOptions])

  // commit 选项拉取：挂载/换分支时初次拉（清空旧 commit 选项 + loading 骨架）；
  // 每次点开 commit 下拉再拉一次最新（静默刷新，不清空，避免下拉打开中闪空）
  const fetchCommitOptions = useCallback(() => {
    if (!needGit || gitProjectId <= 0 || !branch) return Promise.resolve()
    return api
      .GET(API.gitCommitOptions, {
        params: { path: { gitProjectId, branch } },
      })
      .then(({ data, error }) => {
        if (!error && data) setCommitOptions(data.items)
      })
      .catch(() => {})
  }, [needGit, gitProjectId, branch])

  useEffect(() => {
    if (!needGit || gitProjectId <= 0) return
    setCommitOptions([])
    if (!branch) return
    setLoadingCommit(true)
    void fetchCommitOptions().finally(() => setLoadingCommit(false))
  }, [fetchCommitOptions, branch])

  const branchOpts = useMemo(
    () => branchOptions.map((o) => ({ label: o.label, value: o.value })),
    [branchOptions],
  )
  const commitOpts = useMemo(
    () => commitOptions.map((o) => ({ label: o.label, value: o.value })),
    [commitOptions],
  )

  /** 正式部署：走 WS 实时流（进度 + 日志 + 终态），结束后由 status 触发刷新 */
  const realDeploy = () => {
    if (needGit && (!branch || !commit)) {
      toast(t('project.needBranchCommit'))
      return
    }
    if (!stream.ready) {
      toast(t('common.loading'))
      return
    }
    setShowLog(true)
    stream.update({
      projectId: detail.id,
      version: detail.version,
      gitBranch: needGit ? branch : undefined,
      gitCommit: needGit ? commit : undefined,
      config,
      extraValues,
    })
  }

  /** 重置：恢复项目当前的配置/部署参数/分支/commit */
  const reset = () => {
    setConfig(detail.config || '')
    setBranch(detail.gitBranch || '')
    setCommit(detail.gitCommit || '')
    // 不能清空 commitOptions：分支未改时下面的 commit-options effect 不会重跑，
    // 清空会让 commit select 丢 label 查找（显示回落成原始 hash）且点开无列表可取。
    // 分支真的变了由 effect 自行 setCommitOptions([]) 后重取。
    setExtraValues(detail.finalExtraValues ?? [])
  }

  /** 刷新：从服务端重拉项目最新详情，把配置/部署参数/分支/commit 重置为最新值（丢弃本地未部署改动）。
   *  重拉成功后父级 detail 同步更新（其它 Tab/列表一并刷新），失败提示由父级统一弹出 → 此处静默返回 */
  const refresh = async () => {
    setRefreshing(true)
    try {
      const fresh = await onRefresh()
      if (!fresh) return
      setConfig(fresh.config || '')
      setBranch(fresh.gitBranch || '')
      setCommit(fresh.gitCommit || '')
      setExtraValues(fresh.finalExtraValues ?? [])
      toast.success(t('project.configRefreshed'))
    } finally {
      setRefreshing(false)
    }
  }

  /** 配置相对当前部署版本是否有改动 → 决定右侧 diff 列是否展开（有改动才 12/12 并排） */
  const configChanged = (config || '') !== (detail.config || '')
  // 有部署活动（进行中或已结束）才显示「查看/隐藏日志」切换，对齐旧版 hasLog
  const hasLog = stream.loading || stream.status !== 'idle'
  const projectRepoName = detail.repo?.name || detail.name

  /** 项目已开启定时休眠（hibernate_enabled===true）且处于休眠态（hibernate_state==='sleeping'），
   *  同时有副本字段被改成 >0：对账（每 2 分钟一轮）会自动把它缩容回 0，故此处提示用户「这样改不会
   *  立即生效」。判据见 hibernate.shouldWarnReplicaInHibernate。休眠开关统一在页面操作行，本横幅只
   *  提示、不含开关——关掉开关后 hibernate_enabled 变 false，判据失效、横幅随之消失。 */
  const replicaOverZero = useMemo(
    () => shouldWarnReplicaInHibernate(detail, extraValues, hibernateReplicaPaths),
    [detail, extraValues, hibernateReplicaPaths],
  )

  /** 自定义节点字段拆分：TextArea 长文本块移出部署参数网格，进底部「自定义配置」tab；其余字段留网格 */
  const textareaElements = useMemo(
    () => elements.filter((e) => e.type === 'ElementTypeTextArea'),
    [elements],
  )
  const compactElements = useMemo(
    () => elements.filter((e) => e.type !== 'ElementTypeTextArea'),
    [elements],
  )
  const hasTextarea = textareaElements.length > 0
  // 部署时的 TextArea 取值（按 path）：供底部 tab 的 TextArea diff 对比「旧值」，有改动才展开 diff
  //（与 configChanged 对比 detail.config 同一语义；未部署过的新字段无旧值 → 不展开）
  const textareaOldValues = useMemo(() => {
    const map = new Map<string, string>()
    for (const v of detail.finalExtraValues ?? []) map.set(v.path, v.value)
    return map
  }, [detail.finalExtraValues])
  // 底部 tab 的 value：'config'（配置文件编辑器，默认选中）或某个 TextArea 的 path——每个 TextArea 独立
  // 一个 tab、标题取各自 description||path，多 TextArea 时即有多个 tab。无 TextArea 字段时不渲染 tab 栏
  const [bottomTab, setBottomTab] = useState<string>('config')

  /** 更新指定 path 的取值（保留其余项），统一转字符串存储（与 Elements 内部 update 同语义）。
   *  description 由后端在部署落库时按元素定义固化，请求侧不必带（proto 里是可选字段） */
  const updateExtraValue = useCallback((path: string, raw: unknown) => {
    setExtraValues((prev) => [...prev.filter((v) => v.path !== path), { path, value: String(raw) }])
  }, [])

  return (
    <div className="flex h-full min-h-0 flex-col">
      {/* 吸顶头部（对齐旧版 Affix）：pipeline + 项目/分支/commit 级联 + 操作按钮 */}
      <div className="z-10 shrink-0 border-b border-line bg-bg">
        <div className="space-y-2 px-1 pb-2 pt-1">
          {/* pipeline 区域：needGit 时固定占位高度（横幅/占位等高 42px）。
              三种状态高度一致：pipeline 横幅 / 获取不到占位 / 未就绪提示，切换不挤压下方网格 */}
          <div className={needGit ? 'min-h-[42px]' : undefined}>
            {needGit && detail.repoId > 0 && branch && commit && (
              <PipelineInfo repoId={detail.repoId} branch={branch} commit={commit} />
            )}
            {needGit && detail.repoId > 0 && !(branch && commit) && (
              <div className="flex min-h-[42px] items-center gap-2 rounded-md border border-line bg-surface px-3 py-2">
                <Icon name="clock" className="shrink-0 text-[14px] text-faint" />
                <span className="text-[13px] font-medium leading-none text-faint">
                  {branch ? t('project.pipelineNeedCommit') : t('project.pipelineNeedBranch')}
                </span>
              </div>
            )}
          </div>

          <div
            // key 随 active 翻转：失活时重建整格（三个 SearchableSelect 随之重建，open 复位 false），
            // portal 弹层随组件卸载即刻移除，绕开 Radix Presence 150ms 退出动画，无残影
            key={active ? 'sel-active' : 'sel-hidden'}
            className={`grid grid-cols-1 gap-2 ${needGit ? 'md:grid-cols-3 md:gap-0' : 'md:grid-cols-1'}`}
          >
            {/* 三个 SearchableSelect 在 md+ 合并为分段控件（对齐旧版 Row/Col）：左保留圆角、中全 0、右保留圆角。
                骨架高度 h-[38px] 取 trigger 的 used height（min-h-9=36 内容已顶满 34 后加 border 2px = 38） */}
            {loadingRepo ? (
              <SegmentedSkeleton className="md:rounded-r-none" />
            ) : (
              <SearchableSelect
                value={String(detail.repoId)}
                options={[{ value: String(detail.repoId), label: projectRepoName }]}
                onChange={() => {}}
                placeholder={t('project.project')}
                searchPlaceholder={t('project.searchProject')}
                align="center"
                truncateTip
                className="md:rounded-r-none"
              />
            )}
            {needGit &&
              (loadingBranch ? (
                <SegmentedSkeleton className="md:-ml-px md:rounded-none md:border-l-0" />
              ) : (
                <SearchableSelect
                  value={branch}
                  options={branchOpts}
                  onChange={(v) => {
                    setBranch(v as string)
                    setCommit('')
                  }}
                  onOpenChange={(open) => {
                    // 每次点开分支下拉 → 重新拉取最新分支列表（静默刷新，下拉保持打开）
                    if (open) void fetchBranchOptions()
                  }}
                  placeholder={t('project.branch')}
                  searchPlaceholder={t('project.searchBranch')}
                  emptyText={t('common.empty')}
                  align="center"
                  truncateTip
                  className="md:-ml-px md:rounded-none md:border-l-0"
                />
              ))}
            {needGit &&
              (loadingCommit ? (
                <SegmentedSkeleton className="md:-ml-px md:rounded-l-none md:border-l-0" />
              ) : (
                <SearchableSelect
                  value={commit}
                  options={commitOpts}
                  onChange={(v) => setCommit(v as string)}
                  onOpenChange={(open) => {
                    // 每次点开 commit 下拉 → 重新拉取最新 commit 列表（静默刷新，下拉保持打开）
                    if (open) void fetchCommitOptions()
                  }}
                  placeholder={t('project.commit')}
                  searchPlaceholder={t('project.searchCommit')}
                  emptyText={t('common.empty')}
                  align="center"
                  truncateTip
                  className="md:-ml-px md:rounded-l-none md:border-l-0"
                />
              ))}
          </div>

          {/* 操作按钮行：统一 xs 尺寸（部署/重置/刷新/历史主操作 + 部署中/日志切换一致高度） */}
          <div className="flex flex-wrap items-center gap-2">
            <Button size="xs" variant="default" disabled={stream.loading} onClick={realDeploy}>
              {stream.loading && <Icon name="loader" className="size-3.5 animate-spin" />}
              <Icon name="rocket" className="text-[13px]" />
              {t('project.deploy')}
            </Button>
            <Button size="xs" variant="outline" disabled={stream.loading} onClick={reset}>
              <Icon name="refresh" className="text-[13px]" />
              {t('project.reset')}
            </Button>
            {/* 刷新配置：重拉服务端最新项目配置（重置按钮只回到已加载快照，本按钮才重新取数） */}
            <Button
              size="xs"
              variant="outline"
              disabled={stream.loading || refreshing}
              onClick={() => void refresh()}
            >
              {refreshing ? (
                <Icon name="loader" className="size-3.5 animate-spin" />
              ) : (
                <Icon name="refresh-cw" className="text-[13px]" />
              )}
              {t('project.configRefresh')}
            </Button>
            {stream.loading && (
              <Button size="xs" variant="outline" onClick={stream.cancel}>
                <Icon name="close" className="text-[13px]" />
                {t('events.actionCancelDeploy')}
              </Button>
            )}
            {hasLog && (
              <Button size="xs" variant="outline" onClick={() => setShowLog((v) => !v)}>
                <Icon name="logs" className="text-[13px]" />
                {showLog ? t('project.hideLog') : t('project.viewLog')}
              </Button>
            )}
            <ConfigHistory projectId={detail.id} configFileType={configFileType} />
            {/* 定时休眠：仓库级启用休眠才显示，靠右排布。月亮图标在「定时休眠」文案左侧，点开弹层
                编辑时间窗（多选仓库窗子集），其余只读展示生效时段；右侧开关切换启用/停用。编辑入口
                从旧 `?` 只读说明升级为可写弹层——列表行月亮入口下线后，此处是改窗的就地入口；
                开关只管启用/停用，弹层只管时间窗，职责分离。均 POST ToggleHibernate（不触发重部署）。 */}
            {hasHibernate && (
              <div className="ml-auto flex items-center gap-1.5">
                <Popover
                  open={hibOpen}
                  onOpenChange={(o) => {
                    // 打开时现取 z：可拖拽宿主弹窗每次指针交互（bringToFront）会递增 z，须打开时才算
                    if (o) setHibZ(nextZIndex())
                    setHibOpen(o)
                  }}
                >
                  <PopoverTrigger asChild>
                    <button
                      type="button"
                      aria-label={t('project.hibernateWindow')}
                      title={t('project.hibernateWindow')}
                      // -mr-1 抵消 24px 命中区图标自身的 4px 右内边距，让月亮与紧邻的「定时休眠」文字更紧凑
                      className={`${hibEnabled ? hibernateIconActive : iconTriggerCls} -mr-1`}
                    >
                      <Icon name={hibEnabled ? 'moon-fill' : 'moon'} className="size-4" />
                    </button>
                  </PopoverTrigger>
                  <PopoverContent
                    align="end"
                    sideOffset={4}
                    style={{ zIndex: hibZ || undefined }}
                    className="w-[min(480px,90vw)] p-3"
                  >
                    {/* 时间窗（可写）：多选仓库窗子集 + 生效时段一览 */}
                    <HibernateWindowSelect
                      value={hibWindows}
                      repoWindowNames={repoWindowNames}
                      windows={windows}
                      onChange={saveWindows}
                      disabled={hibSaving}
                    />
                    {/* 副本字段（只读）：休眠期间会被置 0 的字段名 + 置 0 说明 */}
                    {replicaLabels.length > 0 && (
                      <div className="mt-2.5 border-t border-line pt-2 text-[11px] leading-snug text-mute">
                        <span className="text-faint">
                          {t('project.hibernateReplicaFields')}:{' '}
                        </span>
                        <span className="font-mono">{replicaLabels.join(', ')}</span>
                        <div className="mt-1 text-faint">
                          {t('project.hibernateReplicaZeroNote')}
                        </div>
                      </div>
                    )}
                  </PopoverContent>
                </Popover>
                <span className="flex items-center gap-1 text-[12px] text-mute">
                  {t('project.hibernate')}
                  {hibWindow ? ` · ${hibWindow}` : ''}
                </span>
                <Switch
                  aria-label={t('project.hibernate')}
                  checked={hibEnabled}
                  disabled={hibSaving}
                  onCheckedChange={toggleHibernate}
                />
              </div>
            )}
          </div>

          {/* 休眠态下把副本改成 >0：提示不会立即生效（对账 ~2 分钟内自动缩容回 0）。非阻断，只在
              项目已开启定时休眠、处休眠态且副本字段被改成正数时出现，初值（副本库内为 0）不触发。
              休眠开关统一放在上方独立休眠行，横幅只做提示；关掉开关后 hibernate_enabled 变 false，
              判据失效、横幅消失（见 hibernate.shouldWarnReplicaInHibernate） */}
          {replicaOverZero && <HibernateReplicaWarning fields={replicaLabels} />}
        </div>
      </div>

      {/* 内容区：部署中/后显示日志，否则显示部署参数 + 配置编辑/diff 并排（对齐旧版 showLog 替换表单） */}
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain p-1">
        {showLog ? (
          // fill：面板占满内容区自适应高度，日志区内滚，不被日志条数撑高
          <DeployLog
            status={stream.status}
            percent={stream.percent}
            logs={stream.logs}
            loading={stream.loading}
            fill
          />
        ) : (
          <>
            {compactElements.length > 0 && (
              // 部署参数网格只含非 TextArea 字段（长文本块下移底部「自定义配置」tab）。
              // 不整块重建 Elements：整块 key 翻转会把 CodeEditor/折叠态等全部重置，
              // 违背 keep-alive 保留表单状态的初衷。失活只重键 select 型字段（见 Elements 内部 key 逻辑）
              <div className="mb-3">
                <Elements
                  elements={compactElements}
                  value={extraValues}
                  onChange={setExtraValues}
                  active={active}
                  variant="compact"
                  groupSettings={groupSettings}
                />
              </div>
            )}

            {hasTextarea && (
              // 底部「配置文件 / 各 TextArea」分段控件。配置文件排在首位并默认选中（主编辑面）；
              // 每个 TextArea 独立一个 tab、标题取各自 description||path，多个 TextArea 即多个 tab。
              // tab 按内容宽排布，容器宽度不足时横向滚动、长标题 max-w 截断省略；滚动条隐藏
              // （scrollbar-none），横向滑动保留、纵向恒不占位。无 TextArea 字段时不渲染
              // 自定义 tab 条替代 shadcn TabsList：定高 30px 容器内，经典（非 overlay）滚动条会在
              // 横向溢出时吃掉 ~15px 纵向空间、把 24px 触发器挤爆成纵向滚动条（双条并存）。
              // 自定义条只保留横向溢出 + 隐藏滚动条，从根上杜绝纵向滚动条。
              // 宽度自适应：w-fit 让条只占内容宽（窄时不拉满容器），max-w-full 在内容超宽时
              // 被容器钳制、转为横向滚动（等价原 shadcn TabsList 的 w-fit max-w-full 语义）
              <div
                role="tablist"
                aria-label={t('project.configFileTab')}
                onKeyDown={handleTablistKeyDown}
                // 轨道色用 border-strong/60（比 raised 实一档）：bg-muted=raised 太浅几乎不可见，
                // 又不把 --raised 全局加深（连带骨架/搜索框等），局部取 border-strong 语义 token
                className="mb-2 flex h-[30px] w-fit max-w-full shrink-0 items-center gap-1 overflow-x-auto rounded-lg bg-line-strong/60 p-[3px] scrollbar-none"
              >
                <BottomTabButton
                  active={bottomTab === 'config'}
                  onClick={() => setBottomTab('config')}
                >
                  {t('project.configFileTab')}
                </BottomTabButton>
                {textareaElements.map((element) => (
                  <BottomTabButton
                    key={element.path}
                    // tab value 用 't:' 前缀，避开与 'config' 哨兵值撞车（若某 TextArea 的 path
                    // 恰为 'config'，无前缀时该 tab 与配置文件 tab 值相同、面板判定也会双显）
                    active={bottomTab === `t:${element.path}`}
                    onClick={() => setBottomTab(`t:${element.path}`)}
                    title={element.description || element.path}
                  >
                    <span className="block max-w-[200px] truncate">
                      {element.description || element.path}
                    </span>
                  </BottomTabButton>
                ))}
              </div>
            )}

            {/* 自定义配置面板区：每个 TextArea 独立一个面板，与配置文件面板同样保持挂载、用 hidden
                切换（display:none 不卸载）。关键：切 tab 不卸载 CodeMirror——挂载瞬间编辑器高度为 0、
                撑不起 flex 容器，会触发整块配置区滚动条重置到顶部；保持挂载后切走/切回都无挂载事件，
                滚动位置不丢。每个面板的编辑器/变更 diff 结构逐项对齐配置文件面板（同一套 grid 定高
                模式：min-h-[500px] flex-1 + grid-rows-[minmax(0,1fr)] 行 + 编辑器 h-full），两面板
                视觉一致。值实时写入 extraValues，切走不丢数据 */}
            {textareaElements.map((element) => {
              // 该 TextArea 当前编辑值：extraValues 按 path，无则回退元素 default（同 Elements 的 display 语义）
              const textareaValue =
                extraValues.find((v) => v.path === element.path)?.value ??
                element.default ??
                ''
              // 相对部署值（finalExtraValues 按 path）有改动才展开右侧 diff（与 configChanged 同语义）
              const textareaChanged =
                textareaOldValues.has(element.path) &&
                textareaOldValues.get(element.path) !== textareaValue
              // 编辑器语言由后端 textarea_language 指定；不在 CodeEditor 支持集里回退 textile（与 Elements 同逻辑）
              const textareaLang = (
                FILE_TYPES as readonly string[]
              ).includes(element.textareaLanguage)
                ? element.textareaLanguage
                : 'textile'
              return (
                <div
                  key={element.path}
                  // 与触发器的 't:' 前缀对应，value 统一带前缀判定（见上方 tab 注释）
                  className={
                    bottomTab !== `t:${element.path}`
                      ? 'hidden'
                      : 'grid min-h-[500px] flex-1 grid-rows-[minmax(0,1fr)]'
                  }
                  style={{
                    gridTemplateColumns: textareaChanged
                      ? 'minmax(0, 1fr) minmax(0, 1fr)'
                      : 'minmax(0, 1fr)',
                  }}
                >
                  <div className="min-h-0 min-w-0">
                    <CodeEditor
                      value={textareaValue}
                      onChange={(v) => updateExtraValue(element.path, v)}
                      language={textareaLang}
                      className="h-full !rounded-r-none !border-r-0"
                    />
                  </div>
                  {textareaChanged && (
                    <div className="min-h-0 min-w-0">
                      <DiffViewer
                        oldValue={textareaOldValues.get(element.path)}
                        newValue={textareaValue}
                        language={textareaLang}
                        initialView="unified"
                        hideToolbar
                        className="h-full overflow-hidden rounded-l-none rounded-r-md border-y border-r border-line"
                        viewportClassName="rounded-l-none rounded-r-md"
                      />
                    </div>
                  )}
                </div>
              )
            })}

            {/* 配置文件面板：配置编辑器与变更 diff 并排。左右两列放同一条 grid 行轨（minmax(0,1fr)）：
                行高定死（flex-1 + min-h-[500px]）后两列即等高，编辑器/diff 用 h-full 填满
                （同 DiffModal 已证实的 grid 定高模式）。不沿用 flex stretch 列：height:100% 对其
                求值不稳、flex-grow 对不定高容器不分配，都会让 diff 塌成内容高、与左侧 codemirror 不等高。
                diff 列不默认展开：配置无改动时编辑器占满全宽，改动后才 12/12 并排。
                bottomTab 非 'config'（选中某 TextArea tab）时隐藏（display 切换，CodeMirror 不卸载） */}
            <div
              className={
                bottomTab !== 'config'
                  ? 'hidden'
                  : 'grid min-h-[500px] flex-1 grid-rows-[minmax(0,1fr)]'
              }
              style={{
                gridTemplateColumns: configChanged
                  ? 'minmax(0, 1fr) minmax(0, 1fr)'
                  : 'minmax(0, 1fr)',
              }}
            >
              <div className="min-h-0 min-w-0">
                <CodeEditor
                  value={config}
                  onChange={setConfig}
                  minHeight="500px"
                  language={configFileType}
                  // 无缝并排：去掉接缝侧（右侧）边框与圆角，与 diff 卡片平贴，不再有白线
                  className="h-full !rounded-r-none !border-r-0"
                />
              </div>
              {configChanged && (
                <div className="min-h-0 min-w-0">
                  {/* 无缝并排：接缝侧（左）无边框无圆角，与编辑器平贴（白线 = 两边框贴一起所致，已去）。
                      上下+右侧补 border-line 与编辑器同款——否则 diff 无边框会比编辑器暗色面
                      上下各高 1px（编辑器表面被 1px 边框内缩）。viewportClassName 把内部滚动
                      容器左圆角也拉平：根节点 overflow-hidden 只裁溢出、裁不掉子元素内部圆角缺口。 */}
                  <DiffViewer
                    oldValue={detail.config || ''}
                    newValue={config}
                    language={configFileType}
                    initialView="unified"
                    // 配置更新页不展示分屏/统一/仅变更/复制工具栏：恒统一视图、仅显示变更
                    hideToolbar
                    className="h-full overflow-hidden rounded-l-none rounded-r-md border-y border-r border-line"
                    viewportClassName="rounded-l-none rounded-r-md"
                  />
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
