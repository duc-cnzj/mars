import {
  lazy,
  Suspense,
  useEffect,
  useRef,
  useState,
  type MouseEvent,
  type ReactNode,
} from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import type { TKey } from '@/i18n/keys'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { copyText } from '@/lib/copy'
import { selectAllOnDoubleClick } from '@/lib/selection'
import { useOverlayZ } from '@/hooks/useOverlayZ'
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
/** i18n 取词最小签名：只要「字面量 key → 文案」这一面（TKey 约束编译期防悬空 key）。
 *  不用 i18next TFunction / ReturnType<useTranslation>：前者的泛型在本仓 strict 下实例化过深，
 *  后者把 useTranslation 的泛型重载原样带出，同样报 TS2589。 */
type TFn = (key: TKey) => string

/**
 * 双击整选 + 复制（卡片标题→空间名称、成员/管理员邮箱共用）：
 * 整选给到视觉反馈（浏览器默认按「词」断选，mars-demo 只选到半截），复制结果走 toast。
 * 提到模块层是因为空间信息弹窗（NamespaceInfoDialog）也要用同一交互。
 */
function makeDoubleClickCopy(t: TFn) {
  return (text: string, doneKey: TKey) => (e: MouseEvent<HTMLElement>) => {
    selectAllOnDoubleClick(e)
    void copyText(text).then((ok) =>
      ok ? toast.success(t(doneKey)) : toast.error(t('common.copyFailed')),
    )
  }
}

/*
 * 卡片上四个弹窗（成员/信息/管理/删除确认）与项目详情弹窗**平级**——后者是可拖拽宿主，
 * z 取 nextZIndex()（从 51 起），而这四个走 shadcn 默认 z-50，项目详情弹窗打开时会被整块压住，
 * 表现为「点了成员没反应」（弹窗确实开了，只是被盖住）。故统一走 useOverlayZ 在打开时置顶。
 *
 * 遮罩同步抬升（DialogContent raiseOverlay）：四者都是顶层（DialogDepthContext depth=0），
 * 默认遮罩恒 z-50，只抬 content 的话被压住的项目详情弹窗不会变暗，两层弹窗像硬叠在一起。
 * 这四个弹窗不是兄弟多开，抬遮罩不会引发兄弟互压（见 dialog.tsx 顶部注释）。
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
  // 空间信息弹窗：点标题左侧图标打开，聚合管理员 / 空间资源总使用量 / 空间访问地址
  const [infoOpen, setInfoOpen] = useState(false)
  // 四个弹窗的 z：打开时置顶，压过同级的项目详情弹窗（可拖拽宿主 z≥51）
  const confirmZ = useOverlayZ(confirmOpen)
  const manageZ = useOverlayZ(manageOpen)
  const membersZ = useOverlayZ(membersOpen)
  const infoZ = useOverlayZ(infoOpen)
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

  const copyOnDoubleClick = makeDoubleClickCopy(t)

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
        {/* 标题左侧图标块 = 「空间信息」入口：管理员 / 空间资源总使用量 / 空间访问地址三块内容收进同一弹窗，
            顶部图标簇不再各占一位（原三个 popover 触发图标已移除）。
            悬停/聚焦换成问号并左右摆动（animate-icon-wobble）——问号是「这里是什么」的语义提示，
            暗示可点开看详情。纯 CSS 显隐切换（group-hover/focus-visible），不引入 React 状态、不额外渲染 */}
        <button
          type="button"
          onClick={() => setInfoOpen(true)}
          aria-label={t('workbench.namespaceInfo')}
          title={t('workbench.namespaceInfo')}
          className="group/info flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary-soft text-primary transition-colors hover:bg-primary/15 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40"
        >
          <Icon
            name="namespace"
            className="text-[16px] group-hover/info:hidden group-focus-visible/info:hidden"
          />
          {/* 问号比默认图标大一号（20px vs 16px）：裸问号笔画只占 24 网格的六成，
              同字号下视觉重量明显轻于空间图标，放大补回来 */}
          <Icon
            name="question"
            className="hidden animate-icon-wobble text-[20px] group-hover/info:block group-focus-visible/info:block"
          />
        </button>
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            {/* 左组：title + 私有 包一个 div（天然宽度，不 flex-1 撑宽，私有紧贴名称）；操作簇由 justify-between 推到最右 */}
            <div className="flex min-w-0 items-center gap-2">
              {/* 双击标题：整选 + 复制空间名称（见 copyOnDoubleClick）；title 提示交互，复制反馈走 toast */}
              <span
                className="min-w-0 truncate text-[14px] font-bold text-ink"
                title={t('workbench.copyNameTip')}
                onDoubleClick={copyOnDoubleClick(ns.name, 'workbench.copyName')}
              >
                {ns.name}
              </span>
              {ns.private && <Tag tone="accent" className="shrink-0">{t('workbench.private')}</Tag>}
            </div>
            {/* 右组贴最右，紧凑图标簇（gap-0 无间距）。拖拽手柄（关注 Tab）插在最左端，与关注星同一交互样式；
                管理员/资源用量/访问地址已收进左侧图标块的「空间信息」弹窗 */}
            <div className="flex shrink-0 items-center gap-0">
              {dragHandle}
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
          {/* 空间 ID 已移入「空间信息」弹窗标题行（名称可改、ID 稳定，排障对账时一眼定位），footer 只留项目数 */}
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
        <DialogContent className="sm:max-w-md" style={{ zIndex: membersZ }} raiseOverlay>
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
                  {/* 双击邮箱：整选 + 复制（成员邮箱常要转发/粘贴，双击即得） */}
                  <span
                    className="min-w-0 flex-1 truncate font-mono text-[12px] text-ink"
                    title={t('workbench.copyMemberEmailTip')}
                    onDoubleClick={copyOnDoubleClick(m.email, 'workbench.copyMemberEmail')}
                  >
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

      {/* 空间信息弹窗：标题左侧图标触发，聚合管理员 / 资源用量 / 访问地址 */}
      <NamespaceInfoDialog ns={ns} open={infoOpen} onOpenChange={setInfoOpen} z={infoZ} />

      {/* 管理弹窗（仅 owner）：私有/成员/转让（描述编辑已上移到卡片内联） */}
      <Dialog open={manageOpen} onOpenChange={(o) => !o && setManageOpen(false)}>
        <DialogContent className="sm:max-w-lg" style={{ zIndex: manageZ }} raiseOverlay>
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
        <DialogContent className="max-w-md" style={{ zIndex: confirmZ }} raiseOverlay>
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
 * 空间信息弹窗：卡片左上图标触发，聚合原先顶部三个图标 popover 的内容——
 * 管理员邮箱（creator_email）/ 空间资源总使用量（懒拉 metricsNamespaceCpuMemory）/ 空间访问地址（懒拉 endpointsNamespace）。
 * 两块远端数据都在弹窗打开时拉取、各带「已加载则短路」的缓存，语义与原 popover 一致（不问不拉、拉过不重拉）。
 */
