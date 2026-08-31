import { lazy, Suspense, useEffect, useRef, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { copyText } from '@/lib/copy'
import { Icon } from '@/components/Icons'
import { Tag } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { Textarea } from '@/components/ui/shadcn/textarea'
import { Switch } from '@/components/ui/shadcn/switch'
import { Avatar, AvatarFallback } from '@/components/ui/shadcn/avatar'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/shadcn/popover'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'
import { useAuth } from '@/features/auth/AuthProvider'
import { MemberInput } from './MemberInput'
// 创建项目弹窗静态依赖 Elements→CodeEditor(CodeMirror 620KB)，懒加载后延迟到点「新建」才拉取
const CreateProjectModal = lazy(() =>
  import('../projects/CreateProjectModal').then((m) => ({ default: m.CreateProjectModal })),
)
import { ProjectRow } from '@/features/projects/ProjectRow'

type NamespaceModel = components['schemas']['types.NamespaceModel']
type ProjectModel = components['schemas']['types.ProjectModel']
type ServiceEndpointModel = components['schemas']['types.ServiceEndpoint']
type MemberModel = components['schemas']['types.MemberModel']

/**
 * 展示用成员列表：后端创建者只写 namespace.creator_email、不写入 members 表
 * （biz.Create 仅 SetCreatorEmail，无成员行插入），新命名空间 members 恒为空，
 * 导致卡片计数「0 成员」把创建者漏掉。这里把创建者合成进展示列表
 * （显式成员已含创建者时不叠加），计数/头像/成员弹窗三者一致。
 * 注意：管理弹窗的可编辑成员仍用原始 members——owner 归属走转让管理，不做 tag 删除入口。
 */
function displayMembers(ns: NamespaceModel): MemberModel[] {
  if (!ns.creatorEmail) return ns.members
  if (ns.members.some((m) => m.email === ns.creatorEmail)) return ns.members
  // 合成行 id 用 0（哨兵，与 NamespaceManager.memberDisplayList 一致）：真实 member id 自增从 1 起，不冲突
  return [{ id: 0, email: ns.creatorEmail }, ...ns.members]
}

/**
 * 命名空间卡片：名称/描述 + 成员头像 + 项目数 + 收藏星 + 管理/删除入口。
 * 收藏切换乐观更新失败回滚；删除需二次确认；管理弹窗（描述/私有/成员/转让）仅 owner 可见。
 */
export function NamespaceCard({
  ns,
  loading = false,
  onToggleFavorite,
  onOpenProject,
  onDeleted,
  onChanged,
  dragHandle,
}: {
  ns: NamespaceModel
  /** 空间刷新中（其他用户部署/删除触发 ReloadProjects）：整卡覆盖层 + spinner，对齐旧版 ItemCard 的 Spin */
  loading?: boolean
  onToggleFavorite: (ns: NamespaceModel) => void
  /** 打开项目详情弹窗：弹窗状态已提升到工作台层（URL ?open= 持久化），卡片只上报点击 */
  onOpenProject: (p: ProjectModel) => void
  /** 删除命名空间成功后回调（携带空间 id，供工作台关闭该空间下已打开的弹窗） */
  onDeleted: (nsId: number) => void
  /** 空间内项目/配置变更回调（携带空间 id，供工作台按空间详情原地刷新）。稳定引用才让 memo 生效 */
  onChanged: (nsId: number) => void
  /** 拖拽排序手柄（关注 Tab 启用时注入）：渲染在右上图标簇最左端，不传入则不显示 */
  dragHandle?: ReactNode
}) {
  const { t } = useTranslation()
  const { user } = useAuth()
  const [busy, setBusy] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [deleting, setDeleting] = useState(false)
  // 创建项目弹窗
  const [createOpen, setCreateOpen] = useState(false)
  // 项目 >6 时折叠：只展示按更新时间排序最新的 6 个，其余折叠（点击展开/收起）
  const [projectsExpanded, setProjectsExpanded] = useState(false)

  // 管理弹窗：私有/成员/转让（描述编辑已上移到卡片内联）
  const [manageOpen, setManageOpen] = useState(false)
  // 成员弹窗：点击底部成员区打开，列出全部成员
  const [membersOpen, setMembersOpen] = useState(false)
  const [isPrivate, setIsPrivate] = useState(ns.private)
  const [membersList, setMembersList] = useState<string[]>([])
  const [transferEmail, setTransferEmail] = useState('')
  const [saving, setSaving] = useState(false)
  // 管理门控对齐后端 access.go：admin 绕过 owner 校验，普通用户仅创建者可见
  const isOwner = (user?.roles.includes('mars_admin') ?? false) || ns.creatorEmail === user?.email
  // 展示用成员（创建者合成进列表），卡片计数/头像/成员弹窗统一用它
  const members = displayMembers(ns)
  // 项目内联列表折叠：>6 时按更新时间降序保留最新 6 个，其余折叠；展开时展示全部（同一排序，折叠/展开不重排）。
  // 小列表（≤6）不排序、不折叠，保持后端原始顺序（避免常见布局变动）
  const projectCount = ns.projects.length
  const foldProjects = projectCount > 6
  const visibleProjects = foldProjects
    ? [...ns.projects]
        .sort(
          (a, b) =>
            (b.updatedAt ? new Date(b.updatedAt).getTime() : 0) -
            (a.updatedAt ? new Date(a.updatedAt).getTime() : 0),
        )
        .slice(0, projectsExpanded ? projectCount : 6)
    : ns.projects
  const foldedCount = foldProjects ? projectCount - 6 : 0

  // 仅在弹窗打开瞬间从当前 ns 快照表单字段，避免父级刷新 ns 时冲掉未保存的编辑
  useEffect(() => {
    if (!manageOpen) return
    setIsPrivate(ns.private)
    setMembersList(ns.members.map((m) => m.email))
    setTransferEmail('')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [manageOpen])

  const copyId = async () => {
    const ok = await copyText(String(ns.id))
    if (ok) toast.success(t('workbench.copyId'))
    else toast.error(t('common.copyFailed'))
  }

  const toggleFavorite = async () => {
    if (busy) return
    setBusy(true)
    const prev = ns.favorite
    onToggleFavorite({ ...ns, favorite: !prev }) // 乐观更新
    try {
      const { error } = await api.POST(API.namespacesFavorite, {
        body: { id: ns.id, favorite: !prev },
      })
      if (error) {
        onToggleFavorite({ ...ns, favorite: prev }) // 失败回滚
        throw new Error(error.message ?? String(error))
      }
      toast.success(!prev ? t('workbench.favoriteSuccess') : t('workbench.unfavoriteSuccess'))
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  const remove = async () => {
    setDeleting(true)
    try {
      const { error } = await api.DELETE(API.namespacesDetail, {
        params: { path: { id: ns.id } },
      })
      if (error) throw new Error(error.message ?? String(error))
      setConfirmOpen(false)
      toast.success(t('workbench.deleteSuccess', { name: ns.name }))
      onDeleted(ns.id)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setDeleting(false)
    }
  }

  /** 一次提交全部空间配置（私有/成员/转让），由后端 update_config 单事务原子落库；
   *  转让管理员非空才随配置一并转让，转让后当前用户不再是 owner，统一关闭弹窗。 */
  const saveConfig = async () => {
    if (saving) return
    setSaving(true)
    try {
      const email = transferEmail.trim()
      const { error } = await api.POST(API.namespacesUpdateConfig, {
        body: {
          id: ns.id,
          private: isPrivate,
          emails: membersList,
          ...(email ? { newAdminEmail: email } : {}),
        },
      })
      if (error) throw new Error(error.message ?? String(error))
      // 按实际改动分派具体消息：成员增删优先（对齐「空间成员添加成功」诉求），
      // 其次私有/转让，无差异改动兜底「空间配置已更新」
      const prevMembers = ns.members.map((m) => m.email)
      const added = membersList.filter((m) => !prevMembers.includes(m))
      const removed = prevMembers.filter((m) => !membersList.includes(m))
      const membersChanged = added.length > 0 || removed.length > 0
      const privateChanged = isPrivate !== ns.private
      if (membersChanged && removed.length === 0) {
        toast.success(t('workbench.membersAdded'))
      } else if (membersChanged && added.length === 0) {
        toast.success(t('workbench.membersRemoved'))
      } else if (membersChanged) {
        toast.success(t('workbench.membersSaved'))
      } else if (privateChanged) {
        toast.success(t('workbench.privateSaved'))
      } else if (email) {
        toast.success(t('workbench.transferSaved'))
      } else {
        toast.success(t('workbench.configSaved'))
      }
      setManageOpen(false)
      onChanged(ns.id)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="group relative flex h-full flex-col gap-3 rounded-lg border border-line bg-surface p-4 transition-[box-shadow,border-color] hover:border-primary/40 hover:shadow-xl hover:shadow-ink/20">
      {/* 头部：左侧 36px 正方形图标块与标题行顶对齐；右侧分上下两行 = 标题行(名称+操作簇) / 描述行 */}
      <div className="group/top flex items-start gap-2.5">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary-soft text-primary">
          <Icon name="namespace" className="text-[16px]" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            {/* 左组：title + 私有 + 复制 包一个 div（天然宽度，不 flex-1 撑宽，私有/复制紧贴名称）；操作簇由 justify-between 推到最右 */}
            <div className="flex min-w-0 items-center gap-2">
              <span className="min-w-0 truncate text-[14px] font-bold text-ink">{ns.name}</span>
              {ns.private && <Tag tone="accent" className="shrink-0">{t('workbench.private')}</Tag>}
              <Button
                variant="ghost"
                size="icon-xs"
                onClick={copyId}
                title={t('workbench.copyId')}
                aria-label={t('workbench.copyId')}
                className="text-faint hover:text-primary"
              >
                <Icon name="copy" className="size-4" />
              </Button>
            </div>
            {/* 管理员 + 空间资源用量 + 空间访问地址 + 关注：右组贴最右，紧凑图标簇（gap-0 无间距，
                每个都是 ghost icon-xs 标准按钮，hover 点亮 primary）。拖拽手柄（关注 Tab）插在最左端，
                与其余图标同一交互样式 */}
            <div className="flex shrink-0 items-center gap-0">
              {dragHandle}
              <NamespaceAdmin email={ns.creatorEmail} />
              <NamespaceCpuMemory namespaceId={ns.id} />
              <NamespaceEndpoints namespaceId={ns.id} />
              {/* 关注星：主题色实心填充（随换肤）；未关注为描边淡色 */}
              <Button
                variant="ghost"
                size="icon-xs"
                onClick={toggleFavorite}
                aria-pressed={ns.favorite}
                className={`${
                  ns.favorite ? 'text-primary hover:text-primary' : 'text-faint'
                }`}
                title={ns.favorite ? t('workbench.unfavorite') : t('workbench.favorite')}
                aria-label={ns.favorite ? t('workbench.unfavorite') : t('workbench.favorite')}
              >
                <Icon name="star" className={`size-4 ${ns.favorite ? 'fill-current' : ''}`} />
              </Button>
            </div>
          </div>
          <NamespaceDescription
            text={ns.description}
            namespaceId={ns.id}
            canEdit={isOwner}
            onChanged={onChanged}
          />
        </div>
      </div>

      {/* 项目内联列表：点击打开项目详情弹窗（可同时开多个）。md 起一行 2 个（对齐旧版 Col md={12}）。
          >6 时折叠（只渲染最新 6 个，底部展开/收起切换） */}
      {projectCount > 0 && (
        <>
          <div className="grid grid-cols-1 gap-1.5 md:grid-cols-2">
            {visibleProjects.map((p) => (
              <ProjectRow key={p.id} project={p} onClick={() => onOpenProject(p)} />
            ))}
          </div>
          {foldProjects && (
            <Button
              variant="dashed"
              size="xs"
              className="w-full"
              onClick={() => setProjectsExpanded((v) => !v)}
            >
              <Icon name={projectsExpanded ? 'collapse' : 'expand'} className="text-[12px]" />
              {projectsExpanded
                ? t('workbench.collapseProjects')
                : t('workbench.expandProjects', { count: foldedCount })}
            </Button>
          )}
        </>
      )}

      {/* 新建项目入口（旧版虚线按钮） */}
      <Button variant="dashed" size="xs" className="w-full" onClick={() => setCreateOpen(true)}>
        <Icon name="plus" className="text-[12px]" />
        {t('workbench.addProject')}
      </Button>

      {/* 底部：成员 + 项目数 + 删除 */}
      <div className="mt-auto flex items-center justify-between border-t border-line pt-3">
        {/* 成员区：点击弹窗展示全部成员 */}
        <button
          type="button"
          onClick={() => setMembersOpen(true)}
          aria-label={t('workbench.membersLabel')}
          title={t('workbench.membersLabel')}
          className="flex items-center gap-2 rounded-md px-1 py-0.5 transition-colors hover:bg-raised focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40"
        >
          <div className="flex -space-x-1.5">
            {members.slice(0, 3).map((m) => (
              <Avatar key={m.id} className="size-5 ring-2 ring-surface">
                <AvatarFallback
                  title={m.email}
                  className="bg-primary-soft font-mono text-[10px] font-bold text-primary"
                >
                  {m.email[0]?.toUpperCase() ?? ''}
                </AvatarFallback>
              </Avatar>
            ))}
          </div>
          <span className="font-mono text-[11px] text-faint">
            {members.length} {t('workbench.members')}
          </span>
        </button>
        <div className="flex items-center gap-1">
          <span className="flex items-center gap-1 rounded-md bg-raised px-2 py-1 font-mono text-[11px] text-mute">
            <Icon name="project" className="text-[12px]" />
            {ns.projects.length}
          </span>
          {/* 空间配置（管理）置于底部，与删除并列 */}
          {isOwner && (
            <Button
              variant="ghost"
              size="icon-xs"
              onClick={() => setManageOpen(true)}
              aria-label={t('workbench.manage')}
              title={t('workbench.manage')}
              className="text-faint hover:text-primary"
            >
              <Icon name="gear" className="size-4" />
            </Button>
          )}
          {/* 删除同样仅 owner 可见（对齐旧版 ItemCard useIsOwned 包裹删除按钮） */}
          {isOwner && (
            <Button
              variant="ghost"
              size="icon-xs"
              onClick={() => setConfirmOpen(true)}
              className="text-faint opacity-60 transition-[background-color,border-color,box-shadow,color,scale,opacity] hover:opacity-100 hover:text-err focus-visible:opacity-100"
              title={t('workbench.deleteNamespace')}
            >
              <Icon name="close" className="size-4" />
            </Button>
          )}
        </div>
      </div>

      {/* 成员弹窗：点击底部成员区打开，列出全部成员（owner 行打「所有者」标记） */}
      <Dialog open={membersOpen} onOpenChange={(o) => !o && setMembersOpen(false)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-[15px]">
              <Icon name="user" className="text-[14px]" />
              {t('workbench.membersLabel')}
              <span className="font-mono text-[12px] text-mute">· {ns.name}</span>
            </DialogTitle>
          </DialogHeader>
          <div className="flex max-h-[50vh] flex-col gap-0.5 overflow-y-auto py-1">
            {members.length === 0 ? (
              <div className="px-2 py-6 text-center text-[13px] text-faint">
                {t('common.empty')}
              </div>
            ) : (
              members.map((m) => (
                <div
                  key={m.id}
                  className="flex items-center gap-2.5 rounded-md px-2 py-1.5 hover:bg-raised"
                >
                  <Avatar className="size-6">
                    <AvatarFallback className="bg-primary-soft font-mono text-[10px] font-bold text-primary">
                      {m.email[0]?.toUpperCase() ?? ''}
                    </AvatarFallback>
                  </Avatar>
                  <span className="min-w-0 flex-1 truncate font-mono text-[12px] text-ink">
                    {m.email}
                  </span>
                  {m.email === ns.creatorEmail && (
                    <Tag tone="accent">{t('workbench.ownerTag')}</Tag>
                  )}
                </div>
              ))
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* 管理弹窗（仅 owner）：私有/成员/转让（描述编辑已上移到卡片内联） */}
      <Dialog open={manageOpen} onOpenChange={(o) => !o && setManageOpen(false)}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-[15px]">
              <Icon name="gear" className="text-[14px]" />
              {t('workbench.manage')}
              <span className="font-mono text-[12px] text-mute">· {ns.name}</span>
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-5 py-1">
            {/* 私有 */}
            <div className="space-y-1.5">
              <div className="text-[12px] text-mute">{t('workbench.privateLabel')}</div>
              <div className="flex items-center justify-between rounded-md border border-line px-3 py-2">
                <span className="text-[13px] text-ink">{t('workbench.private')}</span>
                <Switch checked={isPrivate} onCheckedChange={setIsPrivate} />
              </div>
            </div>

            {/* 成员 */}
            <div className="space-y-1.5">
              <label className="text-[12px] text-mute">{t('workbench.membersLabel')}</label>
              <MemberInput
                value={membersList}
                onChange={setMembersList}
                placeholder={t('workbench.membersPlaceholder')}
              />
              <p className="text-[11px] text-faint">{t('workbench.membersTip')}</p>
            </div>

            {/* 转让所有权 */}
            <div className="space-y-1.5">
              <label className="text-[12px] text-mute">{t('workbench.transferLabel')}</label>
              <Input
                value={transferEmail}
                onChange={(e) => setTransferEmail(e.target.value)}
                placeholder={t('workbench.transferPlaceholder')}
              />
              <p className="text-[11px] text-faint">
                {t('workbench.transferTip', { email: ns.creatorEmail })}
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setManageOpen(false)} disabled={saving}>
              {t('common.cancel')}
            </Button>
            <Button onClick={saveConfig} disabled={saving}>
              {saving && <Icon name="loader" className="size-4 animate-spin" />}
              {t('common.save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除确认 */}
      <Dialog open={confirmOpen} onOpenChange={(o) => !o && setConfirmOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t('workbench.deleteNamespace')}</DialogTitle>
          </DialogHeader>
          <p className="text-[13px] leading-relaxed text-mute">
            {t('workbench.deleteConfirm')}
            <span className="ml-1 font-medium text-ink">{ns.name}</span>？
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" disabled={deleting} onClick={remove}>
              {deleting && <Icon name="loader" className="size-4 animate-spin" />}
              {t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 创建项目弹窗：成功后刷新该卡片项目列表。Suspense fallback=null：弹窗懒加载完成即弹出 */}
      <Suspense fallback={null}>
        <CreateProjectModal
          namespaceId={ns.id}
          namespaceName={ns.name}
          open={createOpen}
          onClose={() => setCreateOpen(false)}
          onChanged={() => onChanged(ns.id)}
        />
      </Suspense>

      {/* 空间刷新中覆盖层：别人对该空间部署/删除时整卡置 loading，刷新完成即消失（对齐旧版 Spin spinning=loading） */}
      {loading && (
        <div className="absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-surface/70 backdrop-blur-[2px]">
          <Icon name="loader" className="size-5 animate-spin text-primary" />
        </div>
      )}
    </div>
  )
}

/**
 * 空间描述（对齐旧版 ItemCard 做法）：点击 Popover，内含 TextArea(rows=5) + 提交按钮的表单。
 * - 有描述：单行省略 + 悬浮卡片看全文；悬浮显示铅笔图标可再编辑。
 * - 无描述且可编辑：悬浮显示「暂无描述，点击添加」，点击弹出编辑表单。
 * - 无描述且不可编辑：「未知」占位。
 * 描述编辑入口放在卡片上，而非底部管理弹窗。
 */
function NamespaceDescription({
  text,
  namespaceId,
  canEdit,
  onChanged,
}: {
  text: string
  namespaceId: number
  canEdit: boolean
  onChanged: (nsId: number) => void
}) {
  const { t } = useTranslation()
  const ref = useRef<HTMLDivElement>(null)
  const [truncated, setTruncated] = useState(false)
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState(text)
  const [saving, setSaving] = useState(false)

  // 依赖只留 text：text 变化重建 + ResizeObserver 兜住尺寸变化。
  // truncated 放依赖会让它每次翻转都重建 observer（反模式，无谓开销）
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const check = () => setTruncated(el.scrollWidth > el.clientWidth)
    check()
    const ro = new ResizeObserver(check)
    ro.observe(el)
    return () => ro.disconnect()
  }, [text])

  const saveDesc = async () => {
    if (saving) return
    setSaving(true)
    try {
      const { error } = await api.POST(API.namespacesUpdateDesc, {
        params: { path: { id: namespaceId } },
        body: { id: namespaceId, desc: draft.trim() },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('workbench.descSaved'))
      setOpen(false)
      onChanged(namespaceId)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  const openEditor = (o: boolean) => {
    setOpen(o)
    if (o) setDraft(text)
  }

  const editor = (
    <PopoverContent side="top" className="w-[min(300px,80vw)] p-2">
      <div className="mb-1 px-1 text-[12px] font-medium">{t('workbench.descLabel')}</div>
      <Textarea
        autoFocus
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        rows={5}
        placeholder={t('workbench.descPlaceholder')}
        className="text-[12px]"
      />
      <div className="mt-2 flex justify-end">
        <Button size="sm" variant="outline" disabled={saving} onClick={saveDesc}>
          {saving && <Icon name="loader" className="size-3.5 animate-spin" />}
          {t('common.save')}
        </Button>
      </div>
    </PopoverContent>
  )

  // 无描述
  if (!text) {
    if (!canEdit) {
      return <div className="text-[12px] text-faint">{t('common.unknown')}</div>
    }
    return (
      // modal：编辑表单打开时外点只关闭，不误触下方项目行
      <Popover modal open={open} onOpenChange={openEditor}>
        <PopoverTrigger asChild>
          <button
            type="button"
            className="flex items-center gap-1 rounded text-[12px] text-faint opacity-0 transition-opacity group-hover/top:opacity-100 hover:text-primary focus-visible:opacity-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 pointer-coarse:opacity-100"
          >
            <Icon name="plus" className="text-[12px]" />
            {t('workbench.addDescription')}
          </button>
        </PopoverTrigger>
        {editor}
      </Popover>
    )
  }

  // 有描述：省略 + 悬浮全文 + 铅笔编辑
  return (
    <div className="flex items-center gap-1">
      <TooltipProvider delayDuration={100}>
        <Tooltip>
          <TooltipTrigger asChild>
            <div
              ref={ref}
              tabIndex={0}
              className="min-w-0 flex-1 truncate rounded text-[12px] text-faint focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40"
            >
              {text}
            </div>
          </TooltipTrigger>
          {truncated && (
            <TooltipContent
              side="top"
              className="max-w-[min(380px,80vw)] whitespace-pre-line break-words text-[12px] leading-relaxed"
            >
              {text}
            </TooltipContent>
          )}
        </Tooltip>
      </TooltipProvider>
      {canEdit && (
        <Popover modal open={open} onOpenChange={openEditor}>
          <PopoverTrigger asChild>
            <button
              type="button"
              aria-label={t('common.edit')}
              className="flex shrink-0 items-center rounded p-0.5 text-faint opacity-0 transition-opacity hover:text-primary focus-visible:opacity-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 group-hover:opacity-100 pointer-coarse:opacity-100"
            >
              <Icon name="pencil" className="size-3" />
            </button>
          </PopoverTrigger>
          {editor}
        </Popover>
      )}
    </div>
  )
}

/**
 * 空间管理员：icon 触发，点击 Popover 展示管理员邮箱（对齐顶部 icon 簇「点击弹层」交互）。
 * 顶部只占一个图标位，不挤占名称区；管理员名称通过弹层查看。
 */
function NamespaceAdmin({ email }: { email: string }) {
  const { t } = useTranslation()
  return (
    // modal：阻止外点透传到下层卡片内容（卡片顶 icon 簇 side=top 翻转后会盖住项目行）
    <Popover modal>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="icon-xs"
          aria-label={t('workbench.adminLabel')}
          title={t('workbench.adminLabel')}
          className="text-faint hover:text-primary"
        >
          <Icon name="crown" className="size-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent side="top" className="w-[max-content] max-w-[min(320px,80vw)] p-2">
        <div className="mb-1 px-1 text-[12px] font-medium">{t('workbench.adminLabel')}</div>
        <div className="flex items-center gap-1.5 px-1 pb-1 font-mono text-[12px]">
          <Icon name="user" className="shrink-0 text-[12px] text-faint" />
          <span className="break-all">{email || t('common.unknown')}</span>
        </div>
      </PopoverContent>
    </Popover>
  )
}

/**
 * 空间级 CPU/内存用量：点击时懒拉取 /api/metrics/namespace/{namespaceId}/cpu_memory。
 * 旧版为 hover Tooltip，触屏/键盘不可达；改为点击弹层满足「Hover vs Tap」（参考旧版 CpuMemory）。
 */
function NamespaceCpuMemory({ namespaceId }: { namespaceId: number }) {
  const { t } = useTranslation()
  const [usage, setUsage] = useState<{ cpu: string; memory: string } | null>(null)

  const fetchUsage = () => {
    if (usage) return
    api
      .GET(API.metricsNamespaceCpuMemory, {
        params: { path: { namespaceId } },
      })
      .then(({ data }) => {
        if (data) setUsage({ cpu: data.cpu, memory: data.memory })
      })
      .catch(() => setUsage({ cpu: '-', memory: '-' }))
  }

  return (
    // modal：阻止外点透传到底层项目行（点 icon 开 popover 后误触项目行打开弹窗）
    <Popover modal onOpenChange={(open) => open && fetchUsage()}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="icon-xs"
          aria-label={t('workbench.spaceCpuMemory')}
          className="text-faint hover:text-primary"
        >
          <Icon name="gauge" className="size-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent side="top" className="w-[min(240px,80vw)] p-2 font-mono text-[11px]">
        <div className="mb-1 px-1 text-[12px] font-medium">{t('workbench.spaceCpuMemory')}</div>
        {usage ? (
          <div className="flex flex-col gap-0.5 px-1 pb-1">
            <span>cpu: {usage.cpu || '-'}</span>
            <span>memory: {usage.memory || '-'}</span>
          </div>
        ) : (
          <div className="flex items-center gap-1.5 px-1 py-1 text-faint">
            <Icon name="loader" className="size-3 animate-spin" />
            {t('common.loading')}
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}

/** 空间端点：hover/点击时拉取 /api/endpoints/namespaces/{namespaceId}，支持 http 链接与复制（参考旧版 ServiceEndpoint） */
function NamespaceEndpoints({ namespaceId }: { namespaceId: number }) {
  const { t } = useTranslation()
  const [eps, setEps] = useState<ServiceEndpointModel[]>([])
  const [loaded, setLoaded] = useState(false)

  const fetchEndpoints = () => {
    if (loaded) return
    api
      .GET(API.endpointsNamespace, {
        params: { path: { namespaceId } },
      })
      .then(({ data }) => {
        setEps(data?.items ?? [])
        setLoaded(true)
      })
      .catch(() => setLoaded(true))
  }

  const copyUrl = async (url: string) => {
    const ok = await copyText(url)
    if (ok) toast.success(t('common.copied'))
    else toast.error(t('common.copyFailed'))
  }

  return (
    // modal：阻止外点透传到底层项目行（点 icon 开 popover 后误触项目行打开弹窗）
    <Popover modal onOpenChange={(open) => open && fetchEndpoints()}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="icon-xs"
          aria-label={t('workbench.endpoints')}
          className="text-faint hover:text-primary"
        >
          <Icon name="link" className="size-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent side="top" className="w-[max-content] max-w-[min(480px,90vw)] p-2">
        <div className="mb-1 px-1 text-[12px] font-medium">{t('workbench.endpoints')}</div>
        {!loaded ? (
          <div className="flex items-center gap-1.5 px-1 py-1 text-[12px] text-faint">
            <Icon name="loader" className="size-3 animate-spin" />
            {t('common.loading')}
          </div>
        ) : eps.length === 0 ? (
          <div className="px-1 py-1 text-[12px] text-faint">{t('common.empty')}</div>
        ) : (
          <div className="flex max-h-48 flex-col gap-0.5 overflow-auto">
            {eps.map((ep, i) => (
              <div
                key={i}
                className="flex items-center gap-1.5 rounded-md px-1 py-1 text-[12px] hover:bg-raised"
              >
                <span className="shrink-0 text-faint">
                  {ep.name}
                  {ep.portName ? `(${ep.portName})` : ''}:
                </span>
                {ep.url.startsWith('http') ? (
                  <a
                    href={ep.url}
                    target="_blank"
                    rel="noreferrer"
                    className="min-w-0 flex-1 truncate text-primary hover:underline"
                  >
                    {ep.url}
                  </a>
                ) : (
                  <span className="min-w-0 flex-1 truncate text-mute">{ep.url}</span>
                )}
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-xs"
                  onClick={() => copyUrl(ep.url)}
                  title={t('common.copied')}
                  className="shrink-0 text-faint hover:text-primary"
                >
                  <Icon name="copy" className="text-[11px]" />
                </Button>
              </div>
            ))}
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
