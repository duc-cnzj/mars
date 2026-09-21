/** 项目详情弹窗的 Tab 可见性规则（纯函数，无 React/DOM 依赖）。
 *
 * 单独成模块而不是写进 ProjectDetailModal：这条规则是「哪些 Tab 需要工作负载在跑」的
 * 唯一事实来源，且组件本体依赖 WebSocket / 拖拽 Provider（useWebsocket 无 Provider 直接抛错），
 * Node 冒烟环境起不来；抽成纯函数后 .smoke 可直接断言本规则，守住「恢复后无法重新部署」的回归。 */

/** 项目详情弹窗的 Tab 键 */
export type TabKey = 'logs' | 'shell' | 'edit' | 'detail' | 'topology'

/** Tab 声明序（全量）：文案由组件按 i18n 补，本层只管顺序与可见性 */
export const ALL_TAB_KEYS: readonly TabKey[] = ['logs', 'shell', 'edit', 'topology', 'detail']

/**
 * 需要「工作负载在跑」的操作类 Tab：日志/终端/拓扑都直连 pod，未部署时无 pod 可连，
 * 故仅在 Deployed/Deploying 时展示。
 *
 * ⚠️ 「部署配置」(edit) **不在**此列：它不依赖运行态——未部署/状态未知的项目正是要在它
 * 里面点「部署」。误删恢复后的项目 deploy_status 被重置为 StatusUnknown（helm release 已
 * 卸载，见 internal/data/namespace.go 的 RestoreDeleted），重新部署是唯一出路；若把 edit
 * 一并藏掉，恢复出来的项目就再没有入口点部署，成了死结。
 */
const WORKLOAD_TABS: ReadonlySet<TabKey> = new Set<TabKey>(['logs', 'shell', 'topology'])

/** 该 Tab 是否依赖运行中的工作负载（日志/终端/拓扑直连 pod） */
export function isWorkloadTab(tab: TabKey): boolean {
  return WORKLOAD_TABS.has(tab)
}

/**
 * 按运行态过滤可见 Tab：canOperate=false 时只留不依赖工作负载的「部署配置 + 详细信息」，
 * 顺序沿用 ALL_TAB_KEYS。canOperate 由调用方按 deployStatus（Deployed/Deploying）判定。
 */
export function visibleTabKeys(canOperate: boolean): TabKey[] {
  return ALL_TAB_KEYS.filter((key) => canOperate || !isWorkloadTab(key))
}
