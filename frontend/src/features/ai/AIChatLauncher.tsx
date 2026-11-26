import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/features/auth/AuthProvider'
import { useDraggableDialog } from '@/hooks/useDraggableDialog'
import { useOverlayZ } from '@/hooks/useOverlayZ'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/shadcn/dialog'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/shadcn/alert-dialog'
import { Button } from '@/components/ui/shadcn/button'
import { AISparklesIcon, Icon } from '@/components/Icons'
import { DialogDragFrame } from '@/components/DialogDragFrame'
import { humanizeDateTime } from '@/lib/humanizeDateTime'
import { toast } from '@/lib/toast'
import { reportAIUsage } from './aiApi'
import {
  streamChat,
  streamConfirm,
  type ChatEvent,
  type ChatMessage as ChatMessageModel,
  type ChatPendingConfirm,
} from './aiChatApi'
import {
  deriveSessionTitle,
  loadChatStore,
  newSessionId,
  saveChatStore,
  type ChatSession,
  type ChatStore,
} from './chatStorage'
import { loadAutoApprove, saveAutoApprove } from './aiPrefs'
import { AIBindingBar } from './AIBindingBar'
import { AIQuickPrompts } from './AIQuickPrompts'
import { ChatMessage, formatArguments, type ToolOutcome } from './ChatMessage'
import { aiBindingKey, type AIBinding } from './AIContext'

/** 空消息/空绑定常量：保证「空态」下引用稳定，避免无谓重渲染。 */
const EMPTY_MESSAGES: ChatMessageModel[] = []
const EMPTY_BINDINGS: AIBinding[] = []

/** 输入框内置命令（整条消息恰好等于 `/命令` 才触发）。 */
const SLASH_COMMANDS = ['new', 'clear'] as const
type SlashCommand = (typeof SLASH_COMMANDS)[number]

/** 命令说明文案的 i18n key（输入 `/` 时的提示行使用）。 */
const SLASH_COMMAND_HINT_KEYS = {
  new: 'ai.chat.commandNewHint',
  clear: 'ai.chat.commandClearHint',
} as const satisfies Record<SlashCommand, string>

/** 识别内置命令；只认整条消息精确命中，`/api 路径` 这类正常提问照旧发给模型。 */
function parseSlashCommand(text: string): SlashCommand | null {
  const match = text.trim().match(/^\/([a-z]+)$/i)
  if (!match) return null
  const name = match[1].toLowerCase()
  return (SLASH_COMMANDS as readonly string[]).includes(name) ? (name as SlashCommand) : null
}

/** 入口问题的归一化：空白折叠成单行，空串按「无问题」处理。 */
function normalizePrompt(prompt?: string): string | undefined {
  const value = prompt?.replace(/\s+/g, ' ').trim()
  return value ? value : undefined
}

/** 光标是否停在首行且无选区：此时 ↑ 才回溯历史，其余情况交给默认的光标移动。 */
function isCaretOnFirstLine(el: HTMLTextAreaElement): boolean {
  if (el.selectionStart !== el.selectionEnd) return false
  return !el.value.slice(0, el.selectionStart).includes('\n')
}

/** 两组绑定是否等价（`空间/项目` 集合比较，顺序无关）。 */
function sameBindings(a: AIBinding[], b: AIBinding[]): boolean {
  if (a.length !== b.length) return false
  const set = new Set(b.map(aiBindingKey))
  return a.every((item) => set.has(aiBindingKey(item)))
}

/** 新建一段空会话（未发言，不落盘）。 */
function makeSession(bindings: AIBinding[]): ChatSession {
  return { id: newSessionId(), title: '', updatedAt: Date.now(), bindings, messages: [] }
}

/** 取当前选中会话；存储损坏时可能找不到，按空态处理。 */
function getActive(store: ChatStore): ChatSession | null {
  return store.sessions.find((session) => session.id === store.activeId) ?? null
}

/**
 * 服务端 AI 对话面板：直接调 `/api/ai/chat` 流式接口；对话历史只保存在浏览器
 * localStorage（按用户 email 分键，服务端零会话状态；每个用户保留多段会话，
 * 超过 7 天未更新自动清理，详见 chatStorage）。
 *
 * 交互：
 * - 空白对话展示常见问题，点击即发送；
 * - 只读工具自动执行，写工具暂停并展示确认条（批准/拒绝后继续同一轮对话）；
 * - 生成中可「停止」中断 SSE，已生成的内容保留；
 * - 输入框 ↑/↓ 像命令行一样回溯当前会话已发送的问题（未发送的草稿会暂存并恢复）；
 * - 输入框支持内置命令（整条消息精确命中才触发）：`/new` 新建对话、`/clear` 清除当前对话；
 * - 场景入口（项目诊断/日志分析）带来的问题在打开时自动发送；
 * - 「历史对话」可切换此前的会话，每段会话恢复当时的绑定范围；
 * - 换绑不会污染旧会话：已有内容时另开一段新会话，旧会话保留在历史里。
 */
