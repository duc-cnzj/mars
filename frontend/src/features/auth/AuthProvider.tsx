import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { api } from '@/api/client'
import {
  getToken,
  removeToken,
  setToken,
  setLogoutUrl,
  getLogoutUrl,
  removeLogoutUrl,
} from '@/api/token'
import { Spinner } from '@/components/ui'
import type { components } from '@/api/schema'

type UserInfo = components['schemas']['auth.InfoResponse']

interface AuthCtxValue {
  user: UserInfo | null
  loading: boolean
  signin: (username: string, password: string) => Promise<UserInfo>
  signout: () => void
  refresh: () => Promise<UserInfo | null>
}

const Ctx = createContext<AuthCtxValue | null>(null)

/** 全局加载态（RequireAuth / GuestRoute 在会话恢复期间的占位） */
function AuthLoading() {
  return (
    <div className="flex h-screen items-center justify-center bg-bg">
      <Spinner />
    </div>
  )
}

/**
 * 认证提供者：负责登录、用户信息拉取、登出。
 * 初始挂载时若本地有 token 则自动恢复会话（GET /api/auth/info）；
 * 恢复失败时清掉无效 token（避免无限回跳登录页）。
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserInfo | null>(null)
  const [loading, setLoading] = useState(true)

  /** 拉取当前用户信息并写回 state（无 token 直接置空） */
  const loadUser = useCallback(async (): Promise<UserInfo | null> => {
    if (!getToken()) {
      setUser(null)
      return null
    }
    const { data } = await api.GET('/api/auth/info')
    if (data) {
      setUser(data)
      if (data.logoutUrl) setLogoutUrl(data.logoutUrl)
      return data
    }
    // 会话恢复失败：清除无效 token，交给守卫回登录页
    if (getToken()) removeToken()
    setUser(null)
    return null
  }, [])

  useEffect(() => {
    // 网络失败时 loadUser 会 reject，捕获避免未处理 rejection；
    // 失败后 user 保持 null、loading 置 false，由 RequireAuth 交回登录页
    void loadUser()
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [loadUser])

  /** 账号密码登录：成功即写 token 并恢复会话；失败抛错（由 Login 弹"用户名或密码不正确"） */
  const signin = useCallback(
    async (username: string, password: string): Promise<UserInfo> => {
      const { data, error } = await api.POST('/api/auth/login', {
        body: { username, password },
      })
      if (error || !data?.token) throw new Error('login failed')
      setToken(data.token)
      const info = await loadUser()
      if (!info) throw new Error('login failed')
      return info
    },
    [loadUser],
  )

  /** 登出：清 token，跳 SSO 登出地址或登录页 */
  const signout = useCallback(() => {
    removeToken()
    setUser(null)
    const url = getLogoutUrl() || '/login'
    removeLogoutUrl()
    window.location.href = url
  }, [])

  const value = useMemo(
    () => ({ user, loading, signin, signout, refresh: loadUser }),
    [user, loading, signin, signout, loadUser],
  )

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

/** 取认证上下文（必须在 AuthProvider 内） */
export function useAuth(): AuthCtxValue {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

/** 路由守卫：未登录重定向 /login（携带原路径 state.from），加载中显示骨架 */
export function RequireAuth({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  const location = useLocation()
  if (loading) return <AuthLoading />
  if (!user) return <Navigate to="/login" replace state={{ from: location }} />
  return <>{children}</>
}

/** 访客守卫：已登录访问 /login 时重定向回首页 */
export function GuestRoute({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <AuthLoading />
  if (user) return <Navigate to="/" replace />
  return <>{children}</>
}

/**
 * 管理员守卫：管理后台路由级门控（mars_admin）。
 * 嵌套在 RequireAuth 内（用户已登录），仅校验角色；非管理员重定向回首页，
 * 防止直接敲 URL 绕过下拉入口的可视化隐藏（可见性与可访问性双保险）。
 */
export function RequireAdmin({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <AuthLoading />
  if (!user?.roles.includes('mars_admin')) return <Navigate to="/" replace />
  return <>{children}</>
}
