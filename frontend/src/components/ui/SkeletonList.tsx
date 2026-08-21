import { Skeleton } from './shadcn/skeleton'

/** 列表行骨架：与事件/文件等"分隔行列表"等高的占位（对齐实际列表结构而非卡片） */
export function SkeletonList({ count = 8, bare = false }: { count?: number; bare?: boolean }) {
  if (bare) {
    return (
      <div className="divide-y divide-line">
        {Array.from({ length: count }, (_, i) => (
          <div key={i} className="flex items-start justify-between gap-4 px-5 py-3.5">
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2">
                <Skeleton className="h-4 w-20" />
                <Skeleton className="h-[22px] w-14 rounded-full" />
                <Skeleton className="h-3 w-28" />
              </div>
              <Skeleton className="mt-2.5 h-3 w-3/4" />
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <Skeleton className="h-8 w-20 rounded-md" />
            </div>
          </div>
        ))}
      </div>
    )
  }
  return (
    <div className="divide-y divide-line overflow-hidden rounded-lg border border-line bg-surface">
      {Array.from({ length: count }, (_, i) => (
        <div key={i} className="flex items-start justify-between gap-4 px-5 py-3.5">
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-[22px] w-14 rounded-full" />
              <Skeleton className="h-3 w-28" />
            </div>
            <Skeleton className="mt-2.5 h-3 w-3/4" />
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <Skeleton className="h-8 w-20 rounded-md" />
          </div>
        </div>
      ))}
    </div>
  )
}
