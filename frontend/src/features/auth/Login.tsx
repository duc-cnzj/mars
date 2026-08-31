import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { setState, isRandomBg, toggleRandomBg } from '@/api/token'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/shadcn/tabs'
import { Icon } from '@/components/Icons'
import marsLogo from '@/assets/marslogo.png'
import { useAuth } from './AuthProvider'
import type { components } from '@/api/schema'

type SettingsResponse = components['schemas']['auth.SettingsResponse']
type OidcSetting = components['schemas']['auth.SettingsResponse_OidcSetting']
type BackgroundResponse = components['schemas']['picture.BackgroundResponse']

const LOGIN_TYPE_KEY = 'mars_login_type'

/** 读取上次登录方式偏好：有 SSO 时默认 SSO */
const getSavedLoginType = (): 'sso' | 'password' => {
  const saved = localStorage.getItem(LOGIN_TYPE_KEY)
  return saved === 'password' ? 'password' : 'sso'
}

const saveLoginType = (type: 'sso' | 'password') => {
  localStorage.setItem(LOGIN_TYPE_KEY, type)
}

/**
 * 登录页：随机壁纸 + 固定 pin + 背景版权角标；
 * 账号密码 + SSO 两个 Tab（SSO 仅在 /api/auth/settings 启用时展示），登录方式记忆。
 * 密码登录失败 toast"用户名或密码不正确"，成功后跳回原路径（含 query）。
 */
export function Login() {
  const { t } = useTranslation()
  const { signin } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [tab, setTab] = useState<'password' | 'sso'>('password')
  const [ssoItems, setSsoItems] = useState<OidcSetting[]>([])
  const [bgInfo, setBgInfo] = useState<BackgroundResponse>()
  const [random, setRandom] = useState(isRandomBg())

  const from = (location.state as { from?: { pathname: string; search: string } } | null)?.from

  /** 拉取背景图（url + copyright）；random 传后端（random_bg=1 随机挑图，否则固定当天首图，对齐旧版） */
  const fetchBg = useCallback(() => {
    void api
      .GET(API.pictureBackground, {
        params: { query: { random: isRandomBg() } },
      })
      .then(({ data }) => {
        if (data) setBgInfo(data)
      })
  }, [])

  useEffect(() => {
    fetchBg()
    // 拉取 SSO 配置（若启用则展示 OIDC 登录 Tab），并恢复登录方式记忆
    void api.GET(API.authSettings).then(({ data }) => {
      if (!data) return
      const items = (data as SettingsResponse).items.filter((i) => i.enabled)
      setSsoItems(items)
      setTab(items.length > 0 ? getSavedLoginType() : 'password')
    })
  }, [fetchBg])

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!username || !password || submitting) return
    setSubmitting(true)
    try {
      await signin(username, password)
      // 跳回原路径（含 query）；无来源则回首页
      navigate(from ? `${from.pathname}${from.search}` : '/', { replace: true })
    } catch {
      toast.error(t('auth.loginFailed'))
    } finally {
      setSubmitting(false)
    }
  }

  const onSso = (item: OidcSetting) => {
    // 记录 state 供回调校验（防 CSRF）
    setState(item.state)
    window.location.href = item.url
  }

  /** 固定/取消固定当前壁纸（对齐旧版：仅切换状态并写入 localStorage，背景下次加载时按新状态生效） */
  const togglePin = () => {
    setRandom(toggleRandomBg())
  }

  const passwordForm = (
    <form onSubmit={onSubmit} className="space-y-4">
      <label className="block">
        <span className="mb-1.5 block text-[12px] text-mute">{t('auth.username')}</span>
        <Input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          autoComplete="username"
          placeholder={t('auth.username')}
          autoFocus
        />
      </label>
      <label className="block">
        <span className="mb-1.5 block text-[12px] text-mute">{t('auth.password')}</span>
        <Input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="current-password"
          placeholder={t('auth.password')}
        />
      </label>
      <Button
        type="submit"
        variant="default"
        className="btn-login h-10 w-full rounded-lg text-[14px] font-semibold text-primary-foreground"
        disabled={submitting}
      >
        {submitting && <Icon name="loader" className="size-4 animate-spin" />}
        {t('auth.signIn')}
      </Button>
    </form>
  )

  const ssoForm = (
    <div className="mt-4 space-y-2">
      {ssoItems.map((item) => (
        <Button key={item.name} variant="outline" className="w-full" onClick={() => onSso(item)}>
          <Icon name="key" className="text-[14px]" />
          {item.name}
        </Button>
      ))}
    </div>
  )

  return (
    <div
      className="relative flex min-h-screen items-center justify-center bg-bg px-4"
      style={{
        backgroundImage: bgInfo?.url ? `url(${bgInfo.url})` : undefined,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }}
    >
      {/* 壁纸固定 pin */}
      <button
        type="button"
        onClick={togglePin}
        title={random ? t('auth.pinWallpaper') : t('auth.unpinWallpaper')}
        className="absolute right-5 top-5 flex h-9 w-9 items-center justify-center rounded-full bg-black/20 text-white/70 backdrop-blur transition-colors hover:bg-black/30 hover:text-white"
      >
        {random ? <Icon name="pin-off" className="size-[18px]" /> : <Icon name="pin" className="size-[18px]" />}
      </button>

      <div className="login-card w-full max-w-sm rounded-xl border border-line bg-surface p-8 shadow-[0_16px_50px_rgba(0,0,0,0.08)]">
        {/* 品牌区：旧版 Mars logo + dank mono 字标 */}
        <div className="mb-8 flex items-center justify-center gap-3">
          <img src={marsLogo} alt="Mars" className="h-11 w-11 rounded-xl" />
          <span
            className="text-[28px] text-ink"
            style={{
              fontFamily: '"dank mono", ui-monospace, monospace',
              fontWeight: 800,
              WebkitTextStroke: '0.6px currentColor',
            }}
          >
            Mars
          </span>
        </div>

        {ssoItems.length > 0 ? (
          <div className="flex flex-col gap-4">
            <Tabs
              value={tab}
              onValueChange={(v) => {
                setTab(v as 'password' | 'sso')
                saveLoginType(v as 'password' | 'sso')
              }}
            >
              <TabsList className="w-full">
                <TabsTrigger value="password" className="flex-1">
                  {t('auth.passwordLogin')}
                </TabsTrigger>
                <TabsTrigger value="sso" className="flex-1">
                  {t('auth.ssoLogin')}
                </TabsTrigger>
              </TabsList>
            </Tabs>
            {tab === 'password' ? passwordForm : ssoForm}
          </div>
        ) : (
          passwordForm
        )}
      </div>

      {/* 背景版权角标 */}
      {bgInfo?.copyright && (
        <div className="absolute bottom-6 left-1/2 -translate-x-1/2 rounded-full bg-black/30 px-5 py-2 text-[13px] text-white/80 backdrop-blur sm:left-auto sm:right-10 sm:translate-x-0">
          {bgInfo.copyright}
        </div>
      )}
    </div>
  )
}
