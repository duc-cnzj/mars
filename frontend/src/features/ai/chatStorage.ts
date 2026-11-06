import type { AIBinding } from './AIContext'
import type { ChatMessage } from './aiChatApi'

/**
 * AI 对话历史的浏览器本地持久化（先存前端本地，不设计数据库）。
 *
 * 按登录用户 email 分键：同一浏览器切换账号不会看到上一个人的对话。
 * 每个用户保存多段会话（会话列表可切换），每段会话带独立绑定：
 * 切回旧会话时恢复当时的空间/项目范围，避免历史内容与当前绑定错配。
 *
 * 清理机制：
 * - 时间：会话带最后更新时间，超过 7 天未更新在读取时直接清除；
 * - 条数：最多保留 30 段会话，单段会话最多 200 条消息（从最近的 user 消息边界截断，
 *   避免把工具调用对切散）；空会话不落盘。
 * 存储失败（隐私模式/配额超限）只影响「刷新后恢复历史」，不影响本次对话。
 */

const STORAGE_PREFIX = 'mars_ai_chat_v2'
/** 单段会话最多保留的消息条数。 */
const MAX_STORED_MESSAGES = 200
/** 每个用户最多保留的会话段数（按最后更新时间取最新）。 */
const MAX_STORED_SESSIONS = 30
/** 会话保留时长：以最后更新时间为准，超过 7 天不再恢复。 */
const HISTORY_TTL_MS = 7 * 24 * 60 * 60 * 1000
/** 会话标题最大长度（超出截断）。 */
const TITLE_MAX_LENGTH = 40

/** 一段本地会话：标题取自首条用户消息，绑定随会话保存。 */
export type ChatSession = {
  id: string
  title: string
  updatedAt: number
  bindings: AIBinding[]
  messages: ChatMessage[]
}

/** 本地存储结构：会话列表 + 当前选中的会话。 */
export type ChatStore = { activeId: string | null; sessions: ChatSession[] }

/** 生成指定用户的对话历史存储键。 */
function chatStorageKey(email: string): string {
  return `${STORAGE_PREFIX}:${email}`
}

/** 生成新的会话 id；优先用 crypto.randomUUID，缺失时退回时间戳 + 随机串。 */
export function newSessionId(): string {
  const cryptoObj = globalThis.crypto
  if (cryptoObj && typeof cryptoObj.randomUUID === 'function') return cryptoObj.randomUUID()
  return `s-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

/** 由消息推导会话标题：取首条用户消息，空白折叠后按 40 字符截断。 */
export function deriveSessionTitle(messages: ChatMessage[]): string {
  const first = messages.find((message) => message.role === 'user' && message.content?.trim())
  if (!first?.content) return ''
  const text = first.content.replace(/\s+/g, ' ').trim()
  return text.length > TITLE_MAX_LENGTH ? `${text.slice(0, TITLE_MAX_LENGTH)}…` : text
}

/**
 * 读取本地会话：丢弃损坏、空、超过保留期与超出段数上限的会话。
 * 记录整体不可解析，或所有会话都已过期时，直接删除存储键。
 */
export function loadChatStore(email: string): ChatStore {
  const empty: ChatStore = { activeId: null, sessions: [] }
  try {
    const key = chatStorageKey(email)
    const raw = localStorage.getItem(key)
    if (!raw) return empty
    const parsed = parseStore(raw)
    if (!parsed) {
      localStorage.removeItem(key)
      return empty
    }
    const now = Date.now()
    const sessions = parsed.sessions
      .filter((session) => session.messages.length > 0 && now - session.updatedAt <= HISTORY_TTL_MS)
      .sort((a, b) => b.updatedAt - a.updatedAt)
      .slice(0, MAX_STORED_SESSIONS)
    if (sessions.length === 0) {
      localStorage.removeItem(key)
      return empty
    }
    const activeId = sessions.some((session) => session.id === parsed.activeId)
      ? parsed.activeId
      : sessions[0].id
    return { activeId, sessions }
  } catch {
    return empty
  }
}

/** 保存会话列表：空会话不落盘；全部为空时删除存储键。 */
export function saveChatStore(email: string, store: ChatStore): void {
  try {
    const key = chatStorageKey(email)
    const sessions = store.sessions
      .filter((session) => session.messages.length > 0)
      .sort((a, b) => b.updatedAt - a.updatedAt)
      .slice(0, MAX_STORED_SESSIONS)
      .map((session) => ({ ...session, messages: pruneMessages(session.messages) }))
    if (sessions.length === 0) {
      localStorage.removeItem(key)
      return
    }
    const activeId = sessions.some((session) => session.id === store.activeId)
      ? store.activeId
      : sessions[0].id
    localStorage.setItem(key, JSON.stringify({ activeId, sessions }))
  } catch {
    // 隐私模式/配额超限：历史不落盘，但本次对话仍然可用。
  }
}

/** 截断单段会话：从 user 消息边界开始，保证不留下孤立的 assistant/tool 片段。 */
function pruneMessages(messages: ChatMessage[]): ChatMessage[] {
  if (messages.length <= MAX_STORED_MESSAGES) return messages
  let start = messages.length - MAX_STORED_MESSAGES
  while (start < messages.length && messages[start].role !== 'user') start++
  if (start >= messages.length) start = messages.length - MAX_STORED_MESSAGES
  return messages.slice(start)
}

/** 解析存储记录；结构不符时返回 null（按损坏数据处理）。 */
function parseStore(raw: string): ChatStore | null {
  const parsed: unknown = JSON.parse(raw)
  if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) return null
  const record = parsed as Record<string, unknown>
  if (!Array.isArray(record.sessions)) return null
  const sessions = record.sessions
    .map(toSession)
    .filter((session): session is ChatSession => session !== null)
  return { activeId: typeof record.activeId === 'string' ? record.activeId : null, sessions }
}

/** 还原单段会话；必要字段缺失时返回 null。 */
function toSession(value: unknown): ChatSession | null {
  if (typeof value !== 'object' || value === null) return null
  const record = value as Record<string, unknown>
  if (typeof record.id !== 'string' || typeof record.updatedAt !== 'number') return null
  if (!Array.isArray(record.messages)) return null
  return {
    id: record.id,
    title: typeof record.title === 'string' ? record.title : '',
    updatedAt: record.updatedAt,
    bindings: Array.isArray(record.bindings) ? record.bindings.filter(isBinding) : [],
    messages: record.messages.filter(isChatMessage),
  }
}

/** 运行时结构校验：绑定只信任 `namespace`（字符串）与可选 `project`（字符串）。 */
function isBinding(value: unknown): value is AIBinding {
  if (typeof value !== 'object' || value === null) return false
  const record = value as Record<string, unknown>
  if (typeof record.namespace !== 'string' || record.namespace === '') return false
  return record.project === undefined || typeof record.project === 'string'
}

/** 运行时结构校验：只信任最基本的消息形状，损坏数据直接丢弃。 */
function isChatMessage(value: unknown): value is ChatMessage {
  if (typeof value !== 'object' || value === null) return false
  const message = value as Record<string, unknown>
  const role = message.role
  if (role !== 'user' && role !== 'assistant' && role !== 'tool') return false
  if (message.content !== undefined && typeof message.content !== 'string') return false
  if (message.tool_call_id !== undefined && typeof message.tool_call_id !== 'string') return false
  if (message.tool_calls !== undefined && !Array.isArray(message.tool_calls)) return false
  return true
}
