import i18n from '@/i18n'

/**
 * 复刻后端 date.ToHumanizeDateTime 的相对时间本地化实现（go-humanize CustomRelTime 语义）。
 *
 * 后端 internal/util/date/date.go 的 magnitudes 中文刻度表被 i18n 词条取代，阈值与档位
 * 划分逐项对齐；方向后缀（以前/以后）也改为词条，实现 zh/en 双语言。后端 `event_at`
 * 写死中文，新 UI 不再消费该字段，改由本函数基于 createdAt 实时计算并跟随 locale。
 */

/** 秒/分/时/天/周/月/年 的毫秒基准（对齐 humanize 常量：Month=30 天、Year=12*Month=360 天） */
const SEC = 1_000
const MIN = 60 * SEC
const HOUR = 60 * MIN
const DAY = 24 * HOUR
const WEEK = 7 * DAY
const MONTH = 30 * DAY
const YEAR = 12 * MONTH

/**
 * 相对时间刻度：d 为档位阈值毫秒（diff 严格小于 d 才选中本档，对应 sort.Search 的 > 语义）；
 * fixed 为无数量的固定词条（"现在"/"1 小时"/"很久"）；key 为带 {{count}} 数量的词条，
 * 数量 = diff / divByMs（Go 整数除法向下取整）。表序按 d 升序，与后端逐项对齐。
 */
const MAGNITUDES = [
  // neutral：方向后缀档（后端该档 Format 无 %s，任何方向都只显示"现在"）
  { d: SEC, fixed: 'events.relative.now', neutral: true },
  { d: 2 * SEC, fixed: 'events.relative.secOne' },
  { d: MIN, key: 'events.relative.sec', divByMs: SEC },
  { d: 2 * MIN, fixed: 'events.relative.minOne' },
  { d: HOUR, key: 'events.relative.min', divByMs: MIN },
  { d: 2 * HOUR, fixed: 'events.relative.hourOne' },
  { d: DAY, key: 'events.relative.hour', divByMs: HOUR },
  { d: 2 * DAY, fixed: 'events.relative.dayOne' },
  { d: WEEK, key: 'events.relative.day', divByMs: DAY },
  { d: 2 * WEEK, fixed: 'events.relative.weekOne' },
  { d: MONTH, key: 'events.relative.week', divByMs: WEEK },
  { d: 2 * MONTH, fixed: 'events.relative.monthOne' },
  { d: YEAR, key: 'events.relative.month', divByMs: MONTH },
  { d: 18 * MONTH, fixed: 'events.relative.yearOne' },
  { d: 2 * YEAR, fixed: 'events.relative.yearTwo' },
  { d: 37 * YEAR, key: 'events.relative.year', divByMs: YEAR },
  { d: Number.MAX_SAFE_INTEGER, fixed: 'events.relative.long' },
] as const

/**
 * 将时间转为相对参考时刻（默认当前）的本地化描述，如"3 分钟以前"/"3 minutes ago"。
 * 与后端行为一致：diff < 1s 显示"现在"（无方向后缀）；超过 37 年显示"很久以前"；
 * 未来时间取"以后"/"from now"方向。input 无法解析为合法日期时返回空串（对齐后端 t==nil 返回空串）。
 */
export function toHumanizeDateTime(input: string | number | Date, now = Date.now()): string {
  const target = new Date(input).getTime()
  if (Number.isNaN(target)) return ''

  const diff = now - target
  const future = diff < 0
  const absDiff = Math.abs(diff)

  // sort.Search 语义：首个 D > absDiff 的档位；理论无命中（diff 超安全整数范围）时取最后一档
  let n = MAGNITUDES.findIndex((m) => m.d > absDiff)
  if (n < 0) n = MAGNITUDES.length - 1
  const mag = MAGNITUDES[n]

  // 用 in 收窄 as const 数组元素的互斥形状（fixed 档无数值、key 档带数量）
  const phrase = 'fixed' in mag
    ? i18n.t(mag.fixed)
    : i18n.t(mag.key, { count: Math.floor(absDiff / mag.divByMs) })

  // neutral 档（"现在"）无方向后缀，任何方向都直接返回，对齐后端 Format 无 %s 的行为
  if ('neutral' in mag) return phrase

  return i18n.t(future ? 'events.relative.fromNow' : 'events.relative.ago', { phrase })
}
