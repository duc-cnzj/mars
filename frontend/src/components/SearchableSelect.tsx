import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Input } from '@/components/ui/shadcn/input'
import { nextZIndex } from '@/lib/zIndex'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/shadcn/popover'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'
import { Icon } from '@/components/Icons'

export interface SearchableSelectOption {
  value: string
  label: string
  /** 选项说明（可选）：下拉行在 label 下叠一行小字；搜索同时匹配 label 与 description */
  description?: string
}

/**
 * 可搜索下拉（shadcn 风格）：
 * 关闭时只渲染 trigger，选项只有在打开后按关键词过滤才进 DOM——
 * 与 radix Select（打开即渲染全部 SelectItem）相比，分支上万时不再卡顿。
 * 打开空关键词时也最多渲染 MAX_RENDERED 条，超出用 limitText 提示搜索。
 *
 * 单选（默认）：选中即关闭；多选（multiple）：勾选/取消保持打开，可连续选择，
 * trigger 里以 chips 展示已选，空选 = 不设值（分支语义为全部）。
 * 自定义（creatable）：搜索框输入回车可把输入原文当自定义值提交（精确命中已有
 * 选项则切换，否则新增），配合 createText 展示「回车添加」提示行。
 */
const MAX_RENDERED = 200
/** 多选时 trigger 里最多展示的 chip 数，超出折叠为 +N；hover +N 可查看被折叠的剩余选项 */
const MAX_TAGS = 10

