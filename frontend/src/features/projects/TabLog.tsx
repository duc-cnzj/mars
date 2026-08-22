import { memo, useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { getToken } from '@/api/token'
import { useWebsocket } from '@/hooks/useWebsocket'
import { AnsiText } from '@/components/AnsiText'
import { copyText } from '@/lib/copy'
import { Icon } from '@/components/Icons'
import { Empty, SkeletonTabLog } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { RadioGroup, RadioGroupItem } from '@/components/ui/shadcn/radio-group'
import { PodStateTag } from './PodStateTag'
import { shortContainerName } from '@/lib/shortContainerName'

type StateContainer = components['schemas']['types.StateContainer']

/** 日志行数上限：超长丢弃最旧行，防止整表 DOM 爆炸（行内再配 content-visibility 跳过屏外布局） */
const MAX_LINES = 5000

/** 断流自动恢复阈值：本次会话只拉到这么少行说明容器可能刚重启，自动重拉 */
const AUTO_RELOAD_THRESHOLD = 10

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

/** 单行日志：行号（绝对序号）+ ANSI 着色文本。一条日志一行，长行不换行（whitespace-pre），
 *  溢出由外层 overflow-auto 提供水平滚动；文本 span shrink-0 保持自然宽，避免被压缩后换行。
 *  memo：仅当行内容变化时重渲染，流式追加时不重绘旧行 */
const LogLine = memo(function LogLine({
  line,
  index,
  highlight,
}: {
  line: string
  index: number
  highlight?: string
}) {
  return (
    <div className="flex whitespace-pre" style={{ contentVisibility: 'auto' }}>
      <span className="w-12 shrink-0 select-none pr-3 text-right tabular-nums text-slate-600">
        {index}
      </span>
      <span className="shrink-0 text-slate-200">
        <AnsiText text={line} highlight={highlight} />
      </span>
    </div>
  )
})

/**
 * 容器日志 Tab：SSE 实时流式日志（stream_logs）+ follow 自动滚底 + 关键字过滤 + ANSI 着色。
 * - pod 事件（ProjectPodEvent）触发 debounce 重拉容器列表，pod 重建后新容器自动出现在列表
 * - 行号 / 下载 / 断流自动恢复 / Cmd(Ctrl)+F 聚焦搜索
 */
export function TabLog({ projectId, projectName }: { projectId: number; projectName: string }) {
  const { t } = useTranslation()
  const { subscribeProjectPodEvent } = useWebsocket()
  const [containers, setContainers] = useState<StateContainer[]>([])
  const [selected, setSelected] = useState<StateContainer | null>(null)
  const [lines, setLines] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [streaming, setStreaming] = useState(false)
  const [streamEnded, setStreamEnded] = useState(false)
  // 日志流恒包含 pod 事件（旧版 show_events=1 硬编码同语义），无开关
  const [keyword, setKeyword] = useState('')
  const [follow, setFollow] = useState(true)
  const abortRef = useRef<AbortController | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const searchRef = useRef<HTMLInputElement>(null)
  const pendingRef = useRef('') // 半行缓冲（SSE 分片可能跨帧截断行尾）
  const appendedRef = useRef(0) // 本次会话累计追加行数（断流恢复判据）
  const droppedRef = useRef(0) // 因超上限丢弃的最旧行数 → 绝对行号基数
  const autoReloadTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const podDebounce = useRef<ReturnType<typeof setTimeout> | null>(null)
  // 选中容器同步镜像：供异步回调/事件里读最新值，避免闭包陈旧
  const selectedRef = useRef<StateContainer | null>(null)
  // 流重启信号：递增 → 强制重启当前容器流（需求 1 就绪翻转 / 需求 2 点击已选中 radio）
  const [streamEpoch, setStreamEpoch] = useState(0)

  const applySelected = useCallback((next: StateContainer | null) => {
    selectedRef.current = next
    setSelected(next)
  }, [])

  /** 选择容器：已选中 → 重启流（需求 2，对齐旧版 Radio onClick 恒 bump timestamp）；未选中 → 切换 */
  const selectContainer = useCallback(
    (c: StateContainer) => {
      const cur = selectedRef.current
      if (cur && cur.pod === c.pod && cur.container === c.container) {
        setStreamEpoch((n) => n + 1)
      } else {
        applySelected(c)
      }
    },
    [applySelected],
  )

  /** 点击已选中的 radio 点/标签：Radix 对已选中项不触发 onValueChange，需单独重启流（需求 2） */
  const handleSameRestart = useCallback((c: StateContainer) => {
    const cur = selectedRef.current
    if (cur && cur.pod === c.pod && cur.container === c.container) setStreamEpoch((n) => n + 1)
  }, [])

  /** 追加日志片段：按 \n 拆成完整行，尾部半行留 pending；超 MAX_LINES 丢弃最旧 */
  const appendLog = useCallback((piece: string) => {
    pendingRef.current += piece
    const parts = pendingRef.current.split('\n')
    pendingRef.current = parts.pop() ?? ''
    if (parts.length === 0) return
    appendedRef.current += parts.length
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

  /** 流结束时把半行尾巴作为最后一行刷出 */
  const flushPending = useCallback(() => {
    if (!pendingRef.current) return
    const tail = pendingRef.current
    pendingRef.current = ''
    appendLog(tail + '\n')
  }, [appendLog])

  const listContainers = useCallback(async () => {
    setLoading(true)
    try {
      const { data, error } = await api.GET('/api/projects/{id}/containers', {
        params: { path: { id: projectId } },
      })
      if (error) throw new Error(error.message ?? String(error))
      const items = data?.items ?? []
      setContainers(items)
      const prev = selectedRef.current
      const fresh = prev ? items.find((c) => c.pod === prev.pod && c.container === prev.container) : null
      if (fresh && prev) {
        // 同一选中容器重建后从非 ready → ready：流可能挂在空转/无日志，bump epoch 重启一次，
        // 让 stream_logs 重连（无 ready 门控后不依赖它首发，只负责重建后刷新）
        if (fresh.ready && !prev.ready) setStreamEpoch((n) => n + 1)
        applySelected(fresh) // 用最新对象（含 ready 翻转），引用变化但身份不变 → 不重复起流
      } else {
        // 选中容器消失（pod 重建/删除）→ 落到新容器；身份变化由 stream effect 触发起流
        applySelected(items[0] ?? null)
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }, [projectId])

  // pod 事件 → debounce 1s 重拉容器列表（pod 重建后新容器自动出现，旧容器被剔除）
  useEffect(() => {
    const unsub = subscribeProjectPodEvent(projectId, () => {
      if (podDebounce.current) clearTimeout(podDebounce.current)
      podDebounce.current = setTimeout(() => {
        podDebounce.current = null
        void listContainers()
      }, 1000)
    })
    return () => {
      if (podDebounce.current) clearTimeout(podDebounce.current)
      unsub()
    }
  }, [subscribeProjectPodEvent, projectId, listContainers])

  const startStream = useCallback(
    async (c: StateContainer) => {
      // 容器未就绪也发送 stream_logs 请求：后端应能返回「未就绪」日志/错误帧，
      // 前端不因 ready=false 静默不发（用户要求「要发送请求才对」）。就绪翻转仍由
      // listContainers bump epoch 触发重启（pod 重建场景），详见 listContainers 注释。
      abortRef.current?.abort()
      const ac = new AbortController()
      abortRef.current = ac
      // 重置本次会话的缓冲/计数
      pendingRef.current = ''
      appendedRef.current = 0
      droppedRef.current = 0
      setLines([])
      setStreamEnded(false)
      setStreaming(true)
      setFollow(true)
      try {
        const res = await fetch(
          `/api/containers/namespaces/${encodeURIComponent(c.namespace)}/pods/${encodeURIComponent(c.pod)}/containers/${encodeURIComponent(c.container)}/stream_logs?showEvents=true`,
          { headers: { Authorization: getToken() }, signal: ac.signal },
        )
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
                appendLog('')
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
          flushPending()
          setStreaming(false)
          setStreamEnded(true)
          // 断流自动恢复：日志很短说明容器可能刚重启，debounce 后重拉容器列表并重开流
          if (appendedRef.current < AUTO_RELOAD_THRESHOLD) {
            if (autoReloadTimer.current) clearTimeout(autoReloadTimer.current)
            autoReloadTimer.current = setTimeout(async () => {
              autoReloadTimer.current = null
              await listContainers()
              // 断流恢复只对「当前仍选中的同一容器」重开流：期间若用户已切换容器，
              // 新流由切换 effect 负责，这里必须跳过——否则 startStream 开头的
              // abortRef.current?.abort() 会把新流杀掉，又用旧闭包把流重开到旧 pod
              const cur = selectedRef.current
              if (cur && cur.pod === c.pod && cur.container === c.container) {
                void startStream(cur)
              }
            }, 1000)
          }
        }
      }
    },
    [appendLog, flushPending, listContainers, selected],
  )

  useEffect(() => {
    void listContainers()
  }, [listContainers])

  // 容器切换 / 重建后就绪翻转（listContainers bump epoch）/ 点击已选中 → 重启流
  useEffect(() => {
    if (!selected) {
      setLines([])
      setStreaming(false)
      return
    }
    void startStream(selected)
    return () => {
      // 切换容器时作废旧容器挂起的断流自动恢复定时器：
      // 否则它 1s 后触发会以旧闭包重开旧 pod 的流
      abortRef.current?.abort()
      if (autoReloadTimer.current) {
        clearTimeout(autoReloadTimer.current)
        autoReloadTimer.current = null
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected?.pod, selected?.container, streamEpoch])

  // 卸载时中断流与恢复定时器
  useEffect(
    () => () => {
      abortRef.current?.abort()
      if (autoReloadTimer.current) clearTimeout(autoReloadTimer.current)
    },
    [],
  )

  // follow 自动滚底（搜索时暂停跟随）
  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    if (keyword) {
      el.scrollTop = 0
      return
    }
    if (follow) el.scrollTop = el.scrollHeight
  }, [lines, keyword, follow])

  // 面板尺寸变化（双击全屏 / 拖拽改高度）时，若处于 follow 状态保持贴底
  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const ro = new ResizeObserver(() => {
      if (follow && !keyword) el.scrollTop = el.scrollHeight
    })
    ro.observe(el)
    return () => ro.disconnect()
  }, [follow, keyword])

  // 用户上滚暂停 follow，滚到底恢复
  const handleScroll = () => {
    const el = scrollRef.current
    if (!el || keyword) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40
    setFollow(atBottom)
  }

  // Cmd / Ctrl + F 聚焦搜索框（组件挂载即日志 Tab 可见）
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'f') {
        e.preventDefault()
        searchRef.current?.focus()
        searchRef.current?.select()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

  // 手动重新加载：重拉容器列表 + 重开当前容器流
  const reloadStream = useCallback(() => {
    void listContainers()
    if (selected) void startStream(selected)
  }, [listContainers, selected, startStream])

  // 把当前日志内容拼成文本 Blob 下载（文件名含 projectId + 时间）
  const downloadLog = () => {
    if (lines.length === 0) return
    const blob = new Blob([lines.join('\n')], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const d = new Date()
    const pad = (n: number) => String(n).padStart(2, '0')
    const ts = `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}-${pad(d.getHours())}${pad(d.getMinutes())}${pad(d.getSeconds())}`
    a.download = `project-${projectId}-logs-${ts}.log`
    a.click()
    URL.revokeObjectURL(url)
    toast.success(t('project.logDownloadSuccess'))
  }

  const kw = keyword.trim().toLowerCase()
  const base = droppedRef.current
  const filtered = kw
    ? lines.map((l, i) => ({ line: l, idx: i })).filter((x) => x.line.toLowerCase().includes(kw))
    : lines.map((l, i) => ({ line: l, idx: i }))
  const isEmpty = kw ? filtered.length === 0 : lines.length === 0

  return (
    <div className="flex h-full min-h-0 flex-col gap-3">
      {/* 容器选择：RadioGroup（对齐旧版 Radio.Group 展示方式，横向排布不换行堆叠） */}
      {containers.length > 0 ? (
        <RadioGroup
          value={selected ? `${selected.pod}|${selected.container}` : ''}
          onValueChange={(v) => {
            const [pod, container] = v.split('|')
            const next = containers.find((c) => c.pod === pod && c.container === container)
            if (next) applySelected(next)
          }}
          className="flex flex-wrap items-center gap-x-4 gap-y-1.5"
        >
          {containers.map((c) => {
            const id = `${c.pod}|${c.container}`
            return (
              <div key={id} className="flex items-center gap-1.5">
                {/* 需求 2：点击已选中的 radio 点/标签（label htmlFor 转发到 button）→ 重启流。
                    Radix 对已选中项不触发 onValueChange，这里单独处理；未选中项的切换仍走 onValueChange */}
                <RadioGroupItem
                  id={id}
                  value={id}
                  className="size-3.5"
                  onClick={() => handleSameRestart(c)}
                />
                <label htmlFor={id} className="cursor-pointer select-none text-[13px] text-ink">
                  {shortContainerName(c.container, projectName)}
                </label>
                {/* 点击标签同样选中 radio；标签内复制按钮 hover 显示、复制完整容器名。
                    点击已选中容器（radio 点/标签/胶囊）重启流（对齐旧版 Radio onClick 恒刷新） */}
                <PodStateTag
                  container={c}
                  projectName={projectName}
                  onClick={() => selectContainer(c)}
                  onCopy={() => {
                    // 复制的是完整 pod 名（标签展示的是 pod 短名）
                    void copyText(c.pod).then((ok) => {
                      if (ok) toast.success(t('project.copyPodSuccess'))
                    })
                  }}
                />
              </div>
            )
          })}
        </RadioGroup>
      ) : loading ? (
        // 容器列表加载中：整块骨架占位（单选行 + 工具条 + 深色日志面板），避免切数据跳动
        <SkeletonTabLog />
      ) : (
        <Empty text={t('project.noContainers')} icon="logs" />
      )}

      {/* 工具条 + 断流提示 + 日志面板：仅在有容器时渲染，无容器只留 Empty 提示 */}
      {containers.length > 0 && (
        <>
          {/* 工具条：路径 + 搜索 + 实时状态 + 下载 + 刷新 */}
          <div className="flex flex-wrap items-center gap-2">
            {selected && (
              <span className="font-mono text-[11px] text-faint">
                {selected.namespace} / {selected.pod} / {selected.container}
              </span>
            )}
            <div className="relative w-60">
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
            <span
              className={`flex items-center gap-1.5 text-[11px] ${
                streaming ? 'text-primary' : 'text-faint'
              }`}
            >
              <span className={`size-1.5 rounded-full ${streaming ? 'animate-pulse bg-primary motion-reduce:animate-none' : 'bg-faint'}`} />
              {streaming ? t('project.logLive') : t('project.logStopped')}
            </span>
            {kw && (
              <span className="font-mono text-[11px] text-faint">
                {filtered.length > 0 ? t('project.logMatch', { count: filtered.length }) : t('project.logNoMatch')}
              </span>
            )}
            <span className="font-mono text-[11px] text-faint">{t('project.logLineCount', { count: lines.length })}</span>
            <div className="ml-auto flex items-center gap-1.5">
              <Button
                size="sm"
                variant="outline"
                disabled={!selected || lines.length === 0}
                onClick={downloadLog}
                className="h-7"
                title={t('project.logDownload')}
              >
                <Icon name="logs" className="text-[12px]" />
                {t('project.logDownload')}
              </Button>
            </div>
          </div>

          {/* 断流提示：流结束且未重开时给出重新加载入口（容器可能重启） */}
          {streamEnded && !streaming && (
            <div className="flex items-center gap-2 rounded-md border border-warn/40 bg-warn-soft px-3 py-1.5 text-[12px] text-warn">
              <span>{t('project.logStreamEnded')}</span>
              <Button
                size="xs"
                variant="outline"
                className="h-6 text-warn hover:text-warn"
                onClick={reloadStream}
              >
                <Icon name="refresh" className="text-[11px]" />
                {t('project.logReload')}
              </Button>
            </div>
          )}

          {/* 日志面板：深色终端风，行号 + ANSI 着色 + 关键字过滤。
              flex-1 自适应弹窗高度（双击全屏 / 拖拽改高度时随容器拉伸），min-h 兜底窄窗。 */}
          <div
            ref={scrollRef}
            onScroll={handleScroll}
            className="min-h-[240px] min-w-0 flex-1 overflow-auto rounded-md bg-black/85 font-mono text-[12px] leading-relaxed"
          >
            <div className="min-h-full w-max min-w-full px-4 py-3">
              {isEmpty ? (
                <span className="text-slate-500">{t('common.empty')}</span>
              ) : (
                filtered.map(({ line, idx }) => (
                  <LogLine key={idx} line={line} index={base + idx + 1} highlight={kw || undefined} />
                ))
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
