import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Icon } from '../../components/icons'
import { Tag, type Tone } from '../../components/ui'
import { Progress } from '@/components/ui/shadcn/progress'
import { websocket } from '../../api/websocket'
import type { DeployStreamStatus, DeployLogLine } from './useDeployStream'
import { TimeCost } from './TimeCost'
import { ContainerLogModal, type ContainerLogTarget } from './ContainerLogModal'

const STATUS_TONE: Record<DeployStreamStatus, Tone> = {
  idle: 'mute',
  deploying: 'info',
  deployed: 'ok',
  failed: 'err',
  canceled: 'warn',
}

const STATUS_KEY = {
  idle: 'project.statusUnknown',
  deploying: 'project.statusDeploying',
  deployed: 'project.statusDeployed',
  failed: 'project.statusFailed',
  canceled: 'project.statusCanceled',
} as const

/** 日志行文字色：按帧 result 结构化分色（成功绿/失败红/取消黄，随主题语义 token），默认白 */
const LINE_COLOR: Record<number, string> = {
  [websocket.ResultType.Error]: 'text-err',
  [websocket.ResultType.DeployedFailed]: 'text-err',
  [websocket.ResultType.Success]: 'text-ok',
  [websocket.ResultType.Deployed]: 'text-ok',
  [websocket.ResultType.DeployedCanceled]: 'text-warn',
}

/**
 * 百分比补间：WS 进度帧是离散跳变（0→30→55→80→100），直接 setState 会让
 * 数字/进度条跳格。这里用 rAF 把 display 值向 target 平滑插值（easeOutCubic），
 * 新帧到达时从中途续接，不重新从 0 拉。
 */
function useSmoothPercent(target: number, duration = 500) {
  const [display, setDisplay] = useState(target)
  const displayRef = useRef(target)
  useEffect(() => {
    const from = displayRef.current
    if (from === target) return
    const start = performance.now()
    let raf = 0
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration)
      const eased = 1 - Math.pow(1 - t, 3)
      const val = Math.round(from + (target - from) * eased)
      displayRef.current = val
      setDisplay(val)
      if (t < 1) raf = requestAnimationFrame(tick)
    }
    raf = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf)
  }, [target, duration])
  return display
}

/**
 * 实时部署日志：进度条 + 终端风格日志行（自动滚底）。
 * 日志行内容来自 WS 帧的 metadata.message，LogWithContainers 结果附容器列表。
 * fill=true 时面板占满父级（TabEdit 替换表单的整块面板），日志区 flex-1 内滚、
 * 不被日志条数撑高；默认内容高度（CreateProjectModal 内联在滚动表单里）日志区 max-h-64。
 */
export function DeployLog({
  status,
  percent,
  logs,
  loading,
  fill = false,
}: {
  status: DeployStreamStatus
  percent: number
  logs: DeployLogLine[]
  loading: boolean
  /** 占满父级 flex 容器（面板自适应高度，日志区内滚不撑开整体） */
  fill?: boolean
}) {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement>(null)
  const [logTarget, setLogTarget] = useState<ContainerLogTarget | null>(null)
  const displayPercent = useSmoothPercent(percent)

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [logs])

  return (
    <section
      className={cn(
        'flex flex-col gap-2 rounded-lg border border-line bg-surface p-3',
        fill && 'min-h-0 flex-1',
      )}
    >
      <div className="flex shrink-0 items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Icon name="rocket" className="text-[14px] text-primary" />
          <Tag tone={STATUS_TONE[status]} dot>
            {t(STATUS_KEY[status])}
          </Tag>
          {loading && <Icon name="loader" className="animate-spin text-[13px] text-primary" />}
        </div>
        <div className="flex items-center gap-3">
          {status === 'deploying' && <TimeCost running />}
          <span className="font-mono text-[11px] text-faint">{displayPercent}%</span>
        </div>
      </div>

      <Progress value={displayPercent} />

      <div
        ref={scrollRef}
        className={cn(
          'overflow-auto rounded-md bg-black/85 px-3 py-2 font-mono text-[12px] leading-relaxed text-white/80',
          // fill：日志区吃掉剩余高度内滚；否则内容高度 + 上限（不撑开内联的部署表单）
          fill ? 'min-h-0 flex-1' : 'max-h-64',
        )}
      >
        {logs.length === 0 ? (
          <span className="text-white/40">{t('common.loading')}…</span>
        ) : (
          logs.map((line, i) => (
            <div
              key={i}
              className={cn(
                'whitespace-pre-wrap break-all',
                LINE_COLOR[line.result] ?? 'text-white/80',
              )}
            >
              <span className="text-primary">▸ </span>
              {line.msg || ' '}
              {line.containers.length > 0 && (
                <span className="ml-1 inline-flex flex-wrap items-center gap-x-2 gap-y-1">
                  {line.containers.map((c, ci) => (
                    <span key={ci} className="inline-flex items-center gap-1 text-white/60">
                      <span>({c.pod}/{c.container})</span>
                      <button
                        type="button"
                        onClick={() =>
                          setLogTarget({ namespace: c.namespace, pod: c.pod, container: c.container })
                        }
                        className="rounded border border-primary/50 px-1 py-px text-[10px] leading-4 text-primary transition-colors hover:bg-primary/20"
                      >
                        {t('project.viewLog')}
                      </button>
                    </span>
                  ))}
                </span>
              )}
            </div>
          ))
        )}
      </div>

      {logTarget && (
        <ContainerLogModal
          open
          onClose={() => setLogTarget(null)}
          title={`${logTarget.pod}/${logTarget.container}`}
          container={logTarget}
        />
      )}
    </section>
  )
}
