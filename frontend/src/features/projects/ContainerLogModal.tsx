import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { getToken } from '../../api/token'
import { AnsiText } from '../../utils/ansi'
import { Icon } from '../../components/icons'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'

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
  const [lines, setLines] = useState<string[]>([])
  const [streaming, setStreaming] = useState(false)
  const [ended, setEnded] = useState(false)
  const abortRef = useRef<AbortController | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const pendingRef = useRef('') // 半行缓冲（SSE 分片可能跨帧截断行尾）
  const droppedRef = useRef(0) // 因超上限丢弃的最旧行数（保留以备行号扩展）

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

    const url =
      `/api/containers/namespaces/${encodeURIComponent(container.namespace)}` +
      `/pods/${encodeURIComponent(container.pod)}` +
      `/containers/${encodeURIComponent(container.container)}/stream_logs?showEvents=false`

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

  // follow 自动滚底
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [lines])

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-4xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 pr-8 text-[15px]">
            <Icon name="logs" className="text-[14px] text-primary" />
            <span className="font-mono">{title}</span>
          </DialogTitle>
        </DialogHeader>

        <div className="flex items-center gap-3 text-[12px]">
          <span
            className={`flex items-center gap-1.5 ${streaming ? 'text-primary' : 'text-faint'}`}
          >
            <span
              className={`size-1.5 rounded-full ${streaming ? 'animate-pulse bg-primary motion-reduce:animate-none' : 'bg-faint'}`}
            />
            {streaming ? t('project.logLive') : t('project.logStopped')}
          </span>
          <span className="font-mono text-[11px] text-faint">
            {t('project.logLineCount', { count: lines.length })}
          </span>
        </div>

        <div
          ref={scrollRef}
          className="max-h-[60vh] overflow-auto overscroll-contain rounded-md bg-black/85 px-3 py-2 font-mono text-[12px] leading-relaxed text-white/80"
        >
          {lines.length === 0 ? (
            <span className="text-white/40">
              {streaming ? t('common.loading') + '…' : ended ? t('common.empty') : ''}
            </span>
          ) : (
            lines.map((line, i) => (
              <div key={i} className="whitespace-pre">
                <AnsiText text={line} />
              </div>
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
