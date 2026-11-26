import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Input } from '@/components/ui/shadcn/input'
import { nextZIndex } from '@/lib/zIndex'
import { useScrollPropagationGuard } from '@/hooks/useScrollPropagationGuard'
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
import { Tag } from '@/components/ui'

export interface SearchableSelectOption {
  value: string
  label: string
  /** 选项说明（可选）：下拉行在 label 后同行展示；搜索同时匹配 label 与 description */
  description?: string
  /** 醒目标记（可选）：渲染为带图标的 warn 色文字，与 description 的 accent chip 在形态与
   *  颜色上双重区分；搜索同样匹配其文案 */
  badge?: string
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
/** 多选时 trigger 里最多展示的 chip 数，超出折叠为 +N；hover +N 可查看全部已选选项 */
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
  /** 弹层开合回调：供宿主在点开时拉取最新选项（如分支/commit 列表），SearchableSelect 自身不请求数据 */
  onOpenChange,
  maxTagCount,
  tagClassName,
  contentClassName,
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
  /** 弹层开合回调（打开/关闭都触发；仅通知，不改变组件自身开合行为） */
  onOpenChange?: (open: boolean) => void
  /** trigger 里最多展示的 chip 数（默认 10）；超出折叠为 +N，hover 查看剩余项。
   *  长文本场景（如 `空间/项目`）调小可避免 trigger 被 chips 撑成多行。 */
  maxTagCount?: number
  /** 追加到 trigger 内每个 chip 的类名（宿主可限制宽度/字号，长值截断后 hover 看 title） */
  tagClassName?: string
  /** 弹层内容类名：宿主可覆盖默认宽度（默认跟随 trigger 宽度）；
   *  紧凑触发器 + 长选项（如 `空间/项目`）时传更宽的定值，避免下拉行被截断 */
  contentClassName?: string
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
  // +N 浮层：hover 展开全部已选，可逐项删除；同样需动态 z 盖过宿主弹窗
  const [tagPopoverOpen, setTagPopoverOpen] = useState(false)
  const [tagPopZ, setTagPopZ] = useState(50)
  // 从 +N 移到浮层内容要跨一小段空隙：延迟关闭，进入内容即取消
  const tagCloseTimerRef = useRef<number | null>(null)
  // 同步镜像浮层开合：hover 在 chip 与内容间来回时不重复抬 z（重复会触发内容重挂载）
  const tagPopoverOpenRef = useRef(false)
  // 选项滚动容器挂滚动传播截断：宿主 modal 的滚动锁会吞掉弹层内滚动（详见 useScrollPropagationGuard）
  const bindScrollList = useScrollPropagationGuard<HTMLDivElement>()

  // 契约：multiple 时 value 为 string[]，否则为 string。TS 无法从 boolean 窄化 value，
  // 单选分支用 as 收敛。selectedValues 固定为 string[]，否则推导类型会污染成联合数组。
  const selectedValues: string[] = multiple
    ? (Array.isArray(value) ? value : [])
    : value === ''
      ? []
      : [value as string]
  const selectedSet = useMemo(() => new Set(selectedValues), [selectedValues])
  /** trigger 内展示的 chip 上限：宿主可按文本长度调小（长 `空间/项目` 场景避免撑成多行）。 */
  const tagLimit = maxTagCount ?? MAX_TAGS
  /** 是否折叠出 +N：是则整个 trigger 作为 hover 热区，进入即展开全部已选。 */
  const hasTagOverflow = multiple && selectedValues.length > tagLimit
  /** 值 → 展示 label（选项里没有时回退到原始值，如自定义分支/手输 stage） */
  const labelOf = (v: string) => options.find((o) => o.value === v)?.label ?? v
  /** 单选 trigger 展示文本：选中项 label 或占位符 */
  const displayLabel =
    selectedValues.length > 0 ? labelOf(selectedValues[0] as string) : (placeholder ?? '')
  /** 未选中：展示的是 placeholder，用 Input 同款淡色（text-ink/30），与选中值形成区分 */
  const isPlaceholder = selectedValues.length === 0

