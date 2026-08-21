import { useCallback, useState } from 'react'
import { themes, type ThemeId } from '../themes'

const THEME_KEY = 'mars.theme'

/** 读取持久化的主题，非法/缺失回退默认（蔚蓝） */
function readTheme(): ThemeId {
  try {
    const raw = localStorage.getItem(THEME_KEY)
    if (raw && themes.some((t) => t.id === raw)) return raw as ThemeId
  } catch {
    /* 隐私模式等场景，忽略 */
  }
  return 'seiji'
}

/** 主题偏好 Hook：默认蔚蓝 + 用户级偏好持久化，切换即写入 localStorage */
export function useTheme() {
  const [theme, setThemeState] = useState<ThemeId>(readTheme)

  const setTheme = useCallback((t: ThemeId) => {
    setThemeState(t)
    try {
      localStorage.setItem(THEME_KEY, t)
    } catch {
      /* 忽略存储失败 */
    }
  }, [])

  return { theme, setTheme }
}
