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
import { useTranslation } from 'react-i18next'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import {
  getToken,
  removeToken,
  setToken,
  setLogoutUrl,
  getLogoutUrl,
  removeLogoutUrl,
} from '@/api/token'
import { Spinner } from '@/components/ui'
import { toast } from '@/lib/toast'
import { alignGrayChannel, isReloadPending, setGrayChannel } from '@/hooks/useGrayChannel'
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

/** 「本次是显式登录成功」旗标键；存 sessionStorage 以便跨整页刷新存活 */
const LOGIN_TOAST_KEY = 'mars_login_success'

/**
 * 标记登录成功，由 AuthProvider 的 effect 统一弹提示（勿在登录函数里直接 toast.success）。
 *
 * 为什么绕这一道：loadUser 内的 alignGrayChannel 在灰度「意图」与本浏览器 cookie 不一致时会
 * 立刻 window.location.reload() 去换灰度通道——而 loadUser 正被登录流程 await，于是刚弹出的
 * toast 会连同整页 JS 内存一起被刷新清掉，表现为「登录成功但没有任何提示」。旗标落在
 * sessionStorage（唯一跨刷新存活的本标签页存储），刷新后会话恢复、user 落地时补弹一次，
 * 无论中途刷没刷新，用户都恰好看见一次。
 */
export function markLoginSuccess(): void {
  sessionStorage.setItem(LOGIN_TOAST_KEY, '1')
}

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
  const { t } = useTranslation()
  const [user, setUser] = useState<UserInfo | null>(null)
  const [loading, setLoading] = useState(true)

  /** 拉取当前用户信息并写回 state（无 token 直接置空） */
  const loadUser = useCallback(async (): Promise<UserInfo | null> => {
    if (!getToken()) {
      setUser(null)
      return null
    }
    const { data } = await api.GET(API.authInfo)
    if (data) {
      setUser(data)
      if (data.logoutUrl) setLogoutUrl(data.logoutUrl)
      // 启动对齐：把后端 isGray（灰度意图）落成本浏览器的灰度路由 cookie（灰度事实）。
      // 这是「意图 → 事实」之间唯一的那根线；与当前 cookie 不一致时会硬刷新一次
      //（cookie 只对下一次文档请求生效，不刷新就永远停在旧通道），详见 alignGrayChannel。
      alignGrayChannel(data.isGray)
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

  // 登录成功提示：由「user 落地」驱动而非登录函数内直接弹。会话建立时若带有登录旗标就补弹
  // 一次并立即清旗标——无论中途是否被 alignGrayChannel 的整页刷新打断，都恰好提示一次；
  // 而普通刷新页面（无旗标）不会被误弹。详见 markLoginSuccess。
  // isReloadPending 必须先判：reload 在路上的话这个 effect 仍会在旧文档上跑一遍，此时若把旗标
  // 消费掉，toast 会随刷新一起消失且旗标没了、新文档不再补弹——必须把旗标留给新文档。
  useEffect(() => {
    if (!user) return
    if (isReloadPending()) return
    if (sessionStorage.getItem(LOGIN_TOAST_KEY) !== '1') return
    sessionStorage.removeItem(LOGIN_TOAST_KEY)
    toast.success(t('auth.loginSuccess'))
  }, [user, t])

  /** 账号密码登录：成功即写 token 并恢复会话；失败抛错（由 Login 弹"用户名或密码不正确"） */
  const signin = useCallback(
    async (username: string, password: string): Promise<UserInfo> => {
      const { data, error } = await api.POST(API.authLogin, {
        body: { username, password },
      })
      if (error || !data?.token) throw new Error('login failed')
      setToken(data.token)
      // 旗标必须在 loadUser 之内 setUser 之前打：toast 由「user 落地」的 effect 触发，
      // 打晚了 effect 已经跑过，就永远不弹
      markLoginSuccess()
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
    // 登出即失去灰度意图：清掉灰度路由 cookie，避免下一个登录者继承上一个会话的通道。
    // 紧跟着就是整页跳转（下一次文档请求），故无需再触发对齐刷新。
    setGrayChannel(false)
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

/**
 * 超级管理员守卫：系统设置路由级门控（is_super_admin 来自 /api/auth/info）。
 * 嵌套在 RequireAdmin 内（已登录且为管理员），仅内置超管放行，普通管理员重定向回首页，
 * 防止直接敲 URL 访问系统设置（可见性与可访问性双保险）。
 */
export function RequireSuperAdmin({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <AuthLoading />
  if (!user?.isSuperAdmin) return <Navigate to="/" replace />
  return <>{children}</>
}
