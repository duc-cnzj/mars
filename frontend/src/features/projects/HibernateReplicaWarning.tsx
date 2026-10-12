import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'

/**
 * 休眠窗口内副本数被改成 >0 的醒目提示横幅（非阻断，只提示不拦截）。
 *
 * 由 TabEdit 在「项目已开启定时休眠、处于休眠态（hibernate_state==='sleeping'）且至少一个副本字段
 * 被设成 >0」时条件渲染——文案点破对账语义：窗口内会被自动缩容为 0（约 2 分钟内生效），但设的值
 * 不会丢、会记为窗口结束后的副本数；并列出休眠期间将被置 0 的具体字段，避免用户面对数量突变时误判。
 * 休眠开关统一放在页面操作行（TabEdit 常驻），本横幅只做提示、不含开关；关掉开关后 hibernate_enabled
 * 变 false，判据失效、横幅随之消失（见 hibernate.shouldWarnReplicaInHibernate）。
 */
export function HibernateReplicaWarning({ fields }: { fields: string[] }) {
  const { t } = useTranslation()
  return (
    <div
      role="alert"
      className="flex items-start gap-2 rounded-md border border-warn bg-warn-soft px-3 py-2 text-[13px] font-medium text-warn"
    >
      <Icon name="alert" className="mt-0.5 size-4 shrink-0" />
      <div className="min-w-0 space-y-1">
        <span className="block">{t('project.hibernateReplicaWarning')}</span>
        {fields.length > 0 && (
          <span className="block">
            {t('project.hibernateReplicaZeroFields')}
            {fields.join(', ')}
          </span>
        )}
      </div>
    </div>
  )
}
