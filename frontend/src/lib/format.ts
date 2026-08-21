import i18n from '@/i18n'

/**
 * 统一日期/数字格式化：走 Intl.*（规范要求，不做硬编码格式串拼接）。
 * 视觉保持与旧版 dayjs 'YYYY-MM-DD HH:mm:ss' 一致——日期用 en-CA（ISO 风格即
 * YYYY-MM-DD），时间用当前 locale 的 24 小时制。换 locale 时自动跟随。
 */

/** 日期：YYYY-MM-DD（en-CA 的 Intl 输出即 ISO 风格，无需手动拼串） */
export function formatDate(ts: string | number | Date): string {
  return new Intl.DateTimeFormat('en-CA', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date(ts))
}

/** 日期+时间：YYYY-MM-DD HH:mm:ss */
export function formatDateTime(ts: string | number | Date): string {
  const date = formatDate(ts)
  const time = new Intl.DateTimeFormat(i18n.language, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(new Date(ts))
  return `${date} ${time}`
}

let numFmtLang: string | null = null
let numFmt: Intl.NumberFormat | null = null

/** 秒数保留 1 位小数（TimeCost 计时高频刷新，缓存 NumberFormat 实例避免每帧重建） */
export function formatSeconds(seconds: number): string {
  if (!numFmt || numFmtLang !== i18n.language) {
    numFmtLang = i18n.language
    numFmt = new Intl.NumberFormat(numFmtLang, {
      minimumFractionDigits: 1,
      maximumFractionDigits: 1,
    })
  }
  return numFmt.format(seconds)
}
