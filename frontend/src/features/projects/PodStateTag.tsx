import { memo } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, RefreshCw } from 'lucide-react'
import type { ReactNode } from 'react'
import type { components } from '../../api/schema'
import { shortContainerName } from './shortContainerName'
import { Icon } from '../../components/icons'

type StateContainer = components['schemas']['types.StateContainer']

/** 彩色小胶囊：对齐旧版 antd Tag（`color=xxx` 即底色的形态） */
function Pill({
  className,
  onClick,
  onCopy,
  copyLabel,
  children,
}: {
  className: string
  onClick?: () => void
  onCopy?: () => void
  copyLabel?: string
  children: ReactNode
}) {
  return (
    <span
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      className={`group relative inline-flex shrink-0 items-center gap-1 whitespace-nowrap rounded-md border border-transparent px-2 py-1 text-[11px] font-medium leading-none transition-[padding] ${onClick ? 'cursor-pointer' : ''} ${onCopy ? 'hover:pr-6' : ''} ${className}`}
    >
      {children}
      {onCopy && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation() // 不触发标签的选中点击
            onCopy()
          }}
          title={copyLabel}
          aria-label={copyLabel}
          className="pointer-events-none absolute right-1.5 top-1/2 -translate-y-1/2 opacity-0 transition-opacity group-hover:pointer-events-auto group-hover:opacity-100"
        >
          <Icon name="copy" className="size-3 shrink-0" />
        </button>
      )}
    </span>
  )
}

/**
 * 容器状态标签（对齐旧版 PodStateTag 行为，非就绪显示 "pod 状态" 并带旋转图标）：
 * isOld=即将停止 / terminating=停止中 / pending=启动中 / 未就绪 / 就绪（主题色底 + pod 名）。
 * 传入 projectName 时，pod 名同样截断 "{项目名}-" 前缀（列表展示用短名，与容器名一致）。
 */
export const PodStateTag = memo(function PodStateTag({
  container,
  projectName,
  onClick,
  onCopy,
}: {
  container: StateContainer
  projectName?: string
  /** 整个标签可点击（容器列表点击标签 = 选中 radio） */
  onClick?: () => void
  /** 提供时在标签末尾渲染复制按钮（完整容器名），默认隐藏、hover 显示 */
  onCopy?: () => void
}) {
  const { t } = useTranslation()
  const pod = projectName ? shortContainerName(container.pod, projectName) : container.pod

  let icon: ReactNode = null
  let content: ReactNode = pod
  let className = 'bg-[#a78bfa] text-white'

  if (container.isOld) {
    className = 'bg-[#fde047] text-white'
    icon = <RefreshCw className="size-3 animate-spin" />
    content = `${pod} ${t('project.podAboutToStop')}`
  } else if (container.terminating) {
    className = 'bg-[#fca5a5] text-white'
    icon = <RefreshCw className="size-3 animate-spin" />
    content = `${pod} ${t('project.podStopping')}`
  } else if (container.pending) {
    className = 'bg-[#67e8f9] text-white'
    icon = <Loader2 className="size-3 animate-spin" />
    content = `${pod} ${t('project.podStarting')}`
  } else if (!container.ready) {
    className = 'bg-[#93c5fd] text-white'
    icon = <RefreshCw className="size-3 animate-spin" />
    content = `${pod} ${t('project.podNotReady')}`
  }

  return (
    <Pill
      className={className}
      onClick={onClick}
      onCopy={onCopy}
      copyLabel={t('project.copyContainerName')}
    >
      {icon}
      {content}
    </Pill>
  )
})
