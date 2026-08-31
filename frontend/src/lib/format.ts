import i18n from '@/i18n'

/**
 * 统一日期/数字格式化：走 Intl.*（规范要求，不做硬编码格式串拼接）。
 * 视觉保持与旧版 dayjs 'YYYY-MM-DD HH:mm:ss' 一致——日期用 en-CA（ISO 风格即
 * YYYY-MM-DD），时间用当前 locale 的 24 小时制。换 locale 时自动跟随。
 */

/** 解析输入为合法 Date，非法（空串/undefined/无法解析）返回 null，避免 Intl.format 抛 RangeError */
function toDate(ts: string | number | Date): Date | null {
  const d = new Date(ts)
  return Number.isNaN(d.getTime()) ? null : d
}

/** 日期格式化器缓存（en-CA 固定，单例复用——列表行渲染高频调用，避免每次 new Intl.DateTimeFormat） */
let dateFmt: Intl.DateTimeFormat | null = null
/** 时间格式化器缓存：跟随 i18n.language 变化重建（对齐 formatSeconds 的缓存模式） */
let timeFmtLang: string | null = null
let timeFmt: Intl.DateTimeFormat | null = null

/** 日期：YYYY-MM-DD（en-CA 的 Intl 输出即 ISO 风格，无需手动拼串）；输入非法时返回空串 */
export function formatDate(ts: string | number | Date): string {
  const d = toDate(ts)
  if (!d) return ''
  if (!dateFmt) {
    dateFmt = new Intl.DateTimeFormat('en-CA', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    })
  }
  return dateFmt.format(d)
}

/** 日期+时间：YYYY-MM-DD HH:mm:ss；输入非法时返回空串 */
export function formatDateTime(ts: string | number | Date): string {
  const d = toDate(ts)
  if (!d) return ''
  const date = formatDate(d)
  if (!timeFmt || timeFmtLang !== i18n.language) {
    timeFmtLang = i18n.language
    timeFmt = new Intl.DateTimeFormat(timeFmtLang, {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    })
  }
  return `${date} ${timeFmt.format(d)}`
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
