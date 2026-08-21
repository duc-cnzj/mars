import { useEffect, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { Loader2 } from 'lucide-react'
import type { components } from '../../api/schema'
import { api } from '../../api/client'
import { getHighlightSyntax } from '../../utils/highlight'
import { copyText } from '../../utils/copy'
import { nextZIndex } from '../../hooks/useDraggableDialog'
import { Icon, type IconName } from '../../components/icons'
import { Tag } from '../../components/ui'
import { Button } from '@/components/ui/shadcn/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'

type ProjectModel = components['schemas']['types.ProjectModel']
type ServiceEndpoint = components['schemas']['types.ServiceEndpoint']

/** 详细信息 Tab：id / cpu / memory / 端点 / git / 镜像 / 日期 / 配置，删除入口。 */
export function TabInfo({
  detail,
  onDeleted,
}: {
  detail: ProjectModel
  onDeleted: () => void
}) {
  const { t } = useTranslation()
  const [endpoints, setEndpoints] = useState<ServiceEndpoint[]>([])
  const [cpu, setCpu] = useState('')
  const [memory, setMemory] = useState('')
  const [metricsLoading, setMetricsLoading] = useState(true)
  const [confirmOpen, setConfirmOpen] = useState(false)
  // 确认框 z-index：盖过可拖拽宿主弹窗（宿主 zIndex 从 51 起），打开时再置顶
  const [confirmZ, setConfirmZ] = useState(() => nextZIndex())
  // 触发源是普通 Button（非 Radix DialogTrigger），onOpenChange(true) 不会触发 → 由 effect 在打开时置顶
  useEffect(() => {
    if (confirmOpen) setConfirmZ(nextZIndex())
  }, [confirmOpen])
  const [deleting, setDeleting] = useState(false)
  // 相关配置默认折叠：配置通常很长，收起保持弹窗清爽，需要时展开看完整预览
  const [configOpen, setConfigOpen] = useState(false)

  useEffect(() => {
    let cancelled = false
    setMetricsLoading(true)
    api
      .GET('/api/projects/{id}/memory_cpu_and_endpoints', {
        params: { path: { id: detail.id } },
      })
      .then(({ data, error }) => {
        if (cancelled) return
        setMetricsLoading(false)
        if (error || !data) return
        setCpu(data.cpu)
        setMemory(data.memory)
        setEndpoints(data.urls)
      })
      .catch(() => {
        if (!cancelled) setMetricsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [detail.id])

  const copyConfig = async () => {
    const ok = await copyText(detail.overrideValues)
    if (ok) toast.success(t('common.copied'))
    else toast.error(t('common.copyFailed'))
  }

  const remove = async () => {
    setDeleting(true)
    try {
      const { error } = await api.DELETE('/api/projects/{id}', {
        params: { path: { id: detail.id } },
      })
      if (error) throw new Error(error.message ?? String(error))
      setConfirmOpen(false)
      toast.success(t('project.deleteSuccess', { name: detail.name }))
      onDeleted()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setDeleting(false)
    }
  }

  const needGit = detail.repo?.needGitRepo

  return (
    <div className="flex flex-col gap-4">
      {/* 关键指标 */}
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
        <MetricCard icon="rocket" label={t('project.id')} value={String(detail.id)} />
        <MetricCard icon="cpu" label={t('project.cpu')} value={metricsLoading ? '…' : cpu} />
        <MetricCard icon="memory" label={t('project.memory')} value={metricsLoading ? '…' : memory} />
      </div>

      {/* 基础信息：show 接口标量元数据（repoId / version / gitProjectId / namespaceId） */}
      <Section icon="key" title={t('project.basicInfo')}>
        <KV k={t('project.repoId')} v={detail.repoId ?? '-'} mono />
        <KV k={t('project.version')} v={detail.version ?? '-'} mono />
        <KV k={t('project.gitProjectId')} v={detail.gitProjectId ?? '-'} mono />
        <KV k={t('project.namespaceId')} v={detail.namespaceId ?? '-'} mono />
      </Section>

      {/* 访问地址：link 图标与外部 ProjectRow URL 图标对齐（旧版 LinkOutlined 同款） */}
      <Section icon="link" title={t('project.endpoints')}>
        {metricsLoading ? (
          <div className="text-[13px] text-faint">{t('common.loading')}</div>
        ) : endpoints.length === 0 ? (
          <div className="text-[13px] text-faint">{t('common.empty')}</div>
        ) : (
          <ul className="flex flex-col gap-1">
            {endpoints.map((ep, i) => (
              <li key={i} className="flex items-center gap-2 text-[13px]">
                <span className="font-mono text-faint">{i + 1}.</span>
                {ep.url.startsWith('http') ? (
                  <a
                    href={ep.url}
                    target="_blank"
                    rel="noreferrer"
                    translate="no"
                    className="break-all text-primary hover:underline"
                  >
                    {ep.url}
                    {ep.portName && <span className="text-faint"> ({ep.portName})</span>}
                  </a>
                ) : (
                  <span className="break-all text-ink" translate="no">
                    {ep.url}
                    {ep.portName && <span className="text-faint"> ({ep.portName})</span>}
                  </span>
                )}
                <Tag tone="mute" dot={false}>
                  {ep.name}
                </Tag>
              </li>
            ))}
          </ul>
        )}
      </Section>

      {/* git 信息 */}
      {needGit && (
        <Section icon="repo" title={t('project.gitInfo')}>
          <KV k={t('project.branch')} v={detail.gitBranch || '-'} mono />
          <KV
            k={t('project.commit')}
            v={
              detail.gitCommitWebUrl ? (
                <a
                  href={detail.gitCommitWebUrl}
                  target="_blank"
                  rel="noreferrer"
                  translate="no"
                  className="text-primary hover:underline"
                >
                  {detail.gitCommitTitle || detail.gitCommit} by {detail.gitCommitAuthor} ·{' '}
                  {detail.gitCommitDate}
                </a>
              ) : (
                detail.gitCommit || '-'
              )
            }
          />
        </Section>
      )}

      {/* 容器镜像 */}
      <Section icon="boxes" title={t('project.dockerImages')}>
        {detail.dockerImage.length === 0 ? (
          <div className="text-[13px] text-faint">{t('common.empty')}</div>
        ) : (
          <ul className="flex flex-col gap-1">
            {detail.dockerImage.map((img, i) => (
              <li key={i} className="break-all font-mono text-[12px] text-ink" translate="no">
                {img}
              </li>
            ))}
          </ul>
        )}
      </Section>

      {/* 日期 */}
      <Section icon="gear" title={t('project.timeline')}>
        <KV k={t('project.createdAt')} v={detail.humanizeCreatedAt} />
        <KV k={t('project.updatedAt')} v={detail.humanizeUpdatedAt} />
      </Section>

      {/* 相关配置：默认折叠（配置通常很长，收起保持弹窗清爽），展开显示完整 YAML 预览。
          折叠按钮 sticky 吸顶（图钉）：展开长配置滚动时标题/折叠按钮钉在滚动区顶部不消失，
          想收起不用滚回顶部；bg-background 与弹窗底色一致，遮盖滚过内容不露底。
          预览不设内部 max-h，随内容撑开，弹窗内容区（max-h-[68vh] overflow-auto）单滚动边界 */}
      {detail.overrideValues && (
        <section className="flex flex-col gap-1.5">
          <button
            type="button"
            onClick={() => setConfigOpen((o) => !o)}
            aria-expanded={configOpen}
            aria-controls={`tabinfo-config-panel-${detail.id}`}
            className="sticky -top-px z-20 flex w-full items-center gap-1.5 bg-background py-1 text-[13px] font-semibold text-ink transition-colors hover:text-primary"
          >
            <Icon name="database" className="text-[14px] text-primary" />
            {t('project.overrideValues')}
            <Icon
              name="chevron-down"
              className={`text-[12px] text-faint transition-transform ${configOpen ? 'rotate-180' : ''}`}
            />
          </button>
          {configOpen && (
            <div id={`tabinfo-config-panel-${detail.id}`} className="flex flex-col gap-1 pl-0.5">
              {/* 旧版同款 prism-material-dark 渲染：language-yaml 触发暗底 #2f2f2f + token 配色；
                  !important 压过未分层 theme CSS 的 pre 默认（Roboto Mono 1em / 1.25em 内边距），
                  保持与当前 12px app mono 代码块一致 */}
              <div className="relative">
                <button
                  type="button"
                  onClick={() => void copyConfig()}
                  aria-label={t('common.copy')}
                  title={t('common.copy')}
                  className="absolute right-1 top-1 z-10 flex size-6 items-center justify-center rounded-md text-white/60 transition-colors hover:bg-white/10 hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
                >
                  <Icon name="copy" className="text-[12px]" />
                </button>
                <pre
                  className="language-yaml rounded-md px-3! py-2! font-mono! text-[12px]! leading-relaxed"
                  dangerouslySetInnerHTML={{ __html: getHighlightSyntax(detail.overrideValues, 'yaml') }}
                />
              </div>
            </div>
          )}
        </section>
      )}

      {/* 删除项目 */}
      <div className="flex justify-end border-t border-line pt-4">
        <Button variant="destructive" onClick={() => setConfirmOpen(true)}>
          <Icon name="close" className="text-[13px]" />
          {t('project.deleteProject')}
        </Button>
      </div>

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent className="sm:max-w-md" style={{ zIndex: confirmZ }}>
          <DialogHeader>
            <DialogTitle>{t('project.deleteProject')}</DialogTitle>
          </DialogHeader>
          <p className="text-[13px] leading-relaxed text-mute">
            {t('project.deleteConfirm')}
            <span className="mx-1 text-err">{detail.namespace?.name ?? namespaceLabel(detail)}</span>
            <span>下的</span>
            <span className="mx-1 font-medium text-ink">{detail.name}</span>
            <span>{t('project.question')}</span>
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" disabled={deleting} onClick={remove}>
              {deleting && <Loader2 className="size-4 animate-spin" />}
              {t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function namespaceLabel(detail: ProjectModel): string {
  return detail.namespace?.name ?? String(detail.namespaceId)
}

/** 指标卡片：图标 + 标签 + 值 */
function MetricCard({ icon, label, value }: { icon: IconName; label: string; value: string }) {
  return (
    <div className="flex items-center gap-2.5 rounded-lg border border-line bg-surface px-3 py-2.5">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-primary-soft text-primary">
        <Icon name={icon} className="text-[15px]" />
      </div>
      <div className="min-w-0">
        <div className="text-[11px] text-faint">{label}</div>
        <div className="truncate font-mono text-[13px] font-medium text-ink">{value}</div>
      </div>
    </div>
  )
}

/** 分组区块：图标标题 + 内容 */
function Section({ icon, title, children }: { icon: IconName; title: string; children: ReactNode }) {
  return (
    <section className="flex flex-col gap-1.5">
      <h4 className="flex items-center gap-1.5 text-[13px] font-semibold text-ink">
        <Icon name={icon} className="text-[14px] text-primary" />
        {title}
      </h4>
      <div className="flex flex-col gap-1 pl-0.5">{children}</div>
    </section>
  )
}

/** 键值对行 */
function KV({ k, v, mono }: { k: string; v: ReactNode; mono?: boolean }) {
  return (
    <div className="flex items-baseline gap-2 text-[13px]">
      <span className="shrink-0 text-faint">{k}:</span>
      <span className={`min-w-0 break-all ${mono ? 'font-mono text-[12px]' : ''} text-ink`}>{v}</span>
    </div>
  )
}
