import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { DiffViewer } from '@/components/DiffViewer'
import { EventTitle } from './EventTitle'

/**
 * 事件改动查看：全屏宽弹窗（还原旧版 Drawer width="100%"），
 * 内容区滚动，diff 沿用旧版 react-diff-viewer 组件。
 * 标题统一走 EventTitle（主题色用户名 + 常规字重灰色消息）。
 */
export function DiffModal({
  open,
  onClose,
  username,
  message,
  oldText,
  newText,
}: {
  open: boolean
  onClose: () => void
  username: string
  message: string
  oldText: string
  newText: string
}) {
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="!h-[calc(100vh-2rem)] !w-[calc(100vw-2rem)] !max-w-[calc(100vw-2rem)] !grid-rows-[auto_minmax(0,1fr)] p-4">
        <DialogHeader className="shrink-0">
          <DialogTitle className="flex items-center">
            <EventTitle username={username} message={message} />
          </DialogTitle>
        </DialogHeader>
        <div className="min-h-0 overflow-auto overscroll-contain">
          {/* 对齐旧版 events：改动（old 非空）分屏，纯增/删（old 为空）合并视图，
              避免窄列把行折行多出空行 */}
          <DiffViewer
            oldValue={oldText}
            newValue={newText}
            language="yaml"
            initialView={oldText && oldText !== '' ? 'split' : 'unified'}
            // h-full：填满受限 body（grid 行 1fr），由 DiffViewer 内部 grow 容器滚动
            className="h-full"
          />
        </div>
      </DialogContent>
    </Dialog>
  )
}
