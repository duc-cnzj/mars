import { Fragment } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { Markdown } from '@/components/Markdown'
import type { ChatMessage as ChatMessageModel, ChatToolCall } from './aiChatApi'

/** 工具执行结果，按 tool_call_id 索引（由对话历史里的 tool 消息派生）。 */
export interface ToolOutcome {
  ok: boolean
  content: string
}

/** 终端黑：命令输入与执行结果固定用深色终端配色，不随亮/暗主题变化。
 *  字体用显式等宽栈而非主题 mono token：终端块刻意脱离主题，不能依赖外层 theme class。 */
const TERMINAL_BLOCK =
  'rounded-md border border-white/10 bg-[#0b0e14] text-[#d4d4d4] [font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace]'

/** JSON 语法高亮的配色：键/字符串/数字/字面量/结构符分色，贴合深色终端观感。 */
const JSON_COLOR = {
  key: 'text-[#7ee787]',
  string: 'text-[#a5d6ff]',
  number: 'text-[#f2cc60]',
  literal: 'text-[#ff7b72]',
  punct: 'text-[#8b949e]',
} as const

/**
 * 单条对话消息的渲染：
 * - user → 右侧气泡（保留换行）；
 * - assistant → Markdown 正文 + 本次请求的每个工具调用卡片；
 * - tool → 孤立结果卡（正常情况由对应 assistant 的调用卡片展示，仅在历史被截断时单独出现）。
 */
export function ChatMessage({
  message,
  outcomes,
  pendingId,
}: {
  message: ChatMessageModel
  /** tool_call_id → 执行结果；缺失表示仍在执行或等待确认。 */
  outcomes: Record<string, ToolOutcome>
  /** 等待用户确认的写工具调用 id（该卡片高亮）。 */
  pendingId?: string
}) {
  const { t } = useTranslation()

  if (message.role === 'user') {
    return (
      <div className="flex justify-end">
        <div className="max-w-[85%] whitespace-pre-wrap break-words rounded-2xl rounded-br-md bg-primary/10 px-3.5 py-2 text-[13px] leading-relaxed text-ink">
          {message.content}
        </div>
      </div>
    )
  }

  if (message.role === 'tool') {
    return (
      <ToolResultCard
        title={t('ai.chat.toolResult')}
        ok
        content={message.content ?? ''}
      />
    )
  }

  return (
    <div className="flex min-w-0 flex-col gap-2">
      {message.content?.trim() ? <Markdown>{message.content}</Markdown> : null}
      {message.tool_calls?.map((call) => (
        <ToolCallCard
          key={call.id}
          call={call}
          outcome={outcomes[call.id]}
          pending={pendingId === call.id}
        />
      ))}
    </div>
  )
}

/** 工具调用卡片：名称/参数/状态；结果默认折叠，避免长输出淹没对话。 */
function ToolCallCard({
  call,
  outcome,
  pending,
}: {
  call: ChatToolCall
  outcome?: ToolOutcome
  pending: boolean
}) {
  const { t } = useTranslation()
  return (
    <div
      className={`rounded-lg border px-3 py-2 text-[12px] ${
        pending ? 'border-primary/60 bg-primary/5' : 'border-line bg-raised/40'
      }`}
    >
      <div className="flex flex-wrap items-center gap-1.5">
        <Icon name="terminal" className="shrink-0 text-primary" />
        <span className="font-medium text-ink">{call.function.name}</span>
        <ToolStatus outcome={outcome} pending={pending} />
      </div>
      {/* 命令输入：终端黑底等宽；JSON 参数按语法高亮，非法 JSON 原文兜底 */}
      <ArgumentsBlock raw={call.function.arguments} />
      {outcome && <ToolResultCard title={t('ai.chat.toolResult')} ok={outcome.ok} content={outcome.content} />}
    </div>
  )
}

/**
 * 工具参数块：合法 JSON 按缩进结构渲染并逐 token 着色；
 * 空参数展示 `{}`，非法 JSON（模型偶尔拼出半截）退回原文，不做猜测修复。
 */