export function AIChatLauncher({
  open,
  onClose,
  initialPrompt,
  initialBinding,
  assistantTitle,
  entryNonce = 0,
}: {
  open: boolean
  onClose: () => void
  /** 场景入口带来的首条问题：打开时自动发送。 */
  initialPrompt?: string
  /** 项目卡片入口带来的默认绑定：作为对话上下文随请求下发。 */
  initialBinding?: AIBinding
  /** 面板标题（如「duc 的 AI 助手」）。 */
  assistantTitle?: string
  /** 入口点击序号：面板已打开时再次点击入口会递增，用于只处理一次新上下文。 */
  entryNonce?: number
}) {
  const { t, i18n } = useTranslation()
  /** 当前界面语言：随每次请求下发，服务端要求模型用同一语言回答。 */
  const localeRef = useRef(i18n.language)
  localeRef.current = i18n.language
  const { user } = useAuth()
  const email = user?.email ?? ''
  const dnd = useDraggableDialog()

  /** 会话列表 + 当前选中会话；唯一的状态源，messages/bindings 均由其派生。 */
  const [store, setStore] = useState<ChatStore>(() => {
    const session = makeSession(initialBinding ? [initialBinding] : [])
    return { activeId: session.id, sessions: [session] }
  })
  /** 流式回调与事件处理器读 ref，避免闭包拿到旧会话列表。 */
  const storeRef = useRef(store)
  const emailRef = useRef(email)
  emailRef.current = email
  /** 历史面板是否展开。 */
  const [historyOpen, setHistoryOpen] = useState(false)
  /** 待删除的会话 id（非空时弹出删除确认框）。 */
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  /** 确认框 z：打开时重新取 nextZIndex()，保证盖过同样在累加 z 的可拖拽宿主弹窗。 */
  const deleteDialogZ = useOverlayZ(deleteTarget !== null)
  /** 已按哪个用户加载过本地历史（换账号时重新读取，不继承上一个人的内存会话）。 */
  const historyEmailRef = useRef<string | null>(null)

  /** 正在流式生成（含等待上游首字节）。 */
  const [streaming, setStreaming] = useState(false)
  const streamingRef = useRef(false)
  const abortRef = useRef<AbortController | null>(null)
  /** 等待确认的写工具；非空时禁止发送新消息、切换会话与换绑。 */
  const [pending, setPending] = useState<ChatPendingConfirm | null>(null)
  const pendingRef = useRef<ChatPendingConfirm | null>(null)
  const [error, setError] = useState<string | null>(null)
  /** 免确认模式：写工具到达时前端自动批准，不再逐次弹确认条（票据校验仍在服务端）。 */
  const [autoApprove, setAutoApprove] = useState(false)
  const autoApproveRef = useRef(false)
  autoApproveRef.current = autoApprove
  /** 已按哪个用户恢复过免确认开关（换账号时重新读取）。 */
  const autoApproveLoadedForRef = useRef<string | null>(null)
  useEffect(() => {
    if (!email || autoApproveLoadedForRef.current === email) return
    autoApproveLoadedForRef.current = email
    const stored = loadAutoApprove(email)
    autoApproveRef.current = stored
    setAutoApprove(stored)
  }, [email])
  /** 本轮 SSE 结束时是否需要自动批准刚收到的 need_confirm。 */
  const autoConfirmRef = useRef(false)
  /** 自动批准走与按钮相同的提交路径；用 ref 避免 run/decide 的循环依赖。 */
  const decideRef = useRef<(approved: boolean) => void>(() => {})
  const [model, setModel] = useState('')
  const [draft, setDraft] = useState('')
  /** 输入区 DOM：↑/↓ 回溯历史时读取光标位置，回填后复位光标。 */
  const textareaRef = useRef<HTMLTextAreaElement | null>(null)
  /** 历史回溯游标：-1 表示未在浏览历史，否则指向 inputHistory 的下标。 */
  const historyCursorRef = useRef(-1)
  /** 进入历史浏览前暂存的草稿，↓ 越过最新一条时恢复。 */
  const stashedDraftRef = useRef('')
  /** 本次 draft 更新来自历史回填：提交后把光标移到末尾。 */
  const caretToEndRef = useRef(false)
  /** 命令提示里当前高亮的下标（↑/↓ 在提示项间移动，不复用历史回溯游标）。 */
  const [commandIndex, setCommandIndex] = useState(0)

  // 回填历史后把光标移到末尾：等本次提交写入 DOM 值后再设置选区。
  useEffect(() => {
    if (!caretToEndRef.current) return
    caretToEndRef.current = false
    const el = textareaRef.current
    if (el) el.setSelectionRange(el.value.length, el.value.length)
  }, [draft])

  /** 退出历史浏览，回到「输入的是草稿」的正常编辑态。 */
  const resetHistoryNav = useCallback(() => {
    historyCursorRef.current = -1
    stashedDraftRef.current = ''
  }, [])

  /** 本轮对话是否已上报首 token 事件（每轮只上报一次）。 */
  const firstTokenRef = useRef(false)
  /** tool_call_id → 运行时结果标志（tool 消息只带内容，ok 由事件补充）。 */
  const [toolStatus, setToolStatus] = useState<Record<string, { ok: boolean }>>({})
  const listRef = useRef<HTMLDivElement | null>(null)

  const activeSession = getActive(store)
  const messages = activeSession?.messages ?? EMPTY_MESSAGES
  const bindings = activeSession?.bindings ?? EMPTY_BINDINGS

  /** 当前会话已发送的用户消息（旧 → 新），↑/↓ 在这份列表上回溯。 */
  const inputHistory = useMemo(
    () =>
      messages
        .filter((message) => message.role === 'user')
        .map((message) => message.content ?? '')
        .filter((content) => content.trim() !== ''),
    [messages],
  )
  const inputHistoryRef = useRef(inputHistory)
  inputHistoryRef.current = inputHistory

  /**
   * 输入 `/` 时的命令提示：仅在「整条输入是一个命令前缀」时出现（`/`、`/cl`），
   * 输入已完成某条命令（`/clear`）或包含空格/路径（`/api/health`）时自动收起。
   */
  const commandMatches = useMemo(() => {
    const match = draft.trim().match(/^\/([a-z]*)$/i)
    if (!match) return []
    const prefix = match[1].toLowerCase()
    return SLASH_COMMANDS.filter((name) => name.startsWith(prefix) && name !== prefix)
  }, [draft])
  /** 高亮项：输入收窄后索引可能越界，读取时按当前候选数收敛，避免取到 undefined。 */
  const activeCommandIndex =
    commandMatches.length === 0 ? 0 : Math.min(commandIndex, commandMatches.length - 1)

  // 输入变化后高亮回到第一条，避免沿用上一次的索引落到别的命令上。
  useEffect(() => {
    setCommandIndex(0)
  }, [draft])

  // 切换会话后只回溯新会话的历史：游标与暂存草稿一并清空。
  useEffect(() => {
    resetHistoryNav()
  }, [activeSession?.id, resetHistoryNav])

  /** 写入会话列表：同步 ref 再更新 state，保证后续同步回调读到最新值。 */
  const applyStore = useCallback((next: ChatStore) => {
    storeRef.current = next
    setStore(next)
  }, [])

  /** 立即落盘当前会话列表（流结束、切会话、卸载时调用）。 */
  const persistNow = useCallback(() => {
    if (emailRef.current) saveChatStore(emailRef.current, storeRef.current)
  }, [])

  /** 重置单轮运行时状态（待确认、错误、工具结果标志），不影响消息本身。 */
  const resetRoundState = useCallback(() => {
    pendingRef.current = null
    setPending(null)
    setError(null)
    setToolStatus({})
    firstTokenRef.current = false
  }, [])

  /** 更新当前会话的消息：同步标题（首次用户消息）与最后更新时间。 */
  const setMessages = useCallback(
    (next: ChatMessageModel[]) => {
      const current = storeRef.current
      const active = getActive(current)
      if (!active || next === active.messages) return
      const updated: ChatSession = {
        ...active,
        messages: next,
        updatedAt: Date.now(),
        title: active.title || deriveSessionTitle(next),
      }
      applyStore({
        activeId: current.activeId,
        sessions: current.sessions.map((session) =>
          session.id === updated.id ? updated : session,
        ),
      })
    },
    [applyStore],
  )

  // 历史只加载一次每个用户：恢复后按 email 读取本地会话（同一用户不重复覆盖内存状态）。
  useEffect(() => {
    if (!email || historyEmailRef.current === email) return
    const switchingUser = historyEmailRef.current !== null
    historyEmailRef.current = email
    const stored = loadChatStore(email)
    const memoryEmpty = storeRef.current.sessions.every(
      (session) => session.messages.length === 0,
    )
    if (stored.sessions.length > 0 && memoryEmpty) {
      applyStore(stored)
      return
    }
    // 换账号且新账号没有历史：丢掉上一个人的内存会话，从空白开始。
    if (switchingUser) {
      const session = makeSession([])
      applyStore({ activeId: session.id, sessions: [session] })
    }
  }, [email, applyStore])

  // 会话变化后防抖落盘：流式增量不应该每个 token 都写一次 localStorage。
  useEffect(() => {
    if (!email || historyEmailRef.current !== email) return
    const timer = window.setTimeout(() => saveChatStore(email, store), 400)
    return () => window.clearTimeout(timer)
  }, [email, store])

  // 面板卸载时立即落盘，避免防抖计时器被取消导致最后一段内容丢失。
  useEffect(() => () => persistNow(), [persistNow])

  // 新内容到达后滚到底部；面板无 DOM 的测试环境里 listRef 为空，跳过即可。
  useEffect(() => {
    const el = listRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [messages, streaming, pending])

  /** 处理一帧 SSE 事件：文本追加、工具卡片更新、暂停确认与错误。 */
  const handleEvent = useCallback(
    (event: ChatEvent) => {
      switch (event.type) {
        case 'meta':
          setModel(event.model)
          break
        case 'delta':
          if (!firstTokenRef.current) {
            firstTokenRef.current = true
            reportAIUsage('chat_first_token')
          }
          setMessages(appendAssistantText(messagesRefOf(storeRef.current), event.text))
          break
        case 'tool_call':
          setMessages(upsertToolCall(messagesRefOf(storeRef.current), event))
          break
        case 'tool_result':
          setToolStatus((prev) => ({ ...prev, [event.id]: { ok: event.ok } }))
          setMessages(appendToolResult(messagesRefOf(storeRef.current), event))
          break
        case 'need_confirm':
          // 服务端回传的是已推进的完整对话状态：直接替换本地历史，避免重复推进。
          setMessages(event.pending.messages)
          pendingRef.current = event.pending
          setPending(event.pending)
          // 免确认模式：本帧流结束后自动批准（不能在这里直接批准，active 流尚未收起）。
          if (autoApproveRef.current) autoConfirmRef.current = true
          break
        case 'error':
          setError(event.message)
          reportAIUsage('chat_error')
          break
        case 'done':
          reportAIUsage('chat_done')
          break
      }
    },
    [setMessages],
  )

  /** 跑一次 SSE 请求：统一管理 AbortController、错误兜底与「生成中」状态。 */
  const run = useCallback(
    async (starter: (signal: AbortSignal) => Promise<void>) => {
      const controller = new AbortController()
      abortRef.current = controller
      streamingRef.current = true
      firstTokenRef.current = false
      setStreaming(true)
      setError(null)
      try {
        await starter(controller.signal)
      } catch (e) {
        // 用户主动停止时 fetch 抛 AbortError：不是错误，保留已生成内容。
        if (!controller.signal.aborted) {
          setError(e instanceof Error ? e.message : String(e))
          reportAIUsage('chat_error')
        }
      } finally {
        if (abortRef.current === controller) {
          abortRef.current = null
          streamingRef.current = false
          setStreaming(false)
          persistNow()
          // need_confirm 帧结束后流已收口，此时才能发起下一次请求（decide 拒绝并发流）。
          // 用户已点「停止」时不自动批准：中断优先于免确认模式。
          const shouldAutoConfirm =
            !controller.signal.aborted && autoConfirmRef.current && pendingRef.current !== null
          autoConfirmRef.current = false
          if (shouldAutoConfirm) decideRef.current(true)
        }
      }
    },
    [persistNow],
  )

  /** 发送一条用户消息；生成中或等待确认时忽略。 */
  const send = useCallback(
    (text: string) => {
      const content = text.trim()
      if (!content || streamingRef.current || pendingRef.current) return
      const active = getActive(storeRef.current)
      if (!active) return
      const next = [...active.messages, { role: 'user', content } as ChatMessageModel]
      setMessages(next)
      resetHistoryNav()
      setDraft('')
      void run((signal) =>
        streamChat(next, active.bindings, localeRef.current, signal, handleEvent),
      )
    },
    [handleEvent, resetHistoryNav, run, setMessages],
  )

  /** 常见问题填充属于新输入而非历史回溯：先退出浏览态，避免 ↑ 拿到错位的游标。 */
  const fillDraft = useCallback(
    (value: string) => {
      resetHistoryNav()
      setDraft(value)
    },
    [resetHistoryNav],
  )

  /** 提交写工具确认结果；消息数组原样回传 need_confirm 帧里的状态。 */
  const decide = useCallback(
    (approved: boolean) => {
      const current = pendingRef.current
      if (!current || streamingRef.current) return
      const activeBindings = getActive(storeRef.current)?.bindings ?? EMPTY_BINDINGS
      pendingRef.current = null
      setPending(null)
      void run((signal) =>
        streamConfirm(
          current.ticket,
          approved,
          current.messages,
          activeBindings,
          localeRef.current,
          signal,
          handleEvent,
        ),
      )
    },
    [handleEvent, run],
  )
  // 自动批准复用 decide：它的并发/票据清理逻辑与手动点击完全一致。
  useEffect(() => {
    decideRef.current = decide
  }, [decide])

  /** 停止生成：中断 SSE，保留已收到的内容。 */
  const stop = useCallback(() => {
    abortRef.current?.abort()
  }, [])

  /** 免确认开关：切换即按当前用户持久化，刷新或重新打开面板后保持上次选择。 */
  const toggleAutoApprove = useCallback((enabled: boolean) => {
    setAutoApprove(enabled)
    if (emailRef.current) saveAutoApprove(emailRef.current, enabled)
  }, [])

  /** 开新对话：有内容时另起一段会话（旧会话保留在历史里）；已是空会话则只关闭历史面板。 */
  const newChat = useCallback(() => {
    if (streamingRef.current || pendingRef.current) return
    setHistoryOpen(false)
    const active = getActive(storeRef.current)
    if (!active || active.messages.length === 0) return
    const session = makeSession(active.bindings)
    const current = storeRef.current
    resetRoundState()
    applyStore({ activeId: session.id, sessions: [session, ...current.sessions] })
  }, [applyStore, resetRoundState])

  /**
   * 删除指定会话（历史面板逐条删除 / 删除当前会话共用）：
   * 删除的是当前会话时切到最近一段剩余会话；全部删光则回到空白会话（沿用被删会话的绑定）。
   */
  const deleteSession = useCallback(
    (id: string) => {
      if (streamingRef.current || pendingRef.current) return
      const current = storeRef.current
      const target = current.sessions.find((session) => session.id === id)
      if (!target) return
      const remaining = current.sessions.filter((session) => session.id !== id)
      if (current.activeId !== id) {
        applyStore({ ...current, sessions: remaining })
        return
      }
      resetRoundState()
      const latest = remaining
        .filter((session) => session.messages.length > 0)
        .sort((a, b) => b.updatedAt - a.updatedAt)[0]
      if (latest) {
        applyStore({ activeId: latest.id, sessions: remaining })
        return
      }
      const session = makeSession(target.bindings)
      applyStore({ activeId: session.id, sessions: [session] })
    },
    [applyStore, resetRoundState],
  )

  /** 删除当前会话：空会话无事可删，忽略；有内容时同样要过二次确认。 */
  const clearChat = useCallback(() => {
    if (streamingRef.current || pendingRef.current) return
    const active = getActive(storeRef.current)
    if (!active || active.messages.length === 0) return
    setDeleteTarget(active.id)
  }, [])

  /** 命令用的一键清除：输入 `/clear` 已是明确意图，直接删除当前会话，不再弹确认框。 */
  const clearChatNow = useCallback(() => {
    if (streamingRef.current || pendingRef.current) return
    const active = getActive(storeRef.current)
    if (!active || active.messages.length === 0) return
    deleteSession(active.id)
  }, [deleteSession])

  /** 执行内置命令：/new 新建对话；/clear 直接清除当前会话（无二次确认）。 */
  const runSlashCommand = useCallback(
    (command: SlashCommand) => {
      if (command === 'new') {
        newChat()
        return
      }
      clearChatNow()
    },
    [newChat, clearChatNow],
  )

  /** 执行命令：清空草稿并交给与按钮一致的处理；生成中/等待确认时保留草稿不执行。 */
  const executeCommand = useCallback(
    (command: SlashCommand) => {
      if (streamingRef.current || pendingRef.current) return
      resetHistoryNav()
      setDraft('')
      runSlashCommand(command)
    },
    [resetHistoryNav, runSlashCommand],
  )

  /** 提交输入框：命中内置命令就在本地执行（不发给模型），否则按普通问题发送。 */
  const submitDraft = useCallback(() => {
    const command = parseSlashCommand(draft)
    if (!command) {
      send(draft)
      return
    }
    executeCommand(command)
  }, [draft, executeCommand, send])

  /** Enter 与「发送」按钮共用：命令提示展开时执行高亮项，避免把半截 `/` 当消息发给模型。 */
  const submitInput = useCallback(() => {
    const picked = commandMatches[activeCommandIndex]
    if (picked) {
      executeCommand(picked)
      return
    }
    submitDraft()
  }, [activeCommandIndex, commandMatches, executeCommand, submitDraft])

  /** 切换历史会话：恢复该会话的消息与绑定（生成中/等待确认时禁止切换）。 */
  const switchSession = useCallback(
    (id: string) => {
      if (streamingRef.current || pendingRef.current) return
      const current = storeRef.current
      if (id === current.activeId) {
        setHistoryOpen(false)
        return
      }
      if (!current.sessions.some((session) => session.id === id)) return
      resetRoundState()
      applyStore({ ...current, activeId: id })
      setHistoryOpen(false)
    },
    [applyStore, resetRoundState],
  )

  /**
   * 换绑语义：空会话原地更新绑定；已有内容时另开一段新会话，旧会话保留在历史里，
   * 避免旧消息被模型当作既定上下文继续在旧范围上操作。
   * 生成中/等待确认时禁止换绑（绑定条同步置灰），避免打断进行中的轮次。
   */
  const handleBindingsChange = useCallback(
    (next: AIBinding[]) => {
      if (streamingRef.current || pendingRef.current) return
      const current = storeRef.current
      const active = getActive(current)
      if (!active || sameBindings(active.bindings, next)) return
      if (active.messages.length === 0) {
        const updated: ChatSession = { ...active, bindings: next, updatedAt: Date.now() }
        applyStore({
          activeId: current.activeId,
          sessions: current.sessions.map((session) =>
            session.id === updated.id ? updated : session,
          ),
        })
        return
      }
      const session = makeSession(next)
      resetRoundState()
      setHistoryOpen(false)
      applyStore({ activeId: session.id, sessions: [session, ...current.sessions] })
      toast.info(t('ai.chat.bindingSwitched'))
    },
    [applyStore, resetRoundState, t],
  )

  // 场景入口：按 nonce 只处理一次新上下文（追加绑定 + 自动发送问题）。
  // 与手动换绑同语义：新增绑定且已有历史时另起会话，保证首条问题只按新范围作答。
  const handledEntryRef = useRef<number | null>(null)
  useEffect(() => {
    if (!open || handledEntryRef.current === entryNonce) return
    handledEntryRef.current = entryNonce
    const busy = streamingRef.current || pendingRef.current !== null
    if (initialBinding && !busy) {
      const current = storeRef.current
      const active = getActive(current)
      const alreadyBound = active?.bindings.some(
        (item) => aiBindingKey(item) === aiBindingKey(initialBinding),
      )
      if (active && !alreadyBound) {
        const nextBindings = [...active.bindings, initialBinding]
        if (active.messages.length > 0) {
          const session = makeSession(nextBindings)
          resetRoundState()
          setHistoryOpen(false)
          applyStore({ activeId: session.id, sessions: [session, ...current.sessions] })
        } else {
          const updated: ChatSession = {
            ...active,
            bindings: nextBindings,
            updatedAt: Date.now(),
          }
          applyStore({
            activeId: current.activeId,
            sessions: current.sessions.map((session) =>
              session.id === updated.id ? updated : session,
            ),
          })
        }
      }
    }
    const prompt = normalizePrompt(initialPrompt)
    if (prompt) send(prompt)
  }, [open, entryNonce, initialBinding, initialPrompt, send, applyStore, resetRoundState])

  /** tool_call_id → 结果（内容取自 tool 消息，ok 取自运行时事件）。 */
  const outcomes = useMemo(() => {
    const map: Record<string, ToolOutcome> = {}
    for (const message of messages) {
      if (message.role === 'tool' && message.tool_call_id) {
        map[message.tool_call_id] = {
          ok: toolStatus[message.tool_call_id]?.ok ?? true,
          content: message.content ?? '',
        }
      }
    }
    return map
  }, [messages, toolStatus])

  /** 历史面板只展示有内容的会话，按最后更新时间倒序。 */
  const historySessions = useMemo(
    () =>
      store.sessions
        .filter((session) => session.messages.length > 0)
        .sort((a, b) => b.updatedAt - a.updatedAt),
    [store.sessions],
  )

  /** 待删除会话（对话框标题需要显示它的标题）。 */
  const deleteTargetSession =
    store.sessions.find((session) => session.id === deleteTarget) ?? null

  if (!open) return null

  return (
    <Dialog
      open={open}
      // 非模态（与项目详情弹窗一致）：不锁页面交互，只保留 X 关闭。
      modal={false}
      onOpenChange={(next) => !next && onClose()}
    >
      <DialogContent
        {...dnd.contentProps}
        showOverlay={false}
        // 打开面板即聚焦输入框（压住 Radix 默认聚焦内容容器/关闭按钮的行为）。
        // 拖动、滚轮等非输入交互不受影响；等待写工具确认时 textarea 已 disabled，焦点自然落空。
        onOpenAutoFocus={(e) => {
          e.preventDefault()
          textareaRef.current?.focus()
        }}
        onPointerDownOutside={(e) => e.preventDefault()}
        onFocusOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={(e) => e.preventDefault()}
        style={dnd.contentStyle}
        className="flex h-[78vh] flex-col gap-0 p-0 sm:max-w-3xl"
      >
        <DialogHeader className="flex-row items-center justify-between gap-2 border-b border-line px-4 py-3 pr-14">
          <DialogTitle
            {...dnd.dragHandleProps}
            className="flex min-w-0 cursor-move items-center gap-2 text-[15px] font-semibold text-ink"
          >
            <AISparklesIcon className="shrink-0 text-primary" />
            <span className="truncate">{assistantTitle ?? t('ai.open')}</span>
            {model && (
              <span className="truncate rounded bg-raised px-1.5 py-0.5 text-[10px] font-normal text-mute">
                {model}
              </span>
            )}
          </DialogTitle>
          <div className="relative z-20 flex items-center gap-1.5">
            <Button
              size="xs"
              variant="outline"
              onClick={newChat}
              disabled={streaming || pending !== null}
              title={t('ai.chat.newChatHint')}
            >
              <Icon name="plus" />
              {t('ai.chat.newChat')}
            </Button>
            <Button
              size="xs"
              variant="ghost"
              onClick={() => setHistoryOpen((value) => !value)}
              disabled={streaming || pending !== null}
              title={t('ai.chat.historyHint')}
              aria-expanded={historyOpen}
            >
              <Icon name="clock" />
              {t('ai.chat.history')}
            </Button>
            <Button
              size="xs"
              variant="ghost"
              onClick={clearChat}
              disabled={streaming || pending !== null}
              title={t('ai.chat.clearChatHint')}
              // 删除当前会话：用危险色 hover 作区分（历史里其他会话不受影响）。
              className="text-mute hover:text-destructive"
            >
              <Icon name="trash" />
              {t('ai.chat.clearChat')}
            </Button>
          </div>
        </DialogHeader>

        {historyOpen && (
          <div className="max-h-56 shrink-0 overflow-y-auto border-b border-line bg-raised/40 px-2 py-2">
            <p className="px-2 pb-1 text-[11px] text-faint">{t('ai.chat.historyHint')}</p>
            {historySessions.length === 0 ? (
              <p className="px-2 py-3 text-center text-[12px] text-mute">
                {t('ai.chat.historyEmpty')}
              </p>
            ) : (
              <ul className="flex flex-col gap-0.5">
                {/* 行右侧让出 24px（容器 px-2 + 行 pr-4），避免删除按钮压在对话框右边缘的移动手势条下面。 */}
                {historySessions.map((session) => (
                  <li
                    key={session.id}
                    className={`flex items-center gap-1 rounded-md pr-4 transition-colors hover:bg-primary/10 ${
                      session.id === store.activeId ? 'bg-primary/10' : ''
                    }`}
                  >
                    <button
                      type="button"
                      onClick={() => switchSession(session.id)}
                      disabled={streaming || pending !== null}
                      className={`flex min-w-0 flex-1 items-center justify-between gap-3 px-2 py-1.5 text-left text-[12px] disabled:cursor-not-allowed disabled:opacity-60 ${
                        session.id === store.activeId ? 'font-medium text-ink' : 'text-mute'
                      }`}
                    >
                      <span className="min-w-0 flex-1 truncate">
                        {session.title || t('ai.chat.untitled')}
                      </span>
                      <span className="shrink-0 text-[11px] text-faint">
                        {humanizeDateTime(session.updatedAt)}
                      </span>
                    </button>
                    <button
                      type="button"
                      onClick={() => setDeleteTarget(session.id)}
                      disabled={streaming || pending !== null}
                      title={t('ai.chat.deleteSession')}
                      aria-label={`${t('ai.chat.deleteSession')}：${
                        session.title || t('ai.chat.untitled')
                      }`}
                      className="shrink-0 rounded p-1 text-faint transition-colors hover:bg-destructive/10 hover:text-destructive disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <Icon name="trash" />
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}

        {/* 逐条删除会话需二次确认：删除即从本地存储移除，不可恢复。 */}
        <AlertDialog
          open={deleteTarget !== null}
          onOpenChange={(next) => !next && setDeleteTarget(null)}
        >
          <AlertDialogContent zIndex={deleteDialogZ}>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {t('ai.chat.deleteSessionConfirm', {
                  title: deleteTargetSession?.title || t('ai.chat.untitled'),
                })}
              </AlertDialogTitle>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
              <Button
                variant="destructive"
                onClick={() => {
                  const target = deleteTarget
                  setDeleteTarget(null)
                  if (target) deleteSession(target)
                }}
              >
                {t('ai.chat.deleteSession')}
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        <div className="border-b border-line px-4 py-2">
          <AIBindingBar
            bindings={bindings}
            onChange={handleBindingsChange}
            disabled={streaming || pending !== null}
          />
        </div>

        <div ref={listRef} className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
          {messages.length === 0 && !streaming ? (
            <div className="flex h-full flex-col items-center justify-center gap-4 text-center">
              <div className="space-y-1">
                <p className="text-[15px] font-medium text-ink">{t('ai.chat.emptyTitle')}</p>
                <p className="text-[12px] text-mute">{t('ai.chat.emptyHint')}</p>
              </div>
              <AIQuickPrompts onFill={fillDraft} />
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {messages.map((message, index) => (
                <ChatMessage
                  key={`${index}-${message.role}-${message.tool_call_id ?? ''}`}
                  message={message}
                  outcomes={outcomes}
                  pendingId={pending?.id}
                />
              ))}
              {streaming && <TypingIndicator />}
            </div>
          )}
        </div>

        {error && (
          <div className="mx-4 mb-2 flex items-start gap-2 rounded-lg border border-destructive/40 bg-destructive/5 px-3 py-2 text-[12px] text-destructive">
            <Icon name="warning-fill" className="mt-0.5 shrink-0" />
            <span className="min-w-0 break-words">{error}</span>
          </div>
        )}

        {pending && (
          <div className="mx-4 mb-2 rounded-lg border border-primary/50 bg-primary/5 p-3">
            <p className="flex items-center gap-1.5 text-[12px] font-medium text-primary">
              <Icon name="shield" />
              {t('ai.chat.confirmTitle')}
            </p>
            <p className="mt-1 break-words text-[12px] text-ink">
              {pending.summary || pending.name}
            </p>
            <pre className="mt-1 max-h-32 overflow-auto whitespace-pre-wrap break-words rounded bg-surface px-2 py-1 text-[11px] leading-snug text-mute">
              {formatArguments(pending.arguments)}
            </pre>
            <p className="mt-1 text-[11px] text-mute">{t('ai.chat.confirmHint')}</p>
            <div className="mt-2 flex justify-end gap-2">
              {/* need_confirm 帧之后服务端立即收流：本轮流未收口前 decide 会忽略提交，
                  按钮同步置灰，避免用户点了却没有反应。 */}
              <Button size="xs" variant="outline" onClick={() => decide(false)} disabled={streaming}>
                {t('ai.chat.reject')}
              </Button>
              <Button size="xs" onClick={() => decide(true)} disabled={streaming}>
                {t('ai.chat.approve')}
              </Button>
            </div>
          </div>
        )}

        <div className="border-t border-line px-4 py-3">
          {/* 一体化输入壳：textarea 无边框，聚焦态由外层 ring 表达；工具条常驻在输入区底部。 */}
          <div className="rounded-xl border border-line bg-surface shadow-sm transition-[border-color,box-shadow] focus-within:border-primary/60 focus-within:ring-2 focus-within:ring-primary/15">
            {/* 输入 `/` 时的命令提示：↑/↓ 选择、Enter 执行、点击等同选中执行。 */}
            {commandMatches.length > 0 && (
              <ul aria-label={t('ai.chat.commandHintLabel')} className="border-b border-line p-1">
                {commandMatches.map((name, index) => (
                  <li key={name}>
                    <button
                      type="button"
                      // 高亮项用 aria-current 标注；用原生 button 语义，读屏可直接激活
                      aria-current={index === activeCommandIndex ? 'true' : undefined}
                      // 阻止默认 focus 转移，执行后输入框仍保持焦点
                      onMouseDown={(e) => e.preventDefault()}
                      onMouseEnter={() => setCommandIndex(index)}
                      onClick={() => executeCommand(name)}
                      className={`flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-[12px] transition-colors ${
                        index === activeCommandIndex
                          ? 'bg-accent text-ink'
                          : 'text-mute hover:bg-accent hover:text-ink'
                      }`}
                    >
                      <span className="font-mono text-ink">{`/${name}`}</span>
                      <span className="min-w-0 truncate">{t(SLASH_COMMAND_HINT_KEYS[name])}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
            <textarea
              ref={textareaRef}
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey && !e.nativeEvent.isComposing) {
                  e.preventDefault()
                  submitInput()
                  return
                }
                // 输入法组合态下方向键用于选词，不接管提示导航与历史回溯。
                if (e.nativeEvent.isComposing) return
                // 命令提示展开时 ↑/↓ 移动高亮，优先于历史回溯。
                if (commandMatches.length > 0 && (e.key === 'ArrowUp' || e.key === 'ArrowDown')) {
                  e.preventDefault()
                  const delta = e.key === 'ArrowDown' ? 1 : -1
                  setCommandIndex((index) =>
                    Math.min(Math.max(index + delta, 0), commandMatches.length - 1),
                  )
                  return
                }
                if (e.key === 'ArrowUp') {
                  const history = inputHistoryRef.current
                  if (history.length === 0) return
                  const cursor = historyCursorRef.current
                  if (cursor < 0) {
                    // 未在浏览历史时，仅当光标停在首行才接管 ↑，多行编辑仍可正常移动光标。
                    const el = textareaRef.current
                    if (el && !isCaretOnFirstLine(el)) return
                    stashedDraftRef.current = draft
                    historyCursorRef.current = history.length - 1
                    e.preventDefault()
                    caretToEndRef.current = true
                    setDraft(history[history.length - 1])
                    return
                  }
                  if (cursor === 0) return
                  historyCursorRef.current = cursor - 1
                  e.preventDefault()
                  caretToEndRef.current = true
                  setDraft(history[cursor - 1])
                  return
                }
                if (e.key === 'ArrowDown' && historyCursorRef.current >= 0) {
                  const history = inputHistoryRef.current
                  const cursor = historyCursorRef.current
                  e.preventDefault()
                  if (cursor < history.length - 1) {
                    historyCursorRef.current = cursor + 1
                    caretToEndRef.current = true
                    setDraft(history[cursor + 1])
                    return
                  }
                  // 越过最新一条：恢复浏览前暂存的草稿，回到正常编辑态。
                  historyCursorRef.current = -1
                  caretToEndRef.current = true
                  setDraft(stashedDraftRef.current)
                }
              }}
              disabled={pending !== null}
              rows={2}
              placeholder={
                pending ? t('ai.chat.confirmPlaceholder') : t('ai.chat.placeholder')
              }
              className="block max-h-40 min-h-[3.5rem] w-full resize-none bg-transparent px-3.5 py-2.5 text-[13px] leading-relaxed text-ink outline-none placeholder:text-faint disabled:opacity-60"
            />
            <div className="flex items-center justify-between gap-3 px-2.5 pb-2">
              {/* 免确认开关与发送按钮同一行：高风险选项紧贴执行入口，勾选后立即生效；
                  开关状态按用户持久化到 localStorage，刷新后保持。 */}
              <label
                className="flex cursor-pointer select-none items-center gap-1.5 text-[11px] text-mute"
                title={t('ai.chat.autoApproveHint')}
              >
                <input
                  type="checkbox"
                  checked={autoApprove}
                  onChange={(e) => toggleAutoApprove(e.target.checked)}
                  className="size-3.5 cursor-pointer accent-destructive"
                />
                <span className={autoApprove ? 'font-medium text-destructive' : ''}>
                  {t('ai.chat.autoApprove')}
                </span>
              </label>
              {streaming ? (
                <Button size="sm" variant="outline" className="h-7 px-2.5 text-[12px]" onClick={stop}>
                  <Icon name="pause" />
                  {t('ai.chat.stop')}
                </Button>
              ) : (
                <Button
                  size="sm"
                  className="h-7 px-3 text-[12px]"
                  onClick={submitInput}
                  disabled={!draft.trim() || pending !== null}
                >
                  {t('ai.chat.send')}
                </Button>
              )}
            </div>
          </div>
          {autoApprove ? (
            <p className="mt-1.5 flex items-start gap-1 text-[11px] text-destructive">
              <Icon name="warning-fill" className="mt-0.5 shrink-0" />
              <span className="min-w-0 break-words">{t('ai.chat.autoApproveActive')}</span>
            </p>
          ) : (
            <p className="mt-1.5 text-[11px] text-faint">{t('ai.chat.inputHint')}</p>
          )}
        </div>

        {/* 四边拖拽移动 + 八向缩放：与项目详情弹窗共用同一外框（不只标题栏能拖） */}
        <DialogDragFrame dnd={dnd} />
      </DialogContent>
    </Dialog>
  )
}

/** 读取当前选中会话的消息（流式回调里用，避免闭包过期）。 */
function messagesRefOf(store: ChatStore): ChatMessageModel[] {
  return getActive(store)?.messages ?? EMPTY_MESSAGES
}

/** 生成中的三点提示（等待首 token 或工具执行期间）。 */
function TypingIndicator() {
  return (
    <div className="flex items-center gap-1.5 px-1 text-[12px] text-mute" aria-live="polite">
      <span className="size-1.5 animate-pulse rounded-full bg-primary" />
      <span className="size-1.5 animate-pulse rounded-full bg-primary [animation-delay:150ms]" />
      <span className="size-1.5 animate-pulse rounded-full bg-primary [animation-delay:300ms]" />
    </div>
  )
}

/** 文本增量追加到当前 assistant 消息；上一帧是工具结果时新开一条 assistant。 */
function appendAssistantText(messages: ChatMessageModel[], text: string): ChatMessageModel[] {
  const last = messages[messages.length - 1]
  if (!last || last.role !== 'assistant') {
    return [...messages, { role: 'assistant', content: text }]
  }
  return [...messages.slice(0, -1), { ...last, content: `${last.content ?? ''}${text}` }]
}

/** 记录模型发起的工具调用：同一轮里合并进末尾的 assistant 消息（按 id 去重）。 */
function upsertToolCall(
  messages: ChatMessageModel[],
  event: Extract<ChatEvent, { type: 'tool_call' }>,
): ChatMessageModel[] {
  const call = {
    id: event.id,
    type: 'function' as const,
    function: { name: event.name, arguments: event.arguments },
  }
  const last = messages[messages.length - 1]
  if (!last || last.role !== 'assistant') {
    return [...messages, { role: 'assistant', tool_calls: [call] }]
  }
  const calls = last.tool_calls ?? []
  const replaced = calls.some((item) => item.id === call.id)
    ? calls.map((item) => (item.id === call.id ? call : item))
    : [...calls, call]
  return [...messages.slice(0, -1), { ...last, tool_calls: replaced }]
}

/** 追加工具结果；已有同一 tool_call_id 的消息时跳过（防重复推进历史）。 */
function appendToolResult(
  messages: ChatMessageModel[],
  event: Extract<ChatEvent, { type: 'tool_result' }>,
): ChatMessageModel[] {
  const exists = messages.some(
    (message) => message.role === 'tool' && message.tool_call_id === event.id,
  )
  if (exists) return messages
  return [...messages, { role: 'tool', tool_call_id: event.id, content: event.content }]
}
