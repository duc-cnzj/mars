import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { RefreshFade, SkeletonList } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Skeleton } from '@/components/ui/shadcn/skeleton'
import { useResourceBoard, REFRESH_INTERVAL_MS } from './useResourceBoard'
import { OverviewCards } from './OverviewCards'
import { DeployTrendPanel } from './DeployTrendPanel'
import { NodeTable } from './NodeTable'
import { NamespaceRank } from './NamespaceRank'
import { TopPods } from './TopPods'

/**
 * 集群总览（管理员后台）
 *
 * 资源总览（健康状态 / CPU / 内存 / 请求率）＋ 每日部署趋势 ＋ 节点资源表 ＋ 命名空间排行 ＋ Top Pod 排行，
 * 资源数据由 useResourceBoard 每 REFRESH_INTERVAL_MS 轮询 /api/admin/cluster/board 提供，
 * 部署趋势由 useDeployTrend 拉 /api/admin/cluster/deploy_trend（日粒度、随手动刷新重挂载取新）。
 * 也可点刷新按钮手动拉新一版资源快照。
 */
export function ResourceBoard() {
  const { t } = useTranslation()
  const {
    overview,
    podCount,
    nodes,
    namespaces,
    pods,
    topSort,
    setTopSort,
    cpuTrend,
    memTrend,
    lastUpdate,
    version,
    refreshing,
    refetching,
    refresh,
  } = useResourceBoard()
  // 已有一版数据：首载（nodes 为空）走骨架屏；重取（手动刷新 / topSort 切换）时遮罩旧帧不闪断
  const hasData = nodes.length > 0

  return (
    <div className="flex flex-col gap-4">
      {/* 页头：标题 + 自动刷新状态 + 手动刷新 */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <h2 className="text-[16px] font-semibold text-ink">{t('cluster.title')}</h2>
          <span className="hidden items-center gap-1 text-[11px] text-faint sm:flex">
            <Icon name="clock" className="size-3" />
            {t('cluster.autoRefresh', { seconds: REFRESH_INTERVAL_MS / 1000 })}
          </span>
          {/* 后端聚合缓存每 30s 刷新一次，告知用户看到的快照最多滞后 30s */}
          <span className="hidden items-center gap-1 text-[11px] text-faint sm:flex">
            <Icon name="database" className="size-3" />
            {t('cluster.backendCache')}
          </span>
        </div>
        <div className="flex items-center gap-3">
          <span className="font-mono text-[11px] text-faint">
            {t('cluster.lastUpdate')}{' '}
            <span className="text-mute">{lastUpdate.toLocaleTimeString()}</span>
          </span>
          <Button variant="outline" size="sm" onClick={refresh} disabled={refreshing}>
            {refreshing ? (
              <Icon name="loader" className="size-3.5 animate-spin" />
            ) : (
              <Icon name="refresh" className="size-3.5" />
            )}
            {t('common.refresh')}
          </Button>
        </div>
      </div>

      {/* 内容区：重取遮罩时降透明度并禁止交互，保持旧帧不闪断（首载骨架不遮罩） */}
      <div className="relative">
        {!hasData ? (
          /* 首载骨架：卡片区 + 列表区占位（对齐看板实际布局：四卡 → 节点表 → 排行双列） */
          <div className="flex flex-col gap-4">
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {Array.from({ length: 4 }, (_, i) => (
                <div key={i} className="rounded-lg border border-line bg-surface p-4">
                  <Skeleton className="h-3 w-14" />
                  <Skeleton className="mt-3 h-6 w-20" />
                </div>
              ))}
            </div>
            {/* 部署趋势面板骨架占位（标题行 + 统计行 + 曲线区），与真数据布局同构避免首载跳动 */}
            <div className="rounded-lg border border-line bg-surface p-4">
              <Skeleton className="h-3 w-24" />
              <div className="mt-3 flex gap-8">
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-5 w-20" />
                <Skeleton className="h-5 w-16" />
              </div>
              <Skeleton className="mt-3 h-40 w-full" />
              <Skeleton className="mt-1.5 h-3 w-full" />
            </div>
            <SkeletonList count={5} />
            <div className="flex flex-col gap-4 lg:flex-row">
              <div className="flex-1">
                <SkeletonList count={5} />
              </div>
              <div className="flex-1">
                <SkeletonList count={5} />
              </div>
            </div>
          </div>
        ) : (
          <div className={refetching ? 'pointer-events-none opacity-40' : undefined}>
            {/* 进入/手动刷新重播渐入；轮询不 bump version → 周期刷新不闪屏 */}
            <RefreshFade version={version} className="flex flex-col gap-4">
              {/* 集群总览四卡 */}
              <div>
                <OverviewCards
                  overview={overview}
                  namespaceCount={namespaces.length}
                  podCount={podCount}
                  cpuTrend={cpuTrend}
                  memTrend={memTrend}
                />
              </div>

              {/* 每日部署趋势（deploy_trend 真数据） */}
              <div>
                <DeployTrendPanel />
              </div>

              {/* 节点资源表 */}
              <div>
                <NodeTable nodes={nodes} />
              </div>

              {/* 命名空间排行 + Top Pod 排行（宽屏双列，窄屏堆叠） */}
              <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
                <NamespaceRank namespaces={namespaces} />
                <TopPods pods={pods} topSort={topSort} onTopSortChange={setTopSort} />
              </div>
            </RefreshFade>
          </div>
        )}
        {/* 重取遮罩：手动刷新 / topSort 维度切换（用户操作，非轮询）时居中 spinner */}
        {refetching && hasData && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-surface/50">
            <Icon name="loader" className="size-5 animate-spin text-faint" />
          </div>
        )}
      </div>
    </div>
  )
}
