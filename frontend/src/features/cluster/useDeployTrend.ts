import { useEffect, useState } from 'react'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'

/** 部署趋势默认窗口：近 30 天（对齐服务端 biz.DeployTrendDefaultDays=30） */
export const DEPLOY_TREND_DAYS = 30
/** 可选窗口：30/60/90 天（后端上限 90，见 biz.DeployTrendMaxDays） */
export const DEPLOY_TREND_RANGES = [30, 60, 90] as const
/** 面板默认选中的窗口档位 */
export type DeployTrendRange = (typeof DEPLOY_TREND_RANGES)[number]

export interface DeployTrend {
  /** 每日部署次数序列（下标 0 = 窗口最早一天，末位 = 今天），与 dates 一一对应 */
  counts: number[]
  /** 各采样点的日期标签（"M/D"，如 "8/4"） */
  dates: string[]
  /** 窗口内总部署次数 */
  total: number
  /** 窗口内日均部署次数（保留 1 位小数） */
  dailyAvg: number
  /** 窗口内单日峰值 */
  peak: number
  /** 峰值所在下标（dates[peakIndex] 为峰值日期） */
  peakIndex: number
}

/** "YYYY-MM-DD" → "M/D"：纯字符串切片，不经 new Date() 二次折算——服务端 date 已是服务端本地
 *  时区的天界标签（biz 按 time.Local 分桶），走浏览器 Date 会按本机时区再切一次，负时区会把天界
 *  回退一天；切片只改展示形状、保留原天界语义。 */
function toMDLabel(date: string): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(date)
  return m ? `${Number(m[2])}/${Number(m[3])}` : date
}

/** 从原始序列折叠出面板读数：总次数 / 日均 / 峰值及其下标。
 *  日均分母取实际窗口长度 counts.length（服务端返回 items 数与请求 days 一致：
 *  30→30 点、90→90 点），不能写死默认 30——切 60/90 窗口时写死会低估日均。 */
function summarize(counts: number[], dates: string[]): DeployTrend {
  const n = counts.length || 1
  const total = counts.reduce((sum, c) => sum + c, 0)
  let peak = 0
  let peakIndex = 0
  counts.forEach((c, i) => {
    if (c > peak) {
      peak = c
      peakIndex = i
    }
  })
  return {
    counts,
    dates,
    total,
    dailyAvg: Math.round((total / n) * 10) / 10,
    peak,
    peakIndex,
  }
}

/** 占位基线：全 0 的 days 帧（日期取浏览器今天往前推），首帧撑起曲线区与 x 轴刻度，
 *  避免等真数据落地前一度空画；请求成功后即被真数据覆盖，属加载过渡非业务读数。 */
function placeholderTrend(days: number): DeployTrend {
  const counts = Array.from({ length: days }, () => 0)
  const dates: string[] = []
  const now = new Date()
  for (let i = 0; i < days; i++) {
    const d = new Date(now)
    d.setDate(now.getDate() - (days - 1 - i))
    dates.push(`${d.getMonth() + 1}/${d.getDate()}`)
  }
  return summarize(counts, dates)
}

/**
 * 每日部署趋势数据源：近 days 天（默认 30、可选 30/60/90）每日部署次数，取数自真端点
 * /api/admin/cluster/deploy_trend（管理员接口，服务端按天聚合 changelog、时区分桶、无部署补 0，
 * 升序返回、长度恒等于请求天数）。days 变化即按新窗口重拉。
 *
 * 只在窗口变化/挂载时拉一次（部署次数是日粒度、非秒级变化，无需像 useResourceBoard 那样 45s
 * 轮询实时快照）：手动刷新由外层 ResourceBoard version bump 重挂载 RefreshFade 子树、连带本面板
 * 重建，effect 重跑即拉到新一版。请求失败静默保留占位帧——看板与其同源同鉴权，若真挂了看板会先
 * 暴露，这里不重复弹错。
 */
export function useDeployTrend(days: number = DEPLOY_TREND_DAYS): DeployTrend {
  const [trend, setTrend] = useState<DeployTrend>(() => placeholderTrend(days))

  useEffect(() => {
    // 窗口切换：先切到新窗口占位（长度/日期对齐），再拉真数据覆盖，避免旧窗口帧残留一帧错长度
    setTrend(placeholderTrend(days))
    let active = true
    void (async () => {
      try {
        const { data, error } = await api.GET(API.adminDeployTrend, {
          params: { query: { days } },
        })
        // 已卸载（手动刷新快速重建）时旧响应不落地，避免过期帧回写
        if (!active || error) return
        const items = data?.items ?? []
        // 服务端已按天升序铺满、天界即服务端本地时区，此处只做形状折叠，不改顺序/时区
        if (items.length > 0) {
          setTrend(summarize(items.map((it) => it.count), items.map((it) => toMDLabel(it.date))))
        }
      } catch {
        // api.GET 网络级失败（断网/DNS/中止）走 reject，而非回 {error}——必须 try/catch 兜住，
        // 否则是未捕获 promise rejection。静默保留占位帧：看板与其同源同鉴权，真挂了看板会先
        // 暴露，这里不重复弹错（对齐 useResourceBoard 的失败静默兜底）。
      }
    })()
    return () => {
      active = false
    }
  }, [days])

  return trend
}
