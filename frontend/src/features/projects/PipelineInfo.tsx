import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2 } from 'lucide-react'
import { api } from '../../api/client'
import { toast } from '../../lib/toast'
import { Icon, type IconName } from '../../components/icons'

type PipelineStatus = 'success' | 'running' | 'failed' | 'manual' | 'unknown'

/** i18n 键需字面量类型才能通过 t() 的类型收窄 */
type PipelineKey =
  | 'project.pipelineSuccess'
  | 'project.pipelineRunning'
  | 'project.pipelineFailed'
  | 'project.pipelineManual'
  | 'project.pipelineUnknown'

/**
 * 各状态视觉（对齐旧版 antd Alert）：
 * - success → 绿底实心对勾圆、running → 琥珀实心叹号三角、failed → 红底实心叉圆
 * - manual → 蓝底卡片 + 时钟（等待手动触发）、unknown → 灰底卡片 + 时钟（状态未知）
 */
const PIPELINE_META: Record<PipelineStatus, { key: PipelineKey; icon: IconName; text: string; bg: string; border: string }> = {
  success: { key: 'project.pipelineSuccess', icon: 'check-circle-fill', text: 'text-ok', bg: 'bg-ok-soft', border: 'border-ok/30' },
  running: { key: 'project.pipelineRunning', icon: 'warning-fill', text: 'text-warn', bg: 'bg-warn-soft', border: 'border-warn/30' },
  failed: { key: 'project.pipelineFailed', icon: 'close-circle-fill', text: 'text-err', bg: 'bg-err-soft', border: 'border-err/30' },
  manual: { key: 'project.pipelineManual', icon: 'clock', text: 'text-info', bg: 'bg-info-soft', border: 'border-info/30' },
  unknown: { key: 'project.pipelineUnknown', icon: 'clock', text: 'text-faint', bg: 'bg-surface', border: 'border-line' },
}

/**
 * git pipeline 状态横幅（对齐旧版 antd Alert）：全宽彩色底 + 状态图标 + 状态文案 +
 * 「查看 pipeline 详情」外链 + 右侧刷新按钮（loading 转圈，hover 品牌色 + 弱放大）。
 * 无 branch/commit 时不渲染；拉取中占位 loading；404 占位「未找到」；请求失败/无状态占位「不可用」+ 可重试。
 * 状态色：ok 绿 / warn 琥珀 / err 红 / info 蓝；unknown 与 loading 保持中性灰。各占位与横幅等高，保持槽位稳定不跳。
 */