export function SearchableSelect({
  value,
  options,
  onChange,
  multiple = false,
  creatable = false,
  placeholder,
  searchPlaceholder,
  emptyText,
  limitText,
  createText,
  className,
  zIndex,
  align = 'start',
  /** 单选 trigger 选中值过长被省略（truncate 生效）时 hover 弹完整文本 tooltip；默认关闭（沿用原生 title） */
  truncateTip = false,
}: {
  value: string | string[]
  options: SearchableSelectOption[]
  onChange: (value: string | string[]) => void
  multiple?: boolean
  /** 允许自定义选项：回车把输入原文作为新值提交（antd Select tags 模式语义） */
  creatable?: boolean
  placeholder?: string
  searchPlaceholder?: string
  emptyText?: string
  /** 命中数超 MAX_RENDERED 时的提示（返回渲染条数与总命中数） */
  limitText?: (shown: number, total: number) => string
  /** creatable 时「回车添加」提示行文案（传入关键词） */
  createText?: (query: string) => string
  className?: string
  /**
   * 弹层 z-index 显式覆盖（默认打开时自动 nextZIndex() 盖过宿主弹窗）。
   * 注意宿主（可拖拽弹窗）z 会随指针交互递增，挂载时算一次会过期，一般无需传。
   */
  zIndex?: number
  /**
   * trigger 内容对齐：默认 start（左对齐）。center 供分段控件用（TabEdit 项目/分支/commit），
   * 只影响文本对齐（text-center），chevron 仍固定最右；不覆盖弹层（PopoverContent align 另管）。
   */
  align?: 'start' | 'center'
  /** 单选 trigger 文本被省略时是否弹 tooltip（默认不弹，保持原生 title 行为；仅单选生效） */
  truncateTip?: boolean
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  // 弹层 z-index：打开时重取（见 onOpenChange），保证始终高于宿主弹窗当前 z
  const [popZ, setPopZ] = useState(0)
  const [query, setQuery] = useState('')
  const [highlighted, setHighlighted] = useState(0)
  const searchRef = useRef<HTMLInputElement>(null)
  // 单选 trigger 截断检测：文本被省略（truncate 生效）时才弹 tooltip，替代原生 title 不截断也弹。
  // 命名带 label 前缀，与下拉列表的 truncated（命中数超限）区分。
  const labelRef = useRef<HTMLSpanElement>(null)
  const [labelTruncated, setLabelTruncated] = useState(false)
  const [tipHover, setTipHover] = useState(false)
  // tooltip content 走 portal 挂 body，须盖过可拖拽宿主弹窗的动态 z-index（同 Elements 注释）
  const [tipZ, setTipZ] = useState(50)

  // 契约：multiple 时 value 为 string[]，否则为 string。TS 无法从 boolean 窄化 value，
  // 单选分支用 as 收敛。selectedValues 固定为 string[]，否则推导类型会污染成联合数组。
  const selectedValues: string[] = multiple
    ? (Array.isArray(value) ? value : [])
    : value === ''
      ? []
      : [value as string]
  const selectedSet = useMemo(() => new Set(selectedValues), [selectedValues])
  /** 值 → 展示 label（选项里没有时回退到原始值，如自定义分支/手输 stage） */
  const labelOf = (v: string) => options.find((o) => o.value === v)?.label ?? v
  /** 单选 trigger 展示文本：选中项 label 或占位符 */
  const displayLabel =
    selectedValues.length > 0 ? labelOf(selectedValues[0] as string) : (placeholder ?? '')

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return options
    return options.filter(
      (o) =>
        o.label.toLowerCase().includes(q) ||
        (o.description ?? '').toLowerCase().includes(q),
    )
  }, [options, query])

  const shown = filtered.slice(0, MAX_RENDERED)
  const truncated = filtered.length > MAX_RENDERED

  // 可自定义且当前输入没有精确命中已有项 → 显示「回车添加」行
  const trimmed = query.trim()
  const createShown =
    creatable &&
    trimmed !== '' &&
    !options.some((o) => o.value === trimmed || o.label === trimmed)

  // 打开时聚焦搜索框、重置关键词与高亮；关闭时清空关键词避免下次残留
  useEffect(() => {
    if (open) {
      setQuery('')
      setHighlighted(0)
      searchRef.current?.focus()
    } else {
      setQuery('')
    }
  }, [open])

  // 截断检测（单选 + truncateTip 开启时）：scrollWidth > clientWidth（+1px 缓冲防亚像素抖动）；
  // ResizeObserver 在 span 尺寸变化（窗口/分段列宽调整）时重算，文字变化由 displayLabel 依赖触发重算。
  // 未开启时直接 return 不挂 observer；未截断或占位文本短时 labelTruncated=false，tooltip 恒不打开。
  useEffect(() => {
    if (multiple || !truncateTip) return
    const el = labelRef.current
    if (!el) return
    const measure = () => setLabelTruncated(el.scrollWidth > el.clientWidth + 1)
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(el)
    return () => ro.disconnect()
  }, [multiple, truncateTip, displayLabel])

  /** 提交一个值：单选直接选中关闭；多选切换选中态保持打开 */
  const select = (v: string) => {
    if (multiple) {
      const next = selectedSet.has(v)
        ? selectedValues.filter((x) => x !== v)
        : [...selectedValues, v]
      onChange(next)
    } else {
      onChange(v)
      setOpen(false)
    }
  }

  /** 自定义值：把输入原文作为新值提交（多选追加，单选选中即关） */
  const createCustom = (q: string) => {
    if (multiple) {
      if (selectedSet.has(q)) return
      onChange([...selectedValues, q])
    } else {
      onChange(q)
      setOpen(false)
    }
    setQuery('')
    setHighlighted(0)
  }

  const onSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setHighlighted((h) => Math.min(h + 1, shown.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setHighlighted((h) => Math.max(h - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (creatable && trimmed) {
        // 可自定义：输入原文优先；精确命中已有项则切换，否则创建自定义值
        if (selectedSet.has(trimmed)) return
        const exact = options.find((o) => o.value === trimmed || o.label === trimmed)
        if (exact) select(exact.value)
        else createCustom(trimmed)
      } else {
        const target = shown[highlighted]
        if (target) select(target.value)
      }
    } else if (e.key === 'Escape') {
      e.preventDefault()
      setOpen(false)
    }
  }

  // modal：打开时把焦点困在弹层内。非 modal 时，上方残留的 radix Select 会在
  // 搜索框聚焦后把焦点抢回它的 trigger，radix Popover 判定焦点逃离即自动关闭。
  return (
    <Popover
      open={open}
      onOpenChange={(o) => {
        // 打开时重取 z-index：可拖拽宿主弹窗每次指针交互（bringToFront）会递增 z，
        // 挂载时算一次会过期，导致弹层被宿主盖住、选项看不见
        if (o) setPopZ(nextZIndex())
        setOpen(o)
      }}
      modal
    >
      <PopoverTrigger asChild>
        <button
          type="button"
          data-slot="searchable-select-trigger"
          className={cn(
            'flex min-h-9 w-full items-center gap-2 rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs outline-none transition-[color,box-shadow] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50',
            // align 决定文本对齐：默认左对齐；center 供分段控件（TabEdit 项目/分支/commit）
            align === 'center' ? 'justify-center text-center' : 'justify-start text-left',
            multiple && 'flex-wrap',
            className,
          )}
        >
          {multiple ? (
            <>
              <span className="flex min-w-0 flex-1 flex-wrap items-center gap-1">
                {selectedValues.length === 0 && (
                  <span className="text-ink/30">{placeholder}</span>
                )}
                {selectedValues.slice(0, MAX_TAGS).map((v) => (
                  <span
                    key={v}
                    className="flex max-w-full items-center gap-0.5 rounded-md bg-secondary px-1.5 py-0.5 text-[12px] text-secondary-foreground"
                  >
                    <span className="truncate" title={labelOf(v)}>
                      {labelOf(v)}
                    </span>
                    {/* 阻止 pointerdown/click 冒泡，避免点 X 反而打开弹层 */}
                    <button
                      type="button"
                      aria-label={t('common.removeTag')}
                      onPointerDown={(e) => e.stopPropagation()}
                      onClick={(e) => {
                        e.stopPropagation()
                        select(v)
                      }}
                      className="text-secondary-foreground/60 hover:text-secondary-foreground"
                    >
                      <Icon name="close" className="size-3" />
                    </button>
                  </span>
                ))}
                {selectedValues.length > MAX_TAGS && (
                  <TooltipProvider>
                    <Tooltip delayDuration={0}>
                      <TooltipTrigger asChild>
                        <span className="cursor-help rounded-md bg-secondary px-1.5 py-0.5 text-[12px] text-secondary-foreground">
                          +{selectedValues.length - MAX_TAGS}
                        </span>
                      </TooltipTrigger>
                      <TooltipContent side="bottom" align="start" className="max-w-[18rem] p-2">
                        <div className="flex flex-wrap gap-1">
                          {selectedValues.slice(MAX_TAGS).map((v) => (
                            <span
                              key={v}
                              className="max-w-full truncate rounded bg-background/15 px-1.5 py-0.5 text-[12px] text-background"
                            >
                              {labelOf(v)}
                            </span>
                          ))}
                        </div>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
              </span>
              <Icon name="chevron-down" className="size-4 shrink-0 opacity-50" />
            </>
          ) : (
            <>
              {/* truncateTip 开启：仅选中值被省略时 hover 弹完整文本 tooltip（受控 open，见下方）。
                  hover 挂在文本 span 本体；仅截断时才真正打开、才需抬 z-index，未截断时 open 恒
                  false、content 不挂载。默认关闭则回到原生 title 行为（不截断也弹，无样式）。 */}
              {truncateTip ? (
                <TooltipProvider delayDuration={100}>
                  <Tooltip open={labelTruncated && tipHover} onOpenChange={() => {}}>
                    <TooltipTrigger asChild>
                      <span
                        ref={labelRef}
                        onMouseEnter={() => {
                          if (labelTruncated) setTipZ(nextZIndex())
                          setTipHover(true)
                        }}
                        onMouseLeave={() => setTipHover(false)}
                        className={cn(
                          'min-w-0 flex-1 truncate',
                          align === 'center' && 'text-center',
                        )}
                      >
                        {displayLabel}
                      </span>
                    </TooltipTrigger>
                    <TooltipContent
                      side="top"
                      // 防御性去掉 text-wrap:balance：balance 会把折行压成多行等宽、首行不占满就换行。
                      // 必须写长属性 textWrapStyle（text-wrap-style）而非 shorthand textWrap（text-wrap:normal）：
                      // text-balance 写的是 longhand，Chrome 长短属性交互怪癖，shorthand 压不掉（同 Elements FieldLabel）
                      style={{ zIndex: tipZ, textWrapStyle: 'auto' }}
                      className="max-w-[min(320px,80vw)] break-words"
                    >
                      {displayLabel}
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              ) : (
                <span
                  // 两种对齐都 flex-1 占满：start 时文字居左、center 时文字居中，
                  // chevron 均被推在最右（用户要求居中的是文本，箭头保持靠右）
                  className={cn('min-w-0 flex-1 truncate', align === 'center' && 'text-center')}
                  title={
                    selectedValues.length > 0
                      ? labelOf(selectedValues[0] as string)
                      : undefined
                  }
                >
                  {displayLabel}
                </span>
              )}
              <Icon name="chevron-down" className="size-4 shrink-0 opacity-50" />
            </>
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="start"
        sideOffset={4}
        style={{ zIndex: zIndex ?? (popZ || undefined) }}
        className="w-[var(--radix-popover-trigger-width)] min-w-[12rem] p-0"
      >
        <div className="border-b border-line p-1.5">
          <div className="relative">
            <Icon name="search" className="absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              ref={searchRef}
              data-slot="searchable-select-input"
              aria-label={searchPlaceholder}
              value={query}
              onChange={(e) => {
                setQuery(e.target.value)
                setHighlighted(0)
              }}
              onKeyDown={onSearchKeyDown}
              placeholder={searchPlaceholder}
              className="h-8 pl-8 text-left"
            />
          </div>
        </div>
        {truncated && limitText && (
          <div className="border-b border-line px-3 py-1.5 text-[12px] text-muted-foreground">
            {limitText(shown.length, filtered.length)}
          </div>
        )}
        {createShown && (
          <button
            type="button"
            data-slot="searchable-select-create"
            onClick={() => createCustom(trimmed)}
            className="flex w-full items-center gap-2 border-b border-line px-3 py-1.5 text-left text-sm text-primary hover:bg-accent"
          >
            <Icon name="plus" className="size-4 shrink-0" />
            {createText ? createText(trimmed) : trimmed}
          </button>
        )}
        {/* 原生 div 滚动：Radix ScrollArea 的 viewport 高度 100% 依赖父级确定高度，
            max-height 不构成确定高度，选项多时会把内容撑高被 overflow-hidden 截断而无法滚动 */}
        <div className="max-h-[min(16rem,45vh)] overflow-y-auto">
          {shown.map((o, i) => {
            const isSel = multiple
              ? selectedSet.has(o.value)
              : o.value === value
            return (
              <button
                key={o.value}
                type="button"
                data-slot="searchable-select-option"
                title={o.label}
                onMouseEnter={() => setHighlighted(i)}
                onClick={() => select(o.value)}
                className={cn(
                  'relative flex w-full items-center gap-2 py-1.5 pr-8 pl-2 text-left text-sm outline-none select-none focus-visible:ring-2 focus-visible:ring-ring/50',
                  i === highlighted ? 'bg-accent text-accent-foreground' : 'text-foreground',
                  isSel && 'font-medium',
                )}
              >
                {/* 有 description 时内层改纵向叠排（label + 小字说明），无则保持单行；外按钮仍 items-center 整体垂直居中 */}
                <span className="flex min-w-0 flex-1 flex-col">
                  <span className="truncate">{o.label}</span>
                  {o.description && (
                    <span className="truncate text-[11px] text-muted-foreground">
                      {o.description}
                    </span>
                  )}
                </span>
                {isSel && (
                  <span className="absolute right-2 flex items-center">
                    <Icon name="check" className="size-4" />
                  </span>
                )}
              </button>
            )
          })}
          {shown.length === 0 && !createShown && (
            <div className="px-2 py-4 text-center text-sm text-muted-foreground">
              {emptyText}
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}
