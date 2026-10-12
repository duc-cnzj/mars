import { useEffect, useState, type ReactNode } from 'react'
import { Trans, useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import type { TKey } from '@/i18n/keys'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { getHighlightSyntax } from '@/lib/highlight'
import { copyText } from '@/lib/copy'
import { selectAllOnDoubleClick } from '@/lib/selection'
import { useOverlayZ } from '@/hooks/useOverlayZ'
import { Icon, type IconName } from '@/components/Icons'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { Switch } from '@/components/ui/shadcn/switch'
import { Tag } from '@/components/ui'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { getHibernateWindows } from '@/api/hibernateWindowsCache'
import { hibernateWindowLabel, replicaFieldLabels } from './hibernate'

type ProjectModel = components['schemas']['types.ProjectModel']
type ServiceEndpoint = components['schemas']['types.ServiceEndpoint']
type HibernateWindowModel = components['schemas']['hibernate.HibernateWindow']

/** 详细信息 Tab：基础信息（项目 ID 等标量）/ cpu / memory / 端点 / git / 镜像 / 日期 / 配置，删除入口。 */
export function TabInfo({
  detail,
  onDeleted,
  onChanged,
}: {
  detail: ProjectModel
  onDeleted: () => void
  /** 项目配置变更成功回调：休眠开关切换后通知父级刷新详情与工作台列表 */
  onChanged: () => void
}) {
  const { t } = useTranslation()
  const [endpoints, setEndpoints] = useState<ServiceEndpoint[]>([])
  const [cpu, setCpu] = useState('')
  const [memory, setMemory] = useState('')
  const [metricsLoading, setMetricsLoading] = useState(true)
  const [confirmOpen, setConfirmOpen] = useState(false)
  // 确认框 z-index：盖过可拖拽宿主弹窗（宿主 zIndex 从 51 起）。
  // 触发源是普通 Button（非 Radix DialogTrigger），onOpenChange(true) 不会触发 → 只能由 effect 在打开时置顶
  const confirmZ = useOverlayZ(confirmOpen)
  const [deleting, setDeleting] = useState(false)
  // 删除二次确认的「输入项目名」防误触：逐字敲对项目名才放行删除按钮
  const [confirmText, setConfirmText] = useState('')
  const confirmMatched = confirmText.trim() === detail.name
  // 相关配置默认折叠：配置通常很长，收起保持弹窗清爽，需要时展开看完整预览
  const [configOpen, setConfigOpen] = useState(false)

  useEffect(() => {
    let cancelled = false
    setMetricsLoading(true)
    api
      .GET(API.projectsMemoryCpuEndpoints, {
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

  // 定时休眠：仓库级能力（marsConfig.hibernate.enabled）决定本区是否展示；
  // 开关态回显自项目级记录（hibernateEnabled），运行态 hibernateState 为 awake/sleeping。
  const hb = detail.repo?.marsConfig?.hibernate
  const hasHibernate = Boolean(hb?.enabled)
  // 全局时间窗：把只读展示的 window_name 映射成「窗名 (时段)」。仅在启用休眠时拉取，
  // 走缓存模块去重；拉取失败/窗口未命中时回退为只显示窗名（helper 内部兜底）。
  const [windows, setWindows] = useState<HibernateWindowModel[]>([])
  useEffect(() => {
    if (!hasHibernate) return
    let cancelled = false
    void getHibernateWindows().then((w) => {
      if (!cancelled) setWindows(w)
    })
    return () => {
      cancelled = true
    }
  }, [hasHibernate])
  // 本地开关态：初始同步 detail，切换时乐观更新（失败回滚），父级刷新后按最新 detail 重新对齐
  const [hibEnabled, setHibEnabled] = useState(detail.hibernateEnabled)
  const [hibSaving, setHibSaving] = useState(false)
  useEffect(() => {
    setHibEnabled(detail.hibernateEnabled)
  }, [detail.id, detail.hibernateEnabled])

  /** 切换项目级休眠开关：仅改开关不触发重部署（副本映射/时间窗来自仓库级配置）；
   *  成功后通知父级刷新详情与工作台列表，失败回滚本地态 */
  const toggleHibernate = async (next: boolean) => {
    setHibEnabled(next)
    setHibSaving(true)
    try {
      const { error } = await api.POST(API.projectsHibernate, {
        params: { path: { projectId: detail.id } },
        body: { projectId: detail.id, enabled: next },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(
        next ? t('project.hibernateToggleOn') : t('project.hibernateToggleOff'),
      )
      onChanged()
    } catch (e) {
      setHibEnabled(!next)
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setHibSaving(false)
    }
  }

  /** 复制文本并统一反馈：相关配置 / 访问地址 / 容器镜像三处共用同一套交互
   *  （失败统一「复制失败」；成功文案按来源区分——doneKey 由调用点传入，
   *   「已复制访问地址！」「已复制镜像地址！」「已复制相关配置！」，不再一律「已复制」） */
  const copyValue = async (value: string, doneKey: TKey) => {
    const ok = await copyText(value)
    if (ok) toast.success(t(doneKey))
    else toast.error(t('common.copyFailed'))
  }

  const remove = async () => {
    setDeleting(true)
    try {
      const { error } = await api.DELETE(API.projectsDetail, {
        params: { path: { id: detail.id } },
      })
      if (error) throw new Error(error.message ?? String(error))
      setConfirmOpen(false)
      setConfirmText('')
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
      {/* 关键指标：仅 CPU / 内存两张卡（项目 ID 已移入下方「基础信息」区，不再占用指标卡位，
          故网格由 3 列收为 2 列） */}
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        <MetricCard icon="cpu" label={t('project.cpu')} value={metricsLoading ? '…' : cpu} />
        <MetricCard icon="memory" label={t('project.memory')} value={metricsLoading ? '…' : memory} />
      </div>

      {/* 基础信息：show 接口标量元数据（id / repoId / version / gitProjectId / namespaceId）。
          项目 ID 作为最基础标识置首行，与其余标量元数据并列展示。 */}
      <Section icon="key" title={t('project.basicInfo')}>
        <KV k={t('project.projectId')} v={detail.id ?? '-'} mono />
        <KV k={t('project.repoId')} v={detail.repoId ?? '-'} mono />
        <KV k={t('project.version')} v={detail.version ?? '-'} mono />
        <KV k={t('project.gitProjectId')} v={detail.gitProjectId ?? '-'} mono />
        <KV k={t('project.namespaceId')} v={detail.namespaceId ?? '-'} mono />
      </Section>

      {/* 定时休眠：仅仓库启用休眠能力时展示；开关切换项目级参与，时间窗/副本字段来自仓库配置（只读） */}
      {hasHibernate && hb && (
        <Section icon="moon" title={t('project.hibernate')}>
          <div className="flex items-center justify-between gap-2 rounded-md border border-line px-3 py-2">
            <span className="flex items-center gap-2 text-[13px] text-ink">
              <span className="text-mute" title={t('project.hibernateTip')}>
                {t('project.hibernate')}
              </span>
              {hibEnabled && (
                <Tag
                  tone={detail.hibernateState === 'sleeping' ? 'warn' : 'ok'}
                  dot={false}
                  className="shrink-0"
                >
                  {detail.hibernateState === 'sleeping'
                    ? t('project.hibernateStateSleeping')
                    : t('project.hibernateStateAwake')}
                </Tag>
              )}
            </span>
            <Switch
              checked={hibEnabled}
              disabled={hibSaving}
              onCheckedChange={toggleHibernate}
            />
          </div>
          <KV
            k={t('project.hibernateWindow')}
            v={hibernateWindowLabel(hb.windowName, windows) || '-'}
            mono
          />
          <KV
            k={t('project.hibernateReplicaFields')}
            v={
              replicaFieldLabels(hb.replicaPaths, detail.repo?.marsConfig?.elements).join(', ') || '-'
            }
            mono
          />
          {/* 一句话说明：休眠时间段内副本字段被置 0，避免用户看到数量变化时误判 */}
          <div className="text-[12px] leading-snug text-faint">
            {t('project.hibernateReplicaZeroNote')}
          </div>
        </Section>
      )}

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
                  </a>
                ) : (
                  <span className="break-all text-ink" translate="no">
                    {ep.url}
                  </span>
                )}
                {(ep.svcName || ep.ingressName) && (
                  <span onDoubleClick={selectAllOnDoubleClick}>
                    <Tag tone="info" dot={false} className="shrink-0">{ep.svcName || ep.ingressName}</Tag>
                  </span>
                )}
                {/* 访问地址复制：对齐 ProjectRow 端点弹层同款按钮（ghost icon-xs，hover 提亮主色） */}
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-xs"
                  onClick={() => void copyValue(ep.url, 'project.copyEndpoint')}
                  aria-label={t('common.copy')}
                  title={t('common.copy')}
                  className="shrink-0 text-faint hover:text-primary"
                >
                  <Icon name="copy" className="text-[11px]" />
                </Button>
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
              // 对齐：items-center 让按钮与文字垂直居中（items-start 会让 24px 按钮顶对齐 18px 行高，
              // 恒定偏 3px，实测）；按钮**紧随文本**（文本不 flex-1 撑满，按钮就贴着一行文本走），
              // 超长镜像名靠 break-all 在可用宽度内折行，按钮始终跟在本行文本之后
              <li key={i} className="flex items-center gap-2 text-[12px]">
                {/* 双击整选（selectAllOnDoubleClick）：镜像含 `.` `/` `:`，浏览器默认双击按「词」断选
                    （registry.uco.com/mars/demo:v1.0.0 只选到一段），改选整个文本节点便于整段复制 */}
                <span
                  className="min-w-0 break-all font-mono text-ink"
                  translate="no"
                  onDoubleClick={selectAllOnDoubleClick}
                >
                  {img}
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-xs"
                  onClick={() => void copyValue(img, 'project.copyImage')}
                  aria-label={t('common.copy')}
                  title={t('common.copy')}
                  className="shrink-0 text-faint hover:text-primary"
                >
                  <Icon name="copy" className="text-[11px]" />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </Section>

      {/* 日期 */}
      <Section icon="gear" title={t('project.timeline')}>
        <KV k={t('project.createdAt')} v={detail.humanizeCreatedAt} />
        <KV k={t('project.updatedAt')} v={detail.humanizeUpdatedAt} />
        {detail.updatedBy && <KV k={t('project.updatedBy')} v={detail.updatedBy} />}
      </Section>

      {/* 相关配置：默认折叠（配置通常很长，收起保持弹窗清爽），展开显示完整 YAML 预览。
          折叠按钮 sticky 吸顶（图钉）：展开长配置滚动时标题/折叠按钮钉在滚动区顶部不消失，
          想收起不用滚回顶部；bg-bg 与弹窗底色一致，遮盖滚过内容不露底。
          预览不设内部 max-h，随内容撑开，弹窗内容区（max-h-[68vh] overflow-auto）单滚动边界 */}
      {detail.overrideValues && (
        <section className="flex flex-col gap-1.5">
          <button
            type="button"
            onClick={() => setConfigOpen((o) => !o)}
            aria-expanded={configOpen}
            aria-controls={`tabinfo-config-panel-${detail.id}`}
            className="sticky -top-px z-20 flex w-full items-center gap-1.5 bg-bg py-1 text-[13px] font-semibold text-ink transition-colors hover:text-primary"
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
                  onClick={() => void copyValue(detail.overrideValues, 'project.copyOverrideValues')}
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
        <Button
          variant="destructive"
          onClick={() => {
            setConfirmText('')
            setConfirmOpen(true)
          }}
        >
          <Icon name="close" className="text-[13px]" />
          {t('project.deleteProject')}
        </Button>
      </div>

      {/* 删除确认：目标回显（空间/项目名）+ 一句「如误删除，可联系管理员恢复」+ 输入项目名放行删除按钮 */}
      <Dialog
        open={confirmOpen}
        onOpenChange={(o) => {
          setConfirmOpen(o)
          if (!o) setConfirmText('')
        }}
      >
        <DialogContent className="sm:max-w-md" style={{ zIndex: confirmZ }}>
          <DialogHeader>
            <DialogTitle>{t('project.deleteProject')}</DialogTitle>
          </DialogHeader>
          {/* 段距与内层容器都改用 gap（对齐空间删除弹窗）：space-y 的 margin 对行内 label 无效 */}
          <div className="flex flex-col gap-2">
            <p className="text-[13px] leading-snug text-mute">
              <Trans
                i18nKey="project.deleteConfirm"
                values={{
                  namespace: detail.namespace?.name ?? namespaceLabel(detail),
                  name: detail.name,
                }}
                components={{
                  // 空间名与项目名统一「红色 + 加粗」：与下方「照抄的名字」同款高亮，
                  // 让用户在确认句里一眼定位到被删对象（原先名字是深色常规字重，混在句子里不显眼）
                  ns: <span className="mx-1 font-semibold text-err" />,
                  name: <span className="mx-1 font-semibold text-err" />,
                }}
              />
            </p>
            <p className="text-[13px] leading-snug text-mute">{t('project.deleteRecoverableTip')}</p>
            {/* ⚠️ 同空间删除弹窗：<label> 默认 inline，space-y 系列的垂直 margin 不生效，必须 flex 块级化 */}
            <div className="flex flex-col gap-2.5">
              <label className="text-[12px] text-mute" htmlFor={`project-delete-confirm-${detail.id}`}>
                {/* 与空间删除弹窗同款：Trans 渲染内嵌标签，待输入的项目名红色加粗，避免用户在长句里
                    看漏要照抄的名字（目标名必须一眼可辨）。select-all 让双击（乃至单击）整段选中
                    名称——原生双击以连字符/点为词边界，ductest-test 只会选中「ductest」，用户照着
                    复制会漏字符；user-select: all 把该 span 视作一个不可分割的选区。 */}
                <Trans
                  i18nKey="project.deleteTypeToConfirm"
                  values={{ name: detail.name }}
                  components={{ name: <span className="font-semibold text-err select-all" /> }}
                />
              </label>
              <Input
                id={`project-delete-confirm-${detail.id}`}
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder={detail.name}
                autoComplete="off"
                spellCheck={false}
                className="font-mono text-[12px]"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" disabled={deleting || !confirmMatched} onClick={remove}>
              {deleting && <Icon name="loader" className="size-4 animate-spin" />}
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
