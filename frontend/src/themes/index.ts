/** 主题注册表：十五套双色风格（9 深 + 6 浅）。name/tagline 词条见 i18n themes.* */

export type ThemeId =
  | 'ring'
  | 'amber'
  | 'seiji'
  | 'glacier'
  | 'dracula'
  | 'nord'
  | 'latte'
  | 'github'
  | 'github-dark'
  | 'github-dimmed'
  | 'chrome-dark'
  | 'bay'
  | 'cherry'
  | 'magenta'
  | 'volcano'

export interface ThemeMeta {
  id: ThemeId
  /** 明暗模式 */
  mode: 'dark' | 'light'
  /** 预览底色 */
  bg: string
  /** 预览强调色 */
  accent: string
  /** 顶栏/底栏在深色模式下的主色混色比例（越小越暗，默认 85）；
   *  仅个别明度天然偏高的主题（如 ring 青蓝）单独调低 */
  barMix?: number
}

export const themes: ThemeMeta[] = [
  // ── 默认主题（浅色 · 蔚蓝）──
  {
    id: 'seiji',
    mode: 'light',
    bg: '#f6f7f9',
    accent: '#4f46e5',
  },
  // ── 品牌置顶（浅色 · 璀璨洋红）──
  {
    id: 'magenta',
    mode: 'light',
    bg: '#fdf2f8',
    accent: '#c92a93',
  },
  // ── 浅色（4）──
  {
    id: 'glacier',
    mode: 'light',
    bg: '#eef4f5',
    accent: '#0d9488',
  },
  {
    id: 'latte',
    mode: 'light',
    bg: '#eff1f5',
    accent: '#1e66f5',
  },
  {
    id: 'github',
    mode: 'light',
    bg: '#f6f8fa',
    accent: '#0969da',
  },
  {
    id: 'bay',
    mode: 'light',
    bg: '#eef5fb',
    accent: '#1f7fd1',
  },
  // ── 深色（9）──
  {
    id: 'ring',
    mode: 'dark',
    bg: '#070a10',
    accent: '#22d3ee',
    // 青蓝明度偏高：85% 混色顶栏仍偏亮，单独压到 60%
    barMix: 60,
  },
  {
    id: 'amber',
    mode: 'dark',
    bg: '#14100c',
    accent: '#e0a63f',
  },
  {
    id: 'dracula',
    mode: 'dark',
    bg: '#282a36',
    accent: '#bd93f9',
  },
  {
    id: 'nord',
    mode: 'dark',
    bg: '#2e3440',
    accent: '#88c0d0',
  },
  {
    id: 'github-dark',
    mode: 'dark',
    bg: '#0d1117',
    accent: '#4493f8',
  },
  {
    id: 'github-dimmed',
    mode: 'dark',
    bg: '#22272e',
    accent: '#539bf5',
  },
  {
    id: 'chrome-dark',
    mode: 'dark',
    bg: '#202124',
    accent: '#8ab4f8',
  },
  {
    id: 'cherry',
    mode: 'dark',
    bg: '#170d12',
    accent: '#ff4d5e',
  },
  {
    id: 'volcano',
    mode: 'dark',
    bg: '#131416',
    accent: '#ff4d45',
  },
]

/** 主题 class 名（CSS 变量作用域） */
export const themeClass = (id: ThemeId): string => `theme-${id}`

/**
 * 顶栏/底栏背景渐变。浅色主题用主色原色（主色 → 主色加强色）；
 * 深色主题把两端各与 #34495e 按 barMix（默认 85%）混色压暗——深色主题的
 * --primary-strong 是比主色更亮的高亮色（如 github-dark #4493f8 → #79c0ff），
 * 直接套原色会在近黑底上亮成一条醒目的色带。
 */
export const barGradient = (id: ThemeId): string => {
  const meta = themes.find((t) => t.id === id)
  if (meta?.mode !== 'dark') {
    return 'linear-gradient(to right, var(--primary), var(--primary-strong))'
  }
  const ratio = meta.barMix ?? 85
  return `linear-gradient(to right, color-mix(in srgb, var(--primary) ${ratio}%, #34495e), color-mix(in srgb, var(--primary-strong) ${ratio}%, #34495e))`
}
