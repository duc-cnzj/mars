import type { components } from '@/api/schema'

type Element = components['schemas']['mars.Element']
type HibernateWindow = components['schemas']['hibernate.HibernateWindow']
type ExtraValue = components['schemas']['websocket.ExtraValue']

/**
 * 副本字段展示名：把休眠配置里的字段 path 映射为其在「自定义配置」中填的「字段描述」，
 * 描述缺失时回退 path 本身。项目侧休眠卡片 / 详情为只读展示，用户看描述比看裸 path 直观。
 */
export function replicaFieldLabels(paths: string[], elements?: Element[]): string[] {
  const descByPath = new Map((elements ?? []).map((el) => [el.path, el.description?.trim()]))
  return paths.map((p) => descByPath.get(p) || p)
}

/**
 * 休眠时间窗展示名（多窗）：把每个窗口名与其启止时段拼成「窗名 (HH:MM-HH:MM)」再以逗号连接。
 * 窗口未命中（被删 / 跨环境未同步）或列表尚未加载时回退为窗名本身——时段缺失不应吞掉窗名。
 */
export function hibernateWindowLabel(names: string[], windows: HibernateWindow[]): string {
  return names
    .map((name) => {
      const w = windows.find((it) => it.name === name)
      return w ? `${name} (${w.start}-${w.end})` : name
    })
    .join(', ')
}

/**
 * 休眠时间窗的时段展示名（多窗）：每个窗名对应的「HH:MM-HH:MM」以逗号连接。窗口未命中
 * （被删 / 跨环境未同步）或列表尚未加载时该窗跳过（不掺入空段）——用于只需时段、不掺窗名的
 * 紧凑场景（如配置更新页开关旁的提示）。与 hibernateWindowLabel 的区别：后者带窗名，用于只读键值展示。
 */
export function hibernateWindowTimeRange(names: string[], windows: HibernateWindow[]): string {
  return names
    .map((name) => {
      const w = windows.find((it) => it.name === name)
      return w ? `${w.start}-${w.end}` : ''
    })
    .filter(Boolean)
    .join(', ')
}

/** 解析 "HH:MM" 为当日分钟数；格式不符或越界（小时 >23 / 分钟 >59）返回 null */
function parseHHMM(value: string): number | null {
  const m = /^(\d{1,2}):(\d{2})$/.exec(value.trim())
  if (!m) return null
  const hours = Number(m[1])
  const minutes = Number(m[2])
  if (hours > 23 || minutes > 59) return null
  return hours * 60 + minutes
}

/**
 * 时间窗 → 24h 轨道上的填充段（left/width 均为 0~100 的百分比），供休眠弹层的迷你时间轴渲染。
 * 支持跨零点（start > end，如 22:00→06:00）拆成「start→24:00」与「00:00→end」两段；
 * 窗名未命中（被删 / 跨环境未同步）、时段缺失或非法、起止相等（零长段）一律跳过——
 * 不画错段：宁可留空轨道，也不渲染一条误导的空段。
 *
 * 关键：多个窗若时段**重叠**（如 22:00–06:00 与 23:00–07:00 的尾段都压在 23:00–24:00），
 * 各自半透明的段叠画会因 alpha 叠加在重叠处**变深**，轨道上冒出「一块特别深的点」。
 * 故这里先求所有区间的**并集**再渲染——重叠区只画一次，整条轨道颜色均匀。
 */
export function windowTrackSegments(
  names: string[],
  windows: HibernateWindow[],
): { left: number; width: number }[] {
  const minutesPerDay = 24 * 60
  const toPercent = (minutes: number) => (minutes / minutesPerDay) * 100
  // 先把每个命中窗口展开成 [起, 止) 分钟区间（跨零点拆两段）
  const intervals: [number, number][] = []
  for (const name of names) {
    const win = windows.find((it) => it.name === name)
    if (!win) continue
    const start = parseHHMM(win.start)
    const end = parseHHMM(win.end)
    if (start === null || end === null || start === end) continue
    if (start < end) {
      intervals.push([start, end])
      continue
    }
    // 跨零点：铺满 start 到午夜，再从 0 点铺到 end
    intervals.push([start, minutesPerDay])
    intervals.push([0, end])
  }
  // 并集：按起点排序后合并重叠/相邻区间，杜绝半透明段叠画在重叠处变深
  intervals.sort((a, b) => a[0] - b[0])
  const merged: [number, number][] = []
  for (const [start, end] of intervals) {
    const last = merged[merged.length - 1]
    if (last && start <= last[1]) {
      last[1] = Math.max(last[1], end)
      continue
    }
    merged.push([start, end])
  }
  return merged.map(([start, end]) => ({
    left: toPercent(start),
    width: toPercent(end - start),
  }))
}

/**
 * 副本字段中是否至少有一个被设成了 >0 的数：项目休眠窗口内这类值会被对账（每 2 分钟一轮）自动
 * 缩容为 0，配置更新页据此提示用户「改了不会立即生效」。副本字段类型保证值可解析为数字
 * （见后端 elementSupportsZero），解析非有限数（空/缺失/非数字）一律按不触发处理，避免误报。
 * replicaPaths 为空 → 恒 false（该仓库未配置副本字段，无从判起）。
 */
export function hasReplicaAboveZero(
  values: ExtraValue[] | undefined,
  replicaPaths: string[],
): boolean {
  if (replicaPaths.length === 0) return false
  const byPath = new Map((values ?? []).map((v) => [v.path, v.value]))
  return replicaPaths.some((p) => {
    const n = Number(byPath.get(p))
    return Number.isFinite(n) && n > 0
  })
}

/**
 * 是否需要提示「休眠窗口内副本被改成 >0」：项目已开启定时休眠（hibernate_enabled===true）、
 * 处于休眠态（hibernate_state==='sleeping'）且至少一个副本字段被设成 >0。配置更新页 TabEdit
 * 据此条件渲染提示横幅。横幅内开关关掉后 hibernate_enabled 变 false，判据即不成立、横幅随之
 * 消失——副本改动不再被对账自动缩容，提示的前提也随之解除。
 */
export function shouldWarnReplicaInHibernate(
  proj: { hibernateEnabled?: boolean; hibernateState?: string },
  values: ExtraValue[] | undefined,
  replicaPaths: string[],
): boolean {
  return (
    proj.hibernateEnabled === true &&
    proj.hibernateState === 'sleeping' &&
    hasReplicaAboveZero(values, replicaPaths)
  )
}