function NamespaceInfoDialog({
  ns,
  open,
  onOpenChange,
  z,
}: {
  ns: NamespaceModel
  open: boolean
  onOpenChange: (open: boolean) => void
  /** 顶层弹窗 z（打开时取 nextZIndex()）：压过同级可拖拽的项目详情弹窗 */
  z: number
}) {
  const { t } = useTranslation()
  const copyOnDoubleClick = makeDoubleClickCopy(t)
  const [usage, setUsage] = useState<{ cpu: string; memory: string } | null>(null)
  const [eps, setEps] = useState<ServiceEndpointModel[]>([])
  const [epsLoaded, setEpsLoaded] = useState(false)

  /** 资源用量：拉不到就落 '-' 占位（不让弹窗卡在 loading） */
  const fetchUsage = () => {
    if (usage) return
    api
      .GET(API.metricsNamespaceCpuMemory, { params: { path: { namespaceId: ns.id } } })
      .then(({ data }) => {
        if (data) setUsage({ cpu: data.cpu, memory: data.memory })
      })
      .catch(() => setUsage({ cpu: '-', memory: '-' }))
  }

  /** 访问地址：失败也置 loaded，落到「暂无」空态 */
  const fetchEndpoints = () => {
    if (epsLoaded) return
    api
      .GET(API.endpointsNamespace, { params: { path: { namespaceId: ns.id } } })
      .then(({ data }) => {
        setEps(data?.items ?? [])
        setEpsLoaded(true)
      })
      .catch(() => setEpsLoaded(true))
  }

  // 打开即并发拉两块（依赖只挂 open：ns 变更会重建卡片，闭包里的 ns.id 恒为当次卡片的空间）
  useEffect(() => {
    if (!open) return
    fetchUsage()
    fetchEndpoints()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const copyUrl = async (url: string) => {
    const ok = await copyText(url)
    if (ok) toast.success(t('common.copied'))
    else toast.error(t('common.copyFailed'))
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {/* 比同文件的成员/管理弹窗(512) 宽两档到 2xl(672)：主体是访问地址，地址整条展开吃横向空间。
          加高走「容器留白」——p-8 替代基类 p-6、gap-6 替代 gap-4，只放大卡片内边距与标题间隔，
          不碰列表行距（行距刚收过，再放大会反弹成上一版的松散感） */}
      <DialogContent className="sm:max-w-2xl gap-6 p-8" style={{ zIndex: z }} raiseOverlay>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-[15px]">
            <Icon name="namespace" className="text-[17px]" />
            {t('workbench.namespaceInfo')}
            {/* 空间名 + ID 都走标签（Tag）而非 · 连缀纯文本：名称是弹窗主体用 accent 底、
                ID 是辅助标识用 mute 底，两个 chip 一深一浅即分层，扫读不再糊成一串。
                dot=false：Tag 的点是状态语义词，这里只是标识，不带状态含义。
                text-[13px] 覆盖 Tag 基类的 11px：全弹窗统一到「标题 15 / 正文 13」两级，
                chip 跟着正文走而非停在徽标层（Tag 用字符串拼接、经 Badge 的 cn() 兜底，
                后写的 text-* / font-* 覆盖基类，与 font-normal 同理） */}
            <Tag tone="accent" dot={false} className="font-mono text-[13px] font-normal">
              {ns.name}
            </Tag>
            <Tag tone="mute" dot={false} className="font-mono text-[13px] font-normal">
              ID {ns.id}
            </Tag>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-3 py-1">
          {/* 管理员：双击邮箱整选 + 复制（与卡片标题、成员弹窗行同一交互） */}
          <div className="space-y-1.5">
            <div className="flex items-center gap-1.5 text-[13px] text-mute">
              <Icon name="crown" className="text-[13px]" />
              {t('workbench.adminLabel')}
            </div>
            <div className="rounded-md border border-line px-3 py-2.5">
              <span
                className="block truncate font-mono text-[13px] text-ink"
                title={t('workbench.copyMemberEmailTip')}
                onDoubleClick={copyOnDoubleClick(
                  ns.creatorEmail,
                  'workbench.copyMemberEmail',
                )}
              >
                {ns.creatorEmail || t('common.unknown')}
              </span>
            </div>
          </div>

          {/* 空间资源总使用量 */}
          <div className="space-y-1.5">
            <div className="flex items-center gap-1.5 text-[13px] text-mute">
              <Icon name="gauge" className="text-[13px]" />
              {t('workbench.spaceCpuMemory')}
            </div>
            <div className="rounded-md border border-line px-3 py-2.5 font-mono text-[13px]">
              {usage ? (
                <div className="flex flex-col gap-0.5 text-ink">
                  <span>cpu: {usage.cpu || '-'}</span>
                  <span>memory: {usage.memory || '-'}</span>
                </div>
              ) : (
                <div className="flex items-center gap-1.5 text-faint">
                  <Icon name="loader" className="size-3 animate-spin" />
                  {t('common.loading')}
                </div>
              )}
            </div>
          </div>

          {/* 空间访问地址：http 链接可直接打开，右侧复制按钮 */}
          <div className="space-y-1.5">
            <div className="flex items-center gap-1.5 text-[13px] text-mute">
              <Icon name="link" className="text-[13px]" />
              {t('workbench.endpoints')}
            </div>
            {/* 外框只留 4px：行自带 px-2 py-1.5（悬停底色要贴边），外框再给 8px 会双层内缩 */}
            <div className="rounded-md border border-line p-1">
              {!epsLoaded ? (
                <div className="flex items-center gap-1.5 px-2 py-1.5 text-[13px] text-faint">
                  <Icon name="loader" className="size-3 animate-spin" />
                  {t('common.loading')}
                </div>
              ) : eps.length === 0 ? (
                <div className="px-2 py-1.5 text-[13px] text-faint">{t('common.empty')}</div>
              ) : (
                <div className="flex max-h-48 flex-col overflow-auto">
                  {eps.map((ep, i) => (
                    <div
                      key={i}
                      className="flex items-center gap-2 rounded-md px-2 py-1 text-[13px] hover:bg-raised"
                    >
                      {/* 名称/端口 + 地址同行不换行：名称是扫读锚点放最左（加粗 ink，不再是原来
                          压成 text-faint 几乎糊掉的那版），与地址之间留 gap-2 断开、不再用冒号粘连；
                          地址 truncate 保证永不折行，完整值靠 title 悬停与右侧复制按钮兜底 */}
                      <span className="shrink-0">
                        <span className="font-medium text-ink">{ep.name}</span>
                        {ep.portName && <span className="text-mute"> · {ep.portName}</span>}
                      </span>
                      {ep.url.startsWith('http') ? (
                        <a
                          href={ep.url}
                          target="_blank"
                          rel="noreferrer"
                          title={ep.url}
                          className="min-w-0 flex-1 truncate font-mono text-primary hover:underline"
                        >
                          {ep.url}
                        </a>
                      ) : (
                        <span className="min-w-0 flex-1 truncate font-mono text-ink" title={ep.url}>
                          {ep.url}
                        </span>
                      )}
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-xs"
                        onClick={() => copyUrl(ep.url)}
                        aria-label={t('common.copy')}
                        title={t('common.copy')}
                        className="shrink-0 text-faint hover:text-primary"
                      >
                        <Icon name="copy" className="text-[13px]" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
