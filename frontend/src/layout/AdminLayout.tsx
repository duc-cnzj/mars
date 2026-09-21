import { useCallback, useState } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/features/auth/AuthProvider'
import { Icon, type IconName } from '@/components/Icons'
import { Tag } from '@/components/ui'
import type { TKey } from '@/i18n/keys'

/**
 * 管理后台侧栏导航项：图标 + 词条键 + 路由目标。
 * 平台级治理视角，按职责分组自上而下：
 * 集群总览（资源治理）→ 核心业务前置（仓库 → 事件 → 令牌）→
 * 资源治理（项目活跃度 → 命名空间 → 空间资源）→ 账号与系统（用户 → 误删恢复 → 系统设置）；
 * 事件/令牌/用户复用普通入口组件（真实 API），操作审计已合并进事件页（管理员即看全平台）。
 */
const ADMIN_NAV: { icon: IconName; labelKey: TKey; to: string; superOnly?: boolean }[] = [
  { icon: 'cluster', labelKey: 'nav.cluster', to: '/admin/cluster' },
  { icon: 'repo', labelKey: 'nav.repo', to: '/admin/repos' },
  { icon: 'pulse', labelKey: 'nav.events', to: '/admin/events' },
  { icon: 'key', labelKey: 'nav.token', to: '/admin/tokens' },
  { icon: 'boxes', labelKey: 'nav.projects', to: '/admin/projects' },
  { icon: 'namespace', labelKey: 'nav.namespaces', to: '/admin/namespaces' },
  { icon: 'gauge', labelKey: 'nav.resources', to: '/admin/resources' },
  { icon: 'users', labelKey: 'nav.users', to: '/admin/users' },
  // 误删恢复仅超级管理员可见：会重建集群侧骨架（k8s namespace + docker secret），
  // 后端 namespace.Authorize 对 Restore 走 RequireSuperAdmin 强制名单。
  // 排在系统设置之前：排障（找回误删）比翻配置更常走，且同为超管专属项挨着更成组
  { icon: 'restore', labelKey: 'nav.restore', to: '/admin/restore', superOnly: true },
  // 系统设置仅超级管理员可见（后端 /api/admin/settings 已按 is_super_admin 门禁）
  { icon: 'gear', labelKey: 'nav.settings', to: '/admin/settings', superOnly: true },
]

/**
 * 侧栏宽度（要调就改这一个地方）：展开自适应 / 收起图标栏。
 *
 * 展开态不用固定宽度：中英文导航文案长度差很大（「空间治理」4 字 vs
 * "Repository Management" 21 字符），固定 13rem 对两边都不合适——英文被
 * truncate 成省略号，中文则白白空出一大截。
 *
 * w-max 让侧栏贴合最长一项；再配一个下限垫高，避免短文案语言（中文实测内容仅
 * 156px）贴得过紧。下限取 11.5rem：实测内容 156 ↔ 原固定宽 208 的中点约 182，
 * 取标准步进 184，即在「不拥挤」与「不空旷」之间居中（调宽度只改这一个数）。
 * 上限 16rem 给极端长文案兜底（超限仍由标签上的 truncate 收尾，不会撑破布局）。
 *
 * 注：注释里不要写出宽度类名的字面量——Tailwind v4 会扫描注释文本，
 * 注释里出现的类名会被一并编进产物（实测踩过一次，产物里多出一条无引用的宽度类）。
 */
const SIDEBAR_EXPANDED = 'w-max min-w-46 max-w-64'
const SIDEBAR_COLLAPSED = 'w-14'
/** 收起态持久化键：跨刷新保留用户偏好（同 useTheme 的 localStorage 模式） */
const SIDEBAR_COLLAPSED_KEY = 'mars.admin.sidebarCollapsed'

/** 读取收起态：localStorage 是系统边界（隐私模式/SSR 会抛），失败一律回展开态 */
function readSidebarCollapsed(): boolean {
  try {
    return localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === '1'
  } catch {
    return false
  }
}

/**
 * 管理后台布局：左侧导航 + 右侧主体（标准后台控制台形态）。
 * 固定视口高度容器（对齐 events/tokens 的 h-[calc(100dvh_-_100px)] 约定；
 * AppLayout 在 /admin/* 下隐藏底栏、文档恰好 100dvh，因此侧栏导航与面包屑不随页面漂移），
 * 侧栏常驻、可折叠成图标栏（收起态持久化）；右侧顶部面包屑固定、主体内部滚动。
 */
