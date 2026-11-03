import { API } from '@/api/endpoints'
import { getToken } from '@/api/token'

/** 最近一次成功使用 AI 的时间戳（毫秒）存储键：入口预热据此判断是否值得唤起。 */
export const AI_LAST_USED_STORAGE_KEY = 'mars_ai_last_used_at'

/** 沙箱运行态：由后端按 Deployment 副本数与就绪数派生。 */
export type AISandboxState = 'running' | 'starting' | 'sleeping' | 'absent'

/** 当前用户的 agent 项目状态快照（/api/ai/status 的 agent 字段）。 */
export interface AIAgentStatus {
  project_id: number
  sandbox: AISandboxState
}

/** /api/ai/status 与 /api/ai/agent/enable 的响应体。 */
export interface AIStatus {
  enabled: boolean
  agent: AIAgentStatus | null
}

/**
 * AI 接口的 HTTP 错误：携带状态码，调用方据此区分 409（部署进行中）等可重试冲突。
 * 后端错误体是纯文本（http.Error），直接作为 message 透出。
 */
export class AIHTTPError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'AIHTTPError'
    this.status = status
  }
}

/**
 * 请求手写 AI 端点（不在 openapi schema 内，故不走 typed client）。
 * 统一挂 Authorization 与 X-Requested-With，与全局 api client 保持同一鉴权约定。
 */
async function requestAI(path: string, init: RequestInit): Promise<AIStatus> {
  const res = await fetch(path, {
    ...init,
    headers: {
      Authorization: getToken(),
      'X-Requested-With': 'XMLHttpRequest',
      ...init.headers,
    },
  })
  if (!res.ok) {
    const text = (await res.text()).trim()
    throw new AIHTTPError(res.status, text || `HTTP ${res.status}`)
  }
  return (await res.json()) as AIStatus
}

/** 查询 AI 功能可用性与当前用户 agent 状态（GET /api/ai/status）。 */
export function fetchAIStatus(): Promise<AIStatus> {
  return requestAI(API.aiStatus, { method: 'GET' })
}

/** 启用/唤醒当前用户的 agent 沙箱（POST /api/ai/agent/enable，后端幂等）。 */
export function enableAIAgent(): Promise<AIStatus> {
  return requestAI(API.aiAgentEnable, { method: 'POST' })
}

/** AI 使用漏斗事件名，与服务端 isAIUsageEvent 白名单保持一致。 */
export type AIUsageEvent =
  | 'entry_click'
  | 'status_disabled'
  | 'enable_start'
  | 'enable_ok'
  | 'enable_fail'
  | 'sandbox_ready'
  | 'sandbox_open'

/**
 * 上报一个 AI 使用漏斗事件（POST /api/ai/usage）。
 * 埋点是尽力而为：网络失败静默吞掉，绝不能影响入口编排的主流程。
 * durationMs 仅 sandbox_ready 使用，表示从点击入口到沙箱就绪的毫秒数。
 */
export function reportAIUsage(event: AIUsageEvent, durationMs?: number): void {
  void postAIEvent(API.aiUsage, { event, duration_ms: durationMs })
}

/**
 * 刷新 agent 沙箱活跃度（POST /api/ai/activity）。
 * AI 弹窗打开且页面可见时定时调用，避免用户正在看日志/终端却被空闲缩容误关；
 * 失败静默，心跳不能变成用户交互的故障点。
 */
export function touchAIActivity(): void {
  void postAIEvent(API.aiActivity)
}

/** 发送无响应体的 AI 辅助请求；所有失败（网络/HTTP）都静默。 */
async function postAIEvent(path: string, body?: unknown): Promise<void> {
  try {
    await fetch(path, {
      method: 'POST',
      headers: {
        Authorization: getToken(),
        'X-Requested-With': 'XMLHttpRequest',
        ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
      },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    // 尽力而为的埋点/心跳，不让网络错误冒泡。
  }
}
