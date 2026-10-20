import { API } from '@/api/endpoints'
import { getToken } from '@/api/token'

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
