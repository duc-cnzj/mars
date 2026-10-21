/** 容器终端（xterm）的复制快捷键判定（纯函数，Node 冒烟可直接断言）。
 *
 * 背景：xterm.js 默认把 Ctrl+C 一律编码成 ^C 下发给容器，Windows 用户选中文本后
 * 按 Ctrl+C 只会中断前台进程、剪贴板没有变化。Windows Terminal 的约定是
 * 「有选区 = 复制，无选区 = 中断」，这里复刻同一语义。
 *
 * 只在 Windows 生效：Linux 终端惯例是 Ctrl+Shift+C 复制（Ctrl+C 恒为 SIGINT），
 * macOS 用 Cmd+C，都不应被本逻辑改写。
 */

/** 是否为 Windows 客户端：UA 在一次会话内不变，模块级判定一次即可。 */
export const IS_WINDOWS =
  typeof navigator !== 'undefined' && /windows/i.test(navigator.userAgent ?? '')

/** 判定 Ctrl+C 是否应按「复制选中文本」处理所需的最小事件形状。 */
export interface CtrlCKeyEvent {
  type: string
  key: string
  ctrlKey: boolean
  shiftKey: boolean
  altKey: boolean
  metaKey: boolean
}

/**
 * 是否把本次 Ctrl+C 拦截为「复制」：
 *   - 必须 keydown（keyup 不再触发第二次复制）
 *   - 必须是纯 Ctrl+C（带 Shift/Alt/Meta 的组合键不抢）
 *   - 必须存在选区；无选区时返回 false，保留 Ctrl+C 的 SIGINT 语义
 *   - 仅在 Windows 生效
 */
export function isWindowsCtrlCCopy(
  ev: CtrlCKeyEvent,
  hasSelection: boolean,
  isWindows: boolean,
): boolean {
  return (
    isWindows &&
    ev.type === 'keydown' &&
    ev.ctrlKey &&
    !ev.shiftKey &&
    !ev.altKey &&
    !ev.metaKey &&
    ev.key.toLowerCase() === 'c' &&
    hasSelection
  )
}
