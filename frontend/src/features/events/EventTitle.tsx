import { Icon } from '@/components/Icons'

/**
 * 事件弹窗统一标题：主题色用户名 chip（user 图标）+ 常规字重灰色消息。
 * 还原旧版 Drawer title（红用户名+图标+13px 消息）视觉，红色改为跟随主题的 primary 强调色。
 */
export function EventTitle({ username, message }: { username: string; message: string }) {
  return (
    <span
      className="flex items-center gap-2 text-left"
      style={{ fontSize: 13, fontWeight: 400, lineHeight: 1.6 }}
    >
      <span className="flex shrink-0 items-center gap-1 rounded bg-primary/10 px-1.5 py-0.5 text-primary">
        <Icon name="user" className="text-[12px]" />
        <span style={{ fontWeight: 500 }}>{username}</span>
      </span>
      <span className="min-w-0 flex-1 truncate text-mute">{message}</span>
    </span>
  )
}
