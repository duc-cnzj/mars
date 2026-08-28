import { lazy, Suspense, useLayoutEffect } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { Toaster } from 'sonner'
import 'sonner/dist/styles.css' // sonner 官方默认样式（明暗主题变量 + 类型色）
import './toast.css' // toast 主题化设计：与 15 个系统主题语义 token 联动（玻璃表面 + 类型色相）
import { Spinner } from './components/ui'
import { ErrorBoundary } from './components/ErrorBoundary'
import { AuthProvider, GuestRoute } from './features/auth/AuthProvider'
import { themeClass, themes } from './themes'
import { useTheme } from './hooks/useTheme'
import { useThemeHotkeys } from './hooks/useThemeHotkeys'

// 登录后的应用壳（RequireAuth + WebSocket + AppLayout）懒加载：登录页只渲染 GuestRoute/Login，
// 不需要顶栏/底栏/集群状态/WebSocket/protobuf 这一坨，懒加载后这些依赖不进登录页。
const AppShell = lazy(() => import('./layout/AppShell').then((m) => ({ default: m.AppShell })))

// 路由级按需加载：登录/工作台/仓库/事件/令牌各自成 chunk，首屏只拉关键路径。
const Login = lazy(() => import('./features/auth/Login').then((m) => ({ default: m.Login })))
const AuthCallback = lazy(() => import('./features/auth/AuthCallback').then((m) => ({ default: m.AuthCallback })))
const Workbench = lazy(() => import('./features/workbench/Workbench').then((m) => ({ default: m.Workbench })))
const Repos = lazy(() => import('./features/repos/Repos').then((m) => ({ default: m.Repos })))
const Events = lazy(() => import('./features/events/Events').then((m) => ({ default: m.Events })))
const AccessTokenManager = lazy(() =>
  import('./features/tokens/AccessTokenManager').then((m) => ({ default: m.AccessTokenManager })),
)
const NotFound = lazy(() => import('./pages/NotFound').then((m) => ({ default: m.NotFound })))
const ResourceBoardDemo = lazy(() =>
  import('./features/admin/ResourceBoardDemo').then((m) => ({ default: m.ResourceBoardDemo })),
)

import './themes/seiji.css'
import './themes/magenta.css'
import './themes/latte.css'
import './themes/mint.css'
import './themes/lavender.css'
import './themes/cherry.css'
import './themes/violet.css'
import './themes/lime.css'

/** 应用根：主题作用域包裹路由，认证 + Toast + 布局全在这层编排 */
export default function App() {
  const { theme, setTheme } = useTheme()
  // 主题快捷键：Ctrl/⌘ + Shift + ，/ 。按注册顺序循环切换（编辑态不接管）
  useThemeHotkeys({ theme, setTheme })

  // 应用主题有明暗之分，换算成 sonner 的明暗模式（dark 深色饱和渐变 / light 浅色粉彩）
  const toastTheme = themes.find((t) => t.id === theme)?.mode ?? 'dark'

  // shadcn 门户组件（Dialog/Dropdown/Select）渲染在 body 层级，
  // 主题 CSS 变量必须同步挂到 body，否则门户元素拿不到变量（透明背景/无样式）。
  // 用 useLayoutEffect 保证在绘制前生效，避免主题闪烁。
  useLayoutEffect(() => {
    const cls = themeClass(theme)
    document.body.classList.add(cls)
    return () => document.body.classList.remove(cls)
  }, [theme])

  return (
    <div className={`h-screen ${themeClass(theme)}`}>
      <BrowserRouter>
        {/* sonner Toaster：挂在主题类 div 内，--nx-* 主题变量可直接级联到 toast 元素 */}
        <Toaster
          richColors
          theme={toastTheme}
          position="bottom-right"
          duration={3000}
          visibleToasts={3}
        />
        <AuthProvider>
          <ErrorBoundary>
            <Suspense
              fallback={
                <div className="flex h-screen items-center justify-center">
                  <Spinner />
                </div>
              }
            >
              <Routes>
                {/* 认证：登录 / OIDC 回调（已登录访问 /login 重定向回首页） */}
                <Route
                  path="/login"
                  element={
                    <GuestRoute>
                      <Login />
                    </GuestRoute>
                  }
                />
                <Route path="/auth/callback" element={<AuthCallback />} />

                {/* 受保护：应用壳 + WebSocket 实时通道（认证后常驻，驱动终端/部署/集群信息） */}
                <Route
                  element={
                    <AppShell theme={theme} onSelectTheme={setTheme} />
                  }
                >
                  <Route index element={<Workbench />} />
                  <Route path="repos" element={<Repos />} />
                  <Route path="events" element={<Events />} />
                  <Route path="tokens" element={<AccessTokenManager />} />
                  {/* 旧 URL 重定向：/access_token_manager 保留兼容老书签 */}
                  <Route path="access_token_manager" element={<Navigate to="/tokens" replace />} />
                </Route>

                {/* 空间资源 demo（临时路由，未接真实后端） */}
                <Route path="/demo/resources" element={<ResourceBoardDemo />} />

                {/* 兜底 404 */}
                <Route path="*" element={<NotFound />} />
              </Routes>
            </Suspense>
          </ErrorBoundary>
        </AuthProvider>
      </BrowserRouter>
    </div>
  )
}
