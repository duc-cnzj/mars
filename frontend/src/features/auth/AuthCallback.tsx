import { useEffect, useRef, type ReactNode } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { getState, removeState, setToken } from '@/api/token'
import { toast } from '@/lib/toast'
import { markLoginSuccess, useAuth } from './AuthProvider'

/** 已成功换取 token 的 OIDC code（sessionStorage：跨整页刷新存活，用于跳过 URL 重放） */
const OIDC_DONE_KEY = 'mars_oidc_done'

/**
 * OIDC 回调页：用 code 换 token（POST /api/auth/exchange），
 * 校验 state 防止 CSRF，成功后写 token、跳主页（成功提示由 AuthProvider 统一弹）。
 */
export function AuthCallback() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  // 只用稳定引用：auth 是 context value，会随 user 变化重建——放进 effect 依赖会导致
  // refresh 成功 setUser 后 context 值变化 → effect 清理+重跑 → 重复 exchange
  // （OIDC code 单次有效则二次 exchange 失败跳登录页、表现为 SSO 登录失败；可复用则无限请求循环）。
  // refresh 本身是稳定的 useCallback。
  const { refresh } = useAuth()

  const code = params.get('code')
  const state = params.get('state')
  // 一次性消费锁：见下方 effect 内注释（StrictMode 双跑防护）
  const handledRef = useRef(false)

  useEffect(() => {
    // StrictMode（开发态）会把挂载 effect 连跑两遍（setup→cleanup→setup）：第一遍在首个
    // await 之前已同步 removeState()，第二遍再比对 getState() 必然失配 →
    // 误报「用户名或密码错误」并踢回登录页。ref 在双跑间保持不变，用它保证只处理一次。
    if (handledRef.current) return
    handledRef.current = true

    const exchange = async () => {
      if (!code) {
        navigate('/login', { replace: true })
        return
      }
      // 本 code 换过 token 了还要再进来一次，只可能是上一步 refresh 触发了整页刷新
      // （灰度通道对齐 alignGrayChannel 会 reload）、浏览器按原 URL 重放本页。此时一次性
      // state 早被消费，再校验必然「失配」→ 误报用户名或密码错误（生产环境同样会中招）。
      // 已换过就直接回家，不重复校验、不重复换。
      if (sessionStorage.getItem(OIDC_DONE_KEY) === code) {
        navigate('/', { replace: true })
        return
      }
      // state 不一致：拒绝，回登录
      if (state !== getState()) {
        toast.error(t('auth.loginFailed'))
        removeState()
        navigate('/login', { replace: true })
        return
      }
      removeState()
      const { data, error } = await api.POST(API.authExchange, { body: { code } })
      if (error || !data?.token) {
        navigate('/login', { replace: true })
        return
      }
      // 先落「已换过」标记再往下走：refresh 随时可能触发整页刷新，标记必须早于它
      sessionStorage.setItem(OIDC_DONE_KEY, code)
      setToken(data.token)
      // 打旗标而非直接弹 toast：refresh 内的灰度通道对齐可能整页刷新，直接弹会被刷掉
      markLoginSuccess()
      await refresh()
      navigate('/', { replace: true })
    }

    void exchange()
  }, [code, state, navigate, toast, t, refresh])

  return <LoginLoading />
}

/** 轻量加载态：登录中提示 */
function LoginLoading(): ReactNode {
  const { t } = useTranslation()
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-3 bg-bg">
      <span className="h-6 w-6 animate-spin rounded-full border-2 border-line border-t-primary" />
      <span className="text-[13px] text-mute">{t('auth.loggingIn')}</span>
    </div>
  )
}
