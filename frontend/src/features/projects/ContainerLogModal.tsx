import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { nextZIndex } from '@/lib/zIndex'
import { getToken } from '@/api/token'
import { containerStreamLogsUrl } from '@/api/endpoints'
import { Icon } from '@/components/Icons'
import { Input } from '@/components/ui/shadcn/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { LogLine } from './TabLog'

/** 日志行数上限：超长丢弃最旧行，防止 DOM 爆炸 */
const MAX_LINES = 5000

/** 解析 grpc-gateway 流式帧（data: {json} / NDJSON），返回 log 片段 */
function decodeLogFrame(line: string): string {
  if (!line) return ''
  const payload = line.startsWith('data:') ? line.slice(5).trim() : line.trim()
  if (!payload.startsWith('{')) return ''
  let ev: { result?: { log?: string }; error?: { code?: number; message?: string }; log?: string }
  try {
    ev = JSON.parse(payload)
  } catch {
    return ''
  }
  if (ev.error) throw new Error(ev.error.message ?? 'stream error')
  return ev.result?.log ?? ev.log ?? ''
}

/** 容器定位信息（与 websocket.Container 结构一致） */
export interface ContainerLogTarget {
  namespace: string
  pod: string
  container: string
}

/**
 * 容器实时日志弹窗：从部署日志行的"查看日志"入口打开，
 * SSE 流式拉取指定容器日志（stream_logs，复用 TabLog 机制），
 * follow 自动滚底 + ANSI 着色 + 行数上限。关闭/卸载时 abort 流。
 * 常被嵌套在可拖拽宿主弹窗（z-51+）内（拓扑/部署面板），打开时取下一个
 * 共享 z-index 盖过宿主，遮罩随 content z 一并抬升（见 dialog.tsx 的 depth 逻辑）。
 */
