/**
 * 平台判断：macOS（含 iOS 触控设备）为 true。
 * 快捷键文案（⌘ vs Ctrl）等多处共用，统一放在这里避免各组件各自声明。
 */
export const isMac = /Mac|iPod|iPhone|iPad/i.test(navigator.platform || navigator.userAgent)
