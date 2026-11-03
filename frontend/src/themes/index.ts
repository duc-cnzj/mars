/** 主题注册表：五主题浅色双色风格。name/tagline 词条见 i18n themes.* */

export type ThemeId = 'seiji' | 'magenta' | 'latte' | 'mint' | 'lavender'

export interface ThemeMeta {
  id: ThemeId
  /** 预览强调色 */
  accent: string
}

export const themes: ThemeMeta[] = [
  // 默认主题（雪白 · 靛蓝）
  {
    id: 'seiji',
    accent: '#4f46e5',
  },
  // 品牌置顶（璀璨洋红）
  {
    id: 'magenta',
    accent: '#c92a93',
  },
  {
    id: 'latte',
    accent: '#1e66f5',
  },
  {
    id: 'mint',
    accent: '#047857',
  },
  {
    id: 'lavender',
    accent: '#7c3aed',
  },
]

/** 主题 class 名（CSS 变量作用域） */
export const themeClass = (id: ThemeId): string => `theme-${id}`

/** 顶栏/底栏背景渐变：所有主题均为浅色，统一用主色原色（主色 → 主色加强色）。 */
export const barGradient = 'linear-gradient(to right, var(--primary), var(--primary-strong))'
