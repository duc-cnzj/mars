import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { invalidateHibernateWindows } from '@/api/hibernateWindowsCache'
import { Icon } from '@/components/Icons'
import { SearchableSelect, type SearchableSelectOption } from '@/components/SearchableSelect'
import { Empty, SkeletonList } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { Switch } from '@/components/ui/shadcn/switch'
import { Textarea } from '@/components/ui/shadcn/textarea'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/shadcn/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'

type WindowModel = components['schemas']['hibernate.HibernateWindow']

/** 弹窗表单草稿：编辑时带 id（并记下原名用于判断是否改名），新建时为空 */
type Draft = {
  id?: number
  /** 编辑时的原始名称：用于判断是否改名，仅改名时展示级联开关 */
  originalName?: string
  /** 改名时是否级联迁移引用旧名的仓库（仅改名时生效） */
  cascade: boolean
  name: string
  start: string
  end: string
  description: string
}

const EMPTY_DRAFT: Draft = { name: '', start: '23:00', end: '06:00', description: '', cascade: false }

/** 列表行「引用仓库」最多展示的 chip 数，超出折叠为 +N（hover 查看全部） */
const MAX_REPO_CHIPS = 6

/**
 * 列表行「引用仓库」chips：空态展示 i18n 文案，非空最多渲染 MAX_REPO_CHIPS 个仓库名、
 * 超出折叠为 +N（hover 显示其余）。名字过长行内截断（title 兜底）避免撑破行布局。
 */
function RepoRefChips({ names }: { names: string[] }) {
  const { t } = useTranslation()
  if (names.length === 0) {
    return <span className="text-[12px] text-faint">{t('hibernateWindows.reposEmpty')}</span>
  }
  return (
    <>
      {names.slice(0, MAX_REPO_CHIPS).map((name) => (
        <span
          key={name}
          className="max-w-[12rem] truncate rounded bg-raised px-1.5 py-0.5 font-mono text-[11px] text-mute"
          title={name}
        >
          {name}
        </span>
      ))}
      {names.length > MAX_REPO_CHIPS && (
        <span
          className="rounded bg-raised px-1.5 py-0.5 text-[11px] text-faint"
          title={names.slice(MAX_REPO_CHIPS).join(', ')}
        >
          +{names.length - MAX_REPO_CHIPS}
        </span>
      )}
    </>
  )
}

/**
 * 全局休眠时间窗管理（后台）：列表 + 新建/编辑 + 删除。
 * 时间窗全局共享，仓库在表单里按名称引用其中一个；列表数据量小，一次拉全、不分页。
 */
