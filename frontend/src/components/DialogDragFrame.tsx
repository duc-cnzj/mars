import { useTranslation } from 'react-i18next'
import { useDraggableDialog, type ResizeDir } from '@/hooks/useDraggableDialog'

/** useDraggableDialog 的返回类型：外框只消费手势 props 与最大化状态。 */
type DraggableDialogState = ReturnType<typeof useDraggableDialog>

/**
 * 可拖拽弹窗的四边交互外框：铺在 DialogContent 内容之上的透明手势层。
 *
 * 为什么需要它：`useDraggableDialog` 的 `dragHandleProps` 只挂在标题栏时，
 * 用户只有按住标题才能移动窗口；旧版 DraggableModal 的手感是「整边可拖」。
 * 这里补齐两层：
 *   1. 四边 24px 移动边条：拖动移动窗口、双击全屏/还原（最大化时保留，拖动由 hook clamp）；
 *   2. 贴边界 6px 细沿 + 四角 16px 缩放把手：对边锚定，拖哪边只有哪边动；
 *      只非最大化时渲染（全屏无可伸缩边）。
 * 关闭钮由 DialogContent 基座渲染（z-20）在最上层，内容区在 padding 内侧，均不被覆盖。
 *
 * 项目详情 / 新建项目 / 控制台 / AI 对话四个弹窗共用本组件，手势与视觉口径一处维护。
 */
export function DialogDragFrame({ dnd }: { dnd: DraggableDialogState }) {
  const { t } = useTranslation()

  return (
    <>
      {MOVE_BANDS.map((position) => (
        <div
          key={position}
          aria-hidden
          title={t('project.dragTitle')}
          {...dnd.dragHandleProps}
          className={`pointer-events-auto absolute cursor-move ${position}`}
        />
      ))}
      {!dnd.isMaximized &&
        RESIZE_HANDLES.map(({ dir, corner, position }) => (
          <div
            key={dir}
            aria-hidden
            title={t(corner ? 'project.resizeCorner' : 'project.resizeEdge')}
            {...dnd.getResizeHandleProps(dir)}
            onDoubleClick={dnd.dragHandleProps.onDoubleClick}
            className={`pointer-events-auto absolute z-10 ${position}`}
          />
        ))}
    </>
  )
}

/** 四条 24px 移动边条：上下铺满宽度、左右铺满高度。 */
const MOVE_BANDS = [
  'inset-x-0 top-0 h-6',
  'inset-x-0 bottom-0 h-6',
  'inset-y-0 left-0 w-6',
  'inset-y-0 right-0 w-6',
] as const

/** 八向缩放把手：四边贴边 6px 细沿 + 四角 16px，方向决定哪边跟手。 */
const RESIZE_HANDLES: ReadonlyArray<{
  dir: ResizeDir
  corner: boolean
  position: string
}> = [
  { dir: 'n', corner: false, position: 'inset-x-0 top-0 h-1.5 cursor-ns-resize' },
  { dir: 's', corner: false, position: 'inset-x-0 bottom-0 h-1.5 cursor-ns-resize' },
  { dir: 'w', corner: false, position: 'inset-y-0 left-0 w-1.5 cursor-ew-resize' },
  { dir: 'e', corner: false, position: 'inset-y-0 right-0 w-1.5 cursor-ew-resize' },
  { dir: 'nw', corner: true, position: 'top-0 left-0 size-4 cursor-nwse-resize' },
  { dir: 'ne', corner: true, position: 'top-0 right-0 size-4 cursor-nesw-resize' },
  { dir: 'sw', corner: true, position: 'bottom-0 left-0 size-4 cursor-nesw-resize' },
  { dir: 'se', corner: true, position: 'bottom-0 right-0 size-4 cursor-nwse-resize' },
]
