import { useEffect } from 'react'
import { themes, type ThemeId } from '@/themes'

/**
 * 主题快捷键：Ctrl/⌘ + Shift + ，按注册顺序循环切到下一个主题。
 * - 用 event.code（Comma）而非 event.key：Shift+, 时 key 是 '<'，code 与键盘布局无关更稳。
 * - 为什么只留一个且选逗号：方向键被 macOS 调度中心/应用窗口（Ctrl+↑/↓）抢、⌘+Shift+[/] 被浏览器
 *   抢去切标签页、⌘+Shift+. 在部分系统/输入法下被吞掉——页面都收不到事件（preventDefault 无效）。
 *   逗号是实测唯一全平台可达、稳定的按键。
 * - 编辑态（input/textarea/select/contentEditable/代码编辑器）一律不接管，避免打字被劫持。
 */
export function useThemeHotkeys({
  theme,
  setTheme,
}: {
  theme: ThemeId
  setTheme: (t: ThemeId) => void
}) {
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null
      const tag = target?.tagName
      const editable =
        tag === 'INPUT' ||
        tag === 'TEXTAREA' ||
        tag === 'SELECT' ||
        target?.isContentEditable ||
        !!target?.closest('.cm-editor, .monaco-editor')
      if (editable) return
      if (!((e.metaKey || e.ctrlKey) && e.shiftKey && !e.altKey)) return
      if (e.code !== 'Comma') return

      const idx = themes.findIndex((t) => t.id === theme)
      const next = themes[(idx + 1) % themes.length].id
      e.preventDefault()
      setTheme(next)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [theme, setTheme])
}
