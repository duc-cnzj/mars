/**
 * 分段控件骨架：外层静态盒（border/高度固定），内层 animate-pulse 呼吸填充。
 * 关键：填充左侧悬空 1px（md:ml-px + md:w-[calc(100%-1px)]）——
 * 后一段经 -ml-px 左叠 1px、其左 1px 正好盖住上一段的右边框（共享竖线），
 * 若填充顶满 w-full 会周期性盖住/露出该竖线造成闪烁；悬空 1px 后竖线全程可见且静态。
 * 高度 h-[38px] 对齐真实 trigger used height。创建/更新部署页共用。
 */
export function SegmentedSkeleton({ className }: { className?: string }) {
  return (
    <div className={`h-[38px] w-full overflow-hidden rounded-md border border-line-strong ${className ?? ''}`}>
      <div className="h-full w-full animate-pulse bg-raised md:ml-px md:w-[calc(100%-1px)]" />
    </div>
  )
}
