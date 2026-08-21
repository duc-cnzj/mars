import { useEffect, type ReactNode } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { api } from '../../api/client'
import { getState, removeState, setToken } from '../../api/token'
import { toast } from '@/lib/toast'
import { useAuth } from './AuthContext'

/**
 * OIDC 回调页：用 code 换 token（POST /api/auth/exchange），
 * 校验 state 防止 CSRF，成功后写 token 并跳主页。
 */
export function AuthCallback() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const auth = useAuth()

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
      await auth.refresh?.()
      if (!cancelled) navigate('/', { replace: true })
    }

    void exchange()
    return () => {
      cancelled = true
    }
  }, [code, state, navigate, toast, t, auth])

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
