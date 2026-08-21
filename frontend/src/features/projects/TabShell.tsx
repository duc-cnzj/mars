import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ChangeEvent,
  type PointerEvent as ReactPointerEvent,
} from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { Loader2 } from 'lucide-react'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import type { components } from '../../api/schema'
import { api } from '../../api/client'
import { getToken } from '../../api/token'
import { websocket } from '../../api/websocket'
import type { FrameHandler } from '../../realtime/useWebsocket'
import { useWebsocket } from '../../realtime/useWebsocket'
import { nextZIndex } from '../../hooks/useDraggableDialog'
import { Icon } from '../../components/icons'
import { Empty, SkeletonTabShell } from '../../components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/shadcn/popover'
import { RadioGroup, RadioGroupItem } from '@/components/ui/shadcn/radio-group'
import { copyText } from '../../utils/copy'
import { PodMetrics } from './PodMetrics'
import { PodStateTag } from './PodStateTag'
import { shortContainerName } from './containerName'

type StateContainer = components['schemas']['types.StateContainer']
type NPC = { namespace: string; pod: string; container: string }
type Pane = { id: string; npc: NPC }

const MAX_PANES = 4

const encoder = new TextEncoder()
const decoder = new TextDecoder()

/**
 * 命令行 Tab：xterm 交互式终端，最多 4 个、2×2「田」字网格布局。
 * 新增顺序固定：①向右边 ②向下边 ③右下，即填充顺序 左上→右上→左下→右下；最多 4 个。
 * 网格内可拖拽分隔条调整行列占比（列分隔条改左右列权重，行分隔条改上下行权重）。
 * 每个分屏是一个独立 WebSocket 会话（session_id `<ns>-<pod>-<container>:<uuid>`），
 * stdout 帧按 slug 路由回对应终端。顶部为容器选择 + 新增终端 + 文件传输 + 强杀 + 资源用量。
 * - pod 事件 → debounce 重拉容器列表（新 pod 自动出现在列表）
 * - WebSocket 断线重连 → 重建所有终端实例
 * - 上传前按 /api/files/max_upload_size 校验大小
 */
