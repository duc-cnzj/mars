import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import marsLogo from '../assets/marslogo.png'
import { ThemeSwitcher } from '../components/ui/ThemeSwitcher'
import { useVersion } from '../hooks/useVersion'
import { getLocale, setLocale, type Locale } from '../i18n'
import { barGradient, type ThemeId } from '../themes'
import { ClusterStatus } from './ClusterStatus'
import { UserMenu } from './UserMenu'

/**
 * 渐变顶栏（还原旧版 AppHeader 布局）：无侧边栏，
 * 左 logo + "Mars" 字标 + 版本号，右 集群状态灯 + 主题/语言切换 + 用户下拉。
 */
export function Topbar({
  theme,
  onSelectTheme,
}: {
  theme: ThemeId
  onSelectTheme: (t: ThemeId) => void
}) {
  const { t } = useTranslation()
  const [locale, setLocaleState] = useState<Locale>(getLocale())
  const version = useVersion()

  const switchLocale = () => {
    const next: Locale = locale === 'zh-CN' ? 'en' : 'zh-CN'
    setLocaleState(next)
    setLocale(next)
  }

  return (
    <header
      className="sticky top-0 z-20 flex h-16 shrink-0 items-center justify-between px-4 text-primary-foreground shadow-sm sm:px-8 lg:px-10"
      style={
        {
          // 渐变跟随主题阶梯：浅色亮渐变、深色主色压暗（barGradient 按明暗/barMix 分支）
          background: barGradient(theme),
          // 前景不再强推白色：子组件用 text-primary-foreground 取各主题自带的前景色。
          // 浅色主题自带白字（渐变底偏深可读）；深色主题自带近黑字（如 ring #04151b），
          // 在压暗后的渐变底上对比更高——原先统一白字在深色渐变上反而亮得刺眼。
        }
      }
    >
      {/* 左：品牌字标（点击回工作台），logo + dank mono 字体与旧版一致 */}
      <Link to="/" className="flex items-center">
        <img src={marsLogo} alt="Mars" className="h-6 w-6" style={{ marginRight: 10 }} />
        <span
          className="text-[18px] font-semibold"
          style={{ fontFamily: '"dank mono", ui-monospace, monospace' }}
        >
          Mars
        </span>
        {version?.version && (
          <span className="ml-1 -mt-2 text-[10px] text-primary-foreground/70">
            {version.version}
          </span>
        )}
      </Link>

      {/* 右：工具 + 用户 + 状态（集群状态在最右）；min-w-0 + 用户名截断，避免小屏溢出 */}
      <div className="flex min-w-0 items-center gap-2">
        <ThemeSwitcher theme={theme} onSelect={onSelectTheme} variant="overlay" />
        <button
          onClick={switchLocale}
          className="flex h-8 items-center rounded-lg border border-primary-foreground/20 bg-primary-foreground/10 px-2.5 font-mono text-[12px] text-primary-foreground transition-colors hover:bg-primary-foreground/15"
          title={t('locale.language')}
        >
          {locale === 'zh-CN' ? 'EN' : '中文'}
        </button>
        <UserMenu />
        <ClusterStatus />
      </div>
    </header>
  )
}
