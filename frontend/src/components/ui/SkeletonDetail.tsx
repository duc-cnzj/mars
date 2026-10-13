import { Skeleton } from './shadcn/skeleton'

/**
 * 项目详情骨架：占位结构与 TabInfo 对齐（指标卡片行 + 分组区块 KV 行 + 底部删除按钮），
 * 用于项目详情弹窗加载态，避免加载切数据时跳动（参考 UX: reserve space, no content jumping）。
 */
export function SkeletonDetail() {
  return (
    <div className="flex flex-col gap-4">
      {/* 关键指标：2 张卡片（图标块 + 标签 + 值）——对齐 TabInfo（项目 ID 已移入基础信息，
          指标卡只剩 CPU / 内存，网格由 3 列收为 2 列） */}
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {Array.from({ length: 2 }, (_, i) => (
          <div
            key={i}
            className="flex items-center gap-2.5 rounded-lg border border-line bg-surface px-3 py-2.5"
          >
            <Skeleton className="h-8 w-8 shrink-0 rounded-md" />
            <div className="min-w-0 flex-1 space-y-1.5">
              <Skeleton className="h-3 w-12" />
              <Skeleton className="h-3.5 w-20" />
            </div>
          </div>
        ))}
      </div>

      {/* 分组区块：图标标题行 + KV 行（首个区块 5 行、其余 2 行，贴近基础信息/休眠/时间线）。
          首块 5 行对齐 TabInfo 基础信息区（id / repoId / version / gitProjectId / namespaceId）；
          次块（2 行）对齐 TabInfo 定时休眠区（时间窗 / 副本字段两行 KV，仓库启用休眠时才渲染，
          骨架按常见态预留）。共 4 块覆盖详情页主要分组，避免加载完成时区块数从少变多跳动 */}
      {Array.from({ length: 4 }, (_, s) => (
        <div key={s} className="flex flex-col gap-1.5">
          <div className="flex items-center gap-1.5">
            <Skeleton className="h-3.5 w-3.5 rounded" />
            <Skeleton className="h-3.5 w-24" />
          </div>
          <div className="flex flex-col gap-1 pl-0.5">
            {Array.from({ length: s === 0 ? 5 : 2 }, (_, r) => (
              <div key={r} className="flex items-baseline gap-2">
                <Skeleton className="h-3 w-14" />
                <Skeleton className="h-3 w-32" />
              </div>
            ))}
          </div>
        </div>
      ))}

      {/* 底部删除按钮 */}
      <div className="flex justify-end border-t border-line pt-4">
        <Skeleton className="h-8 w-28 rounded-md" />
      </div>
    </div>
  )
}
