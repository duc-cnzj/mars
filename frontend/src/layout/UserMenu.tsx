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
import type zhCN from '@/i18n/locales/zh-CN'

/** 词条扁平 key 联合类型，保证导航键在词条内（同旧 Sidebar 的 FlatKeys 手法） */
type TKey = FlatKeys<typeof zhCN>

/** 将嵌套词条对象递归展开为点号路径的联合类型 */
type FlatKeys<T, P extends string = ''> = {
  [K in keyof T & string]: T[K] extends Record<string, unknown>
    ? FlatKeys<T[K], `${P}${K}.`>
    : `${P}${K}`
}[keyof T & string]

/** 用户下拉菜单项：图标 + 词条键 + 路由目标 */
const NAV_ITEMS: { icon: IconName; labelKey: TKey; to: string }[] = [
  { icon: 'grid', labelKey: 'nav.workbench', to: '/' },
  { icon: 'pulse', labelKey: 'nav.events', to: '/events' },
  { icon: 'repo', labelKey: 'nav.repo', to: '/repos' },
  { icon: 'key', labelKey: 'nav.token', to: '/tokens' },
]

/**
 * 用户下拉（shadcn DropdownMenu + Avatar）：Avatar + 用户名触发，
 * 菜单承载全部导航入口（工作台/事件/仓库/令牌/接口文档）与登出。
 * Radix 原生：portal 定位 + 焦点圈定 + 键盘导航 + 选中即关。
 */
export function UserMenu() {
  const { t } = useTranslation()
  const { user, signout } = useAuth()

  // 仓库仅 mars_admin 可见（旧版管理员门控）；事件已对普通用户开放（后端按 operator_email 归属过滤，各自只看自己的事件）
  const isAdmin = user?.roles.includes('mars_admin') ?? false
  const navItems = NAV_ITEMS.filter((it) => (it.to === '/repos' ? isAdmin : true))

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-primary-foreground transition-colors hover:bg-primary-foreground/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-foreground/40"
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
            <Icon name="gear" className="size-4" />
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
