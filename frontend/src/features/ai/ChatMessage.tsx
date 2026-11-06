import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { Markdown } from '@/components/Markdown'
import type { ChatMessage as ChatMessageModel, ChatToolCall } from './aiChatApi'

/** 工具执行结果，按 tool_call_id 索引（由对话历史里的 tool 消息派生）。 */
export interface ToolOutcome {
  ok: boolean
  content: string
}

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
      <pre className="mt-1.5 overflow-x-auto rounded bg-surface px-2 py-1 text-[11px] leading-snug text-mute">
        {formatArguments(call.function.arguments)}
      </pre>
      {outcome && <ToolResultCard title={t('ai.chat.toolResult')} ok={outcome.ok} content={outcome.content} />}
    </div>
  )
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
      <pre className="mt-1 max-h-60 overflow-auto whitespace-pre-wrap break-words rounded bg-surface px-2 py-1 leading-snug text-mute">
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
