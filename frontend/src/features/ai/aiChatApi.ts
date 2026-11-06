import { API } from '@/api/endpoints'
import { getToken } from '@/api/token'
import type { AIBinding } from './AIContext'

/**
 * 服务端 AI 对话的 HTTP 客户端（手写 fetch + ReadableStream）。
 *
 * 后端契约见 doc/ai_agent.md「服务端 AI 对话」：
 * - GET  /api/ai/chat/status  → { enabled, model }
 * - POST /api/ai/chat         → text/event-stream
 * - POST /api/ai/chat/confirm → text/event-stream（提交写工具确认结果）
 *
 * 不用 typed client：SSE 响应不是 JSON，openapi schema 也不覆盖这三个端点。
 * 对话历史只保存在浏览器（localStorage），服务端不落任何会话状态。
 */

/** 对话消息角色：system 由服务端注入，前端永远不发。 */
export type ChatRole = 'user' | 'assistant' | 'tool'

/** 模型请求的一次工具调用（OpenAI function calling 结构，原样回传）。 */
export interface ChatToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    /** 参数原文（JSON 字符串），服务端按原文哈希校验确认票据。 */
    arguments: string
  }
}

/** 与后端 chatMessage 一一对应的线上消息结构。 */
export interface ChatMessage {
  role: ChatRole
  content?: string
  tool_calls?: ChatToolCall[]
  tool_call_id?: string
}

/** /api/ai/chat/status 的响应。 */
export interface AIChatStatus {
  enabled: boolean
  model?: string
}

/** need_confirm 帧：写工具暂停，携带票据与已推进的完整对话状态。 */
export interface ChatPendingConfirm {
  ticket: string
  id: string
  name: string
  arguments: string
  summary: string
  messages: ChatMessage[]
}

/** 服务端 SSE 事件（按 event 名收敛后的判别联合）。 */
export type ChatEvent =
  | { type: 'meta'; model: string }
  | { type: 'delta'; text: string }
  | { type: 'tool_call'; id: string; name: string; arguments: string; readOnly: boolean }
  | { type: 'tool_result'; id: string; name: string; ok: boolean; content: string }
  | { type: 'need_confirm'; pending: ChatPendingConfirm }
  | { type: 'error'; message: string }
  | { type: 'done'; rounds: number; elapsedMs: number }

/** 查询服务端对话是否可用（未配置时前端隐藏入口）。 */
export async function fetchAIChatStatus(): Promise<AIChatStatus> {
  const res = await fetch(API.aiChatStatus, {
    method: 'GET',
    headers: { Authorization: getToken(), 'X-Requested-With': 'XMLHttpRequest' },
  })
  if (!res.ok) {
    throw new Error((await res.text()).trim() || `HTTP ${res.status}`)
  }
  return (await res.json()) as AIChatStatus
}

/**
 * 发起一轮对话并以回调接收 SSE 事件；signal 用于「停止生成」。
 * locale 是当前界面语言，服务端据此要求模型用同一语言回答。
 */
export function streamChat(
  messages: ChatMessage[],
  bindings: AIBinding[],
  locale: string,
  signal: AbortSignal,
  onEvent: (event: ChatEvent) => void,
): Promise<void> {
  return streamSSE(API.aiChat, { messages, bindings, locale }, signal, onEvent)
}

/**
 * 提交写工具确认结果并继续同一轮对话（messages 必须原样回传 need_confirm 帧里的值）。
 * locale 必须与首轮一致：服务端续跑时会重新拼装 system prompt。
 */
export function streamConfirm(
  ticket: string,
  approved: boolean,
  messages: ChatMessage[],
  bindings: AIBinding[],
  locale: string,
  signal: AbortSignal,
  onEvent: (event: ChatEvent) => void,
): Promise<void> {
  return streamSSE(
    API.aiChatConfirm,
    { ticket, approved, messages, bindings, locale },
    signal,
    onEvent,
  )
}

/** 发送 JSON 请求并逐帧解析 text/event-stream。 */
async function streamSSE(
  path: string,
  body: unknown,
  signal: AbortSignal,
  onEvent: (event: ChatEvent) => void,
): Promise<void> {
  const res = await fetch(path, {
    method: 'POST',
    headers: {
      Authorization: getToken(),
      'Content-Type': 'application/json',
      'X-Requested-With': 'XMLHttpRequest',
    },
    body: JSON.stringify(body),
    signal,
  })
  if (!res.ok) {
    throw new Error((await res.text()).trim() || `HTTP ${res.status}`)
  }
  if (!res.body) {
    throw new Error('当前浏览器不支持流式响应')
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  // 帧之间以空行分隔；心跳注释帧（": ping"）由 parseFrame 忽略。
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      const event = parseFrame(buffer.slice(0, boundary))
      buffer = buffer.slice(boundary + 2)
      if (event) onEvent(event)
      boundary = buffer.indexOf('\n\n')
    }
  }
  // 流结束时刷新解码器：多字节字符可能残留在最后一块里。
  buffer += decoder.decode()
  const tail = parseFrame(buffer)
  if (tail) onEvent(tail)
}

/** 把一帧 SSE 文本解析成事件；注释帧与不完整帧返回 null。 */
function parseFrame(block: string): ChatEvent | null {
  let event = ''
  const data: string[] = []
  for (const raw of block.split('\n')) {
    const line = raw.endsWith('\r') ? raw.slice(0, -1) : raw
    if (line.startsWith(':')) continue
    if (line.startsWith('event:')) {
      event = line.slice('event:'.length).trim()
      continue
    }
    if (line.startsWith('data:')) data.push(line.slice('data:'.length).replace(/^ /, ''))
  }
  if (event === '' || data.length === 0) return null
  const payload = JSON.parse(data.join('\n')) as Record<string, unknown>
  switch (event) {
    case 'meta':
      return { type: 'meta', model: String(payload.model ?? '') }
    case 'delta':
      return { type: 'delta', text: String(payload.text ?? '') }
    case 'tool_call':
      return {
        type: 'tool_call',
        id: String(payload.id ?? ''),
        name: String(payload.name ?? ''),
        arguments: String(payload.arguments ?? ''),
        readOnly: payload.read_only === true,
      }
    case 'tool_result':
      return {
        type: 'tool_result',
        id: String(payload.id ?? ''),
        name: String(payload.name ?? ''),
        ok: payload.ok === true,
        content: String(payload.content ?? ''),
      }
    case 'need_confirm':
      return {
        type: 'need_confirm',
        pending: {
          ticket: String(payload.ticket ?? ''),
          id: String(payload.id ?? ''),
          name: String(payload.name ?? ''),
          arguments: String(payload.arguments ?? ''),
          summary: String(payload.summary ?? ''),
          messages: (payload.messages ?? []) as ChatMessage[],
        },
      }
    case 'error':
      return { type: 'error', message: String(payload.message ?? '') }
    case 'done':
      return {
        type: 'done',
        rounds: Number(payload.rounds ?? 0),
        elapsedMs: Number(payload.elapsed_ms ?? 0),
      }
    default:
      return null
  }
}
