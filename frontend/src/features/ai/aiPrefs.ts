/**
 * AI 面板的用户偏好持久化（按登录用户 email 分键）。
 *
 * 目前只有「免确认」开关：它是高风险选项，同一浏览器切换账号时不能串用，
 * 因此与对话历史一样按 email 隔离，而不是用全局键。
 * 存储失败（隐私模式/配额超限）只影响下次打开时的默认值，不影响本次会话。
 */

const AUTO_APPROVE_PREFIX = 'mars_ai_auto_approve_v1'

/** 生成指定用户的免确认开关存储键。 */
export function autoApproveStorageKey(email: string): string {
  return `${AUTO_APPROVE_PREFIX}:${email}`
}

/** 读取免确认开关状态；未设置或读取失败时返回 false（安全默认）。 */
export function loadAutoApprove(email: string): boolean {
  try {
    return localStorage.getItem(autoApproveStorageKey(email)) === 'true'
  } catch {
    return false
  }
}

/** 保存免确认开关状态，刷新后恢复。 */
export function saveAutoApprove(email: string, enabled: boolean): void {
  try {
    localStorage.setItem(autoApproveStorageKey(email), String(enabled))
  } catch {
    // 存不下只影响下次打开时的默认值。
  }
}
