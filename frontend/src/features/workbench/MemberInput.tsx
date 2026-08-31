import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'

/**
 * 成员标签输入（对齐邮件邀请惯例：GitHub/Gmail chip 输入）。
 * - 已有成员渲染为可删除 tag；内嵌输入框回车/逗号/粘贴自动转 tag
 * - 退格删最后一个、点击容器聚焦、删除按钮单个移除
 * 工作台命名空间卡与管理员后台命名空间治理两处共用（成员编辑表单同一交互）。
 */
export function MemberInput({
  value,
  onChange,
  placeholder,
}: {
  value: string[]
  onChange: (emails: string[]) => void
  placeholder?: string
}) {
  const { t } = useTranslation()
  const inputRef = useRef<HTMLInputElement>(null)
  const [draft, setDraft] = useState('')

  /** 追加一批邮箱：去掉首尾分隔符 + 按邮箱大小写不敏感去重 */
  const addEmails = (raws: string[]) => {
    const emails = raws
      .map((s) => s.trim().replace(/^[,，;；\s]+|[,，;；\s]+$/g, ''))
      .filter(Boolean)
    if (emails.length === 0) return
    // 批次内先折叠自身重复（粘贴 "a@x.com a@x.com" 这类同批重名），再对现有 value 去重——
    // 只对 value 去重会把同批第二个 a@x.com 直接放行 → 两个 tag 共享 key={email}，React 重排错乱
    const deduped = emails.filter(
      (e, i) => emails.findIndex((x) => x.toLowerCase() === e.toLowerCase()) === i,
    )
    const added = deduped.filter((e) => !value.some((v) => v.toLowerCase() === e.toLowerCase()))
    if (added.length > 0) onChange([...value, ...added])
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.nativeEvent.isComposing) return // 输入法组词回车/逗号不触发提交
    if (e.key === 'Enter' || e.key === ',' || e.key === '，') {
      e.preventDefault()
      addEmails([draft])
      setDraft('')
    } else if (e.key === 'Backspace' && draft === '' && value.length > 0) {
      onChange(value.slice(0, -1)) // 空输入退格删最后一个 tag
    }
  }

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    const parts = e.clipboardData.getData('text').split(/[\s,，;；]+/).filter(Boolean)
    if (parts.length > 1) {
      e.preventDefault() // 粘贴整段邮箱：直接拆分成多个 tag
      addEmails(parts)
    }
  }

  return (
    <div
      onClick={() => inputRef.current?.focus()}
      className="flex min-h-[52px] cursor-text flex-wrap items-center gap-1.5 rounded-md border border-line-strong bg-transparent px-2.5 py-2 transition-[color,box-shadow] focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50"
    >
      {value.map((email, i) => (
        <span
          key={email}
          className="inline-flex h-6 max-w-full items-center gap-1 rounded-full bg-primary-soft pl-2.5 pr-1.5 font-mono text-[11px] text-primary"
        >
          <span className="truncate">{email}</span>
          <button
            type="button"
            aria-label={`${t('workbench.membersRemove')} ${email}`}
            onClick={(e) => {
              e.stopPropagation()
              onChange(value.filter((_, j) => j !== i))
            }}
            className="flex size-4 shrink-0 items-center justify-center rounded-full text-primary/50 transition-colors hover:bg-primary/15 hover:text-primary"
          >
            <Icon name="close" className="text-[10px]" />
          </button>
        </span>
      ))}
      <input
        ref={inputRef}
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={handleKeyDown}
        onPaste={handlePaste}
        placeholder={value.length === 0 ? placeholder : ''}
        aria-label={t('workbench.membersLabel')}
        className="h-6 min-w-[10ch] flex-1 bg-transparent text-[13px] text-ink outline-none placeholder:text-ink/30"
      />
    </div>
  )
}
