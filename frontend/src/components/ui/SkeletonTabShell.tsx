import { Skeleton } from './shadcn/skeleton'

/**
 * 命令行骨架：占位结构与 TabShell 对齐（容器单选行 + 工具条 + 深色终端面板），
 * 用于容器列表加载态，避免加载切数据时跳动（reserve space, no content jumping）。
 * 深色面板里的占位条用半透明白，保证在 bg-black/85 上可见。
 */
export function SkeletonTabShell() {
  return (
    <div className="flex h-full min-h-0 flex-col gap-2">
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

      {/* 工具条：新增终端 + 分隔线 + 上传/下载/强杀 + 右侧资源用量图表（按钮 xs 高 h-6） */}
      <div className="flex flex-wrap items-center gap-1">
        <Skeleton className="h-6 w-20 rounded-md" />
        <Skeleton className="mx-0.5 h-3.5 w-px" />
        <Skeleton className="h-6 w-24 rounded-md" />
        <Skeleton className="h-6 w-24 rounded-md" />
        <Skeleton className="h-6 w-16 rounded-md" />
        <div className="ml-auto flex min-w-0 flex-1 items-stretch justify-end gap-1">
          <Skeleton className="h-6 w-48 rounded-md" />
        </div>
      </div>

      {/* 深色终端面板：与真实终端同高（flex-1 随弹窗自适应），默认单格占位 */}
      <div className="flex min-h-[160px] flex-1 flex-col overflow-hidden rounded-lg border border-line bg-black/85">
        {/* 终端标题栏：terminal 图标 + 命名空间/pod/容器路径 */}
        <div className="flex items-center gap-2 border-b border-white/10 px-3 py-1.5">
          <Skeleton className="size-3 rounded-sm bg-white/20" />
          <Skeleton className="h-2.5 w-44 bg-white/20" />
        </div>
        {/* 终端内容区：命令行占位行 */}
        <div className="flex flex-col gap-2.5 px-4 py-3">
          {Array.from({ length: 6 }, (_, i) => (
            <Skeleton
              key={i}
              className="h-3 bg-white/20"
              style={{ width: `${70 - (i % 3) * 16}%` }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