function ArgumentsBlock({ raw }: { raw: string }) {
  const trimmed = raw.trim()
  let parsed: unknown = {}
  let valid = true
  if (trimmed !== '') {
    try {
      parsed = JSON.parse(trimmed)
    } catch {
      valid = false
    }
  }
  return (
    <pre
      className={`mt-1.5 overflow-x-auto px-2 py-1.5 text-[11px] leading-snug ${TERMINAL_BLOCK}`}
    >
      {valid ? <JsonPretty value={parsed} depth={0} /> : raw}
    </pre>
  )
}

/** JSON 值递归渲染（两空格缩进，与 JSON.stringify(_, null, 2) 对齐）：按类型分色。 */
function JsonPretty({ value, depth }: { value: unknown; depth: number }) {
  const pad = '  '.repeat(depth)
  const inner = '  '.repeat(depth + 1)
  if (value === null) return <span className={JSON_COLOR.literal}>null</span>
  if (Array.isArray(value)) {
    if (value.length === 0) return <span className={JSON_COLOR.punct}>[]</span>
    return (
      <>
        <span className={JSON_COLOR.punct}>[</span>
        {value.map((item, index) => (
          <Fragment key={index}>
            {'\n'}
            {inner}
            <JsonPretty value={item} depth={depth + 1} />
            {index < value.length - 1 ? <span className={JSON_COLOR.punct}>,</span> : null}
          </Fragment>
        ))}
        {'\n'}
        {pad}
        <span className={JSON_COLOR.punct}>]</span>
      </>
    )
  }
  if (typeof value === 'string') {
    return <span className={JSON_COLOR.string}>{JSON.stringify(value)}</span>
  }
  if (typeof value === 'number') {
    return <span className={JSON_COLOR.number}>{String(value)}</span>
  }
  if (typeof value === 'boolean') {
    return <span className={JSON_COLOR.literal}>{String(value)}</span>
  }
  if (typeof value === 'object') {
    const entries = Object.entries(value as Record<string, unknown>)
    if (entries.length === 0) return <span className={JSON_COLOR.punct}>{'{}'}</span>
    return (
      <>
        <span className={JSON_COLOR.punct}>{'{'}</span>
        {entries.map(([key, item], index) => (
          <Fragment key={key}>
            {'\n'}
            {inner}
            <span className={JSON_COLOR.key}>{JSON.stringify(key)}</span>
            <span className={JSON_COLOR.punct}>{': '}</span>
            <JsonPretty value={item} depth={depth + 1} />
            {index < entries.length - 1 ? <span className={JSON_COLOR.punct}>,</span> : null}
          </Fragment>
        ))}
        {'\n'}
        {pad}
        <span className={JSON_COLOR.punct}>{'}'}</span>
      </>
    )
  }
  return <span className={JSON_COLOR.literal}>null</span>
}

/** 工具状态徽标：等待确认 / 执行中 / 成功 / 失败。 */
function ToolStatus({ outcome, pending }: { outcome?: ToolOutcome; pending: boolean }) {
  const { t } = useTranslation()
  if (pending) return <span className="text-primary">{t('ai.chat.toolPending')}</span>
  if (!outcome) return <span className="text-mute">{t('ai.chat.toolRunning')}</span>
  if (outcome.ok) {
    return (
      <span className="inline-flex items-center gap-1 text-emerald-600">
        <Icon name="circle-check" />
        {t('ai.chat.toolOK')}
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 text-destructive">
      <Icon name="circle-x" />
      {t('ai.chat.toolError')}
    </span>
  )
}

/** 可折叠的工具结果块。 */
function ToolResultCard({ title, ok, content }: { title: string; ok: boolean; content: string }) {
  return (
    <details className="mt-1.5 text-[11px]">
      <summary className={`cursor-pointer select-none ${ok ? 'text-mute' : 'text-destructive'}`}>
        {title}
      </summary>
      {/* 命令结果：与输入同一套终端黑底，失败结果仍靠上方 summary 的红色区分 */}
      <pre
        className={`mt-1 max-h-60 overflow-auto whitespace-pre-wrap break-words px-2 py-1.5 leading-snug ${TERMINAL_BLOCK}`}
      >
        {content}
      </pre>
    </details>
  )
}

/** 参数格式化：合法 JSON 缩进展示，非法原文兜底（模型偶尔会拼出半截参数）。
 *  工具卡片与确认条共用同一份展示口径。 */
export function formatArguments(raw: string): string {
  if (!raw.trim()) return '{}'
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}
