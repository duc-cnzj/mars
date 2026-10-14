import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { Icon } from '@/components/Icons'
import { Tag, type Tone } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Switch } from '@/components/ui/shadcn/switch'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/shadcn/dialog'
import { getHibernateWindows } from '@/api/hibernateWindowsCache'
import { hibernateWindowLabel, replicaFieldLabels } from '@/features/projects/hibernate'

type RepoModel = components['schemas']['types.RepoModel']
type ProjectModel = components['schemas']['types.ProjectModel']
type HibernateConfig = components['schemas']['mars.Hibernate']
type Element = components['schemas']['mars.Element']
type HibernateWindow = components['schemas']['hibernate.HibernateWindow']
/** deployStatus 枚举字面量（schema 仅声明文件，运行时不能 import 枚举） */
type DeployStatus = 'StatusUnknown' | 'StatusDeploying' | 'StatusDeployed' | 'StatusFailed'

const PAGE_SIZE = 15

/** 部署状态 → 标签色（已部署 ok / 部署中 warn / 失败 err / 未知 mute），对齐治理页 tone 约定 */
const STATUS_TONE: Record<DeployStatus, Tone> = {
  StatusDeployed: 'ok',
  StatusDeploying: 'warn',
  StatusFailed: 'err',
  StatusUnknown: 'mute',
}

/** 部署状态 → 文案键（编译期字面量联合，词条改名即报错） */
const STATUS_KEY: Record<
  DeployStatus,
  'project.statusDeployed' | 'project.statusDeploying' | 'project.statusFailed' | 'project.statusUnknown'
> = {
  StatusDeployed: 'project.statusDeployed',
  StatusDeploying: 'project.statusDeploying',
  StatusFailed: 'project.statusFailed',
  StatusUnknown: 'project.statusUnknown',
}

/** commit 短 sha：列表行只留前 7 位，完整 sha 放 title 悬浮可见 */
const shortSha = (sha?: string) => (sha ? sha.slice(0, 7) : '')

/** 行内 meta 项：淡色标签 + 主体值（值默认 ink，mono 走等宽栈便于比对 sha/分支/镜像） */
function Meta({ label, children, mono, title }: { label: string; children: ReactNode; mono?: boolean; title?: string }) {
  return (
    <span className="inline-flex min-w-0 items-center gap-1" title={title}>
      <span className="shrink-0 text-faint">{label}</span>
      <span className={`truncate ${mono ? 'font-mono text-ink' : 'text-ink'}`}>{children}</span>
    </span>
  )
}

/**
 * 仓库「查看项目」弹窗：列出使用该仓库的项目（名称 + 命名空间名 + 休眠态），
 * 并支持逐个项目启停定时休眠（仅切换开关，不触发重部署）。
 *
 * 数据来源两路：项目列表走 GET /api/projects?repoId=（items 已带 namespace 与
 * hibernate_enabled/hibernate_state）；仓库的休眠能力位（marsConfig.hibernate.enabled）
 * 列表行不含 marsConfig（data.projectRepo.List 只 Select 子集、不 eager load repo），
 * 故需单独拉 GET /api/repos/{id}。能力未开时全部开关禁用 + 顶部提示。
 */