export function TabShell({
  projectId,
  projectName,
  resizeAt,
}: {
  projectId: number
  projectName: string
  /** 宿主弹窗尺寸变化信号：缩放/最大化/还原时 bump，用于终端 refit（对齐旧版 resizeAt 机制） */
  resizeAt?: number
}) {
  const { t } = useTranslation()
  const { ready, send, subscribe, subscribeProjectPodEvent } = useWebsocket()
  const [containers, setContainers] = useState<StateContainer[]>([])
  // 首次拉取容器列表未返回前为 true：渲染骨架占位，避免误显「暂无任何容器」（对齐 TabLog）
  const [loading, setLoading] = useState(true)
  const [target, setTarget] = useState<NPC | null>(null)
  const [panes, setPanes] = useState<Pane[]>([])
  const [forceDeleting, setForceDeleting] = useState(false)
  // 强杀小确认框（popover 走 portal 挂 body）须盖过可拖拽宿主弹窗的 z-51+：打开时取下一个共享 z-index
  const [confirmZ, setConfirmZ] = useState(() => nextZIndex())
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [dlPath, setDlPath] = useState('')
  const [dlOpen, setDlOpen] = useState(false)
  // 下载弹层（popover 走 portal 挂 body）须盖过可拖拽宿主弹窗的 z-51+：打开时取下一个共享 z-index
  const [dlZ, setDlZ] = useState(() => nextZIndex())
  const [maxUpload, setMaxUpload] = useState({ bytes: 0, humanizeSize: '' })
  // 田字网格行列权重：colWeights=[左,右] / rowWeights=[上,下]，拖拽分隔条按比例转移
  const [colWeights, setColWeights] = useState<number[]>([1, 1])
  const [rowWeights, setRowWeights] = useState<number[]>([1, 1])
  const [resizeRev, setResizeRev] = useState(0)
  const [reconnectRev, setReconnectRev] = useState(0)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const podDebounce = useRef<ReturnType<typeof setTimeout> | null>(null)
  const dragRef = useRef<{
    axis: 'col' | 'row'
    start: number
    containerSize: number
    weights: number[]
  } | null>(null)
  const handlersRef = useRef<{ move: (e: PointerEvent) => void; up: (e: PointerEvent) => void } | null>(null)
  const hasConnectedRef = useRef(false)
  // 目标容器同步镜像：供事件回调里读最新值，避免闭包陈旧
  const targetRef = useRef<NPC | null>(null)
  useEffect(() => {
    targetRef.current = target
  }, [target])

  /** 重建 shell 会话：新 session_id → 后端新开 shell 连接（点击已选中容器复用，与 WS 重连同一机制） */
  const reconnect = useCallback(() => {
    setReconnectRev((r) => r + 1)
  }, [])

  /** 选择容器：已选中 → 重建 shell 会话（重新建立连接）；未选中 → 切换目标 */
  const selectContainer = useCallback(
    (c: StateContainer) => {
      const cur = targetRef.current
      const npc = { namespace: c.namespace, pod: c.pod, container: c.container }
      if (cur && cur.pod === c.pod && cur.container === c.container) {
        reconnect()
      } else {
        setTarget(npc)
      }
    },
    [reconnect],
  )

  /** 点击已选中的 radio 点/标签：Radix 对已选中项不触发 onValueChange，需单独重建连接 */
  const handleSameReconnect = useCallback(
    (c: StateContainer) => {
      const cur = targetRef.current
      if (cur && cur.pod === c.pod && cur.container === c.container) reconnect()
    },
    [reconnect],
  )

  const reload = useCallback(() => {
    setLoading(true)
    api
      .GET('/api/projects/{id}/containers', { params: { path: { id: projectId } } })
      .then(({ data }) => setContainers(data?.items ?? []))
      .finally(() => setLoading(false))
  }, [projectId])

  useEffect(() => {
    reload()
  }, [reload])

  // pod 事件 → debounce 1s 重拉容器列表（pod 重建后新容器自动出现）
  useEffect(() => {
    const unsub = subscribeProjectPodEvent(projectId, () => {
      if (podDebounce.current) clearTimeout(podDebounce.current)
      podDebounce.current = setTimeout(() => {
        podDebounce.current = null
        reload()
      }, 1000)
    })
    return () => {
      if (podDebounce.current) clearTimeout(podDebounce.current)
      unsub()
    }
  }, [subscribeProjectPodEvent, projectId, reload])

  // WebSocket 断线重连成功 → 旧终端会话已失效，重建所有终端实例并重拉容器
  useEffect(() => {
    if (ready) {
      if (hasConnectedRef.current) {
        setReconnectRev((r) => r + 1)
        reload()
      } else {
        hasConnectedRef.current = true
      }
    }
  }, [ready, reload])

  // 上传大小上限：挂载时拉取一次
  useEffect(() => {
    api
      .GET('/api/files/max_upload_size')
      .then(({ data }) => {
        if (data) setMaxUpload({ bytes: data.bytes, humanizeSize: data.humanizeSize })
      })
      .catch(() => {
        /* 拉取失败不阻塞上传，按无上限处理 */
      })
  }, [])

  // 默认选中第一个可用容器；选中项被移除时回退到首个
  useEffect(() => {
    if (containers.length === 0) {
      setTarget(null)
      return
    }
    setTarget((prev) => {
      if (prev && containers.some((c) => c.pod === prev.pod && c.container === prev.container)) {
        return prev
      }
      const first = containers[0]
      return { namespace: first.namespace, pod: first.pod, container: first.container }
    })
  }, [containers])

  // 分屏：剔除已消失容器的 pane；全空则自动补一个
  useEffect(() => {
    setPanes((prev) => {
      const valid = prev.filter((p) =>
        containers.some((c) => c.pod === p.npc.pod && c.container === p.npc.container),
      )
      if (valid.length > 0) return valid
      if (containers.length === 0) return []
      const first = containers[0]
      return [
        { id: crypto.randomUUID(), npc: { namespace: first.namespace, pod: first.pod, container: first.container } },
      ]
    })
  }, [containers])

  // 分屏数量变化（新增/删除）→ 行列重置为等分
  useEffect(() => {
    setColWeights([1, 1])
    setRowWeights([1, 1])
  }, [panes.length])

  // 同一容器也允许再开一个终端（每个 pane 独立 WS 会话），最多 MAX_PANES 格
  const addPane = () => {
    if (!target || panes.length >= MAX_PANES) return
    setPanes((prev) => {
      if (prev.length >= MAX_PANES) return prev
      setResizeRev((r) => r + 1)
      return [...prev, { id: crypto.randomUUID(), npc: target }]
    })
  }

  const removePane = (id: string) => {
    setPanes((prev) => prev.filter((p) => p.id !== id))
    setResizeRev((r) => r + 1)
  }

  // ---- 分屏拖拽 resize：田字网格中，列分隔条改左右列权重、行分隔条改上下行权重 ----
  const startResize = (axis: 'col' | 'row', e: ReactPointerEvent) => {
    const container = containerRef.current
    if (!container) return
    const containerSize = axis === 'col' ? container.clientWidth : container.clientHeight
    if (containerSize <= 0) return
    e.preventDefault()
    e.stopPropagation()
    dragRef.current = {
      axis,
      start: axis === 'col' ? e.clientX : e.clientY,
      containerSize,
      weights: (axis === 'col' ? colWeights : rowWeights).slice(),
    }
    const move = (ev: PointerEvent) => {
      const d = dragRef.current
      if (!d) return
      const pos = d.axis === 'col' ? ev.clientX : ev.clientY
      const total = d.weights.reduce((a, b) => a + b, 0)
      if (total <= 0) return
      const dUnits = ((pos - d.start) / d.containerSize) * total
      const next = d.weights.slice()
      next[0] = Math.max(0.05, next[0] + dUnits)
      next[1] = Math.max(0.05, next[1] - dUnits)
      if (d.axis === 'col') setColWeights(next)
      else setRowWeights(next)
    }
    const up = () => {
      dragRef.current = null
      window.removeEventListener('pointermove', move)
      window.removeEventListener('pointerup', up)
      handlersRef.current = null
      setResizeRev((r) => r + 1) // 拖拽结束 → 各终端 refit
    }
    handlersRef.current = { move, up }
    window.addEventListener('pointermove', move)
    window.addEventListener('pointerup', up)
  }

  // 卸载时清理可能残留的拖拽监听
  useEffect(
    () => () => {
      if (handlersRef.current) {
        window.removeEventListener('pointermove', handlersRef.current.move)
        window.removeEventListener('pointerup', handlersRef.current.up)
      }
    },
    [],
  )

  const forceDelete = async () => {
    if (!target) return
    setForceDeleting(true)
    try {
      const { error } = await api.POST(
        '/api/containers/namespaces/{namespace}/pods/{pod}/force_delete',
        {
          params: { path: { namespace: target.namespace, pod: target.pod } },
          body: { namespace: target.namespace, pod: target.pod, gracePeriodSeconds: '0' },
        },
      )
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('project.forceDeleteSuccess', { pod: target.pod }))
      reload()
    } catch (e) {
      toast.error(
        t('project.forceDeleteFailed', { pod: target.pod }) + (e instanceof Error ? `: ${e.message}` : ''),
      )
    } finally {
      setForceDeleting(false)
    }
  }

  /** 上传文件到目标容器：先校验大小（max_upload_size），POST /api/files（multipart）→ id → copy_to_pod */
  const onFileSelected = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = '' // 允许再次选择同一文件
    if (!file || !target || uploading) return
    if (maxUpload.bytes > 0 && file.size > maxUpload.bytes) {
      toast.error(t('project.uploadTooLarge', { size: maxUpload.humanizeSize }))
      return
    }
    setUploading(true)
    try {
      const fd = new FormData()
      fd.append('file', file)
      const upRes = await fetch('/api/files', {
        method: 'POST',
        headers: { Authorization: getToken() },
        body: fd,
      })
      if (!upRes.ok) throw new Error((await upRes.text()) || `HTTP ${upRes.status}`)
      const { id } = (await upRes.json()) as { id: number }
      const { data, error } = await api.POST('/api/containers/copy_to_pod', {
        body: {
          fileId: String(id),
          namespace: target.namespace,
          pod: target.pod,
          container: target.container,
        },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('project.copyToPodSuccess', { path: data?.podFilePath ?? '' }))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    } finally {
      setUploading(false)
    }
  }

  /** 从目标容器下载文件：POST /api/copy_from_pod 返回二进制流，触发浏览器下载 */
  const onDownload = async () => {
    const path = dlPath.trim()
    if (!path || !target || downloading) return
    setDownloading(true)
    try {
      const res = await fetch('/api/copy_from_pod', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: getToken(),
        },
        body: JSON.stringify({
          namespace: target.namespace,
          pod: target.pod,
          container: target.container,
          filepath: path,
        }),
      })
      if (!res.ok) throw new Error((await res.text()) || `HTTP ${res.status}`)
      const blob = await res.blob()
      const cd = res.headers.get('Content-Disposition') ?? ''
      const encoded = /filename\*=utf-8''([^;]+)/.exec(cd)?.[1]
      const plain = /filename="?([^";]+)"?/.exec(cd)?.[1]
      const name = decodeURIComponent(encoded ?? plain ?? '') || path.split('/').pop() || 'download'
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = name
      a.click()
      URL.revokeObjectURL(url)
      toast.success(t('project.downloadSuccess'))
      setDlOpen(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    } finally {
      setDownloading(false)
    }
  }

  // 田字网格定位（CSS Grid 单父容器，所有 pane 为直接子节点）：
  // 用稳定 pane-id 作 key，增删时 React 只复用/卸载对应实例，其余 pane 不重建 → 会话/数据不丢。
  // 列模板 3 轨（col0|gutter|col1），行模板 3 轨（row0|gutter|row1）；不足 2 列/3 行时退化为单轨。
  // gutter 轨刻意做窄（6px），终端间距小；分隔条手柄在轨内居中。
  // 3 个终端时最后一行只有 1 个 pane：横跨整行宽度，列分隔条仅作用于上行 → 拖动 1-2 滑块不影响 3 号终端宽度。
  const nCols = panes.length > 1 ? 3 : 1
  const nRows = panes.length > 2 ? 3 : 1
  const gridColOf = (i: number) => {
    if (nCols !== 3) return 1
    if (panes.length === 3 && i === 2) return '1 / span 3'
    return (i % 2) * 2 + 1
  }
  const rowPos = (i: number) => (nRows === 3 ? Math.floor(i / 2) * 2 + 1 : 1)
  const GAP = '0.375rem'
  const gridTemplateColumns = nCols === 3 ? `${colWeights[0]}fr ${GAP} ${colWeights[1]}fr` : '1fr'
  const gridTemplateRows = nRows === 3 ? `${rowWeights[0]}fr ${GAP} ${rowWeights[1]}fr` : '1fr'

  if (containers.length === 0) {
    // 首次拉取未返回前用骨架占位（对齐 TabLog 加载态），避免误显「暂无任何容器」闪跳
    if (loading) return <SkeletonTabShell />
    return <Empty text={t('project.noContainers')} />
  }

  return (
    <div className="flex h-full min-h-0 flex-col gap-2">
      {/* 容器选择：RadioGroup（与日志页 TabLog 视觉一致） */}
      <RadioGroup
        value={target ? `${target.pod}|${target.container}` : ''}
        onValueChange={(v) => {
          const [pod, container] = v.split('|')
          const next = containers.find((c) => c.pod === pod && c.container === container)
          if (next) setTarget({ namespace: next.namespace, pod: next.pod, container: next.container })
        }}
        className="flex flex-wrap items-center gap-x-4 gap-y-1.5"
      >
        {containers.map((c) => {
          const id = `${c.pod}|${c.container}`
          return (
            <div key={id} className="flex items-center gap-1.5">
              {/* 点击已选中的 radio 点/标签（label htmlFor 转发到 button）→ 重建 shell 连接 */}
              <RadioGroupItem
                id={id}
                value={id}
                className="size-3.5"
                onClick={() => handleSameReconnect(c)}
              />
              <label htmlFor={id} className="cursor-pointer select-none text-[13px] text-ink">
                {shortContainerName(c.container, projectName)}
              </label>
              {/* 点击标签同样选中 radio；标签内复制按钮 hover 显示、复制完整容器名。
                  点击已选中容器（radio 点/标签/胶囊）重建 shell 连接 */}
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

      {/* 单行工具栏（沿用旧版布局）：新增终端（最多 4 个）+ 上传/下载 + 强杀，字体小号，
          资源用量图表内联在右侧（旧版是按钮 + 图表同一行） */}
      <div className="flex flex-wrap items-center gap-1">
        <Button
          size="xs"
          variant="outline"
          disabled={!target || panes.length >= MAX_PANES}
          onClick={addPane}
          title={panes.length >= MAX_PANES ? t('project.maxTerminals') : undefined}
        >
          <Icon name="plus" className="text-[11px]" />
          {t('project.addTerminal')}
        </Button>

        {target && (
          <>
            <span className="mx-0.5 h-3.5 w-px bg-line" aria-hidden />
            <input ref={fileInputRef} type="file" className="hidden" onChange={onFileSelected} />
            <Button
              size="xs"
              variant="outline"
              disabled={uploading}
              onClick={() => fileInputRef.current?.click()}
            >
              {uploading ? <Loader2 className="size-3 animate-spin" /> : <Icon name="copy" className="text-[11px]" />}
              {t('project.uploadToPod')}
            </Button>

            {/* 下载文件：沿用旧版 Popconfirm 方式——点击按钮弹出路径输入，不再把 input 内联到一行 */}
            <Popover open={dlOpen} onOpenChange={(o) => { if (o) setDlZ(nextZIndex()); setDlOpen(o) }}>
              <PopoverTrigger asChild>
                <Button size="xs" variant="outline" disabled={downloading}>
                  {downloading ? <Loader2 className="size-3 animate-spin" /> : <Icon name="logs" className="text-[11px]" />}
                  {t('project.downloadFromPod')}
                </Button>
              </PopoverTrigger>
              <PopoverContent align="start" sideOffset={6} className="w-80 p-3" style={{ zIndex: dlZ }}>
                <div className="flex flex-col gap-2">
                  <div className="text-[12px] font-medium text-ink">{t('project.downloadFromPod')}</div>
                  <Input
                    autoFocus
                    aria-label={t('project.filePathPlaceholder')}
                    value={dlPath}
                    onChange={(e) => setDlPath(e.target.value)}
                    placeholder={t('project.filePathPlaceholder')}
                    className="h-7 font-mono text-[11px]"
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') void onDownload()
                    }}
                  />
                  <div className="flex justify-end gap-1.5">
                    <Button size="xs" variant="outline" onClick={() => setDlOpen(false)}>
                      {t('common.cancel')}
                    </Button>
                    <Button
                      size="xs"
                      variant="default"
                      disabled={downloading || !dlPath.trim()}
                      onClick={onDownload}
                    >
                      {downloading && <Loader2 className="size-3 animate-spin" />}
                      {t('project.downloadFromPod')}
                    </Button>
                  </div>
                </div>
              </PopoverContent>
            </Popover>

            {/* 强杀容器：沿用下载弹层同款 Popover 小确认框（对齐旧版 Popconfirm），不再用全屏确认框 */}
            <Popover
              open={confirmOpen}
              onOpenChange={(o) => {
                if (o) setConfirmZ(nextZIndex())
                setConfirmOpen(o)
              }}
            >
              <PopoverTrigger asChild>
                <button
                  type="button"
                  className="inline-flex items-center gap-1 rounded-md border border-err/40 px-2 py-0.5 text-[11px] text-err transition-colors hover:bg-err-soft"
                >
                  <Icon name="power" className="text-[11px]" />
                  {t('project.forceDeletePod')}
                </button>
              </PopoverTrigger>
              <PopoverContent align="start" sideOffset={6} className="w-80 p-3" style={{ zIndex: confirmZ }}>
                <div className="flex flex-col gap-2">
                  <div className="text-[12px] font-medium text-ink">{t('project.forceDeleteTitle')}</div>
                  <div className="text-[11px] leading-snug text-mute">
                    {t('project.forceDeleteDesc', { pod: target.pod })}
                  </div>
                  <div className="flex justify-end gap-1.5">
                    <Button size="xs" variant="outline" onClick={() => setConfirmOpen(false)}>
                      {t('common.cancel')}
                    </Button>
                    <Button
                      size="xs"
                      variant="default"
                      className="bg-destructive text-white hover:bg-destructive/90"
                      disabled={forceDeleting}
                      onClick={() => {
                        void forceDelete()
                        setConfirmOpen(false)
                      }}
                    >
                      {forceDeleting && <Loader2 className="size-3 animate-spin" />}
                      {t('project.forceDeletePod')}
                    </Button>
                  </div>
                </div>
              </PopoverContent>
            </Popover>
          </>
        )}

        {/* 选中 pod 的实时资源用量：内联在工具栏右侧（沿用旧版单行布局） */}
        {target && (
          <div className="ml-auto flex min-w-0 flex-1 items-stretch justify-end gap-1">
            <PodMetrics namespace={target.namespace} pod={target.pod} />
          </div>
        )}
      </div>

      {/* 分屏终端网格：2×2「田」字布局，填充顺序 左上→右上→左下→右下（新增顺序 右→下→右下）。
          CSS Grid 单父容器：pane 按 gridColumn/gridRow 定位，增删时其余 pane 依稳定 key 复用实例、
          不重建不重连（避免丢终端数据）；列/行分隔条轨道可拖拽改左右列/上下行权重。
          3 个终端时末行 pane 横跨整行（col gutter 只覆盖上行），拖动 1-2 列滑块不影响 3 号终端。
          每格内含 ShellTerminal（自带 ResizeObserver，分隔条拖拽/弹窗缩放时实时 refit） */}
      {ready ? (
        <div
          ref={containerRef}
          className="grid min-h-[160px] min-w-0 flex-1"
          style={{ gridTemplateColumns, gridTemplateRows }}
        >
          {panes.map((p, i) => (
            <div
              key={`${p.id}:${reconnectRev}`}
              className="min-h-0 min-w-0"
              style={{ gridColumn: gridColOf(i), gridRow: rowPos(i), minWidth: 0, minHeight: 0 }}
            >
              <ShellTerminal
                {...p.npc}
                send={send}
                subscribe={subscribe}
                resizeAt={`${resizeRev}:${resizeAt ?? 0}`}
                onClose={() => removePane(p.id)}
              />
            </div>
          ))}
          {nCols === 3 && (
            <div
              key="col-gutter"
              role="separator"
              aria-orientation="vertical"
              onPointerDown={(e) => startResize('col', e)}
              className="group flex cursor-col-resize items-center justify-center"
              style={{ gridColumn: 2, gridRow: panes.length === 3 ? 1 : `1 / span ${nRows}` }}
            >
              <div className="h-full w-1 rounded-full bg-line transition-colors group-hover:bg-primary group-active:bg-primary" />
            </div>
          )}
          {nRows === 3 && (
            <div
              key="row-gutter"
              role="separator"
              aria-orientation="horizontal"
              onPointerDown={(e) => startResize('row', e)}
              className="group flex cursor-row-resize items-center justify-center"
              style={{ gridColumn: `1 / span ${nCols}`, gridRow: 2 }}
            >
              <div className="h-1 w-full rounded-full bg-line transition-colors group-hover:bg-primary group-active:bg-primary" />
            </div>
          )}
        </div>
      ) : (
        <div className="flex items-center gap-2 rounded-lg border border-dashed border-line bg-surface px-4 py-6 text-[13px] text-mute">
          <Icon name="loader" className="animate-spin text-[14px] text-primary" />
          {t('common.loading')}…
        </div>
      )}
    </div>
  )
}

function ShellTerminal({
  namespace,
  pod,
  container,
  send,
  subscribe,
  resizeAt,
  onClose,
}: NPC & {
  send: (bytes: Uint8Array) => void
  subscribe: (slug: string, handler: FrameHandler) => () => void
  resizeAt: number | string
  onClose: () => void
}) {
  const { t } = useTranslation()
  const hostRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<Terminal | null>(null)
  const fitRef = useRef<FitAddon | null>(null)
  const sessionId = useMemo(
    () => `${namespace}-${pod}-${container}:${crypto.randomUUID()}`,
    [namespace, pod, container],
  )

  // 容器尺寸变化即 refit：覆盖弹窗缩放/最大化/还原、分屏拖拽、窗口 resize 等一切路径，
  // rAF 节流避免同帧多次 fit；保证终端始终贴合宿主宽高（resizeAt 信号保留作兜底）
  useEffect(() => {
    const host = hostRef.current
    if (!host) return
    let raf = 0
    const ro = new ResizeObserver((entries) => {
      const entry = entries[0]
      if (!entry) return
      const { width, height } = entry.contentRect
      if (width <= 0 || height <= 0) return
      cancelAnimationFrame(raf)
      raf = requestAnimationFrame(() => {
        try {
          fitRef.current?.fit()
        } catch {
          /* 容器尚未渲染完成时忽略 */
        }
      })
    })
    ro.observe(host)
    return () => {
      cancelAnimationFrame(raf)
      ro.disconnect()
    }
  }, [])

  // 分屏布局尺寸变化（拖拽/新增/删除/方向切换）后自适应
  useEffect(() => {
    if (!fitRef.current) return
    const timer = setTimeout(() => {
      try {
        fitRef.current?.fit()
      } catch {
        /* 容器尚未渲染完成时忽略 */
      }
    }, 80)
    return () => clearTimeout(timer)
  }, [resizeAt])

  useEffect(() => {
    const host = hostRef.current
    if (!host) return

    const term = new Terminal({
      fontSize: 14,
      fontFamily: 'var(--font-mono-family)', // 跟随主题 mono 栈，与其余代码/日志一致
      cursorBlink: true,
      rows: 25,
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.open(host)
    termRef.current = term
    fitRef.current = fit

    // stdout / toast 帧：后端把 session_id 作为 slug 回投，帧体解码为 WsHandleShellResponse
    const unsubscribe = subscribe(sessionId, (meta, raw) => {
      let res: websocket.WsHandleShellResponse
      try {
        res = websocket.WsHandleShellResponse.decode(raw)
      } catch {
        return
      }
      if (meta.result === websocket.ResultType.Error && meta.message) {
        toast.error(meta.message)
        return
      }
      const tm = res.terminalMessage
      if (!tm) return
      if (tm.op === 'stdout' && tm.data) term.write(tm.data)
      if (tm.op === 'toast' && tm.data) toast.error(decoder.decode(tm.data))
    })

    // 初始化会话
    send(
      websocket.WsHandleExecShellInput.encode({
        type: websocket.Type.HandleExecShell,
        container: { namespace, pod, container },
        sessionId,
      }).finish(),
    )

    let resizeTimer: ReturnType<typeof setTimeout> | null = null
    term.onData((str) => {
      send(
        websocket.TerminalMessageInput.encode({
          type: websocket.Type.HandleExecShellMsg,
          message: { sessionId, op: 'stdin', data: encoder.encode(str), width: 0, height: 0 },
        }).finish(),
      )
    })
    term.onResize(({ cols, rows }) => {
      if (resizeTimer) clearTimeout(resizeTimer)
      resizeTimer = setTimeout(() => {
        send(
          websocket.TerminalMessageInput.encode({
            type: websocket.Type.HandleExecShellMsg,
            message: { sessionId, op: 'resize', data: new Uint8Array(0), width: cols, height: rows },
          }).finish(),
        )
      }, 200)
    })

    // 等布局稳定后自适应尺寸
    const fitTimer = setTimeout(() => {
      try {
        fit.fit()
      } catch {
        /* 容器尚未渲染完成时忽略 */
      }
    }, 60)
    term.focus()

    return () => {
      clearTimeout(fitTimer)
      if (resizeTimer) clearTimeout(resizeTimer)
      send(
        websocket.TerminalMessageInput.encode({
          type: websocket.Type.HandleCloseShell,
          message: { sessionId, op: '', data: new Uint8Array(0), width: 0, height: 0 },
        }).finish(),
      )
      unsubscribe()
      termRef.current = null
      fitRef.current = null
      term.dispose()
    }
  }, [sessionId, namespace, pod, container, send, subscribe, toast])

  return (
    <div className="flex h-full flex-col overflow-hidden rounded-lg border border-line bg-black/85">
      <div className="flex items-center gap-2 border-b border-white/10 px-3 py-1.5">
        <Icon name="terminal" className="text-[12px] text-primary" />
        <span className="truncate font-mono text-[11px] text-white/70">
          {namespace}/{pod}/{container}
        </span>
        <button
          type="button"
          onClick={onClose}
          className="ml-auto rounded p-1 text-white/40 transition-colors hover:bg-white/10 hover:text-white"
          title={t('common.close')}
        >
          <Icon name="close" className="text-[12px]" />
        </button>
      </div>
      <div ref={hostRef} className="min-h-0 flex-1 px-1 py-1" aria-label={t('project.tabShell')} />
    </div>
  )
}
