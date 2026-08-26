import { useCallback, useEffect, useMemo, useState, type KeyboardEvent, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { Icon } from '@/components/Icons'
import { CodeEditor, FILE_TYPES } from '@/components/CodeEditor'
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
  onDeployed,
}: {
  detail: ProjectModel
  /** 当前是否为激活 Tab：失活（keep-alive 隐藏）时通过 key 重建三个 SearchableSelect，
   *  强制关闭可能开着、portal 挂在 body 上的弹层——否则切走瞬间残留成幽灵下拉（闪一帧淡出） */
  active?: boolean
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

  // 分支选项拉取：挂载时初次拉（loading 骨架），每次点开分支下拉再拉一次最新。
  // 点开时的刷新不置 loading —— 骨架会把打开的下拉换掉导致闪关，故静默更新选项
  const fetchBranchOptions = useCallback(() => {
    if (!needGit || gitProjectId <= 0) return Promise.resolve()
    return api
      .GET('/api/git/projects/{gitProjectId}/branch_options', {
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
      .GET('/api/git/projects/{gitProjectId}/branches/{branch}/commit_options', {
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

  /** 配置相对当前部署版本是否有改动 → 决定右侧 diff 列是否展开（有改动才 12/12 并排） */
  const configChanged = (config || '') !== (detail.config || '')
  // 有部署活动（进行中或已结束）才显示「查看/隐藏日志」切换，对齐旧版 hasLog
  const hasLog = stream.loading || stream.status !== 'idle'
  const projectRepoName = detail.repo?.name || detail.name

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

  /** 更新指定 path 的取值（保留其余项），统一转字符串存储（与 Elements 内部 update 同语义） */
  const updateExtraValue = useCallback((path: string, raw: unknown) => {
    setExtraValues((prev) => [
      ...prev.filter((v) => v.path !== path),
      { path, value: String(raw) },
    ])
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

/**
 * 底部配置 tab 条的方向键导航（roving tabindex，对齐原 shadcn Tabs 的键盘行为）：
 * ←/→ 循环移动焦点并激活（聚焦即切换），Home/End 跳首尾。事件在按钮上触发后冒泡到
 * tablist 容器统一处理；target 可能是按钮内 span，向上找 [role=tab] 定位当前项。
 */
function handleTablistKeyDown(e: KeyboardEvent<HTMLDivElement>) {
  const tabs = Array.from(e.currentTarget.querySelectorAll<HTMLButtonElement>('[role="tab"]'))
  if (tabs.length === 0) return
  const curEl = (e.target as HTMLElement).closest<HTMLButtonElement>('[role="tab"]')
  const cur = curEl ? tabs.indexOf(curEl) : -1
  let next = cur
  switch (e.key) {
    case 'ArrowRight':
      next = cur < 0 ? 0 : (cur + 1) % tabs.length
      break
    case 'ArrowLeft':
      next = cur < 0 ? tabs.length - 1 : (cur - 1 + tabs.length) % tabs.length
      break
    case 'Home':
      next = 0
      break
    case 'End':
      next = tabs.length - 1
      break
    default:
      return
  }
  e.preventDefault()
  tabs[next].focus()
  tabs[next].click()
}

/**
 * 底部配置 tab 条的单按钮：尺寸对齐上方部署按钮（Button size=xs：h-6 + text-xs），
 * 选中态用主题色柔色变体（primary-soft 底 + primary 文字），比实心填充轻、比无色默认态鲜明。
 * 长标题由外部包 span truncate + max-w 截断；按钮自身 shrink-0，容器不足时横向滚动不挤压。
 * roving tabindex：仅选中项可 Tab 聚焦（tabIndex=0），其余 -1，方向键在容器 onKeyDown 统一移动。
 * 不用 shadcn TabsTrigger：其基座 h-[calc(100%-1px)]/py-1/text-sm 需多层 ! 压掉，
 * 且宿主 TabsList 定高 + 溢出时经典滚动条会挤出纵向滚动条，自定义条从根上规避。
 */
function BottomTabButton({
  active,
  onClick,
  children,
  title,
}: {
  active: boolean
  onClick: () => void
  children: ReactNode
  /** 完整标题（长文本 tab 的 tooltip） */
  title?: string
}) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      tabIndex={active ? 0 : -1}
      title={title}
      onClick={onClick}
      className={`flex h-6 shrink-0 items-center gap-1.5 rounded-md px-2 text-xs font-medium whitespace-nowrap transition-colors ${
        active
          ? 'bg-primary/20 text-primary'
          : 'text-foreground/60 hover:text-foreground'
      }`}
    >
      {children}
    </button>
  )
}
