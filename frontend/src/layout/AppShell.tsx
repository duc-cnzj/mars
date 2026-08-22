import { RequireAuth } from '@/features/auth/AuthProvider'
import { ProvideWebsocket } from '@/hooks/useWebsocket'
import { AppLayout } from './AppLayout'
import type { ThemeId } from '@/themes'

/**
 * 登录后的应用壳（受保护路由的外壳）：认证守卫 + WebSocket 实时通道 + 整体布局。
 * App.tsx 对它是 lazy 引入——应用壳（顶栏/底栏/集群状态/WebSocket/protobuf）只在
 * 通过认证进入主界面时才下载，登录页不加载这一坨。
 */
export function AppShell({
  theme,
  onSelectTheme,
}: {
  theme: ThemeId
  onSelectTheme: (t: ThemeId) => void
}) {
  return (
    <RequireAuth>
      <ProvideWebsocket>
        <AppLayout theme={theme} onSelectTheme={onSelectTheme} />
      </ProvideWebsocket>
    </RequireAuth>
  )
}
