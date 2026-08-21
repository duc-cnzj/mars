import { Outlet, useLocation } from 'react-router-dom'
import { Topbar } from './Topbar'
import { Footer } from './Footer'
import { ErrorBoundary } from '../components/ErrorBoundary'
import type { ThemeId } from '../themes'

/**
 * 应用整体布局（旧版风格）：全宽渐变顶栏 + 居中内容区 + 渐变底栏。
 * 无侧边栏，导航收在顶栏用户下拉，品牌字标点击回工作台。
 *
 * 页面随文档整体滚动，footer 位于内容末尾（不固定到底部）——对齐旧版 antd Layout 行为。
 * 需要容器内滚动的页面（events/repos/tokens）自足高度 calc(100dvh - 顶栏 - padding)，
 * 不依赖布局提供有界高度。侧边距响应式收窄（50px 固定值在小屏会吃掉近半屏）。
 */
export function AppLayout({
  theme,
  onSelectTheme,
}: {
  theme: ThemeId
  onSelectTheme: (t: ThemeId) => void
}) {
  const location = useLocation()

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
      <Footer theme={theme} />
    </div>
  )
}