  /** 取消 +N 浮层的延迟关闭。 */
  const cancelTagClose = () => {
    if (tagCloseTimerRef.current === null) return
    window.clearTimeout(tagCloseTimerRef.current)
    tagCloseTimerRef.current = null
  }

  /** 打开 +N 浮层：主选择弹层展开时不叠加第二个浮层；仅关闭态重新取 z 盖过可拖拽宿主弹窗。 */
  const openTagPopover = () => {
    cancelTagClose()
    if (open || tagPopoverOpenRef.current) return
    tagPopoverOpenRef.current = true
    setTagPopZ(nextZIndex())
    setTagPopoverOpen(true)
  }

  /** 关闭 +N 浮层：同步镜像 ref，否则下次 hover 会被 openTagPopover 的守卫挡掉。 */
  const closeTagPopover = () => {
    cancelTagClose()
    tagPopoverOpenRef.current = false
    setTagPopoverOpen(false)
  }

  /** 指针离开 +N 或浮层内容：延迟关闭，给指针穿过空隙进入内容的时间。 */
  const scheduleTagClose = () => {
    cancelTagClose()
    tagCloseTimerRef.current = window.setTimeout(() => {
      tagCloseTimerRef.current = null
      tagPopoverOpenRef.current = false
      setTagPopoverOpen(false)
    }, 150)
  }

  /** 卸载时清掉挂起的关闭定时器。 */
  useEffect(() => () => cancelTagClose(), [])

  // 已选降到 chip 上限以内时 +N 消失，顺带收起浮层，避免下次超限时突然以展开态出现。
  useEffect(() => {
    if (hasTagOverflow) return
    tagPopoverOpenRef.current = false
    setTagPopoverOpen(false)
  }, [hasTagOverflow])

