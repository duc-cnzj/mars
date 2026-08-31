import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { Icon } from '@/components/Icons'
import { Tag, SegmentedSkeleton } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { SearchableSelect } from '@/components/SearchableSelect'
import { useDraggableDialog } from '@/hooks/useDraggableDialog'
import { useWheelRedirect } from '@/hooks/useWheelRedirect'
import { useConfetti } from '@/hooks/useConfetti'
import { CodeEditor, FILE_TYPES } from '@/components/CodeEditor'
import { BottomTabButton, handleTablistKeyDown } from './configTabs'
import { Elements } from './Elements'
import { DeployLog } from './DeployLog'
import { PipelineInfo } from './PipelineInfo'
import { useDeployStream } from './useDeployStream'

type GitOption = components['schemas']['git.Option']
type Element = components['schemas']['mars.Element']
type GroupSetting = components['schemas']['mars.GroupSetting']
type ExtraValue = components['schemas']['websocket.ExtraValue']

/**
 * 创建项目弹窗（布局对齐配置更新 TabEdit/旧版 DeployProjectForm）：
 * - 吸顶头部：pipeline 槽 + 项目/分支/commit 分段三控件 + 操作按钮行（部署/取消/查看日志）
 * - 部署参数（Elements）+ 配置编辑在内容区；部署时日志替换表单（fill 占满 + 查看/隐藏日志切换）
 * - WS 实时部署流（CreateProject），成功后彩蛋 + toast + 刷新 + 关闭
 */
