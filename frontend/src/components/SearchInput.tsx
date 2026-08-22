import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { isMac } from '@/lib/platform'
import { Input } from '@/components/ui/shadcn/input'
import { Kbd } from '@/components/ui/Kbd'
import { Icon } from './Icons'

/**
 * 通用搜索框：左侧放大镜 + 右侧快捷键徽标（⌘K / Ctrl K）/ 有值时变清除按钮。
 * 内嵌 ctrl/cmd+k 全局快捷键聚焦（对齐 workbench 交互）。
 */
export function SearchInput({
  value,
  onChange,
  placeholder,
  className,
}: {
  value: string
  onChange: (v: string) => void
  placeholder?: string
  className?: string
}) {
  const { t } = useTranslation()
  const ref = useRef<HTMLInputElement>(null)

  // ctrl/cmd + k 聚焦搜索框
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault()
        ref.current?.focus()
        ref.current?.select()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  const hasValue = value.length > 0

  return (
    <div className={cn('relative', className)}>
      <Icon
        name="search"
        className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-faint"
      />
      <Input
        ref={ref}
        aria-label={t('common.search')}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={cn(
          // shadcn 基座（bg-transparent / md:text-sm / placeholder:text-ink/30 /
          // transition-[color,box-shadow]）在 v4 排序里会后于本组覆盖类，需用 ! 强制；
          // 圆角/边框/内边距/focus bg 已确认能自然胜出。focus ring 沿用基座 ring-[3px] ring-ring/50 保持全站一致。
          // Ctrl K 徽标更宽，非 mac 需要更多右侧空间，避免与文本重叠
          isMac
            ? 'h-9 rounded-lg border-line !bg-raised/60 pl-9 pr-14 !text-[13px]'
            : 'h-9 rounded-lg border-line !bg-raised/60 pl-9 pr-16 !text-[13px]',
          'placeholder:!text-ink/30 !transition-[border-color,background-color,box-shadow] duration-150',
          'hover:border-line-strong',
          'focus-visible:bg-surface',
        )}
      />
      {hasValue ? (
        <button
          type="button"
          aria-label={t('common.clearSearch')}
          onClick={() => {
            onChange('')
            ref.current?.focus()
          }}
          className="absolute right-2 top-1/2 flex size-5 -translate-y-1/2 items-center justify-center rounded-full text-faint transition-colors hover:bg-raised hover:text-ink"
        >
          <Icon name="close" className="size-3.5" />
        </button>
      ) : (
        <button
          type="button"
          aria-label={t('common.focusSearch')}
          onClick={() => ref.current?.focus()}
          className="absolute right-2.5 top-1/2 -translate-y-1/2"
        >
          {isMac ? (
            // ⌘ 符号在等宽小字号里渲染得很小，单独放大并用无衬线字体显示
            <Kbd className="flex items-center gap-0.5 text-[11px]">
              <span className="font-sans text-[12px] leading-none">⌘</span>
              K
            </Kbd>
          ) : (
            <Kbd className="text-[11px]">Ctrl K</Kbd>
          )}
        </button>
      )}
    </div>
  )
}