export function RepoProjectsDialog({
  open,
  repo,
  onClose,
}: {
  open: boolean
  /** 列表行的 RepoModel（取 id/name）；能力位走详情另拉，不依赖本对象字段 */
  repo: RepoModel | null
  onClose: () => void
}) {
  const { t } = useTranslation()
  const [items, setItems] = useState<ProjectModel[]>([])
  const [count, setCount] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  /** 仓库级休眠能力位：null=加载中，false=未开（开关全禁用），true=可切换 */
  const [capable, setCapable] = useState<boolean | null>(null)
  /** 仓库级休眠配置（时间窗名 + 副本字段）：开关「具体做了什么」的说明来源，能力未开时为空 */
  const [hb, setHb] = useState<HibernateConfig | null>(null)
  /** 仓库自定义配置字段：把副本字段 path 映射成用户填的「字段描述」，展示更直观 */
  const [elements, setElements] = useState<Element[]>([])
  /** 全局休眠时间窗：把仓库声明的 window_names 映射成「窗名 (HH:MM-HH:MM)」列表展示 */
  const [windows, setWindows] = useState<HibernateWindow[]>([])
  /** 正在切换的项目 id（>0 表示某行开关进行中，禁用该行避免重复提交） */
  const [togglingId, setTogglingId] = useState(0)

  const repoId = repo?.id ?? 0

  const load = useCallback(
    async (p: number, append = false) => {
      if (repoId <= 0) return
      if (append) setLoadingMore(true)
      else setLoading(true)
      try {
        const { data, error } = await api.GET(API.projects, {
          params: { query: { repoId, page: p, pageSize: PAGE_SIZE } },
        })
        if (error) throw new Error(error.message ?? String(error))
        if (!data) return
        setItems((prev) => (append ? [...prev, ...data.items] : data.items))
        setCount(data.count)
      } catch (e) {
        toast.error(e instanceof Error ? e.message : String(e))
      } finally {
        setLoading(false)
        setLoadingMore(false)
      }
    },
    [repoId],
  )

  // 打开（或切换仓库）时重置并加载：项目列表第 1 页 + 仓库详情（休眠能力位）
  useEffect(() => {
    if (!open || repoId <= 0) return
    setItems([])
    setCount(0)
    setPage(1)
    setCapable(null)
    setHb(null)
    setElements([])
    setTogglingId(0)
    void load(1, false)
    // 仓库详情：能力位（开关是否可点）+ 休眠配置（时间窗/副本字段，供说明「开关到底做了什么」）
    api
      .GET(API.reposDetail, { params: { path: { id: repoId } } })
      .then(({ data }) => {
        const cfg = data?.item?.marsConfig
        setCapable(cfg?.hibernate?.enabled ?? false)
        setHb(cfg?.hibernate ?? null)
        setElements(cfg?.elements ?? [])
      })
      .catch(() => setCapable(false))
    // 全局时间窗：把仓库声明的 window_names 映射成「窗名 (时段)」；缓存模块去重，不额外打请求风暴
    void getHibernateWindows().then(setWindows)
  }, [open, repoId, load])

  // 切换项目级休眠开关：乐观更新，失败回滚并提示（语义对齐 TabEdit.toggleHibernate）。
  // 开启前必须已选至少一个时间窗（空选 = 未选，不继承仓库全量）——本弹窗无选窗入口，
  // 未选则拒绝开启并提示去项目详情配置，不翻开关、不发请求
  const toggle = async (p: ProjectModel, next: boolean) => {
    if (togglingId !== 0) return
    if (next && (p.hibernateWindowNames?.length ?? 0) === 0) {
      toast.error(t('project.hibernateWindowRequired'))
      return
    }
    setTogglingId(p.id)
    setItems((prev) => prev.map((it) => (it.id === p.id ? { ...it, hibernateEnabled: next } : it)))
    try {
      const { error } = await api.POST(API.projectsHibernate, {
        params: { path: { projectId: p.id } },
        // 带上项目已选定的窗口子集，避免仅切开关时把选择清空（空数组 = 未选，后端会拒）
        body: { projectId: p.id, enabled: next, windowNames: p.hibernateWindowNames ?? [] },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(next ? t('project.hibernateToggleOn') : t('project.hibernateToggleOff'))
    } catch (e) {
      setItems((prev) => prev.map((it) => (it.id === p.id ? { ...it, hibernateEnabled: !next } : it)))
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setTogglingId(0)
    }
  }

  const hasMore = items.length < count

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-[880px]">
        <DialogHeader>
          <DialogTitle>{t('repos.projectsTitle', { name: repo?.name ?? '' })}</DialogTitle>
        </DialogHeader>

        {/* 一句话说明开关语义：让管理员知道「拨这个开关到底会发生什么」，不是无标识的裸开关 */}
        <p className="text-[12px] leading-snug text-faint">{t('repos.projectsHibernateHint')}</p>

        {/* 仓库未启用休眠能力：提示 + 全部开关禁用（后端 ToggleHibernate 也会拒，提前告知） */}
        {capable === false && (
          <div className="flex items-start gap-2 rounded-md border border-warn/60 bg-warn-soft px-3 py-2 text-[13px] text-ink">
            <Icon name="alert" className="mt-0.5 size-4 shrink-0 text-warn" />
            <span>{t('repos.projectsNoHibernate')}</span>
          </div>
        )}

        {/* 能力开：明列该仓库的休眠配置——时间窗 + 置零的副本字段——具体回答「这个开关做什么」 */}
        {capable === true && hb && (
          <div className="grid gap-1.5 rounded-md border border-line bg-raised/40 px-3 py-2 text-[12px]">
            <div className="flex flex-wrap items-baseline gap-2">
              <span className="w-16 shrink-0 text-faint">{t('project.hibernateWindow')}</span>
              <span className="font-mono text-ink">{hibernateWindowLabel(hb.windowNames ?? [], windows) || '-'}</span>
            </div>
            <div className="flex flex-wrap items-baseline gap-2">
              <span className="w-16 shrink-0 text-faint">{t('project.hibernateReplicaFields')}</span>
              <span className="font-mono text-ink">
                {replicaFieldLabels(hb.replicaPaths, elements).join(', ') || '-'}
              </span>
            </div>
            <p className="text-faint">{t('project.hibernateReplicaZeroNote')}</p>
          </div>
        )}

        <div className="max-h-[62vh] overflow-y-auto">
          {loading ? (
            <div className="flex items-center justify-center gap-1.5 py-8 text-[13px] text-faint">
              <Icon name="loader" className="size-4 animate-spin" />
              {t('common.loading')}
            </div>
          ) : items.length === 0 ? (
            <div className="py-8 text-center text-[13px] text-faint">{t('repos.projectsEmpty')}</div>
          ) : (
            <div className="divide-y divide-line">
              {items.map((p) => (
                <div key={p.id} className="flex items-start justify-between gap-4 px-1 py-3.5">
                  <div className="min-w-0 flex-1">
                    {/* 行 1：部署状态 + 项目名 + 休眠态（开启时才有）+ 项目 ID——一眼看清「是什么、活着没、是谁」 */}
                    <div className="flex flex-wrap items-center gap-2">
                      <Tag tone={STATUS_TONE[p.deployStatus] ?? 'mute'} dot={false}>
                        {t(STATUS_KEY[p.deployStatus] ?? 'project.statusUnknown')}
                      </Tag>
                      <span className="truncate text-[13px] font-medium text-ink">{p.name}</span>
                      {p.hibernateEnabled && (
                        <Tag tone={p.hibernateState === 'sleeping' ? 'warn' : 'ok'} dot={false}>
                          {p.hibernateState === 'sleeping'
                            ? t('project.hibernateStateSleeping')
                            : t('project.hibernateStateAwake')}
                        </Tag>
                      )}
                      <span className="font-mono text-[11px] text-faint">#{p.id}</span>
                    </div>
                    {/* 行 2：归属空间 · 分支 · 提交 sha · 版本——判断「这个项目在干什么」的抓手 */}
                    <div className="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-[12px]">
                      <span className="inline-flex items-center gap-1">
                        <Icon name="namespace" className="size-3.5 text-faint" />
                        <span className="text-ink">{p.namespace?.name ?? '-'}</span>
                      </span>
                      {p.gitBranch && <Meta label={t('project.branch')} mono>{p.gitBranch}</Meta>}
                      {p.gitCommit && (
                        <Meta label={t('project.commit')} mono title={p.gitCommit}>
                          {shortSha(p.gitCommit)}
                        </Meta>
                      )}
                      {p.version > 0 && <Meta label={t('project.version')} mono>{`v${p.version}`}</Meta>}
                    </div>
                    {/* 行 3：最近一次提交标题/作者/日期——「最近在改什么、谁在改」 */}
                    {(p.gitCommitTitle || p.gitCommitAuthor || p.gitCommitDate) && (
                      <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-0.5 text-[12px] text-faint">
                        {p.gitCommitTitle && (
                          <span className="min-w-0 truncate" title={p.gitCommitTitle}>
                            {p.gitCommitAuthor
                              ? t('repos.projectsCommitBy', { title: p.gitCommitTitle, author: p.gitCommitAuthor })
                              : p.gitCommitTitle}
                          </span>
                        )}
                        {!p.gitCommitTitle && p.gitCommitAuthor && <span>{p.gitCommitAuthor}</span>}
                        {p.gitCommitDate && <span>{p.gitCommitDate}</span>}
                      </div>
                    )}
                    {/* 行 4：镜像 · 部署时间 · 最近更新/更新人——「跑的是什么、什么时候动的、谁动的」 */}
                    <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-[12px]">
                      {(p.dockerImage?.length ?? 0) > 0 && (
                        <Meta label={t('repos.projectsImage')} mono title={p.dockerImage.join('\n')}>
                          {p.dockerImage[0]}
                          {p.dockerImage.length > 1 ? `  +${p.dockerImage.length - 1}` : ''}
                        </Meta>
                      )}
                      {p.humanizeCreatedAt && <Meta label={t('project.createdAt')}>{p.humanizeCreatedAt}</Meta>}
                      {p.humanizeUpdatedAt && (
                        <Meta label={t('project.updatedAt')}>
                          {p.humanizeUpdatedAt}
                          {p.updatedBy ? ` · ${p.updatedBy}` : ''}
                        </Meta>
                      )}
                    </div>
                  </div>
                  {/* 开关带文字标签：不做无标识裸开关——标签说明「这个开关是什么」，配置面板说明「它会做什么」。
                      标签与 Switch 平级（不包进 <label>）：<label> 会把整行空白点击转发给开关，导致误切换 */}
                  <div className="flex shrink-0 flex-col items-end gap-1">
                    <div className="flex items-center gap-1.5">
                      <span className="text-[12px] text-mute">{t('project.hibernate')}</span>
                      <Switch
                        aria-label={`${p.name} ${t('project.hibernate')}`}
                        checked={p.hibernateEnabled}
                        // 任一行切换进行中即禁用全部开关：toggle 的 `togglingId !== 0` 门闸是全局的，
                        // 只禁用当前行会让其余行「看着可点、点了没反应」（死开关）
                        disabled={capable !== true || togglingId !== 0}
                        onCheckedChange={(next) => toggle(p, next)}
                      />
                    </div>
                    <span className="text-[11px] text-faint">
                      {p.hibernateEnabled ? t('project.hibernateToggleOn') : t('project.hibernateToggleOff')}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
          {/* 分页：count 未取尽时给「加载更多」，追加下一页（不打断已渲染行） */}
          {hasMore && !loading && (
            <div className="flex justify-center py-3">
              <Button
                size="sm"
                variant="outline"
                disabled={loadingMore}
                onClick={() => {
                  const next = page + 1
                  setPage(next)
                  void load(next, true)
                }}
              >
                {loadingMore && <Icon name="loader" className="size-4 animate-spin" />}
                {t('repos.projectsLoadMore')}
              </Button>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {t('common.close')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