  /** 从 +N 浮层删除单个已选值（多选语义：过滤后直接提交）。 */
  const removeValue = (v: string) => onChange(selectedValues.filter((x) => x !== v))

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return options
    return options.filter(
      (o) =>
        o.label.toLowerCase().includes(q) ||
        (o.description ?? '').toLowerCase().includes(q) ||
        (o.badge ?? '').toLowerCase().includes(q),
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

  // 非 modal：不再给 body 切 pointer-events，避免弹层开合瞬间浏览器重算 :hover、
  // 在非模态宿主弹窗（项目详情 modal={false}）边缘露出卡片时闪一下 hover。
  // 焦点链不依赖 modal 焦点陷阱：打开自动聚焦搜索框、键盘上下/回车选择、Escape 关闭。
  return (
    <Popover
      open={open}
      onOpenChange={(o) => {
        // 打开时重取 z-index：可拖拽宿主弹窗每次指针交互（bringToFront）会递增 z，
        // 挂载时算一次会过期，导致弹层被宿主盖住、选项看不见
        if (o) {
          setPopZ(nextZIndex())
          // 主选择弹层与 +N 浮层互斥，避免两层浮层叠在一起
          closeTagPopover()
        }
        setOpen(o)
        // 通知宿主：点开时可拉取最新选项（分支/commit 列表）
        onOpenChange?.(o)
      }}
    >
      <PopoverTrigger asChild>
        <button
          type="button"
          data-slot="searchable-select-trigger"
          // 折叠出 +N 时整条 trigger 都是热区：hover 任意位置（chips / 空白 / +N）都展开全部已选，
          // 不必精准瞄准 +N 那一小块。
          onPointerEnter={hasTagOverflow ? openTagPopover : undefined}
          onPointerLeave={hasTagOverflow ? scheduleTagClose : undefined}
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
                {selectedValues.slice(0, tagLimit).map((v) => (
                  <span
                    key={v}
                    className={cn(
                      'flex max-w-full items-center gap-0.5 rounded-md bg-secondary px-1.5 py-0.5 text-[12px] text-secondary-foreground',
                      tagClassName,
                    )}
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
                {hasTagOverflow && (
                  <Popover
                    open={tagPopoverOpen}
                    onOpenChange={(next) => {
                      if (next) {
                        openTagPopover()
                        return
                      }
                      closeTagPopover()
                    }}
                  >
                    {/* +N 仅作浮层定位锚点：hover 热区挂在外层 trigger 上（进入整条触发器即展开）。 */}
                    <PopoverTrigger asChild>
                      <span
                        // 在 +N 上按/点不要穿透到外层选择弹层
                        onPointerDown={(e) => e.stopPropagation()}
                        onClick={(e) => e.stopPropagation()}
                        className="cursor-default rounded-md bg-secondary px-1.5 py-0.5 text-[12px] text-secondary-foreground"
                      >
                        +{selectedValues.length - tagLimit}
                      </span>
                    </PopoverTrigger>
                    <PopoverContent
                      // key 随 z 变化：Radix Popper 只在内容挂载时把 computed z 拷到浮层 wrapper，
                      // 内容未卸载时要重新挂载才会采用新的 z（宿主置顶后重开场景）。
                      key={tagPopZ}
                      side="bottom"
                      align="start"
                      sideOffset={6}
                      style={{ zIndex: tagPopZ }}
                      // hover 展开的浮层不抢焦点，避免打断用户正在操作的输入框
                      onOpenAutoFocus={(e) => e.preventDefault()}
                      onCloseAutoFocus={(e) => e.preventDefault()}
                      // 进入内容只取消关闭：不能再抬 z（会改 key 触发重挂载，使已取得的按钮节点脱离 DOM）
                      onPointerEnter={cancelTagClose}
                      onPointerLeave={scheduleTagClose}
                      // 浮层 portal 到 body，但 React 合成事件仍沿组件树冒泡到外层 trigger 按钮，
                      // 不拦住的话点行内 X 会顺带打开外层选择弹层
                      onClick={(e) => e.stopPropagation()}
                      className="w-[18rem] max-w-[calc(100vw-2rem)] p-1"
                    >
                      {/* 每行一个名称 + 删除按钮；清单含上方已露出的 chip，便于统一管理 */}
                      <ul className="max-h-60 overflow-y-auto">
                        {selectedValues.map((v) => (
                          <li
                            key={v}
                            className="flex items-center gap-2 rounded px-1.5 py-1 text-[12px] text-foreground hover:bg-accent"
                          >
                            <span className="min-w-0 flex-1 truncate" title={labelOf(v)}>
                              {labelOf(v)}
                            </span>
                            <button
                              type="button"
                              aria-label={t('common.removeTag')}
                              onClick={() => removeValue(v)}
                              className="shrink-0 rounded p-0.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                            >
                              <Icon name="close" className="size-3" />
                            </button>
                          </li>
                        ))}
                      </ul>
                    </PopoverContent>
                  </Popover>
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
                          isPlaceholder && 'text-ink/30',
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
                  className={cn(
                    'min-w-0 flex-1 truncate',
                    align === 'center' && 'text-center',
                    isPlaceholder && 'text-ink/30',
                  )}
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
        className={cn(
          'w-[var(--radix-popover-trigger-width)] min-w-[12rem] overflow-hidden p-0',
          contentClassName,
        )}
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
            max-height 不构成确定高度，选项多时会把内容撑高被 overflow-hidden 截断而无法滚动。
            ref 回调见 bindScrollList：截断滚轮传播，避免宿主 modal 的 RemoveScroll 吞掉列表滚动 */}
        <div
          ref={bindScrollList}
          data-slot="searchable-select-list"
          className="max-h-[min(16rem,45vh)] overflow-y-auto"
        >
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
                {/* label 后同行放标记与说明：badge 用 warn 色文字（不带底色，避免与 description 的
                    accent chip 撞成「双标签」），description 仍是 accent chip，超长行内截断。
                    外按钮仍 items-center 整体垂直居中 */}
                <span className="flex min-w-0 flex-1 items-center gap-1.5 overflow-hidden">
                  <span className="min-w-0 truncate">{o.label}</span>
                  {o.badge && (
                    <span className="inline-flex shrink-0 items-center gap-0.5 text-[11px] font-medium text-warn">
                      <Icon name="shield" className="size-3" />
                      {o.badge}
                    </span>
                  )}
                  {o.description && (
                    <Tag tone="accent" dot={false} className="min-w-0 max-w-[55%]">
                      <span className="min-w-0 truncate">{o.description}</span>
                    </Tag>
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
