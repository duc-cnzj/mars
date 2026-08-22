import { Skeleton } from './shadcn/skeleton'
import { SegmentedSkeleton } from '@/components/ui/SegmentedSkeleton'

/**
 * 配置更新骨架：占位结构与 TabEdit 对齐（吸顶头部 pipeline 槽 + 三段 select 分段控件 + 按钮行；
 * 内容区部署参数占位 + 配置编辑器大块），用于项目详情弹窗加载态，避免加载切数据时跳动
 * （reserve space, no content jumping）。
 * 此前 edit tab 加载误用 SkeletonDetail（指标卡片/KV 分组/删除按钮），结构与配置编辑页完全不符，
 * 加载中「配置」区看起来缺失。本骨架按 TabEdit 真实结构铺位。
 */
export function SkeletonTabEdit() {
  return (
    <div className="flex h-full min-h-0 flex-col">
      {/* 吸顶头部：pipeline 占位槽 + 三段 select 分段控件 + 操作按钮行 */}
      <div className="z-10 shrink-0 border-b border-line bg-bg">
        <div className="space-y-2 px-1 pb-2 pt-1">
          {/* pipeline 槽：与 PipelineInfo/占位横幅等高（42px）的骨架条（图标 + 状态文案 + 右侧刷新） */}
          <div className="flex min-h-[42px] items-center gap-2 rounded-md border border-line bg-surface px-3 py-2">
            <Skeleton className="size-4 shrink-0 rounded-full" />
            <Skeleton className="h-3.5 w-32" />
            <Skeleton className="ml-auto size-6 shrink-0 rounded-md" />
          </div>
          <div className="grid grid-cols-1 gap-2 md:grid-cols-3 md:gap-0">
            <SegmentedSkeleton className="md:rounded-r-none" />
            <SegmentedSkeleton className="md:-ml-px md:rounded-none md:border-l-0" />
            <SegmentedSkeleton className="md:-ml-px md:rounded-l-none md:border-l-0" />
          </div>
          {/* 操作按钮行：部署/重置/历史（Button xs 高，占位宽 24 对齐） */}
          <div className="flex flex-wrap items-center gap-2">
            <Skeleton className="h-7 w-24 rounded-md" />
            <Skeleton className="h-7 w-24 rounded-md" />
            <Skeleton className="h-7 w-24 rounded-md" />
          </div>
        </div>
      </div>

      {/* 内容区：部署参数占位（3 列字段行）+ 配置编辑器大块（flex-1 随弹窗自适应） */}
      <div className="flex min-h-0 flex-1 flex-col overflow-hidden p-1">
        <div className="mb-3 grid grid-cols-1 gap-2 sm:grid-cols-2 md:grid-cols-3">
          {Array.from({ length: 3 }, (_, i) => (
            <div key={i} className="flex items-center gap-2">
              <Skeleton className="h-3 w-16 shrink-0" />
              <Skeleton className="h-8 min-w-0 flex-1 rounded-md" />
            </div>
          ))}
        </div>
        <div className="flex min-h-[300px] flex-1 flex-col overflow-hidden rounded-md border border-line bg-surface">
          <div className="flex flex-col gap-2.5 px-3 py-3">
            {Array.from({ length: 8 }, (_, i) => (
              <Skeleton
                key={i}
                className="h-3"
                style={{ width: `${85 - (i % 4) * 14}%` }}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
