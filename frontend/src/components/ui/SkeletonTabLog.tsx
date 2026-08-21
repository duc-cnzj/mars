import { Skeleton } from './shadcn/skeleton'

/**
 * 容器日志骨架：占位结构与 TabLog 对齐（容器单选行 + 工具条 + 深色日志面板），
 * 用于容器列表加载态，避免加载切数据时跳动（reserve space, no content jumping）。
 * 深色面板里的占位条用半透明白，保证在 bg-black/85 上可见。
 */
export function SkeletonTabLog() {
  return (
    <div className="flex h-full min-h-0 flex-col gap-3">
      {/* 容器单选行：radio 圆点 + 容器名 + 状态胶囊 */}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5">
        {Array.from({ length: 3 }, (_, i) => (
          <div key={i} className="flex items-center gap-1.5">
            <Skeleton className="size-3.5 shrink-0 rounded-full" />
            <Skeleton className="h-3.5 w-24" />
            <Skeleton className="h-5 w-28 rounded-md" />
          </div>
        ))}
      </div>

      {/* 工具条：路径 + 搜索框 + 状态 + 下载按钮 */}
      <div className="flex flex-wrap items-center gap-2">
        <Skeleton className="h-3 w-44" />
        <Skeleton className="h-7 w-60 rounded-md" />
        <Skeleton className="h-3 w-14" />
        <div className="ml-auto flex items-center gap-1.5">
          <Skeleton className="h-7 w-24 rounded-md" />
        </div>
      </div>

      {/* 深色日志面板：与真实日志区同高（flex-1 随弹窗自适应），逐行占位 */}
      <div className="min-h-[240px] flex-1 overflow-hidden rounded-md bg-black/85">
        <div className="flex flex-col gap-2.5 px-4 py-3">
          {Array.from({ length: 7 }, (_, i) => (
            <Skeleton
              key={i}
              className="h-3 bg-white/20"
              style={{ width: `${50 - (i % 4) * 12}%` }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