export function PipelineInfo({
  repoId,
  branch,
  commit,
}: {
  repoId: number
  branch: string
  commit: string
}) {
  const { t } = useTranslation()
  const [status, setStatus] = useState<PipelineStatus | null>(null)
  const [webUrl, setWebUrl] = useState('')
  const [loading, setLoading] = useState(false)
  // 请求失败/返回无 pipeline 状态 → true，占位「不可用」（区别于「未就绪」的静默不渲染）
  const [unavailable, setUnavailable] = useState(false)
  // 404（commit 不存在 / 该 commit 无 pipeline）→ true，占位「未找到」（区别于临时不可用，重试无意义故无刷新）
  const [notFound, setNotFound] = useState(false)
  // 请求序号：分支/commit 切换或手动刷新并发时，只认最后一次响应（丢弃过期）
  const reqRef = useRef(0)

  const fetchInfo = useCallback(
    (notify = false) => {
      if (!repoId || !branch || !commit) {
        setStatus(null)
        setUnavailable(false)
        setNotFound(false)
        return
      }
      const req = ++reqRef.current
      setLoading(true)
      // 重新请求先清「不可用」/「未找到」，避免换分支后残留旧占位
      setUnavailable(false)
      setNotFound(false)
      api
        .GET('/api/git/repos/{repoId}/branches/{branch}/commits/{commit}/pipeline_info', {
          params: { path: { repoId, branch, commit } },
        })
        .then(({ data, error, response }) => {
          if (req !== reqRef.current) return
          setLoading(false)
          // 404：commit 不存在 / 该 commit 无 pipeline → 专门的「未找到」状态（区别于临时不可用，重试无意义）
          if (response?.status === 404) {
            setStatus(null)
            setNotFound(true)
            return
          }
          if (error || !data) {
            // 初始加载失败 → 标不可用（占位展示）；手动刷新失败 → 保留上次状态，仅弹错误 toast
            if (!notify) setUnavailable(true)
            if (notify) toast.error(t('project.pipelineRefreshFailed'))
            return
          }
          if (data.status in PIPELINE_META) {
            setStatus(data.status as PipelineStatus)
            setUnavailable(false)
            setWebUrl(data.webUrl ?? '')
            // 仅手动刷新成功才弹 toast；初始加载与过期响应静默
            if (notify) toast.success(t('project.pipelineRefreshed'))
          } else {
            // 返回的状态不在枚举内 → 视为不可用
            setStatus(null)
            setUnavailable(true)
          }
        })
        .catch(() => {
          if (req !== reqRef.current) return
          setLoading(false)
          // 网络异常：初始失败标不可用；手动刷新失败保留上次状态
          if (!notify) setUnavailable(true)
          if (notify) toast.error(t('project.pipelineRefreshFailed'))
        })
    },
    [repoId, branch, commit],
  )

  useEffect(() => {
    void fetchInfo()
  }, [fetchInfo])

  // 404（commit 不存在 / 无 pipeline）：红色「未找到」（与横幅等高；区别于「不可用」不给刷新按钮，重试无意义）
  if (notFound) {
    return (
      <div className="flex min-h-[42px] items-center gap-2 rounded-md border border-err/30 bg-err-soft px-3 py-2">
        <Icon name="clock" className="shrink-0 text-[14px] text-err" />
        <span className="text-[13px] font-medium leading-none text-err">
          {t('project.pipelineNotFound')}
        </span>
      </div>
    )
  }

  // 拉取中（首载/重试/换分支后重新获取）：loading 占位（与横幅等高，保持槽位稳定不跳）
  if (!status && loading) {
    return (
      <div className="flex min-h-[42px] items-center gap-2 rounded-md border border-line bg-surface px-3 py-2">
        <Loader2 className="size-4 shrink-0 animate-spin text-faint" />
        <span className="text-[13px] font-medium leading-none text-faint">{t('common.loading')}</span>
      </div>
    )
  }

  // 获取不到 pipeline：琥珀色「不可用」+ 可重试刷新（与横幅等高，保持槽位稳定不跳）
  if (!status) {
    if (!unavailable) return null
    return (
      <div className="flex items-center gap-2 rounded-md border border-warn/30 bg-warn-soft px-3 py-2">
        <Icon name="clock" className="shrink-0 text-[14px] text-warn" />
        <span className="text-[13px] font-medium leading-none text-warn">
          {t('project.pipelineUnavailable')}
        </span>
        <button
          type="button"
          onClick={() => void fetchInfo(true)}
          title={t('project.pipelineRefresh')}
          aria-label={t('project.pipelineRefresh')}
          className="ml-auto flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-md text-faint transition-colors hover:bg-raised hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
        >
          <Icon name="refresh" className={`text-[14px] ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>
    )
  }

  const meta = PIPELINE_META[status]
  return (
    <div className={`flex items-center gap-2 rounded-md border ${meta.bg} ${meta.border} px-3 py-2`}>
      <Icon name={meta.icon} className={`shrink-0 text-[14px] ${meta.text}`} />
      <span className={`text-[13px] font-medium leading-none ${meta.text}`}>{t(meta.key)}</span>
      {webUrl && (
        <a
          href={webUrl}
          target="_blank"
          rel="noreferrer"
          className={`flex items-center gap-1 text-[12px] leading-none transition-colors hover:underline ${meta.text}`}
        >
          {t('project.pipelineLink')}
          <Icon name="external" className="text-[12px]" />
        </a>
      )}
      <button
        type="button"
        onClick={() => void fetchInfo(true)}
        title={t('project.pipelineRefresh')}
        aria-label={t('project.pipelineRefresh')}
        className="ml-auto flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-md text-faint transition-colors hover:bg-raised hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
      >
        <Icon name="refresh" className={`text-[14px] ${loading ? 'animate-spin' : ''}`} />
      </button>
    </div>
  )
}
