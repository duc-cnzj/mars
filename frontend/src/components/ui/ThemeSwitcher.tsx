import { useTranslation } from 'react-i18next'
import { Check, ChevronDown } from 'lucide-react'
import { themes, type ThemeId } from '../../themes'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/shadcn/dropdown-menu'

/** 主题 id → i18n 词条 key（名字 / 一句定位），字面量联合保持 t() 类型校验 */
type ThemeNameKey =
  | 'themes.ring.name'
  | 'themes.amber.name'
  | 'themes.seiji.name'
  | 'themes.glacier.name'
  | 'themes.dracula.name'
  | 'themes.nord.name'
  | 'themes.latte.name'
  | 'themes.github.name'
  | 'themes.githubDark.name'
  | 'themes.githubDimmed.name'
  | 'themes.chromeDark.name'
  | 'themes.bay.name'
  | 'themes.cherry.name'
  | 'themes.magenta.name'
  | 'themes.volcano.name'
type ThemeTaglineKey =
  | 'themes.ring.tagline'
  | 'themes.amber.tagline'
  | 'themes.seiji.tagline'
  | 'themes.glacier.tagline'
  | 'themes.dracula.tagline'
  | 'themes.nord.tagline'
  | 'themes.latte.tagline'
  | 'themes.github.tagline'
  | 'themes.githubDark.tagline'
  | 'themes.githubDimmed.tagline'
  | 'themes.chromeDark.tagline'
  | 'themes.bay.tagline'
  | 'themes.cherry.tagline'
  | 'themes.magenta.tagline'
  | 'themes.volcano.tagline'

const THEME_NAME_KEY: Record<ThemeId, ThemeNameKey> = {
  ring: 'themes.ring.name',
  amber: 'themes.amber.name',
  seiji: 'themes.seiji.name',
  glacier: 'themes.glacier.name',
  dracula: 'themes.dracula.name',
  nord: 'themes.nord.name',
  latte: 'themes.latte.name',
  github: 'themes.github.name',
  'github-dark': 'themes.githubDark.name',
  'github-dimmed': 'themes.githubDimmed.name',
  'chrome-dark': 'themes.chromeDark.name',
  bay: 'themes.bay.name',
  cherry: 'themes.cherry.name',
  magenta: 'themes.magenta.name',
  volcano: 'themes.volcano.name',
}
const THEME_TAGLINE_KEY: Record<ThemeId, ThemeTaglineKey> = {
  ring: 'themes.ring.tagline',
  amber: 'themes.amber.tagline',
  seiji: 'themes.seiji.tagline',
  glacier: 'themes.glacier.tagline',
  dracula: 'themes.dracula.tagline',
  nord: 'themes.nord.tagline',
  latte: 'themes.latte.tagline',
  github: 'themes.github.tagline',
  'github-dark': 'themes.githubDark.tagline',
  'github-dimmed': 'themes.githubDimmed.tagline',
  'chrome-dark': 'themes.chromeDark.tagline',
  bay: 'themes.bay.tagline',
  cherry: 'themes.cherry.tagline',
  magenta: 'themes.magenta.tagline',
  volcano: 'themes.volcano.tagline',
}

/**
 * 主题切换器：下拉框，当前主题显示色点 + 名称，打开列出全部主题。
 * 每个选项：色点 + 名称 + 一句定位 + 明暗标记 + 当前项勾选。
 * variant：
 * - surface（默认）：常规表面底色（白/深底）
 * - overlay：用于渐变 Header 上的半透明白底 + 白字适配
 */
export function ThemeSwitcher({
  theme,
  onSelect,
  variant = 'surface',
}: {
  theme: ThemeId
  onSelect: (t: ThemeId) => void
  variant?: 'surface' | 'overlay'
}) {
  const { t } = useTranslation()
  const onAccent = variant === 'overlay'
  const current = themes.find((th) => th.id === theme) ?? themes[0]

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          aria-label={t('themes.switchTo')}
          className={`flex h-8 items-center gap-2 rounded-lg px-2.5 text-[12px] transition-colors ${
            onAccent
              ? 'border border-primary-foreground/20 bg-primary-foreground/10 text-primary-foreground hover:bg-primary-foreground/15'
              : 'border border-line bg-raised text-ink hover:bg-surface'
          }`}
        >
          <span
            className="h-3.5 w-3.5 shrink-0 rounded-full border-2"
            style={{
              background: current.accent,
              borderColor: onAccent ? 'rgba(255,255,255,0.4)' : 'var(--border-strong)',
            }}
          />
          <span className="hidden min-[560px]:block">
            {t(THEME_NAME_KEY[current.id])}
          </span>
          <ChevronDown className="size-3.5 shrink-0 opacity-60" />
        </button>
      </DropdownMenuTrigger>
      {/* z-[9999] 盖过一切：可拖拽弹窗的共享 z 计数器从 51 起每开/置顶一次 +1，
          固定高位保证主题选择浮层永远在最上（如项目详情弹窗开着时也能打开并盖住它） */}
      <DropdownMenuContent align="end" sideOffset={6} className="z-[9999] w-64 p-1">
        <div className="px-2 pt-1 pb-1.5 text-[10px] font-medium uppercase tracking-wider text-faint">
          {t('themes.switchTo')}
        </div>
        <DropdownMenuSeparator />
        {themes.map((th) => {
          const active = th.id === theme
          return (
            <DropdownMenuItem
              key={th.id}
              onSelect={() => onSelect(th.id)}
              className="flex items-center gap-2.5 py-2"
            >
              <span
                className="h-3.5 w-3.5 shrink-0 rounded-full border-2 border-line-strong"
                style={{ background: th.accent }}
              />
              <span className="flex min-w-0 flex-1 flex-col">
                <span className="text-[12px] leading-tight font-medium">
                  {t(THEME_NAME_KEY[th.id])}
                </span>
                <span className="truncate text-[10px] leading-tight text-mute">
                  {t(THEME_TAGLINE_KEY[th.id])}
                </span>
              </span>
              <span className={`shrink-0 text-[10px] ${active ? 'text-primary' : 'text-mute'}`}>
                {t(th.mode === 'dark' ? 'themes.dark' : 'themes.light')}
              </span>
              {active && <Check className="size-3.5 shrink-0 text-primary" />}
            </DropdownMenuItem>
          )
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
