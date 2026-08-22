import { useTranslation } from 'react-i18next'
import { Icon, type IconName } from '@/components/Icons'

/** deployStatus 枚举字面量（schema 仅声明文件，运行时不能 import 枚举） */
type DeployStatus = 'StatusUnknown' | 'StatusDeploying' | 'StatusDeployed' | 'StatusFailed'

type DeployStatusMeta = {
  icon: IconName
  className: string
  i18nKey:
    | 'project.statusUnknown'
    | 'project.statusDeploying'
    | 'project.statusDeployed'
    | 'project.statusFailed'
}

/** 状态 → 语义色 icon（对齐旧版 DeployStatus 的 two-tone 语义色：Unknown 琥珀问号 / Deploying 时钟 / Deployed 绿勾 / Failed 红叉） */
const STATUS_META: Record<DeployStatus, DeployStatusMeta> = {
  StatusUnknown: { icon: 'circle-question', className: 'text-warn', i18nKey: 'project.statusUnknown' },
  StatusDeploying: { icon: 'clock', className: 'text-info', i18nKey: 'project.statusDeploying' },
  StatusDeployed: { icon: 'circle-check', className: 'text-ok', i18nKey: 'project.statusDeployed' },
  StatusFailed: { icon: 'circle-x', className: 'text-err', i18nKey: 'project.statusFailed' },
}

/**
 * 部署状态紧凑图标（旧版 icon 表达，替代占空间的文字 Tag）。
 * 颜色 + aria-label + hover title 三通道传达状态，避免"仅靠颜色"（ui-ux-pro-max Color Only 规则）。
 * 状态图标对屏幕阅读器有意义，故覆盖自定义 Icon 默认的 aria-hidden，转 role="img" 供读出。
 */
export function DeployStatusIcon({ status }: { status: DeployStatus }) {
  const { t } = useTranslation()
  const { icon, className, i18nKey } = STATUS_META[status]
  const label = t(i18nKey)
  return (
    <span className="flex shrink-0 items-center" title={label}>
      <Icon name={icon} className={`size-4 ${className}`} role="img" aria-hidden={false} aria-label={label} />
    </span>
  )
}
