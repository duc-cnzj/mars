import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { Icon } from '@/components/Icons'
import { CodeEditor } from '@/components/CodeEditor'
import { DiffViewer } from '@/components/DiffViewer'
import { useConfetti } from '@/hooks/useConfetti'
import { Button } from '@/components/ui/shadcn/button'
import { SearchableSelect } from '@/components/SearchableSelect'
import { ConfigHistory } from './ConfigHistory'
import { DeployLog } from './DeployLog'
import { Elements } from './Elements'
import { PipelineInfo } from './PipelineInfo'
import { SegmentedSkeleton } from '@/components/ui'
import { useDeployStream } from './useDeployStream'

type ProjectModel = components['schemas']['types.ProjectModel']
type GitOption = components['schemas']['git.Option']
type Element = components['schemas']['mars.Element']
type ExtraValue = components['schemas']['websocket.ExtraValue']

/**
 * 配置更新 Tab（布局对齐旧版 DeployProjectForm）：
 * - 吸顶头部：pipeline 状态 + 项目/分支/commit 级联选择 + 操作按钮行（部署/重置/取消/查看日志/历史）
 * - 部署参数（Elements）3 列紧凑排布
 * - 配置编辑器与变更 diff 并排（配置有改动时 12/12 分栏，diff 统一视图、仅显示变更、无工具栏），最小 500px，随内容正常滚动
 * - 部署时日志替换表单（查看/隐藏日志切换），实时进度 + 日志行
 * 正式部署走 WS 实时部署流（WebApply）。
 */
export function TabEdit({
  detail,
  onChanged,
  onDeployed,
}: {
  detail: ProjectModel
  onChanged: () => void
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
  const [configFileType, setConfigFileType] = useState('yaml')
  const [branchOptions, setBranchOptions] = useState<GitOption[]>([])
  const [commitOptions, setCommitOptions] = useState<GitOption[]>([])
  const [loadingRepo, setLoadingRepo] = useState(false)
  const [loadingBranch, setLoadingBranch] = useState(false)
  const [loadingCommit, setLoadingCommit] = useState(false)
  // 部署日志占位开关：部署中/后展示日志（替换表单），对齐旧版「查看/隐藏日志」切换
  const [showLog, setShowLog] = useState(false)

  // 拉取 repo 的 mars 配置：elements 动态表单 + 编辑器语言。
  // loadingRepo 期间项目 select 骨架占位（本页「项目」API）
  useEffect(() => {
    if (detail.repoId <= 0) return
    let alive = true
    setLoadingRepo(true)
    api
      .GET('/api/repos/{id}', { params: { path: { id: detail.repoId } } })
      .then(({ data }) => {
        if (!alive || !data?.item?.marsConfig) return
        const mc = data.item.marsConfig
        setElements(mc.elements ?? [])
        setConfigFileType(mc.configFileType || 'yaml')
      })
      .finally(() => {
        if (alive) setLoadingRepo(false)
      })
    return () => {
      alive = false
    }
  }, [detail.repoId])

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

  useEffect(() => {
    if (!needGit || gitProjectId <= 0) return
    setLoadingBranch(true)
    api
      .GET('/api/git/projects/{gitProjectId}/branch_options', {
        params: { path: { gitProjectId }, query: { repoId: detail.repoId } },
      })
      .then(({ data, error }) => {
        if (!error && data) setBranchOptions(data.items)
      })
      .finally(() => setLoadingBranch(false))
  }, [needGit, gitProjectId, detail.repoId])

  useEffect(() => {
    if (!needGit || gitProjectId <= 0 || !branch) return
    setLoadingCommit(true)
    setCommitOptions([])
    api
      .GET('/api/git/projects/{gitProjectId}/branches/{branch}/commit_options', {
        params: { path: { gitProjectId, branch } },
      })
      .then(({ data, error }) => {
        if (!error && data) setCommitOptions(data.items)
      })
      .finally(() => setLoadingCommit(false))
  }, [needGit, gitProjectId, branch])

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

  /** 配置相对当前部署版本是否有改动 → 决定右侧 diff 列是否展开（有改动才 12/12 并排） */
  const configChanged = (config || '') !== (detail.config || '')
  // 有部署活动（进行中或已结束）才显示「查看/隐藏日志」切换，对齐旧版 hasLog
  const hasLog = stream.loading || stream.status !== 'idle'
  const projectRepoName = detail.repo?.name || detail.name

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

          <div className={`grid grid-cols-1 gap-2 ${needGit ? 'md:grid-cols-3 md:gap-0' : 'md:grid-cols-1'}`}>
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
                  placeholder={t('project.commit')}
                  searchPlaceholder={t('project.searchCommit')}
                  emptyText={t('common.empty')}
                  align="center"
                  truncateTip
                  className="md:-ml-px md:rounded-l-none md:border-l-0"
                />
              ))}
          </div>

          {/* 操作按钮行：统一 xs 尺寸（部署/重置/历史主操作 + 部署中/日志切换一致高度） */}
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
          </div>
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
            {elements.length > 0 && (
              <div className="mb-3">
                <Elements elements={elements} value={extraValues} onChange={setExtraValues} />
              </div>
            )}

            {/* 左右两列放同一条 grid 行轨（minmax(0,1fr)）：行高定死（flex-1 + min-h-[500px]）
                后两列即等高，编辑器/diff 用 h-full 填满（同 DiffModal 已证实的 grid 定高模式）。
                不沿用 flex stretch 列：height:100% 对其求值不稳、flex-grow 对不定高容器不分配，
                都会让 diff 塌成内容高、与左侧 codemirror 不等高。
                diff 列不默认展开：配置无改动时编辑器占满全宽，改动后才 12/12 并排 */}
            <div
              className="grid min-h-[500px] flex-1 grid-rows-[minmax(0,1fr)]"
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
