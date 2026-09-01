import { memo, useCallback, useEffect, useRef, useState, type CSSProperties } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/features/auth/AuthProvider'
import { Icon } from '@/components/Icons'
import { SEARCH_DEBOUNCE_MS } from '@/lib/constants'
import { SearchInput } from '@/components/SearchInput'
import { StatCard } from '@/components/StatCard'
import { Empty, RefreshFade, SkeletonList, Tag } from '@/components/ui'
import { Avatar, AvatarFallback } from '@/components/ui/shadcn/avatar'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/shadcn/alert-dialog'
import { Button } from '@/components/ui/shadcn/button'
import { copyText } from '@/lib/copy'
import { formatDateTime } from '@/lib/format'
import { humanizeDateTime } from '@/lib/humanizeDateTime'
import { toast } from '@/lib/toast'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import type { components } from '@/api/schema'

type UserModel = components['schemas']['user.UserModel']
type UserStats = components['schemas']['user.UserStats']

/** 每页条数（服务端分页，滚动触底追加下一页） */
const PAGE_SIZE = 15

/** 用户行（React.memo）：行级 props 全部稳定（user 引用 + useCallback handler），
 *  列表状态变化（关键词输入/只看管理员/排序切换/滚动揭示）时行不重渲染，杜绝 O(n) 整表刷新。 */