export function AdminLayout() {
  const { t } = useTranslation()
  const location = useLocation()
  const { user } = useAuth()
  const [collapsed, setCollapsed] = useState(readSidebarCollapsed)

  // 当前登录用户可见的导航项：superOnly 项仅内置超管可见（普通管理员不显示系统设置）
  const nav = ADMIN_NAV.filter((it) => !it.superOnly || user?.isSuperAdmin)

  // 当前页导航项（按路由前缀匹配），驱动面包屑第二级
  const current = nav.find((it) => location.pathname.startsWith(it.to))

  /** 切换折叠态并持久化（写入失败忽略——只影响下次刷新是否恢复） */
  const toggle = useCallback(() => {
    setCollapsed((c) => {
      const next = !c
      try {
        localStorage.setItem(SIDEBAR_COLLAPSED_KEY, next ? '1' : '0')
      } catch {
        /* 忽略存储失败 */
      }
      return next
    })
  }, [])

  return (
    <div className="flex h-[calc(100dvh_-_100px)] gap-4">
      {/* 左侧导航：后台标识 + 功能项 + 折叠开关。
          ⚠️ 这里刻意不加 transition-[width]：展开态宽度是 max-content（内在尺寸关键字），与 3.5rem
          之间不可插值，浏览器对不可插值的过渡是「不启动」（直接硬跳），留着只会让人误以为有动画
          （原来两个固定宽度之间才过渡得起来）。要保留动画就得放弃自适应宽度——那是本次改动的反面
          目标，故取「无动画但中英文文案都不被截」。 */}
      <aside
        className={`flex shrink-0 flex-col rounded-lg border border-line bg-surface p-2 ${
          collapsed ? SIDEBAR_COLLAPSED : SIDEBAR_EXPANDED
        }`}
      >
        <div className={`mb-2 flex items-center gap-2 py-2 ${collapsed ? 'justify-center px-0' : 'px-2'}`}>
          <Icon name="shield" className="size-4 shrink-0 text-primary" />
          {!collapsed && <span className="truncate text-[13px] font-medium text-ink">{t('nav.admin')}</span>}
        </div>
        <nav className="flex min-h-0 flex-1 flex-col gap-0.5">
          {nav.map((it) => (
            <NavLink
              key={it.to}
              to={it.to}
              end
              // 收起时只有图标，title 兜底可悬停读名；超管专属项把标识一并并入 title，
              // 否则收起态下「超管」角标被隐藏，该标识就彻底看不见了
              title={collapsed ? (it.superOnly ? `${t(it.labelKey)} · ${t('nav.superOnly')}` : t(it.labelKey)) : undefined}
              className={({ isActive }) =>
                `flex items-center gap-2 rounded-md py-1.5 text-[13px] transition-colors ${
                  collapsed ? 'justify-center px-0' : 'px-2'
                } ${
                  isActive ? 'bg-primary-soft text-primary' : 'text-mute hover:bg-bg hover:text-ink'
                }`
              }
            >
              <Icon name={it.icon} className="size-4 shrink-0" />
              {!collapsed && <span className="truncate">{t(it.labelKey)}</span>}
              {/* 超管专属角标：ml-auto 右对齐使不同长度的标签下角标仍成列；
                  shrink-0 保证窄宽度下优先压缩标签而非挤出角标。
                  tone 用 accent（=primary）而非 warn：表达的是「特权项」不是「有风险」，
                  且随 8 套主题换肤；与恢复页页头同一语义标识保持同色 */}
              {!collapsed && it.superOnly && (
                <Tag tone="accent" dot={false} className="ml-auto shrink-0 px-1.5">
                  {t('nav.superOnly')}
                </Tag>
              )}
            </NavLink>
          ))}
        </nav>
        {/* 折叠开关：展开时显示文案，收起时图标居中 */}
        <button
          type="button"
          onClick={toggle}
          title={collapsed ? t('admin.expandSidebar') : t('admin.collapseSidebar')}
          className={`mt-2 flex items-center gap-2 rounded-md py-1.5 text-[13px] text-mute transition-colors hover:bg-bg hover:text-ink ${
            collapsed ? 'justify-center px-0' : 'px-2'
          }`}
        >
          <Icon name={collapsed ? 'expand' : 'collapse'} className="size-4 shrink-0" />
          {!collapsed && <span>{t('admin.collapseSidebar')}</span>}
        </button>
      </aside>

      {/* 右侧主体：面包屑固定 + 内容内部滚动；main 本身无内边距（面包屑贴边紧凑），
          padding 下沉到内容区滚动容器（main > div p-4）——面包屑顶满、内容区缩进呼吸感，
          左缘再靠父级 gap-4 与侧栏隔开 */}
      <main className="flex min-w-0 flex-1 flex-col">
        <nav className="mb-3 flex shrink-0 items-center gap-1 text-[12px]" aria-label="breadcrumb">
          <span className="text-faint">{t('nav.admin')}</span>
          {current && (
            <>
              <Icon name="chevron-right" className="size-3 text-faint" />
              <span className="text-mute">{t(current.labelKey)}</span>
            </>
          )}
        </nav>
        {/* 内容区缩进 p-4（面包屑保持贴边，紧凑头部 + 内容呼吸感）；
            padding 加在滚动容器上：滚动到底内容停在 padding 之上不贴底 */}
        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
