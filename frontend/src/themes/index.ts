/** 主题注册表：八主题双色风格（5 浅 + 3 深）。name/tagline 词条见 i18n themes.* */

export type ThemeId =
  | 'seiji'
  | 'magenta'
  | 'latte'
  | 'mint'
  | 'lavender'
  | 'cherry'
  | 'violet'
  | 'lime'

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
  // ── 亮色（5）──
  // 默认主题（雪白 · 靛蓝）
  {
    id: 'seiji',
    mode: 'light',
    bg: '#f6f7f9',
    accent: '#4f46e5',
  },
  // 品牌置顶（璀璨洋红）
  {
    id: 'magenta',
    mode: 'light',
    bg: '#fdf2f8',
    accent: '#c92a93',
  },
  {
    id: 'latte',
    mode: 'light',
    bg: '#eff1f5',
    accent: '#1e66f5',
  },
  {
    id: 'mint',
    mode: 'light',
    bg: '#f3faf6',
    accent: '#047857',
  },
  {
    id: 'lavender',
    mode: 'light',
    bg: '#f7f6fc',
    accent: '#7c3aed',
  },
  // ── 暗色（3）──
  {
    id: 'cherry',
    mode: 'dark',
    bg: '#170d12',
    accent: '#ff4d5e',
  },
  {
    id: 'violet',
    mode: 'dark',
    bg: '#0e0d1b',
    accent: '#a78bfa',
    // 混色剂为 var(--bg)（近黑）时 75% 的顶栏亮度≈旧 #34495e 60% 效果，且近黑文字对比约 4.2
    barMix: 75,
  },
  {
    id: 'lime',
    mode: 'dark',
    bg: '#0f1008',
    accent: '#a3e635',
    // 荧光黄绿明度全库最高，85% 会亮到刺眼；与 var(--bg) 混色 62% 即达近黑文字 AA（≈4.5），为压暗下限
    barMix: 62,
  },
]

/** 主题 class 名（CSS 变量作用域） */
export const themeClass = (id: ThemeId): string => `theme-${id}`

/**
 * 顶栏/底栏背景渐变。浅色主题用主色原色（主色 → 主色加强色）；
 * 深色主题把两端各与主题自带的近黑底色 --bg 按 barMix（默认 85%）混色压暗——
 * 深色主题的 --primary-strong 是比主色更亮的高亮色（如 github-dark #4493f8 → #79c0ff），
 * 直接套原色会在近黑底上亮成一条醒目的色带。
 * 混色剂用主题自身 --bg 而非固定 #34495e：后者是蓝灰，混进 lime/violet 会把顶栏
 * 染成偏灰白的钢蓝色（barMix 越低越明显），与主题色相不符（用户反馈顶栏右侧发白）。
 */
export const barGradient = (id: ThemeId): string => {
  const meta = themes.find((t) => t.id === id)
  if (meta?.mode !== 'dark') {
    return 'linear-gradient(to right, var(--primary), var(--primary-strong))'
  }
  const ratio = meta.barMix ?? 85
  return `linear-gradient(to right, color-mix(in srgb, var(--primary) ${ratio}%, var(--bg)), color-mix(in srgb, var(--primary-strong) ${ratio}%, var(--bg)))`
}