export function CreateProjectModal({
  namespaceId,
  namespaceName,
  open,
  onClose,
  onChanged,
}: {
  namespaceId: number
  /** 空间名称：title 徽标展示（替换原先的数字 id，让用户看清创建在哪个空间） */
  namespaceName: string
  open: boolean
  onClose: () => void
  onChanged: () => void
}) {
  const { t } = useTranslation()
  const dnd = useDraggableDialog()
  const fireConfetti = useConfetti()

  const [projectOptions, setProjectOptions] = useState<GitOption[]>([])
  const [branchOptions, setBranchOptions] = useState<GitOption[]>([])
  const [commitOptions, setCommitOptions] = useState<GitOption[]>([])

  const [repoId, setRepoId] = useState(0)
  const [projectName, setProjectName] = useState('')
  const [needGitRepo, setNeedGitRepo] = useState(false)
  const [gitProjectId, setGitProjectId] = useState(0)
  const [branch, setBranch] = useState('')
  const [commit, setCommit] = useState('')
  const [config, setConfig] = useState('')
  const [extraValues, setExtraValues] = useState<ExtraValue[]>([])
  const [elements, setElements] = useState<Element[]>([])
  const [groupSettings, setGroupSettings] = useState<GroupSetting[]>([])
  const [configFileType, setConfigFileType] = useState('yaml')
  // 底部「配置文件 / 各 TextArea」tab：'config'（配置文件编辑器，默认选中）或某个 TextArea 的 path
  //（'t:' 前缀避开与 'config' 哨兵值撞车，同 TabEdit 约定）
  const [bottomTab, setBottomTab] = useState<string>('config')

  const [loadingProject, setLoadingProject] = useState(false)
  const [loadingConfig, setLoadingConfig] = useState(false)
  const [loadingBranch, setLoadingBranch] = useState(false)
  const [loadingCommit, setLoadingCommit] = useState(false)
  const [clusterBad, setClusterBad] = useState(false)
  // 部署日志占位开关：部署中/后展示日志（替换表单），对齐 TabEdit 查看/隐藏日志切换
  const [showLog, setShowLog] = useState(false)

  // 集群健康：status=bad 时禁止部署（安全门控，纯前端）。
  // 只在弹窗打开时拉取——本组件随每张命名空间卡片挂载，若挂载即拉，
  // 首页有 N 张卡片就并发 N 次 /api/cluster_info（切页回首页重拉成请求风暴）。
  // 打开时重置 clusterBad，避免上次打开的 bad 状态残留。
  useEffect(() => {
    if (!open) return
    setClusterBad(false)
    let alive = true
    void api.GET(API.clusterInfo).then(({ data }) => {
      if (alive && data?.item && data.item.status === 'bad') setClusterBad(true)
    })
    return () => {
      alive = false
    }
  }, [open])

  // 打开弹窗：重置表单并拉取 repo 下拉
  useEffect(() => {
    if (!open) return
    setRepoId(0)
    setProjectName('')
    setNeedGitRepo(false)
    setGitProjectId(0)
    setBranch('')
    setCommit('')
    setConfig('')
    setExtraValues([])
    setElements([])
    setGroupSettings([])
    setConfigFileType('yaml')
    setBranchOptions([])
    setCommitOptions([])
    setShowLog(false)
    setLoadingProject(true)
    api
      .GET(API.gitProjectOptions)
      .then(({ data, error }) => {
        if (!error && data) setProjectOptions(data.items)
      })
      .finally(() => setLoadingProject(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  // 选中 repo → 拉取 mars 配置（elements/config/语言）
  const selectRepo = async (option: GitOption) => {
    const id = Number(option.value)
    setRepoId(id)
    setProjectName(option.label)
    setNeedGitRepo(option.needGitRepo)
    setGitProjectId(option.gitProjectId)
    setBranch('')
    setCommit('')
    setBranchOptions([])
    setCommitOptions([])
    setConfig('')
    setExtraValues([])
    setElements([])
    setLoadingConfig(true)
    try {
      const { data, error } = await api.GET(API.reposDetail, {
        params: { path: { id } },
      })
      if (error) throw new Error(error.message ?? String(error))
      const marsConfig = data?.item?.marsConfig
      if (marsConfig) {
        setConfig(marsConfig.configFileValues ?? '')
        setElements(marsConfig.elements ?? [])
        setGroupSettings(marsConfig.groupSettings ?? [])
        setExtraValues(
          (marsConfig.elements ?? []).map((e) => ({
            path: e.path,
            value: e.default,
          })),
        )
        setConfigFileType(marsConfig.configFileType || 'yaml')
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setLoadingConfig(false)
    }
  }

  // 分支级联：选中 repo 且 needGitRepo 时拉取分支
  useEffect(() => {
    if (!needGitRepo || gitProjectId <= 0 || repoId <= 0) return
    setLoadingBranch(true)
    setBranchOptions([])
    api
      .GET(API.gitBranchOptions, {
        params: { path: { gitProjectId }, query: { repoId } },
      })
      .then(({ data, error }) => {
        if (!error && data) setBranchOptions(data.items)
      })
      .finally(() => setLoadingBranch(false))
  }, [needGitRepo, gitProjectId, repoId])

  // commit 级联：选中分支后拉取 commit
  useEffect(() => {
    if (!needGitRepo || gitProjectId <= 0 || !branch) return
    setLoadingCommit(true)
    setCommitOptions([])
    api
      .GET(API.gitCommitOptions, {
        params: { path: { gitProjectId, branch } },
      })
      .then(({ data, error }) => {
        if (!error && data) setCommitOptions(data.items)
      })
      .finally(() => setLoadingCommit(false))
  }, [needGitRepo, gitProjectId, branch])

  // 实时部署流：name 用选中 repo 的 label（未选中为空串，内部自动重订阅）
  const stream = useDeployStream(namespaceId, projectName)

  // TextArea 长文本块移出部署参数网格，进底部「配置文件 / 各 TextArea」tab（结构对齐 TabEdit）
  const textareaElements = useMemo(
    () => elements.filter((e) => e.type === 'ElementTypeTextArea'),
    [elements],
  )
  const compactElements = useMemo(
    () => elements.filter((e) => e.type !== 'ElementTypeTextArea'),
    [elements],
  )
  const hasTextarea = textareaElements.length > 0

  /** 更新指定 path 的取值（保留其余项），统一转字符串存储（与 Elements 内部 update 同语义，同 TabEdit） */
  const updateExtraValue = useCallback((path: string, raw: unknown) => {
    setExtraValues((prev) => [
      ...prev.filter((v) => v.path !== path),
      { path, value: String(raw) },
    ])
  }, [])

  // 整块弹窗区域滚轮重定向：标题/吸顶头等不可滚动区滚轮转发到内容区滚动条，不穿透滚主页面
  //（有嵌套滚动如 CodeMirror/日志则原生，详见 useWheelRedirect）
  const { dialogRef, contentRef } = useWheelRedirect()

  // 部署终态：成功 → 彩蛋 + toast + 刷新 + 关闭（对齐旧版创建流程）；
  // 失败/取消 → 终态 toast 后保留日志面板供排查（对齐成功侧一致性）。
  // 成功前把 showLog 置 false（对齐 TabEdit：部署成功后隐藏日志、回到配置表单）；
  // 随后 onClose() 关闭弹窗，本行在关闭前保证状态一致（若将来改为成功后停留，配置即回显）
  useEffect(() => {
    if (stream.status === 'deployed') {
      fireConfetti()
      toast.success(t('project.deploySuccess', { name: projectName }))
      setShowLog(false)
      onChanged()
      onClose()
    } else if (stream.status === 'failed') {
      toast.error(t('project.deployFailed', { name: projectName }))
    } else if (stream.status === 'canceled') {
      toast.info(t('project.deployCanceled', { name: projectName }))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stream.status])

  // repo 下拉：label + description 给 SearchableSelect（分段内选项带说明行）
  const repoOpts = useMemo(
    () =>
      projectOptions.map((o) => ({
        value: o.value,
        label: o.label,
        description: o.description,
      })),
    [projectOptions],
  )
  const branchOpts = useMemo(
    () => branchOptions.map((o) => ({ label: o.label, value: o.value })),
    [branchOptions],
  )
  const commitOpts = useMemo(
    () => commitOptions.map((o) => ({ label: o.label, value: o.value })),
    [commitOptions],
  )

  const startDeploy = () => {
    if (clusterBad) {
      toast.error(t('project.clusterResourceShortage'))
      return
    }
    if (repoId <= 0 || !projectName) {
      toast(t('project.selectProjectFirst'))
      return
    }
    if (needGitRepo && (!branch || !commit)) {
      toast(t('project.needBranchCommit'))
      return
    }
    if (!stream.ready) {
      toast(t('common.loading'))
      return
    }
    setShowLog(true)
    stream.create({
      repoId,
      gitBranch: needGitRepo ? branch : undefined,
      gitCommit: needGitRepo ? commit : undefined,
      config,
      extraValues,
    })
  }

  const deployDisabled = repoId <= 0 || loadingConfig || clusterBad || stream.loading
  // 有部署活动（进行中或已结束）才显示「查看/隐藏日志」切换，对齐 TabEdit hasLog
  const hasLog = stream.loading || stream.status !== 'idle'

  // 非模态（对齐 ProjectDetailModal/旧版 DraggableModal）：Radix modal Dialog 会给 body 加
  // pointer-events:none 锁死弹窗外点击，且遮罩挡着没法与其他项目弹窗并存；
  // modal={false} + showOverlay={false} 去掉锁与遮罩——弹窗可拖动/缩放/最大化、可多开，弹窗外仍可操作
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()} modal={false}>
      <DialogContent
        ref={dialogRef}
        {...dnd.contentProps}
        // 只允许 X / ESC 关闭（用户：外点不关）。modal={false} 时 Radix 默认 pointerdown 落在外
        // 会走 DismissableLayer dismiss 关掉弹窗，须 preventDefault 拦截；焦点移出另走
        // onFocusOutside 路径，也一并拦（对齐 ProjectDetailModal 的处理与注释）
        showOverlay={false}
        onPointerDownOutside={(e) => e.preventDefault()}
        onFocusOutside={(e) => e.preventDefault()}
        style={dnd.contentStyle}
        className="sm:max-w-5xl h-[70vh] flex flex-col"
      >
        <DialogHeader>
          <DialogTitle
            {...dnd.dragHandleProps}
            className="flex cursor-move select-none items-center gap-2 pr-8 text-[15px]"
            title={t('project.dragTitle')}
          >
            <Icon name="rocket" className="text-[15px] text-primary" />
            {t('project.createProject')}
            <Tag tone="accent" dot={false}>
              {namespaceName || namespaceId}
            </Tag>
          </DialogTitle>
        </DialogHeader>

        {/* 吸顶头部（对齐 TabEdit/旧版 Affix）：pipeline + 项目/分支/commit 分段 + 操作按钮 */}
        <div className="z-10 shrink-0 border-b border-line bg-bg">
          <div className="space-y-2 px-1 pb-2 pt-1">
            {/* pipeline 槽：needGitRepo 时固定占位高度（横幅/占位等高 42px，切换不挤压下方网格） */}
            <div className={needGitRepo ? 'min-h-[42px]' : undefined}>
              {needGitRepo && repoId > 0 && branch && commit && (
                <PipelineInfo repoId={repoId} branch={branch} commit={commit} />
              )}
              {needGitRepo && repoId > 0 && !(branch && commit) && (
                <div className="flex min-h-[42px] items-center gap-2 rounded-md border border-line bg-surface px-3 py-2">
                  <Icon name="clock" className="shrink-0 text-[14px] text-faint" />
                  <span className="text-[13px] font-medium leading-none text-faint">
                    {branch ? t('project.pipelineNeedCommit') : t('project.pipelineNeedBranch')}
                  </span>
                </div>
              )}
            </div>

            {/* 三个 SearchableSelect 在 md+ 合并为分段控件（对齐 TabEdit）：左保留圆角、中全 0、右保留圆角。
                骨架高度 h-[38px] 取 trigger 的 used height（min-h-9=36 内容已顶满 34 后加 border 2px = 38） */}
            <div className={`grid grid-cols-1 gap-2 ${needGitRepo ? 'md:grid-cols-3 md:gap-0' : 'md:grid-cols-1'}`}>
              {loadingProject ? (
                <SegmentedSkeleton className="md:rounded-r-none" />
              ) : (
                <SearchableSelect
                  // repoId=0（未选择）时传空串：SearchableSelect 只在 value==='' 时显示 placeholder，
                  // 传 "0" 会被当成已选值回退显示原始数字「0」而非占位文案
                  value={repoId > 0 ? String(repoId) : ''}
                  options={repoOpts}
                  onChange={(v) => {
                    const opt = projectOptions.find((o) => o.value === v)
                    if (opt) void selectRepo(opt)
                  }}
                  placeholder={t('project.selectProject')}
                  searchPlaceholder={t('common.search')}
                  emptyText={t('common.empty')}
                  align="center"
                  className="md:rounded-r-none"
                />
              )}
              {needGitRepo &&
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
                    searchPlaceholder={t('common.search')}
                    emptyText={t('common.empty')}
                    align="center"
                    className="md:-ml-px md:rounded-none md:border-l-0"
                  />
                ))}
              {needGitRepo &&
                (loadingCommit ? (
                  <SegmentedSkeleton className="md:-ml-px md:rounded-l-none md:border-l-0" />
                ) : (
                  <SearchableSelect
                    value={commit}
                    options={commitOpts}
                    onChange={(v) => setCommit(v as string)}
                    placeholder={t('project.commit')}
                    searchPlaceholder={t('common.search')}
                    emptyText={t('common.empty')}
                    align="center"
                    className="md:-ml-px md:rounded-l-none md:border-l-0"
                  />
                ))}
            </div>

            {/* 操作按钮行：统一 xs 尺寸（部署/取消/日志切换一致高度） */}
            <div className="flex flex-wrap items-center gap-2">
              <Button
                size="xs"
                variant={clusterBad ? 'destructive' : 'default'}
                disabled={deployDisabled}
                onClick={startDeploy}
              >
                {stream.loading && <Icon name="loader" className="size-3.5 animate-spin" />}
                <Icon name="rocket" className="text-[13px]" />
                {clusterBad ? t('project.clusterResourceShortage') : t('project.deploy')}
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
            </div>
          </div>
        </div>

        {/* 内容区：部署中/后显示日志（fill 占满），否则部署参数 + 配置编辑（对齐旧版 showLog 替换表单） */}
        <div ref={contentRef} className="flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain p-1">
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
              {/* 部署参数网格只含非 TextArea 字段（长文本块下移底部「自定义配置」tab，对齐 TabEdit） */}
              {compactElements.length > 0 && (
                <div className="mb-3">
                  <Section title={t('project.configElements')}>
                    <Elements
                      elements={compactElements}
                      value={extraValues}
                      onChange={setExtraValues}
                      variant="compact"
                      groupSettings={groupSettings}
                    />
                  </Section>
                </div>
              )}

              {loadingConfig ? (
                <div className="py-6 text-center text-[13px] text-faint">
                  {t('common.loading')}
                </div>
              ) : (
                <>
                  {/* 底部「配置文件 / 各 TextArea」tab 条（与 TabEdit 一致的 pill 轨道样式）。
                      配置文件排在首位并默认选中（主编辑面）；每个 TextArea 独立一个 tab、标题取各自
                      description||path。定高 30px 灰色轨道 + 激活主题色块（见 configTabs.BottomTabButton
                      默认 pill）。无 TextArea 字段时不渲染 */}
                  {hasTextarea && (
                    <div
                      role="tablist"
                      aria-label={t('project.configFileTab')}
                      onKeyDown={handleTablistKeyDown}
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
                          // tab value 用 't:' 前缀，避开与 'config' 哨兵值撞车（同 TabEdit 约定）
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

                  {/* 每个 TextArea 独立一个面板：grid 定高（min-h-[280px] flex-1 + grid-rows + h-full，
                      对齐设计稿 editor.tall 高度语义）。创建场景无部署旧值、恒不展开 diff（TabEdit 的
                      改动并排 diff 在创建侧不适用）。与配置文件面板用 hidden 切换（display:none 不卸载
                      CodeMirror，切走/切回不重置编辑器） */}
                  {textareaElements.map((element) => {
                    // 该 TextArea 当前编辑值：extraValues 按 path，无则回退元素 default（同 Elements 的 display 语义）
                    const textareaValue =
                      extraValues.find((v) => v.path === element.path)?.value ??
                      element.default ??
                      ''
                    // 编辑器语言由后端 textarea_language 指定；不在 CodeEditor 支持集里回退 textile（同 Elements 逻辑）
                    const textareaLang = (
                      FILE_TYPES as readonly string[]
                    ).includes(element.textareaLanguage)
                      ? element.textareaLanguage
                      : 'textile'
                    return (
                      <div
                        key={element.path}
                        className={
                          bottomTab !== `t:${element.path}`
                            ? 'hidden'
                            : 'grid min-h-[280px] flex-1 grid-rows-[minmax(0,1fr)]'
                        }
                      >
                        <div className="min-h-0 min-w-0">
                          <CodeEditor
                            value={textareaValue}
                            onChange={(v) => updateExtraValue(element.path, v)}
                            language={textareaLang}
                            className="h-full"
                          />
                        </div>
                      </div>
                    )
                  })}

                  {/* 配置文件面板：与 TextArea 面板同一套 grid 定高结构，创建场景恒单列（无 diff 并排） */}
                  <div
                    className={
                      bottomTab !== 'config'
                        ? 'hidden'
                        : 'grid min-h-[280px] flex-1 grid-rows-[minmax(0,1fr)]'
                    }
                  >
                    <div className="min-h-0 min-w-0">
                      <CodeEditor
                        value={config}
                        onChange={setConfig}
                        language={configFileType}
                        className="h-full"
                      />
                    </div>
                  </div>
                </>
              )}
            </>
          )}
        </div>

        {/* 四边交互分层（与 ProjectDetailModal 一致）：
            1. 外层 24px 边条（dragHandleProps）：拖拽移动 + 双击全屏/还原（保留移动窗口功能，含最大化时）
            2. 贴边界 6px 细沿 + 四角 16px（getResizeHandleProps(dir)）：resize，对边锚定；只非最大化时渲染
            关闭钮（DialogContent 基座内 render、z-20）在最上不受遮挡 */}
        <div aria-hidden title={t('project.dragTitle')} {...dnd.dragHandleProps} className="pointer-events-auto absolute inset-x-0 top-0 h-6 cursor-move" />
        <div aria-hidden title={t('project.dragTitle')} {...dnd.dragHandleProps} className="pointer-events-auto absolute inset-x-0 bottom-0 h-6 cursor-move" />
        <div aria-hidden title={t('project.dragTitle')} {...dnd.dragHandleProps} className="pointer-events-auto absolute inset-y-0 left-0 w-6 cursor-move" />
        <div aria-hidden title={t('project.dragTitle')} {...dnd.dragHandleProps} className="pointer-events-auto absolute inset-y-0 right-0 w-6 cursor-move" />
        {!dnd.isMaximized && (
          <>
            <div aria-hidden title={t('project.resizeEdge')} {...dnd.getResizeHandleProps('n')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute inset-x-0 top-0 z-10 h-1.5 cursor-ns-resize" />
            <div aria-hidden title={t('project.resizeEdge')} {...dnd.getResizeHandleProps('s')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute inset-x-0 bottom-0 z-10 h-1.5 cursor-ns-resize" />
            <div aria-hidden title={t('project.resizeEdge')} {...dnd.getResizeHandleProps('w')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute inset-y-0 left-0 z-10 w-1.5 cursor-ew-resize" />
            <div aria-hidden title={t('project.resizeEdge')} {...dnd.getResizeHandleProps('e')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute inset-y-0 right-0 z-10 w-1.5 cursor-ew-resize" />
            <div aria-hidden title={t('project.resizeCorner')} {...dnd.getResizeHandleProps('nw')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute top-0 left-0 z-10 size-4 cursor-nwse-resize" />
            <div aria-hidden title={t('project.resizeCorner')} {...dnd.getResizeHandleProps('ne')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute top-0 right-0 z-10 size-4 cursor-nesw-resize" />
            <div aria-hidden title={t('project.resizeCorner')} {...dnd.getResizeHandleProps('sw')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute bottom-0 left-0 z-10 size-4 cursor-nesw-resize" />
            <div aria-hidden title={t('project.resizeCorner')} {...dnd.getResizeHandleProps('se')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute bottom-0 right-0 z-10 size-4 cursor-nwse-resize" />
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-1.5">
      <h4 className="text-[13px] font-semibold text-ink">{title}</h4>
      <div className="flex flex-col gap-2">{children}</div>
    </section>
  )
}
