import { useTranslation } from 'react-i18next'
import { themes, type ThemeId } from '@/themes'
import { isMac } from '@/lib/platform'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/shadcn/dropdown-menu'
import { Icon } from '@/components/Icons'

/** 主题 id → i18n 词条 key（名字 / 一句定位），字面量联合保持 t() 类型校验。
 *  顺序与 themes/index.ts 注册表一致（决定切换器/快捷键的显示顺序） */
type ThemeNameKey =
  | 'themes.seiji.name'
  | 'themes.magenta.name'
  | 'themes.latte.name'
  | 'themes.mint.name'
  | 'themes.lavender.name'
  | 'themes.cherry.name'
  | 'themes.violet.name'
  | 'themes.lime.name'
type ThemeTaglineKey =
  | 'themes.seiji.tagline'
  | 'themes.magenta.tagline'
  | 'themes.latte.tagline'
  | 'themes.mint.tagline'
  | 'themes.lavender.tagline'
  | 'themes.cherry.tagline'
  | 'themes.violet.tagline'
  | 'themes.lime.tagline'

const THEME_NAME_KEY: Record<ThemeId, ThemeNameKey> = {
  seiji: 'themes.seiji.name',
  magenta: 'themes.magenta.name',
  latte: 'themes.latte.name',
  mint: 'themes.mint.name',
  lavender: 'themes.lavender.name',
  cherry: 'themes.cherry.name',
  violet: 'themes.violet.name',
  lime: 'themes.lime.name',
}
const THEME_TAGLINE_KEY: Record<ThemeId, ThemeTaglineKey> = {
  seiji: 'themes.seiji.tagline',
  magenta: 'themes.magenta.tagline',
  latte: 'themes.latte.tagline',
  mint: 'themes.mint.tagline',
  lavender: 'themes.lavender.tagline',
  cherry: 'themes.cherry.tagline',
  violet: 'themes.violet.tagline',
  lime: 'themes.lime.tagline',
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
        {/* 设计上不需要 focus-visible 焦点环：显式 outline-none 压掉全局 outline、不设 ring——
            Radix 关闭下拉会把焦点还给 trigger，焦点环会硬切弹出（闪烁）；对齐 UserMenu trigger */}
        <button
          type="button"
          aria-label={t('themes.switchTo')}
          className={`flex h-8 items-center gap-2 rounded-lg px-2.5 text-[12px] transition-colors focus-visible:outline-none ${
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
          <Icon name="chevron-down" className="size-3.5 shrink-0 opacity-60" />
        </button>
      </DropdownMenuTrigger>
      {/* z-[9999] 盖过一切：可拖拽弹窗的共享 z 计数器从 51 起每开/置顶一次 +1，
          固定高位保证主题选择浮层永远在最上（如项目详情弹窗开着时也能打开并盖住它） */}
      <DropdownMenuContent align="end" sideOffset={6} className="z-[9999] w-64 p-1">
        <div className="flex items-center justify-between gap-2 px-2 pt-1 pb-1.5 text-[10px]">
          <span className="font-medium uppercase tracking-wider text-faint">
            {t('themes.switchTo')}
          </span>
          <kbd className="flex h-5 items-center gap-0.5 whitespace-nowrap rounded border border-primary/25 bg-primary/10 px-1.5 font-mono text-[10px] leading-none text-primary">
            {isMac ? <span className="font-sans text-[9px]">⌘</span> : <span>Ctrl</span>}
            <span>+</span>
            <span>Shift</span>
            <span>+</span>
            <span>，</span>
          </kbd>
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
              {active && <Icon name="check" className="size-3.5 shrink-0 text-primary" />}
            </DropdownMenuItem>
          )
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