export function ContainerLogModal({
  open,
  onClose,
  title,
  container,
}: {
  open: boolean
  onClose: () => void
  title: string
  container: ContainerLogTarget
}) {
  const { t } = useTranslation()
  const [z, setZ] = useState(() => nextZIndex())
  useEffect(() => {
    if (open) setZ(nextZIndex())
  }, [open])
  const [lines, setLines] = useState<string[]>([])
  const [streaming, setStreaming] = useState(false)
  const [ended, setEnded] = useState(false)
  const abortRef = useRef<AbortController | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const pendingRef = useRef('') // 半行缓冲（SSE 分片可能跨帧截断行尾）
  const droppedRef = useRef(0) // 因超上限丢弃的最旧行数（行号 = dropped + idx + 1）
  // 关键字过滤（对齐 TabLog：不区分大小写，命中行高亮 + 计数）
  const [keyword, setKeyword] = useState('')
  const searchRef = useRef<HTMLInputElement>(null)

  /** 追加日志片段：按 \n 拆成完整行，尾部半行留 pending；超 MAX_LINES 丢弃最旧 */
  const appendLog = useCallback((piece: string) => {
    pendingRef.current += piece
    const parts = pendingRef.current.split('\n')
    pendingRef.current = parts.pop() ?? ''
    if (parts.length === 0) return
    setLines((prev) => {
      const next = prev.length ? [...prev, ...parts] : parts
      if (next.length > MAX_LINES) {
        const dropped = next.length - MAX_LINES
        droppedRef.current += dropped
        return next.slice(dropped)
      }
      return next
    })
  }, [])

  // 打开且目标存在时：重置状态并启动流；关闭/目标切换时 abort 旧流
  useEffect(() => {
    if (!open || !container) {
      setLines([])
      setStreaming(false)
      setKeyword('')
      return
    }
    abortRef.current?.abort()
    const ac = new AbortController()
    abortRef.current = ac
    pendingRef.current = ''
    droppedRef.current = 0
    setLines([])
    setStreaming(true)
    setEnded(false)
    setKeyword('')

    const url = containerStreamLogsUrl(container.namespace, container.pod, container.container, false)

    ;(async () => {
      try {
        const res = await fetch(url, {
          headers: { Authorization: getToken() },
          signal: ac.signal,
        })
        if (!res.ok) throw new Error((await res.text()) || `HTTP ${res.status}`)
        const reader = res.body?.getReader()
        if (!reader) return
        const decoder = new TextDecoder()
        let buffer = ''
        for (;;) {
          const { done, value } = await reader.read()
          if (done) break
          buffer += decoder.decode(value, { stream: true })
          let idx = buffer.indexOf('\n')
          while (idx >= 0) {
            const line = buffer.slice(0, idx)
            buffer = buffer.slice(idx + 1)
            if (line && !line.startsWith(':') && line.trim()) {
              try {
                const piece = decodeLogFrame(line)
                if (piece) appendLog(piece)
              } catch {
                // 单帧解析失败（如日志帧内嵌 error）不打断整条流
              }
            }
            idx = buffer.indexOf('\n')
          }
        }
      } catch (err) {
        if ((err as Error).name !== 'AbortError') {
          toast.error(err instanceof Error ? err.message : String(err))
        }
      } finally {
        if (ac.signal === abortRef.current?.signal) {
          setStreaming(false)
          setEnded(true)
        }
      }
    })()

    return () => ac.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, container.namespace, container.pod, container.container, appendLog])

  // follow 自动滚底（搜索时不跟随，避免把视图拽回底部）
  useEffect(() => {
    if (keyword) return
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [lines, keyword])

  // Cmd / Ctrl + F 聚焦搜索框（对齐 TabLog）；仅弹窗打开时挂，避免与宿主弹窗日志 Tab 抢焦点
  useEffect(() => {
    if (!open) return
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'f') {
        e.preventDefault()
        searchRef.current?.focus()
        searchRef.current?.select()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [open])

  // 关键字过滤（对齐 TabLog 语义）：命中行保留原行号，高亮命中关键字
  const kw = keyword.trim().toLowerCase()
  const filtered = kw
    ? lines.map((l, i) => ({ line: l, idx: i })).filter((x) => x.line.toLowerCase().includes(kw))
    : lines.map((l, i) => ({ line: l, idx: i }))
  const isEmpty = kw ? filtered.length === 0 : lines.length === 0

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-4xl" style={{ zIndex: z }}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 pr-8 text-[15px]">
            <Icon name="logs" className="text-[14px] text-primary" />
            <span className="font-mono">{title}</span>
          </DialogTitle>
        </DialogHeader>

        <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5 text-[12px]">
          <span
            className={`flex items-center gap-1.5 ${streaming ? 'text-primary' : 'text-faint'}`}
          >
            <span
              className={`size-1.5 rounded-full ${streaming ? 'animate-pulse bg-primary motion-reduce:animate-none' : 'bg-faint'}`}
            />
            {streaming ? t('project.logLive') : t('project.logStopped')}
          </span>
          <div className="relative w-56">
            <Icon
              name="search"
              className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-[12px] text-faint"
            />
            <Input
              ref={searchRef}
              aria-label={t('events.searchPlaceholder')}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              placeholder={t('events.searchPlaceholder')}
              className="h-7 pl-8 pr-7 text-[12px]"
            />
            {keyword && (
              <button
                type="button"
                onClick={() => setKeyword('')}
                className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded p-0.5 text-faint hover:text-ink"
                title={t('common.close')}
              >
                <Icon name="close" className="text-[12px]" />
              </button>
            )}
          </div>
          {kw && (
            <span className="font-mono text-[11px] text-faint">
              {filtered.length > 0 ? t('project.logMatch', { count: filtered.length }) : t('project.logNoMatch')}
            </span>
          )}
          <span className="ml-auto font-mono text-[11px] text-faint">
            {t('project.logLineCount', { count: lines.length })}
          </span>
        </div>

        <div
          ref={scrollRef}
          className="max-h-[60vh] overflow-auto overscroll-contain rounded-md bg-black/85 px-3 py-2 font-mono text-[12px] leading-relaxed text-white/80"
        >
          {isEmpty ? (
            <span className="text-white/40">
              {streaming
                ? // 流仍在拉取（可能尚未收到首行）：优先展示 Loading，避免 kw 下误报「无匹配」
                  t('common.loading') + '…'
                : kw
                  ? t('project.logNoMatch')
                  : ended
                    ? t('common.empty')
                    : ''}
            </span>
          ) : (
            filtered.map(({ line, idx }) => (
              <LogLine
                key={idx}
                line={line}
                index={droppedRef.current + idx + 1}
                highlight={kw || undefined}
              />
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
