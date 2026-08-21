import { useTranslation } from 'react-i18next'
import { CircleCheck, CircleQuestionMark, CircleX, Clock } from 'lucide-react'

/** deployStatus 枚举字面量（schema 仅声明文件，运行时不能 import 枚举） */
type DeployStatus = 'StatusUnknown' | 'StatusDeploying' | 'StatusDeployed' | 'StatusFailed'

type DeployStatusMeta = {
  Icon: typeof CircleCheck
  className: string
  i18nKey:
    | 'project.statusUnknown'
    | 'project.statusDeploying'
    | 'project.statusDeployed'
    | 'project.statusFailed'
}

/** 状态 → 语义色 icon（对齐旧版 DeployStatus 的 two-tone 语义色：Unknown 琥珀问号 / Deploying 时钟 / Deployed 绿勾 / Failed 红叉） */
const STATUS_META: Record<DeployStatus, DeployStatusMeta> = {
  StatusUnknown: { Icon: CircleQuestionMark, className: 'text-warn', i18nKey: 'project.statusUnknown' },
  StatusDeploying: { Icon: Clock, className: 'text-info', i18nKey: 'project.statusDeploying' },
  StatusDeployed: { Icon: CircleCheck, className: 'text-ok', i18nKey: 'project.statusDeployed' },
  StatusFailed: { Icon: CircleX, className: 'text-err', i18nKey: 'project.statusFailed' },
}

/**
 * 部署状态紧凑图标（旧版 icon 表达，替代占空间的文字 Tag）。
 * 颜色 + aria-label + hover title 三通道传达状态，避免"仅靠颜色"（ui-ux-pro-max Color Only 规则）。
 */
export function DeployStatusIcon({ status }: { status: DeployStatus }) {
  const { t } = useTranslation()
  const { Icon, className, i18nKey } = STATUS_META[status]
  const label = t(i18nKey)
  return (
    <span className="flex shrink-0 items-center" title={label}>
      <Icon size={16} className={className} aria-label={label} />
    </span>
  )
}