const UserRow = memo(function UserRow({
  user,
  isOpen,
  toggling,
  resetOpen,
  resetting,
  canManage,
  onToggle,
  onClose,
  onConfirm,
  onReset,
  onResetClose,
  onResetConfirm,
  onCopyEmail,
  className,
  style,
}: {
  user: UserModel
  isOpen: boolean
  toggling: boolean
  resetOpen: boolean
  resetting: boolean
  /** 当前登录用户是否为超管：false 时本行只读（普通管理员只能查看，不能改他人权限） */
  canManage: boolean
  onToggle: (u: UserModel) => void
  onClose: () => void
  onConfirm: () => void
  onReset: (u: UserModel) => void
  onResetClose: () => void
  onResetConfirm: () => void
  onCopyEmail: (email: string) => void
  /** RefreshFade 经 cloneElement 注入的渐入 class/延迟——须转发到根元素才生效 */
  className?: string
  style?: CSSProperties
}) {
  const { t } = useTranslation()
  const isAdmin = user.roles.includes('admin')
  // 超管标识来自后端（user.UserModel.is_super_admin，由服务端按内置超管邮箱判定），前端不写死邮箱
  const isSuper = user.isSuperAdmin
  // 角色来源：roles_override=false（未被后台接管）= 生效角色来自最近一次 SSO 登录同步；
  // true（已被后台接管）= 生效角色来自超管手动设置。仅在非超管管理员上展示来源徽标，
  // 内置超管是固定身份，无来源语义。
  const sourceOverride = user.rolesOverride
  return (
    <div
      className={`grid grid-cols-1 gap-2 border-b border-line px-4 py-2.5 last:border-b-0 sm:grid-cols-2 lg:grid-cols-[minmax(0,2fr)_minmax(0,1.5fr)_7rem_15rem] lg:items-center ${className ?? ''}`}
      style={style}
    >
      {/* 用户：头像（系统无头像字段，统一图标占位）+ 姓名 + 邮箱 */}
      <div className="flex min-w-0 items-center gap-2.5">
        <Avatar className="size-8 shrink-0">
          <AvatarFallback className="bg-primary-soft text-primary">
            <Icon name="user" className="size-4" />
          </AvatarFallback>
        </Avatar>
        <div className="min-w-0">
          <div className="truncate text-[13px] font-medium text-ink">{user.name}</div>
          <div className="flex min-w-0 items-center gap-1">
            <span className="truncate font-mono text-[11px] text-faint">{user.email}</span>
            <button
              type="button"
              onClick={() => onCopyEmail(user.email)}
              aria-label={t('common.copy')}
              title={t('common.copy')}
              className="shrink-0 text-faint opacity-60 transition-opacity hover:text-ink hover:opacity-100"
            >
              <Icon name="copy" className="size-3" />
            </button>
          </div>
        </div>
      </div>

      {/* 角色：超级管理员单独打标（超管=最高身份，不再叠加管理员标）；普通用户由空 roles 推导
          （服务端不再返回 user）。非超管管理员附带来源徽标：roles_override=false=SSO 带来，
          true=后台手动设置 */}
      <div className="flex flex-wrap items-center gap-1">
        {isSuper ? (
          <Tag tone="accent" dot={false}>
            {t('users.superAdmin')}
          </Tag>
        ) : isAdmin ? (
          <Tag tone="accent" dot={false}>
            {t('users.roleAdmin')}
          </Tag>
        ) : (
          <Tag tone="mute" dot={false}>
            {t('users.roleUser')}
          </Tag>
        )}
        {isAdmin && !isSuper && (
          <Tag tone={sourceOverride ? 'warn' : 'info'} dot={false}>
            {t(sourceOverride ? 'users.roleSourceManual' : 'users.roleSourceSSO')}
          </Tag>
        )}
      </div>

      {/* 最近登录：从未登录显示占位；有登录记录则相对时间为主（humanizeDateTime 跟随
          locale），精确时间放原生 tooltip 可溯源。lastLogin 可能缺失（后端省略字段），
          formatDateTime/humanizeDateTime 对非法输入均返回空串，不再抛 RangeError */}
      <span className="text-[11px] text-ink">
        {user.lastLogin ? (
          <time dateTime={user.lastLogin} title={formatDateTime(user.lastLogin)}>
            {humanizeDateTime(user.lastLogin)}
          </time>
        ) : (
          <span className="text-faint">{t('users.neverLoggedIn')}</span>
        )}
      </span>

      {/* 操作：仅超管可修改他人权限（需求：普通管理员只能查看不能改）；二次确认；内置超管是固定身份，
          操作栏以「-」占位（无降级入口，连禁用按钮都不放）；其余按钮统一尺寸 + 固定等宽 w-28，
          按语义分层变体——「设为管理员」主操作 default 实心，「移除管理员」危险动作 destructive 红色实心
          （与确认弹窗红键同语义），「恢复同步」次级操作 outline 描边，杜绝「有的按钮大有的按钮小」 */}
      {!canManage ? (
        <span className="text-[11px] text-faint">{t('users.viewOnly')}</span>
      ) : isSuper ? (
        <span className="text-[11px] text-faint">-</span>
      ) : isAdmin ? (
        <div className="flex items-center gap-1.5">
          {/* 恢复 SSO 同步：仅后台手动接管中的管理员显示（rolesOverride=true），
              解除接管标记把该用户交还给 SSO 角色同步（下一次登录生效） */}
          {sourceOverride && (
            <AlertDialog open={resetOpen} onOpenChange={(o) => !o && !resetting && onResetClose()}>
              <Button
                size="sm"
                variant="outline"
                onClick={() => onReset(user)}
                title={t('users.restoreSSOSync')}
                className="w-28"
              >
                <Icon name="restore" className="size-3.5" />
                {t('users.restoreSSOSyncShort')}
              </Button>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>{t('users.restoreSSOSyncConfirmTitle', { name: user.name })}</AlertDialogTitle>
                  <AlertDialogDescription>{t('users.restoreSSOSyncConfirmDesc')}</AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel disabled={resetting}>{t('common.cancel')}</AlertDialogCancel>
                  {/* 普通 Button 而非 AlertDialogAction：后者点击默认关闭弹窗，会「还没生效就消失」 */}
                  <Button disabled={resetting} onClick={onResetConfirm}>
                    {resetting && <Icon name="loader" className="size-4 animate-spin" />}
                    {t('users.restoreSSOSync')}
                  </Button>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}
          <AlertDialog open={isOpen} onOpenChange={(o) => !o && !toggling && onClose()}>
            <Button size="sm" variant="destructive" onClick={() => onToggle(user)} className="w-28">
              <Icon name="shield-minus" className="size-3.5" />
              {t('users.demote')}
            </Button>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{t('users.demoteConfirmTitle', { name: user.name })}</AlertDialogTitle>
                <AlertDialogDescription>{t('users.demoteConfirmDesc')}</AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel disabled={toggling}>{t('common.cancel')}</AlertDialogCancel>
                {/* 普通 Button 而非 AlertDialogAction：后者点击默认关闭弹窗，会「还没生效就消失」 */}
                <Button variant="destructive" disabled={toggling} onClick={onConfirm}>
                  {toggling && <Icon name="loader" className="size-4 animate-spin" />}
                  {t('users.demote')}
                </Button>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      ) : (
        <AlertDialog open={isOpen} onOpenChange={(o) => !o && !toggling && onClose()}>
          <Button size="sm" variant="default" onClick={() => onToggle(user)} className="w-28">
            <Icon name="shield-plus" className="size-3.5" />
            {t('users.promote')}
          </Button>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{t('users.promoteConfirmTitle', { name: user.name })}</AlertDialogTitle>
              <AlertDialogDescription>{t('users.promoteConfirmDesc')}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={toggling}>{t('common.cancel')}</AlertDialogCancel>
              {/* 普通 Button 而非 AlertDialogAction：后者点击默认关闭弹窗，会「还没生效就消失」 */}
              <Button disabled={toggling} onClick={onConfirm}>
                {toggling && <Icon name="loader" className="size-4 animate-spin" />}
                {t('users.promote')}
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </div>
  )
})

/**
 * 用户管理（管理员后台）
 *
 * 数据来自 /api/admin/users（服务端搜索 + 只看管理员过滤 + 全量口径统计）：
 * - 顶部三卡统计来自服务端 stats（全量口径，不受搜索/角色过滤影响）
 * - 关键词搜索（300ms 防抖）与「只看管理员」均走服务端 query
 * - 最近登录列点击切换升/降序（服务端排序）；行内一键复制邮箱
 * - 行内「设为管理员 / 移除管理员」走角色管理接口（PUT /api/admin/users/{email}/role），
 *   均带 AlertDialog 二次确认（超级管理员不可降级）
 * - 无限下拉（服务端分页，滚动触底追加下一页）；搜索/筛选/排序变化自动回顶部
 * - ⚠️ 系统无封号/禁用状态，本页不提供该能力
 */
export function UserManager() {
  const { t } = useTranslation()
  // 当前登录用户是否超管：普通管理员进入本页只读（后端 ToggleAdmin 同样有超管门卫，前端仅隐藏入口）
  const canManage = useAuth().user?.isSuperAdmin === true
  const [users, setUsers] = useState<UserModel[]>([])
  const [count, setCount] = useState(0)
  const [stats, setStats] = useState<UserStats>({ total: 0, admins: 0, regular: 0 })
  const [keyword, setKeyword] = useState('')
  // 防抖后的关键词：避免每次击键都打后端
  const [debouncedKeyword, setDebouncedKeyword] = useState('')
  const [adminOnly, setAdminOnly] = useState(false)
  const [page, setPage] = useState(1)
  const [initialLoading, setInitialLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  // 渐入版本号：整页取数成功 +1，RefreshFade 依 key 重挂载列表重播渐入（追加页不重播，避免闪断）
  const [version, setVersion] = useState(0)
  // 最近登录列排序方向（默认降序 = 最近登录在前，点击表头切换升/降），作为 sort 参数交给服务端
  const [loginSort, setLoginSort] = useState<'asc' | 'desc'>('desc')
  // 提权/降权确认弹窗：null=关闭；非空=确认变更该用户管理员角色（受控弹窗，请求成功前不关闭）
  const [toggleTarget, setToggleTarget] = useState<UserModel | null>(null)
  // 角色变更进行中（防重复点击；进行中禁止关闭弹窗，失败保持打开可重试）
  const [toggling, setToggling] = useState(false)
  // 恢复 SSO 同步确认弹窗：null=关闭；非空=确认解除该用户后台手动接管（受控弹窗，请求成功前不关闭）
  const [resetTarget, setResetTarget] = useState<UserModel | null>(null)
  // 解除接管进行中（防重复点击；进行中禁止关闭弹窗，失败保持打开可重试）
  const [resetting, setResetting] = useState(false)
  // 列表滚动容器（IntersectionObserver 的 root）与底部哨兵
  const scrollRef = useRef<HTMLDivElement>(null)
  const sentinelRef = useRef<HTMLDivElement>(null)
  // 请求忙碌锁：阻止触底加载在请求中/刷新时连环翻页
  const busyRef = useRef(false)
  const refreshingRef = useRef(false)
  // 请求去重基准：记录上次真正发起的 (过滤条件, 页码)，跳过「新条件×旧页码」过期请求与重复请求。
  // 初始值必须「不可能等于」首屏状态——null ≠ filterKey、0 ≠ 1 的 page，
  // 否则首屏挂载被去重判定吞掉，列表永远不请求（骨架屏常驻）
  const lastFilterKeyRef = useRef<string | null>(null)
  const lastPageRef = useRef(0)
  // 最新过滤条件快照（每次渲染同步）：校验在途响应落地时是否仍属当前过滤。
  // 拦截「旧筛选的追加页晚到、被 append 进新筛选结果」的跨条件污染——追加页不可取消，
  // 改落地校验丢弃过期响应
  const filterKeyRef = useRef('')
  const hasMore = users.length < count
  // 有旧数据时的重取 loading：首载 users 为空走骨架，不进遮罩；搜索/筛选/刷新重取则遮罩旧列表
  const refetching = initialLoading && users.length > 0

  // 关键词防抖：输入停顿 300ms 后才把新关键词交给 fetch effect
  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedKeyword(keyword), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [keyword])

  // 过滤条件指纹：搜索 / 只看管理员 / 排序方向任一变化都视为新的过滤条件
  //（排序切换走服务端重新拉取，故也纳入指纹）
  const filterKey = `${debouncedKeyword}|${adminOnly}|${loginSort}`
  // 每渲染同步最新过滤条件（供 fetchList 落地校验，见 filterKeyRef 声明注释）
  filterKeyRef.current = filterKey

  // 过滤条件变化 → 回到顶部 + 重置第 1 页（由 page 变化驱动下面的 fetch）
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: 0 })
    setPage(1)
  }, [filterKey])

  /** 拉取第 p 页用户清单：append=true 追加到列表尾（无限滚动），否则整体替换（首屏/搜索/刷新）。
   *  搜索/只看管理员/最近登录排序交给服务端，stats 为全量口径统计。 */
  const fetchList = useCallback(
    async (p: number, append: boolean) => {
      busyRef.current = true
      if (append) setLoadingMore(true)
      else setInitialLoading(true)
      try {
        const { data, error: err } = await api.GET(API.adminUsers, {
          params: {
            query: {
              page: p,
              pageSize: PAGE_SIZE,
              search: debouncedKeyword.trim() || undefined,
              role: adminOnly ? 'admin' : undefined,
              sort: loginSort,
            },
          },
        })
        if (err) throw new Error(err.message ?? String(err))
        if (!data) return
        // 在途响应落地时若过滤条件已切换（旧筛选的追加页晚到）→ 丢弃，不 append 进新筛选结果
        if (filterKeyRef.current !== filterKey) return
        setError('')
        setUsers((prev) => (append ? [...prev, ...data.items] : data.items))
        setCount(data.count)
        setStats(data.stats)
        // 仅整页取数（非追加）重播渐入：追加页复用已有动画，避免整表重挂闪断
        if (!append) setVersion((v) => v + 1)
      } catch (e) {
        // 追加页失败不清空已加载列表，toast 提示即可；首屏/整页失败落 error 空态
        if (append) toast.error(e instanceof Error ? e.message : String(e))
        else setError(e instanceof Error ? e.message : String(e))
      } finally {
        busyRef.current = false
        setInitialLoading(false)
        setLoadingMore(false)
      }
    },
    [debouncedKeyword, adminOnly, loginSort],
  )

  useEffect(() => {
    // 刷新已直接拉取第 1 页，跳过 setPage(1) 触发的重复请求
    if (page === 1 && refreshingRef.current) return
    // 过滤条件刚变、page 尚未重置到 1（filterKey effect 已 setPage(1)，本 commit 里拿到的还是旧 page）：
    // 跳过这次「新条件×旧页码」的过期请求，真正的第 1 页由 setPage(1) 触发的下一次 effect 承担
    if (lastFilterKeyRef.current !== filterKey && page !== 1) return
    // 同 (filterKey, page) 去重：刷新完成后的 effect 重跑不会重复拉第 1 页
    if (lastFilterKeyRef.current === filterKey && lastPageRef.current === page) return
    lastFilterKeyRef.current = filterKey
    lastPageRef.current = page
    void fetchList(page, page > 1)
  }, [page, filterKey, fetchList])

  /** 手动刷新：回到第 1 页拉最新一版用户清单（保留当前搜索/筛选/排序）。
   *  useCallback 稳定引用，供行 memo 与 toggleAdmin 复用。
   *  重置期间同步锁 busyRef，阻止触底加载在刷新时连环翻页。 */
  const refresh = useCallback(async () => {
    if (refreshingRef.current) return
    refreshingRef.current = true
    setRefreshing(true)
    busyRef.current = true
    // 回到顶部：让哨兵离开视口，刷新后不会自动加载后续页
    if (scrollRef.current) scrollRef.current.scrollTop = 0
    setPage(1)
    // 预标记本次刷新要拉的 (filterKey, 1)，刷新完成后的 effect 重跑据此去重
    lastFilterKeyRef.current = filterKey
    lastPageRef.current = 1
    await fetchList(1, false)
    busyRef.current = false
    refreshingRef.current = false
    setRefreshing(false)
  }, [filterKey, fetchList])

  // 底部哨兵进入列表容器视口（含 300px 预加载区）且非忙碌/非刷新时 → 翻下一页（服务端分页）
  useEffect(() => {
    const sentinel = sentinelRef.current
    const root = scrollRef.current
    if (!sentinel || !root) return
    const io = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !busyRef.current && !refreshing) {
          setPage((p) => p + 1)
        }
      },
      { root, rootMargin: '300px' },
    )
    io.observe(sentinel)
    return () => io.disconnect()
  }, [hasMore, initialLoading, loadingMore, refreshing])

  // 弹窗目标/进行中状态同步 ref：供 useCallback 稳定读取最新值，避免 handler 每渲染重建令行 memo 失效
  const toggleTargetRef = useRef<UserModel | null>(null)
  const togglingRef = useRef(false)
  const resetTargetRef = useRef<UserModel | null>(null)
  const resettingRef = useRef(false)
  toggleTargetRef.current = toggleTarget
  togglingRef.current = toggling
  resetTargetRef.current = resetTarget
  resettingRef.current = resetting

  /** 添加/删除管理员：调角色管理接口（超级管理员不可降级由服务端保证）。
   *  受控弹窗：请求成功才关闭，失败保持打开（toast 报错）供重试，杜绝「还没生效弹窗先消失」。
   *  useCallback 稳定引用：行 memo 的 onConfirm prop 不因父级状态抖动而重建。 */
  const toggleAdmin = useCallback(async () => {
    const target = toggleTargetRef.current
    if (!target || togglingRef.current) return
    const makeAdmin = !target.roles.includes('admin')
    setToggling(true)
    try {
      const { error: err } = await api.PUT(API.adminUserRole, {
        params: { path: { email: target.email } },
        body: { email: target.email, admin: makeAdmin },
      })
      if (err) throw new Error(err.message ?? String(err))
      toast.success(t(makeAdmin ? 'users.promoteSuccess' : 'users.demoteSuccess'))
      setToggleTarget(null)
      void refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setToggling(false)
    }
  }, [refresh, t])

  /** 恢复 SSO 同步：解除后台手动接管（roles_override 置回 false），该用户从下一次登录起
   *  恢复按 SSO 角色同步（手动设置的权限将被覆盖）。受控弹窗：请求成功才关闭，失败保持
   *  打开（toast 报错）供重试；成功后行内角色来源徽标随 refresh 回到「来自 SSO」。 */
  const resetRolesOverride = useCallback(async () => {
    const target = resetTargetRef.current
    if (!target || resettingRef.current) return
    setResetting(true)
    try {
      const { error: err } = await api.PUT(API.adminUserRolesOverride, {
        params: { path: { email: target.email } },
      })
      if (err) throw new Error(err.message ?? String(err))
      toast.success(t('users.restoreSSOSyncSuccess'))
      setResetTarget(null)
      void refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setResetting(false)
    }
  }, [refresh, t])

  /** 复制用户邮箱（邀请成员/联系用户），成功 toast 反馈（useCallback 稳定引用） */
  const copyEmail = useCallback(async (email: string) => {
    const ok = await copyText(email)
    if (ok) toast.success(t('users.copyEmail'))
    else toast.error(t('common.retry'))
  }, [t])

  /** 打开提权/降权确认弹窗（setToggleTarget 恒稳定，useCallback 仅为行 memo 的 prop 引用一致） */
  const handleToggle = useCallback((u: UserModel) => setToggleTarget(u), [])

  /** 关闭确认弹窗（受控弹窗 onOpenChange(false) → 关） */
  const handleClose = useCallback(() => setToggleTarget(null), [])

  /** 打开恢复 SSO 同步确认弹窗（setResetTarget 恒稳定，useCallback 仅为行 memo 的 prop 引用一致） */
  const handleReset = useCallback((u: UserModel) => setResetTarget(u), [])

  /** 关闭恢复确认弹窗（受控弹窗 onOpenChange(false) → 关） */
  const handleResetClose = useCallback(() => setResetTarget(null), [])

  return (
    <div className="flex h-full flex-col gap-4">
      {/* 页头：标题 + 搜索 + 刷新 */}
      <div className="flex shrink-0 flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <h2 className="text-[16px] font-semibold text-ink">{t('users.title')}</h2>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {/* 搜索统一放标题行右上角（对齐 Events/治理/工作台） */}
          <SearchInput
            value={keyword}
            onChange={setKeyword}
            placeholder={t('users.searchPlaceholder')}
            className="w-72"
          />
          <Button variant="outline" size="sm" onClick={refresh} disabled={refreshing}>
            {refreshing ? (
              <Icon name="loader" className="size-4 animate-spin" />
            ) : (
              <Icon name="refresh" className="size-4" />
            )}
            {t('common.refresh')}
          </Button>
        </div>
      </div>

      {/* 顶部三卡统计（服务端全量口径） */}
      <div className="grid shrink-0 grid-cols-1 gap-3 sm:grid-cols-3">
        <StatCard label={t('users.total')} value={stats.total} icon="users" tone="mute" />
        <StatCard label={t('users.admins')} value={stats.admins} icon="shield" tone="accent" />
        <StatCard label={t('users.regular')} value={stats.regular} icon="user" tone="ok" />
      </div>

      {/* 工具栏：只看管理员 + 结果计数（SearchInput 内置 ⌘K 聚焦快捷键，已上移到标题行） */}
      <div className="flex shrink-0 flex-wrap items-center gap-3">
        <Button
          size="sm"
          variant={adminOnly ? 'default' : 'outline'}
          aria-pressed={adminOnly}
          onClick={() => setAdminOnly((v) => !v)}
        >
          <Icon name="shield" className="size-3.5" />
          {t('users.filterAdmin')}
        </Button>
        <span className="text-[12px] text-faint">{t('users.resultCount', { count })}</span>
      </div>

      {/* 用户列表：固定表头 + 内部滚动容器（无限下拉 root） */}
      <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-line bg-surface">
        <div className="hidden grid-cols-[minmax(0,2fr)_minmax(0,1.5fr)_7rem_15rem] items-center gap-2 border-b border-line px-4 py-2 text-[11px] font-medium text-faint lg:grid">
          <span>{t('users.user')}</span>
          <span>{t('users.role')}</span>
          <button
            type="button"
            onClick={() => setLoginSort((v) => (v === 'desc' ? 'asc' : 'desc'))}
            className="flex w-fit items-center gap-0.5 hover:text-ink"
            title={
              loginSort === 'desc' ? t('users.sortDesc') : t('users.sortAsc')
            }
          >
            {t('users.lastLogin')}
            <Icon
              name={loginSort === 'desc' ? 'chevron-down' : 'chevron-up'}
              className="size-3 opacity-70"
            />
          </button>
          <span>{t('users.action')}</span>
        </div>

        <div ref={scrollRef} className="relative min-h-0 flex-1 overflow-y-auto">
        {/* 内容区：重取遮罩时降透明度并禁止交互，保持旧帧不闪断（首载骨架不遮罩） */}
        <div className={refetching ? 'pointer-events-none opacity-40' : undefined}>
        {initialLoading && users.length === 0 ? (
          <SkeletonList count={8} bare />
        ) : error && users.length === 0 ? (
          <Empty icon="alert" text={error} />
        ) : users.length === 0 ? (
          <Empty
            icon="users"
            text={keyword ? t('users.searchEmpty', { kw: keyword.trim() }) : t('common.empty')}
          />
        ) : (
          <RefreshFade version={version}>
          {users.map((u) => (
            <UserRow
              key={u.email}
              user={u}
              isOpen={toggleTarget?.email === u.email}
              toggling={toggling}
              resetOpen={resetTarget?.email === u.email}
              resetting={resetting}
              canManage={canManage}
              onToggle={handleToggle}
              onClose={handleClose}
              onConfirm={toggleAdmin}
              onReset={handleReset}
              onResetClose={handleResetClose}
              onResetConfirm={resetRolesOverride}
              onCopyEmail={copyEmail}
            />
          ))}
          </RefreshFade>
        )}
        {/* 无限下拉哨兵：进入视口翻下一页；无更多数据时显示到底提示 */}
        <div ref={sentinelRef} className="flex h-10 items-center justify-center gap-2">
          {loadingMore ? (
            <span className="flex items-center gap-1.5 text-[12px] text-mute">
              <Icon name="loader" className="animate-spin text-[13px]" />
              {t('common.loadingMore')}
            </span>
          ) : users.length > 0 && !hasMore ? (
            <span className="text-[11px] text-faint">{t('common.noMore')}</span>
          ) : null}
        </div>
        </div>
        {/* 重取遮罩：有旧数据时的 loading（搜索/只看管理员/排序/刷新），居中 spinner */}
        {refetching && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-surface/50">
            <Icon name="loader" className="size-5 animate-spin text-faint" />
          </div>
        )}
        </div>
      </section>
    </div>
  )
}
