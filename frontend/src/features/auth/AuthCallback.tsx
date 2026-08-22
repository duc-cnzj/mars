import { useEffect, type ReactNode } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { api } from '@/api/client'
import { getState, removeState, setToken } from '@/api/token'
import { toast } from '@/lib/toast'
import { useAuth } from './AuthProvider'

/**
 * OIDC 回调页：用 code 换 token（POST /api/auth/exchange），
 * 校验 state 防止 CSRF，成功后写 token 并跳主页。
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

  useEffect(() => {
    let cancelled = false

    const exchange = async () => {
      if (!code) {
        navigate('/login', { replace: true })
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
      const { data, error } = await api.POST('/api/auth/exchange', { body: { code } })
      if (error || !data?.token) {
        navigate('/login', { replace: true })
        return
      }
      setToken(data.token)
      await refresh()
      if (!cancelled) navigate('/', { replace: true })
    }

    void exchange()
    return () => {
      cancelled = true
    }
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
