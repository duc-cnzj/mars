import { Skeleton } from './shadcn/skeleton'

/**
 * 卡片网格骨架：占位结构与工作台卡片对齐（图标块 + 标题行 + 内容行 + 底部条），
 * gap-3 与页面网格一致，避免加载切数据时跳动（参考 UX: reserve space, no content jumping）。
 */
export function SkeletonGrid({ count = 6 }: { count?: number }) {
  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: count }, (_, i) => (
        <div
          key={i}
          className="flex h-full flex-col gap-3 rounded-lg border border-line bg-surface p-4"
        >
          {/* 头部：左侧图标块 + 右侧标题/描述行 */}
          <div className="flex items-start gap-2.5">
            <Skeleton className="h-9 w-9 shrink-0" />
            <div className="min-w-0 flex-1 space-y-2">
              <Skeleton className="h-4 w-1/2" />
              <Skeleton className="h-3 w-2/3" />
            </div>
          </div>
          {/* 内容行占位 */}
          <div className="space-y-2">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-5/6" />
          </div>
          {/* 底部条（成员/计数） */}
          <div className="mt-auto flex items-center justify-between border-t border-line pt-3">
            <Skeleton className="h-5 w-20" />
            <Skeleton className="h-5 w-12" />
          </div>
        </div>
      ))}
    </div>
  )
}
