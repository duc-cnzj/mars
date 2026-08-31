import { Outlet, useLocation } from 'react-router-dom'
import { Topbar } from './Topbar'
import { Footer } from './Footer'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import type { ThemeId } from '@/themes'

/**
 * 应用整体布局（旧版风格）：全宽渐变顶栏 + 居中内容区 + 渐变底栏。
 * 无侧边栏，导航收在顶栏用户下拉，品牌字标点击回工作台。
 *
 * 页面随文档整体滚动，footer 位于内容末尾（不固定到底部）——对齐旧版 antd Layout 行为。
 * 需要容器内滚动的页面（events/tokens）自足高度 calc(100dvh - 顶栏 - padding)，
 * 不依赖布局提供有界高度；仓库管理已迁入管理后台（/admin/repos），由 AdminLayout 提供有界滚动容器。
 * 管理后台（/admin/*）为控制台形态：隐藏底栏，文档高度锁定 100dvh，侧栏导航与面包屑常驻固顶。
 * 侧边距响应式收窄（50px 固定值在小屏会吃掉近半屏）。
 */
export function AppLayout({
  theme,
  onSelectTheme,
}: {
  theme: ThemeId
  onSelectTheme: (t: ThemeId) => void
}) {
  const location = useLocation()
  // 管理后台（/admin/*）是控制台形态：隐藏营销底栏，文档高度恰好锁死 100dvh 不产生页面滚动，
  // AdminLayout 的侧栏导航 + 面包屑才能真正常驻固顶（否则 footer 会把文档撑出视口高，导航随页漂移）。
  const isAdmin = location.pathname.startsWith('/admin')

  return (
    // 页面随文档整体滚动，footer 位于内容末尾（不固定到底部）——对齐旧版 antd Layout 行为
    <div className="flex min-h-screen flex-col bg-bg text-ink">
      <Topbar theme={theme} onSelectTheme={onSelectTheme} />
      <main className="flex-1 px-4 pt-6 pb-3 sm:px-6 lg:px-10">
        <div className="h-full w-full">
          {/* 错误边界按页面出口重建（key=pathname）：页面崩溃切页即恢复，
              但不卸载应用壳——Topbar/ClusterStatus/websocket 常驻，切页不重新拉取 */}
          <ErrorBoundary key={location.pathname}>
            <Outlet />
          </ErrorBoundary>
        </div>
      </main>
      {/* 管理后台无底栏（控制台形态）；工作台/事件/令牌保留营销底栏 */}
      {!isAdmin && <Footer theme={theme} />}
    </div>
  )
}
