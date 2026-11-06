import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import marsLogo from '@/assets/marslogo.png'
import { AISparklesIcon } from '@/components/Icons'
import { ThemeSwitcher } from '@/components/ui/ThemeSwitcher'
import { useVersion } from '@/hooks/useVersion'
import { getLocale, setLocale, type Locale } from '@/i18n'
import { barGradient, type ThemeId } from '@/themes'
import { ClusterStatus } from './ClusterStatus'
import { UserMenu } from './UserMenu'

/**
 * 渐变顶栏（还原旧版 AppHeader 布局）：无侧边栏，
 * 左 logo + "Mars" 字标 + 版本号，右 AI 助手入口 + 集群状态灯 + 主题/语言切换 + 用户下拉。
 */
export function Topbar({
  theme,
  onSelectTheme,
  onOpenAI,
  aiEnabled = true,
}: {
  theme: ThemeId
  onSelectTheme: (t: ThemeId) => void
  /** AI 助手入口：打开服务端对话面板 */
  onOpenAI: () => void
  /** AI 功能是否启用：明确未启用时隐藏入口，避免点击后才提示配置缺失 */
  aiEnabled?: boolean
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
          // 渐变跟随主题主色：所有主题均为浅色，统一主色 → 主色加强色
          background: barGradient,
          // 前景不再强推白色：子组件用 text-primary-foreground 取各主题自带的前景色。
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
        {/* AI 助手入口：全站唯一入口。星芒图标负责识别，大屏补文字标签降低认知成本；
            窄屏只留图标，避免右侧工具簇（语言/主题/用户/集群状态）溢出。
            AI 未启用时隐藏：入口常驻但点下去才报「请联系管理员」是纯噪音。 */}
        {aiEnabled && (
          <button
            onClick={onOpenAI}
            className="flex h-8 items-center gap-1.5 rounded-lg px-2 text-primary-foreground transition-colors hover:bg-primary-foreground/10"
            title={t('ai.open')}
            aria-label={t('ai.open')}
          >
            {/* 七彩流动星芒：颜色沿星形循环流动，不随顶栏前景色变灰 */}
            <AISparklesIcon className="text-[22px]" />
            <span className="hidden text-[12px] font-medium lg:inline">{t('ai.open')}</span>
          </button>
        )}
        {/* 语言切换：无边框纯文字按钮，位于主题切换左侧；与主题/用户触发器同高，hover 浅底提示可点 */}
        <button
          onClick={switchLocale}
          className="flex h-8 items-center rounded-lg px-2.5 font-mono text-[12px] text-primary-foreground transition-colors hover:bg-primary-foreground/10"
          title={t('locale.language')}
        >
          {locale === 'zh-CN' ? t('locale.en') : t('locale.zh')}
        </button>
        <ThemeSwitcher theme={theme} onSelect={onSelectTheme} variant="overlay" />
        <UserMenu />
        <ClusterStatus />
      </div>
    </header>
  )
}