export function HibernateWindowManager() {
  const { t } = useTranslation()

  const [items, setItems] = useState<WindowModel[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  // null=关闭；非空=编辑该条（新建时 id 为空）
  const [draft, setDraft] = useState<Draft | null>(null)
  const [saving, setSaving] = useState(false)
  // 待删除项：null=关闭；非空=确认删除（受控，删除成功前不关闭）
  const [deleteTarget, setDeleteTarget] = useState<WindowModel | null>(null)
  const [deleting, setDeleting] = useState(false)

  // 一键配置引用仓库：null=关闭；非空=配置该窗引用哪些仓库（集合语义：勾选=引用）
  const [configTarget, setConfigTarget] = useState<WindowModel | null>(null)
  // 已勾选的仓库 ID（SearchableSelect 值为 string[]，提交前转 number）
  const [configSelected, setConfigSelected] = useState<string[]>([])
  // 候选仓库选项（懒拉 adminHibernateRepos，仅打开弹窗时拉一次）
  const [repoOptions, setRepoOptions] = useState<SearchableSelectOption[]>([])
  const [configLoading, setConfigLoading] = useState(false)
  const [configSaving, setConfigSaving] = useState(false)

  const fetchList = useCallback(async () => {
    try {
      const { data, error } = await api.GET(API.hibernateWindows)
      if (error) throw new Error(error.message ?? String(error))
      setItems(data?.items ?? [])
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }, [])

  useEffect(() => {
    void fetchList().finally(() => setLoading(false))
  }, [fetchList])

  const refresh = async () => {
    setRefreshing(true)
    await fetchList()
    setRefreshing(false)
  }

  const openCreate = () => setDraft({ ...EMPTY_DRAFT })

  const openEdit = (w: WindowModel) =>
    setDraft({
      id: w.id,
      originalName: w.name,
      cascade: false,
      name: w.name,
      start: w.start,
      end: w.end,
      description: w.description,
    })

  const submit = async () => {
    if (!draft) return
    if (!draft.name.trim()) {
      toast.error(t('hibernateWindows.nameRequired'))
      return
    }
    if (!draft.start || !draft.end) {
      toast.error(t('hibernateWindows.timeRequired'))
      return
    }
    setSaving(true)
    try {
      const body = {
        name: draft.name.trim(),
        start: draft.start,
        end: draft.end,
        description: draft.description.trim(),
      }
      if (draft.id == null) {
        const { error } = await api.POST(API.adminHibernateWindows, { body })
        if (error) throw new Error(error.message ?? String(error))
        toast.success(t('hibernateWindows.createSuccess'))
      } else {
        // cascade 仅在改名时有意义：未改名时后端会走普通更新（迁移数 0），随草稿一并透传即可。
        const { data, error } = await api.PUT(API.adminHibernateWindowDetail, {
          body: { id: draft.id, ...body, cascade: draft.cascade },
          params: { path: { id: draft.id } },
        })
        if (error) throw new Error(error.message ?? String(error))
        const migrated = data?.updatedRepoCount ?? 0
        toast.success(
          migrated > 0
            ? t('hibernateWindows.cascadeSuccess', { count: migrated })
            : t('hibernateWindows.updateSuccess'),
        )
      }
      // 增/改成功 → 立即失效全局时间窗缓存，避免其它页面（仓库表单 / 项目卡片 / 详情）读到旧时段
      invalidateHibernateWindows()
      setDraft(null)
      await refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  const remove = async () => {
    if (!deleteTarget || deleting) return
    setDeleting(true)
    try {
      const { error, response } = await api.DELETE(API.adminHibernateWindowDetail, {
        params: { path: { id: deleteTarget.id } },
      })
      // 400：时间窗仍被仓库按名引用而拒删（后端守卫）——给可本地化的专门提示并保留弹窗，
      // 待用户去仓库配置解除引用后重试；不同于其它错误的原始消息 toast，此处需明确告知处置动作。
      if (response?.status === 400) {
        toast.error(t('hibernateWindows.deleteInUse'))
        return
      }
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('hibernateWindows.deleteSuccess'))
      // 删除成功 → 失效缓存（否则消费方仍可能按窗名回显已被删的时段）
      invalidateHibernateWindows()
      setDeleteTarget(null)
      await refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setDeleting(false)
    }
  }

  // 打开「配置引用仓库」弹窗：清空选择后懒拉候选仓库（已配休眠段的仓库，含禁用），
  // 并以该窗当前引用的仓库名预勾选（repoNames 与候选同源，正常情况下全部命中）。
  const openConfig = async (w: WindowModel) => {
    setConfigTarget(w)
    setConfigSelected([])
    setConfigLoading(true)
    try {
      const { data, error } = await api.GET(API.adminHibernateRepos)
      if (error) throw new Error(error.message ?? String(error))
      const opts: SearchableSelectOption[] = (data?.items ?? []).map((r) => ({
        value: String(r.id),
        label: r.name,
      }))
      setRepoOptions(opts)
      const referenced = new Set(w.repoNames ?? [])
      setConfigSelected(opts.filter((o) => referenced.has(o.label)).map((o) => o.value))
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
      setConfigTarget(null)
    } finally {
      setConfigLoading(false)
    }
  }

  // 保存引用集合：PUT body.repoIds 为目标集合（集合语义，缺则加入、多则移除）。
  // 后端 400（仓库未配休眠段、或移除会让某启用仓库失去最后一个窗）走原始消息 toast。
  const saveConfig = async () => {
    if (!configTarget || configSaving) return
    setConfigSaving(true)
    try {
      const { data, error } = await api.PUT(API.adminHibernateWindowRepos, {
        params: { path: { id: configTarget.id } },
        body: { id: configTarget.id, repoIds: configSelected.map(Number) },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(
        t('hibernateWindows.configReposSuccess', { count: data?.repoNames?.length ?? 0 }),
      )
      setConfigTarget(null)
      await refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setConfigSaving(false)
    }
  }

  return (
    <div className="flex flex-col gap-3">
      {/* 工具栏：标题 + 刷新 + 新建 */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-col gap-0.5">
          <h2 className="text-[16px] font-semibold text-ink">{t('hibernateWindows.title')}</h2>
          <span className="text-[12px] text-faint">{t('hibernateWindows.subtitle')}</span>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button size="sm" variant="outline" disabled={refreshing} onClick={refresh}>
            {refreshing ? (
              <Icon name="loader" className="size-4 animate-spin" />
            ) : (
              <Icon name="refresh" className="size-4" />
            )}
            {t('common.refresh')}
          </Button>
          <Button size="sm" variant="default" onClick={openCreate}>
            <Icon name="plus" className="size-4" />
            {t('hibernateWindows.create')}
          </Button>
        </div>
      </div>

      {/* 列表 */}
      <div className="rounded-lg border border-line bg-surface">
        {loading ? (
          <SkeletonList count={4} bare />
        ) : items.length === 0 ? (
          <div className="flex items-center justify-center p-8">
            <Empty text={t('hibernateWindows.empty')} icon="clock" />
          </div>
        ) : (
          <div className="divide-y divide-line">
            {items.map((w, index) => (
              <div
                key={w.id}
                className="animate-list-in flex items-start justify-between gap-4 px-5 py-4 transition-colors hover:bg-raised"
                style={{ animationDelay: `${Math.min(index, 10) * 30}ms` }}
              >
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-[13px] font-medium text-ink">{w.name}</span>
                    <span className="inline-flex items-center gap-1 rounded bg-raised px-1.5 py-0.5 font-mono text-[12px] text-primary">
                      <Icon name="clock" className="size-3" />
                      {w.start} → {w.end}
                    </span>
                  </div>
                  {w.description && (
                    <div className="mt-1.5 text-[12px] text-mute">{w.description}</div>
                  )}
                  {/* 引用仓库：展示哪些仓库用了这个窗（内容可能多，chips 折叠见 RepoRefChips） */}
                  <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
                    <span className="text-[12px] text-faint">{t('hibernateWindows.reposLabel')}</span>
                    <RepoRefChips names={w.repoNames ?? []} />
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <Button size="sm" variant="outline" onClick={() => openConfig(w)}>
                    {t('hibernateWindows.configRepos')}
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => openEdit(w)}>
                    {t('hibernateWindows.edit')}
                  </Button>
                  <AlertDialog
                    open={deleteTarget?.id === w.id}
                    onOpenChange={(o) => !o && !deleting && setDeleteTarget(null)}
                  >
                    <Button size="sm" variant="destructive" onClick={() => setDeleteTarget(w)}>
                      {t('hibernateWindows.delete')}
                    </Button>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>
                          {t('hibernateWindows.deleteConfirm', { name: w.name })}
                        </AlertDialogTitle>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel disabled={deleting}>{t('common.cancel')}</AlertDialogCancel>
                        {/* 普通 Button 而非 AlertDialogAction：后者点击默认关闭弹窗，会「还没删除成功就消失」 */}
                        <Button variant="destructive" disabled={deleting} onClick={remove}>
                          {deleting && <Icon name="loader" className="size-4 animate-spin" />}
                          {t('hibernateWindows.delete')}
                        </Button>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 新建 / 编辑弹窗 */}
      <Dialog open={draft != null} onOpenChange={(o) => !o && setDraft(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {draft?.id == null ? t('hibernateWindows.create') : t('hibernateWindows.edit')}
            </DialogTitle>
          </DialogHeader>
          {draft && (
            <div className="flex flex-col gap-4">
              <label className="flex flex-col gap-1.5">
                <span className="text-[12px] font-medium text-mute">{t('hibernateWindows.name')}</span>
                <Input
                  value={draft.name}
                  maxLength={64}
                  placeholder={t('hibernateWindows.namePlaceholder')}
                  onChange={(e) => setDraft({ ...draft, name: e.target.value })}
                />
              </label>
              {/* 级联开关仅在「编辑且改名」时出现：未改名时改的是时段/备注，不涉及引用迁移 */}
              {draft.id != null && draft.name.trim() !== (draft.originalName ?? '') && (
                <div className="flex flex-col gap-1.5 rounded-md border border-line bg-raised px-3 py-2.5">
                  {/* 用 div 而非 label：Radix Switch 的 <button> 属可标注元素，label 会让整行点击都切换开关 */}
                  <div className="flex items-center gap-3">
                    <Switch
                      checked={draft.cascade}
                      onCheckedChange={(v) => setDraft({ ...draft, cascade: v })}
                      aria-label={t('hibernateWindows.cascade')}
                    />
                    <span className="text-[12px] font-medium text-ink">
                      {t('hibernateWindows.cascade')}
                    </span>
                  </div>
                  <span className="text-[12px] text-faint">{t('hibernateWindows.cascadeTip')}</span>
                </div>
              )}
              <div className="flex gap-3">
                <label className="flex flex-1 flex-col gap-1.5">
                  <span className="text-[12px] font-medium text-mute">{t('hibernateWindows.start')}</span>
                  <Input
                    type="time"
                    value={draft.start}
                    onChange={(e) => setDraft({ ...draft, start: e.target.value })}
                  />
                </label>
                <label className="flex flex-1 flex-col gap-1.5">
                  <span className="text-[12px] font-medium text-mute">{t('hibernateWindows.end')}</span>
                  <Input
                    type="time"
                    value={draft.end}
                    onChange={(e) => setDraft({ ...draft, end: e.target.value })}
                  />
                </label>
              </div>
              <span className="text-[12px] text-faint">{t('hibernateWindows.windowTip')}</span>
              <label className="flex flex-col gap-1.5">
                <span className="text-[12px] font-medium text-mute">
                  {t('hibernateWindows.description')}
                </span>
                <Textarea
                  value={draft.description}
                  rows={2}
                  placeholder={t('hibernateWindows.descriptionPlaceholder')}
                  onChange={(e) => setDraft({ ...draft, description: e.target.value })}
                />
              </label>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDraft(null)}>
              {t('common.cancel')}
            </Button>
            <Button variant="default" disabled={saving} onClick={submit}>
              {saving && <Icon name="loader" className="size-4 animate-spin" />}
              {t('hibernateWindows.save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 一键配置引用仓库弹窗：把一个时间窗配置到多个仓库（集合语义：勾选=引用） */}
      <Dialog
        open={configTarget != null}
        onOpenChange={(o) => !o && !configSaving && setConfigTarget(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {t('hibernateWindows.configReposTitle', { name: configTarget?.name ?? '' })}
            </DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-3">
            <span className="text-[12px] text-faint">{t('hibernateWindows.configReposHint')}</span>
            {configLoading ? (
              <SkeletonList count={2} bare />
            ) : repoOptions.length === 0 ? (
              <div className="flex items-center justify-center py-6">
                <Empty text={t('hibernateWindows.configReposEmpty')} icon="repo" />
              </div>
            ) : (
              <SearchableSelect
                multiple
                value={configSelected}
                options={repoOptions}
                onChange={(v) => setConfigSelected(v as string[])}
                placeholder={t('hibernateWindows.configReposPlaceholder')}
                searchPlaceholder={t('common.search')}
                emptyText={t('common.empty')}
              />
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" disabled={configSaving} onClick={() => setConfigTarget(null)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="default"
              disabled={configSaving || configLoading || repoOptions.length === 0}
              onClick={saveConfig}
            >
              {configSaving && <Icon name="loader" className="size-4 animate-spin" />}
              {t('hibernateWindows.save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
