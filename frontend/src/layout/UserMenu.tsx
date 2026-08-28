import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/shadcn/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/shadcn/avatar'
import { Icon, type IconName } from '@/components/Icons'
import { useAuth } from '@/features/auth/AuthProvider'
import type { TKey } from '@/i18n/keys'

/** 用户下拉菜单项：图标 + 词条键 + 路由目标 */
const NAV_ITEMS: { icon: IconName; labelKey: TKey; to: string }[] = [
  { icon: 'grid', labelKey: 'nav.workbench', to: '/' },
  { icon: 'pulse', labelKey: 'nav.events', to: '/events' },
  { icon: 'key', labelKey: 'nav.token', to: '/tokens' },
  // 管理后台（mars_admin 门控）：仓库管理/集群资源等后台页统一收进后台侧栏，下拉只留总入口
  { icon: 'shield', labelKey: 'nav.admin', to: '/admin' },
]

/**
 * 用户下拉（shadcn DropdownMenu + Avatar）：Avatar + 用户名触发，
 * 菜单承载导航入口（工作台/事件/令牌/管理后台）与接口文档外链、登出；
 * 仓库管理/集群资源等后台页统一收进管理后台侧栏，此处只留总入口。
 * Radix 原生：portal 定位 + 焦点圈定 + 键盘导航 + 选中即关。
 */
export function UserMenu() {
  const { t } = useTranslation()
  const { user, signout } = useAuth()

  // 管理后台入口仅 mars_admin 可见；事件已对普通用户开放（后端按 operator_email 归属过滤，各自只看自己的事件）
  const isAdmin = user?.roles.includes('mars_admin') ?? false
  const navItems = NAV_ITEMS.filter((it) => (it.to === '/admin' ? isAdmin : true))

  return (
    <DropdownMenu>
      {/* 设计上不需要 focus-visible 焦点环：显式 outline-none 压掉全局 outline、不设 ring——
          Radix 关闭下拉会把焦点还给 trigger，焦点环会硬切弹出（闪烁）；对齐 ThemeSwitcher trigger */}
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-primary-foreground transition-colors hover:bg-primary-foreground/10 focus-visible:outline-none"
        >
          <Avatar className="size-[22px]">
            {user?.avatar && <AvatarImage src={user.avatar} alt={user?.name ?? ''} />}
            <AvatarFallback className="bg-primary-foreground/15 text-primary-foreground">
              <Icon name="user" className="size-3.5" />
            </AvatarFallback>
          </Avatar>
          <span className="max-w-[22vw] truncate text-[13px]">{user?.name ?? ''}</span>
          <Icon name="chevron-down" className="text-[12px] opacity-70" />
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="z-[9998] w-56">
        {/* 用户信息 */}
        <DropdownMenuLabel className="px-2.5 py-2">
          <div className="text-[13px] font-medium text-ink">{user?.name ?? '-'}</div>
          <div className="truncate text-[11px] text-faint">{user?.email ?? ''}</div>
        </DropdownMenuLabel>

        <DropdownMenuSeparator />

        {/* 导航入口 */}
        {navItems.map((it) => (
          <DropdownMenuItem key={it.to} asChild className="cursor-pointer">
            <Link to={it.to}>
              <Icon name={it.icon} className="size-4" />
              {t(it.labelKey)}
            </Link>
          </DropdownMenuItem>
        ))}

        {/* 接口文档（外链） */}
        <DropdownMenuItem asChild className="cursor-pointer">
          <a href="/docs/index.html" target="_blank" rel="noreferrer">
            <Icon name="book" className="size-4" />
            {t('header.docs')}
          </a>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        {/* 登出 */}
        <DropdownMenuItem variant="destructive" className="cursor-pointer" onClick={signout}>
          <Icon name="power" className="size-4" />
          {t('auth.signOut')}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
