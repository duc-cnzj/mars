import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getToken } from '../../api/token'
import { AreaSpark } from '../../components/charts/AreaSpark'

/** 后端窗口：timeSpan=150s / tick=5s → 最多 30 个采样点，环形缓冲上限对齐该值 */
const MAX_POINTS = 30

interface PodSample {
  cpu: number
  memory: number
  humanizeCpu: string
  humanizeMemory: string
  time: string
}

type Series = { v: number; human: string; time: string }

/**
 * 前端自备单位格式化，不依赖后端 humanize 字符串——旧版后端 ≥1000m 的 CPU 输出裸数值
 * "1.500"（无单位），UAT 可能是旧版，前端必须保证每个值都带单位。
 * cpu = 毫核；memory = ScaledValue(3) 即千字节(KB)，×1000 还原字节后按 1024 进制 humanize（对齐 go-humanize）。
 */
const fmtCpu = (milli: number) => `${milli} m`
const fmtMem = (kb: number) => {
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB']
  let v = kb * 1000
  let u = units[0]
  for (let i = 1; v >= 1024 && i < units.length; i++) {
    v /= 1024
    u = units[i]
  }
  return `${v >= 100 ? Math.round(v) : v.toFixed(1)} ${u}`
}

/**
 * Pod 实时资源用量：SSE 流式拉取（grpc-gateway server-streaming，chunked JSON，
 * 每条消息独立 chunk）。解析后环形累积 CPU/内存采样，渲染两个迷你面积图。
 * 断流（pod 消失/鉴权失败）时静默降级为失败态，不打断页面。
 */
export function PodMetrics({ namespace, pod }: { namespace: string; pod: string }) {
  const { t } = useTranslation()
  const [cpu, setCpu] = useState<Series[]>([])
  const [memory, setMemory] = useState<Series[]>([])
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    setCpu([])
    setMemory([])
    setFailed(false)

    const ctrl = new AbortController()
    const url = `/api/metrics/namespace/${namespace}/pods/${pod}/stream?time=${Date.now()}`
    let buf = ''

    const push = (kind: 'cpu' | 'memory') => (r: PodSample) => {
      const sample = {
        v: r.cpu,
        human: fmtCpu(r.cpu),
        time: r.time,
      } as Series
      const set = kind === 'cpu' ? setCpu : setMemory
      if (kind === 'memory') {
        sample.v = r.memory
        sample.human = fmtMem(r.memory)
      }
      set((l) => [...l, sample].slice(-MAX_POINTS))
    }
    const pushCpu = push('cpu')
    const pushMem = push('memory')

    const emit = (line: string) => {
      // grpc-gateway 流式帧外层包 result：{ "result": { cpu, memory, ... } }（对齐旧版 PodMetrics 读 r.result.cpu）
      const r = JSON.parse(line) as { result?: PodSample } & Partial<PodSample>
      const s = (r.result ?? r) as PodSample
      if (typeof s.cpu === 'number') pushCpu(s)
      if (typeof s.memory === 'number') pushMem(s)
    }

    ;(async () => {
      try {
        const res = await fetch(url, {
          headers: { Authorization: getToken() },
          signal: ctrl.signal,
        })
        if (!res.ok || !res.body) throw new Error(`stream ${res.status}`)
        const reader = res.body.getReader()
        const dec = new TextDecoder()
        for (;;) {
          const { done, value } = await reader.read()
          if (done) break
          buf += dec.decode(value, { stream: true })
          // 新行分隔的消息
          let nl: number
          while ((nl = buf.indexOf('\n')) >= 0) {
            const line = buf.slice(0, nl).trim()
            buf = buf.slice(nl + 1)
            if (line) {
              try {
                emit(line)
              } catch {
                /* 半包/脏行跳过 */
              }
            }
          }
          // 无换行的单条消息（grpc-gateway 每消息一个 chunk）
          if (buf.trim()) {
            try {
              emit(buf.trim())
              buf = ''
            } catch {
              /* 暂存，等待下一 chunk 拼全 */
            }
          }
        }
      } catch (e) {
        if (!ctrl.signal.aborted) setFailed(true)
      }
    })()

    return () => ctrl.abort()
  }, [namespace, pod])

  const latest = (l: Series[]) => l[l.length - 1]

  // 沿用旧版布局：两张迷你面积图并排内联在工具栏右侧（不再单独占一行），
  // 高度贴合工具栏行高；pod 上下文以 title 提示，避免额外视觉开销。
  return (
    <div className="flex w-full min-w-0 items-stretch gap-1" title={`${namespace}/${pod}`}>
      {failed ? (
        <div className="flex h-6 min-w-0 flex-1 items-center px-1 text-[11px] text-faint">
          {t('project.metricsUnavailable')}
        </div>
      ) : (
        <>
          <div className="min-w-0 flex-1">
            <AreaSpark
              label={t('project.metricsCpu')}
              value={latest(cpu)?.human ?? '—'}
              points={cpu.map((s) => s.v)}
              labels={cpu.map((s) => s.human)}
              hint={t('project.metricsCpuHint')}
              color="var(--primary)"
              height={9}
            />
          </div>
          <div className="min-w-0 flex-1">
            <AreaSpark
              label={t('project.metricsMemory')}
              value={latest(memory)?.human ?? '—'}
              points={memory.map((s) => s.v)}
              labels={memory.map((s) => s.human)}
              color="var(--ok)"
              height={9}
            />
          </div>
        </>
      )}
    </div>
  )
}
