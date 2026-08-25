import { lazy, Suspense, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import {
  SkeletonDetail,
  SkeletonTabEdit,
  SkeletonTabLog,
  SkeletonTabShell,
} from '@/components/ui'
import { useDraggableDialog } from '@/hooks/useDraggableDialog'
import { useWheelRedirect } from '@/hooks/useWheelRedirect'
import { useWebsocket } from '@/hooks/useWebsocket'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/shadcn/tabs'
import { TabInfo } from './TabInfo'
import { TabLog } from './TabLog'
import { TopologyTab } from '../topology/TopologyTab'

// 命令行 Tab 依赖 xterm（约 300KB），按需加载直到真正打开 Shell 才拉取
const TabShell = lazy(() => import('./TabShell').then((m) => ({ default: m.TabShell })))
// 配置更新 Tab 静态依赖 TabEdit→CodeEditor(CodeMirror 607KB)/DiffViewer/react-diff-viewer，
// 与 TabShell 同策略：懒加载，直到真正打开「配置」Tab 才拉取，弹窗打开不预载
const TabEdit = lazy(() => import('./TabEdit').then((m) => ({ default: m.TabEdit })))

type ProjectModel = components['schemas']['types.ProjectModel']
type TabKey = 'logs' | 'shell' | 'edit' | 'detail' | 'topology'

/**
 * 项目详情弹窗：忠实还原旧版 DraggableModal 的 Tab 结构。
 * 容器日志 / 命令行 / 配置更新 / 拓扑 仅在 Deployed/Deploying 时展示，详细信息始终存在。
 * 打开时拉取项目最新详情，成功后按需刷新；部署成功后从配置 Tab 自动切到拓扑 Tab。
 */
export function ProjectDetailModal({
  project,
  namespaceName,
  open,
  frontAt = 0,
  onClose,
  onDeleted,
  onChanged,
}: {
  project: ProjectModel
  namespaceName: string
  open: boolean
  /** 外部重复点击已打开的该项目卡片时递增：让对应弹窗置顶 */
  frontAt?: number
  onClose: () => void
  onDeleted: () => void
  onChanged: () => void
}) {
  const { t } = useTranslation()
  // 弹窗尺寸变化（缩放/最大化/还原）信号：bump 后传给命令行 Tab 触发终端 refit
  const [resizeAt, setResizeAt] = useState(0)
  const dnd = useDraggableDialog(() => setResizeAt((n) => n + 1))
  // 整块弹窗区域滚轮重定向：标题/Tab 栏等不可滚动区滚轮转发到内容区滚动条，不穿透滚主页面
  const { dialogRef, contentRef } = useWheelRedirect()

  // 重复点击已打开的卡片 → 把该弹窗置顶（frontAt 递增一次置顶一次）
  useEffect(() => {
    if (frontAt > 0) dnd.bringToFront()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [frontAt])
  const { joinProjectPodEvent } = useWebsocket()
  const [detail, setDetail] = useState<ProjectModel | null>(null)
  const [loading, setLoading] = useState(false)
  // 初始 tab 直接按列表项状态定（与下方「打开即同步默认 Tab」的 open effect 同源）：
  // 避免先渲染默认 tab 再靠 effect 切换造成闪跳/重复挂载。
  const [tab, setTab] = useState<TabKey>(() =>
    project.deployStatus === 'StatusDeployed' || project.deployStatus === 'StatusDeploying'
      ? 'logs'
      : 'detail',
  )

  const reload = async () => {
    setLoading(true)
    try {
      const { data, error } = await api.GET('/api/projects/{id}', {
        params: { path: { id: project.id } },
      })
      if (error) throw new Error(error.message ?? String(error))
      if (!data?.item) return
      setDetail(data.item)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }

  /** 配置更新成功后：刷新弹窗内详情 + 通知父级刷新工作台列表 */
  const handleChanged = () => {
    void reload()
    onChanged()
  }

  useEffect(() => {
    if (open) {
      // 打开即同步选定默认 Tab：列表项已带部署状态，无需等 detail 异步加载后再跳转（避免“详细信息→容器日志”的闪烁）
      setTab(canOperate ? 'logs' : 'detail')
      void reload()
    } else {
      setDetail(null)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, project.id])

  // 弹窗打开时加入该项目 pod 事件订阅，关闭时退出（与旧版 useProjectRoom 的挂载/卸载模式一致）
  useEffect(() => {
    if (!open) return
    joinProjectPodEvent(project.namespaceId, project.id, true)
    return () => joinProjectPodEvent(project.namespaceId, project.id, false)
  }, [open, project.id, project.namespaceId, joinProjectPodEvent])

  // detail 未加载完时退回列表项状态（两者同源，列表卡片显示的就是它）
  const canOperate =
    (detail ?? project).deployStatus === 'StatusDeployed' ||
    (detail ?? project).deployStatus === 'StatusDeploying'

  const tabItems: { key: TabKey; label: string }[] = [
    ...(canOperate
      ? [
          { key: 'logs' as const, label: t('project.tabLogs') },
          { key: 'shell' as const, label: t('project.tabShell') },
          { key: 'edit' as const, label: t('project.tabEdit') },
          { key: 'topology' as const, label: t('project.tabTopology') },
        ]
      : []),
    { key: 'detail' as const, label: t('project.tabDetail') },
  ]

  // 兜底：detail 加载后若状态已不可操作（列表快照后刚失败），把停留在操作类 Tab 的选中收回详细信息
  useEffect(() => {
    if (!open || !detail) return
    if (!canOperate && tab !== 'detail') setTab('detail')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, detail])

  return (
    <Dialog
      open={open}
      // 非模态（对齐旧版 DraggableModal）：Radix modal Dialog 会给 body 加 pointer-events:none，
      // 锁死弹窗外所有点击（点了卡片没反应 → 没法开多个）。modal={false} 去掉这个锁。
      // ESC/外点/焦点外移等 dismiss 路径全部在 DialogContent 显式拦截，只保留 X 关闭。
      modal={false}
      onOpenChange={(o) => !o && onClose()}
    >
      <DialogContent
        ref={dialogRef}
        {...dnd.contentProps}
        // 无 mask（对齐旧版 DraggableModal）：页面不被压暗、不拦截外部点击，
        // 才能边开着详情弹窗边点其他卡片再开多个。只允许点击 X 关闭。
        // 用 onPointerDownOutside 而非 onInteractOutside：后者连内部 Popover 打开时
        // focus 到搜索框（在 dialog content 外）也会触发，会与 Radix FocusScope 打架导致下拉开不出来
        showOverlay={false}
        // 三条 dismiss 路径全拦，弹窗只能点 X 关闭：
        // 1. 外部 pointerdown → preventDefault 拦掉 dismiss；
        // 2. 点卡片等可聚焦元素把焦点移出弹窗，Radix 另走 onFocusOutside 的 dismiss 路径
        //   （pointerdown 被「拦截」跳过时 isDeferred 先复位，focusin 不再被跳过 → 误关旧弹窗），
        //   onPointerDownOutside 拦不住焦点路径，需一并 preventDefault onFocusOutside；
        // 3. ESC 默认走 onEscapeKeyDown 关闭，preventDefault 一并拦掉（用户要求：只能 X 关）。
        onPointerDownOutside={(e) => e.preventDefault()}
        onFocusOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={(e) => e.preventDefault()}
        style={dnd.contentStyle}
        className="sm:max-w-5xl h-[70vh] flex flex-col"
      >
        <DialogHeader>
          <DialogTitle
            {...dnd.dragHandleProps}
            className="relative flex cursor-move items-center justify-center px-8"
            title={t('project.dragTitle')}
          >
            {/* 项目名整体居中；命名空间悬浮贴其右上（绝对定位，不参与居中）。
                整体标记 data-no-drag：该区域文本可选中复制，拖拽让位给命名空间/空白区 */}
            <span className="relative min-w-0 cursor-text" data-no-drag>
              <span className="block truncate text-[18px] font-semibold text-ink">
                {project.name}
              </span>
              <span
                className="absolute left-full top-0 ml-0.5 -mt-1 whitespace-nowrap text-[10px] leading-none text-primary"
                style={{ fontFamily: '"dank mono", ui-monospace, monospace' }}
              >
                {namespaceName}
              </span>
            </span>
          </DialogTitle>
        </DialogHeader>

        <Tabs value={tab} onValueChange={(v) => setTab(v as TabKey)}>
          <TabsList variant="line" className="w-full border-b border-line">
            {tabItems.map((it) => (
              // 去掉焦点态主题色边框（TabsTrigger 基座自带 focus-visible:border-ring/ring/outline）
              <TabsTrigger
                key={it.key}
                value={it.key}
                className="focus-visible:!border-transparent focus-visible:!ring-0 focus-visible:!outline-none"
              >
                {it.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>

        <div ref={contentRef} className="mt-1 min-h-0 flex-1 overflow-auto overscroll-contain">
          {loading && !detail ? (
            tab === 'logs' ? (
              <SkeletonTabLog />
            ) : tab === 'shell' ? (
              <SkeletonTabShell />
            ) : tab === 'edit' ? (
              // 配置更新骨架对齐 TabEdit 结构（吸顶头 + 配置编辑器），不再误用 SkeletonDetail
              <SkeletonTabEdit />
            ) : tab === 'topology' ? (
              // 拓扑骨架：复用详细信息骨架占位（资源树由 Tab 自身拉取，弹窗详情只挡首帧）
              <SkeletonDetail />
            ) : (
              <SkeletonDetail />
            )
          ) : !detail ? (
            <div className="py-8 text-center text-[13px] text-mute">{t('common.empty')}</div>
          ) : (
            tabItems.map((it) => {
              const active = tab === it.key
              // 仅 TabEdit 保持挂载（切走 hidden 不卸载）：部署流/表单状态在 tab 间切换时保留
              //（用户：部署中切 tab 部署页不能被 destroy）。
              // 其余 tab 非激活即卸载销毁、切回重建：TabLog/TabInfo 重新拉取最新数据，
              // shell 销毁 xterm/WS 会话，topology 销毁常驻轮询 + pod 事件订阅的资源树
              //（用户：除 TabEdit 外其余 tab 一律销毁，见 tab 声明处注释）。
              // wrapper 统一 h-full overflow-auto：TabEdit/TabShell 是 h-full 内部自滚内容，
              // TabInfo/TabLog 是内容高、由 wrapper 滚动——与原先外层 overflow-auto 行为一致。
              if (!active && it.key !== 'edit') return null
              return (
                <div
                  key={it.key}
                  className={active ? 'h-full overflow-auto overscroll-contain' : 'hidden'}
                >
                  {it.key === 'logs' && <TabLog projectId={detail.id} projectName={detail.name} />}
                  {it.key === 'shell' && (
                    <Suspense fallback={<SkeletonTabShell />}>
                      <TabShell projectId={detail.id} projectName={detail.name} resizeAt={resizeAt} />
                    </Suspense>
                  )}
                  {it.key === 'edit' && (
                    <Suspense fallback={<SkeletonTabEdit />}>
                      {/* 部署成功后从配置 Tab 自动切到拓扑 Tab（用户在配置 Tab 盯日志，成功后跳拓扑看终态资源树）。
                          守卫：仅当前仍在配置 Tab 才跳——用户切到日志/详情等 Tab 时部署在后台完成，不把他拽回拓扑。
                          onDeployed 在 TabEdit 的 [stream.status] effect 触发，inline 箭头捕获最新 tab，父重渲即新闭包 */}
                      <TabEdit
                        detail={detail}
                        active={active}
                        onChanged={handleChanged}
                        onDeployed={() => {
                          if (tab === 'edit') setTab('topology')
                        }}
                      />
                    </Suspense>
                  )}
                  {it.key === 'topology' && <TopologyTab project={detail} resizeAt={resizeAt} />}
                  {it.key === 'detail' && <TabInfo detail={detail} onDeleted={onDeleted} />}
                </div>
              )
            })
          )}
        </div>
        {/* 四边交互分层（盖住 dialog 的 p-6 padding 环）：
            1. 外层 24px 边条（dragHandleProps）：拖拽移动 + 双击全屏/还原——保留旧版整边拖动手感；
               最大化时也保留（双击边条还原全屏，拖动被 clamp 钳制）
            2. 贴边界 6px 细沿 + 四角 16px（getResizeHandleProps(dir)）：resize，对边锚定拖哪边只有哪边动；
               只非最大化时渲染（全屏无可伸缩边）。细化后伸缩命中区小、不再抢占边条的移动手势。
            关闭钮（基座 render、z-20）在最上不受影响；内容区在 padding 内侧不被覆盖。 */}
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
