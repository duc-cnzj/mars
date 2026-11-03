import { createContext, useContext, type ReactNode } from 'react'

/**
 * AI 默认上下文：AI 未明确指定范围时，默认按该绑定作答（由 `mai --bind` 注入会话）。
 * - 只填 `namespace` 表示绑定整个命名空间（含其下全部项目）；
 * - 同时填 `project` 表示只绑定「空间/项目」。
 */
export interface AIBinding {
  namespace: string
  project?: string
}

/**
 * 绑定项的文本形式：有项目时为 `空间/项目`，仅命名空间时为 `空间`。
 * 与 `mai --bind` 的两种取值范围一一对应，也用作选择器的 value 与比较依据。
 */
export function aiBindingKey(binding: AIBinding): string {
  return binding.project ? `${binding.namespace}/${binding.project}` : binding.namespace
}

/** 打开 AI 助手的统一入口：允许调用方带一段场景上下文（部署失败等）进入诊断。 */
export interface AIEntry {
  /** 是否可用：明确未启用时为 false，调用方据此隐藏入口（查询失败按可用处理）。 */
  available: boolean
  /**
   * 打开 AI 助手。
   * - prompt 非空时以 `mai '<prompt>'` 进入交互式 TUI 并自动发送；
   * - binding 非空时以 `mai --bind '<空间>'`（或 `'<空间>/<项目>'`）进入，只绑定默认范围、不代发问题。
   * 两者可同时提供（先绑定，再发问题）。
   */
  open: (prompt?: string, binding?: AIBinding) => void
}

/** 默认空实现：脱离 Provider 的组件（如独立挂载的弹窗冒烟）调用时静默无副作用。 */
const noopEntry: AIEntry = { available: false, open: () => {} }

const AIContext = createContext<AIEntry>(noopEntry)

/** 在布局层提供 AI 入口，供页面内的场景入口（项目部署失败等）复用同一套唤醒编排。 */
export function AIEntryProvider({ value, children }: { value: AIEntry; children: ReactNode }) {
  return <AIContext.Provider value={value}>{children}</AIContext.Provider>
}

/** 读取 AI 入口；未包裹 Provider 时返回空实现，入口按钮自然不展示。 */
export function useAIEntry(): AIEntry {
  return useContext(AIContext)
}
