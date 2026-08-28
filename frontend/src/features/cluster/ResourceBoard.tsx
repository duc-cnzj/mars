import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { RefreshFade } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { useResourceBoard, REFRESH_INTERVAL_MS } from './useResourceBoard'
import { OverviewCards } from './OverviewCards'
import { NodeTable } from './NodeTable'
import { NamespaceRank } from './NamespaceRank'
import { TopPods } from './TopPods'

/**
 * 集群资源看板（管理员后台）
 *
 * 集群总览（健康状态 / CPU / 内存 / 请求率）＋ 节点资源表 ＋ 命名空间排行 ＋ Top Pod 排行，
 * 数据由 useResourceBoard 每 REFRESH_INTERVAL_MS 轮询 /api/admin/cluster/board 提供，
 * 也可点刷新按钮手动拉新一版。
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
    refresh,
  } = useResourceBoard()

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

      {/* 内容区：进入/手动刷新重播渐入；轮询不 bump version → 每 3s 不闪屏 */}
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
  )
}
