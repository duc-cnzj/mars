import { type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon, type IconName } from '@/components/Icons'

/**
 * 空态占位：icon 按页面语义传入（默认数据库图标），避免所有空态都长一个样。
 * action 插槽放引导 CTA（如「新建空间」「清除搜索」），空态不止是说明，还要给下一步。
 */
export function Empty({
  text,
  icon = 'database',
  action,
}: {
  text?: string
  icon?: IconName
  action?: ReactNode
}) {
  const { t } = useTranslation()
  const label = text ?? t('common.empty')
  return (
    <div
      role="status"
      aria-label={label}
      className="flex flex-col items-center justify-center gap-2 py-12 text-faint"
    >
      <Icon name={icon} className="text-[28px]" />
      <span className="text-[13px]">{label}</span>
      {action && <div className="mt-1">{action}</div>}
    </div>
  )
}
