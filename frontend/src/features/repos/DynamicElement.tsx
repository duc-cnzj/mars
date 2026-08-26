import { useMemo, useRef, useState } from 'react'
import type { DragEvent } from 'react'
import { useTranslation } from 'react-i18next'
import type { components } from '@/api/schema'
import { SearchableSelect } from '@/components/SearchableSelect'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { Switch } from '@/components/ui/shadcn/switch'
import { CodeEditor, FILE_TYPES } from '@/components/CodeEditor'
import { SelectFileType } from './SelectFileType'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/shadcn/select'
import { Icon } from '@/components/Icons'
import { cn } from '@/lib/utils'
import {
  buildSections,
  DEFAULT_GROUP,
  groupLabel,
  groupOf,
  type GroupSection,
} from '@/lib/groupSections'

type Element = components['schemas']['mars.Element']
type GroupSetting = components['schemas']['mars.GroupSetting']

/** 需要"选择器选项"的类型：default 必须命中其中一项（对齐旧版隐藏/校验逻辑） */
export const SELECTIVE_TYPES = new Set([
  'ElementTypeSelect',
  'ElementTypeNumberSelect',
  'ElementTypeRadio',
  'ElementTypeNumberRadio',
])

/** 默认值必填的类型（对齐旧版 isDefaultRequired）；RepoFormModal 提交校验复用同一份 */
export const DEFAULT_REQUIRED_TYPES = new Set([
  'ElementTypeInputNumber',
  'ElementTypeRadio',
  'ElementTypeNumberRadio',
  'ElementTypeSelect',
  'ElementTypeNumberSelect',
  'ElementTypeSwitch',
])

/** 类型下拉选项：value 为 schema 枚举字符串，label 走 i18n（as const 保住 key 字面量类型） */
const ELEMENT_TYPE_KEYS = [
  { value: 'ElementTypeUnknown', key: 'repos.elementTypeUnknown' },
  { value: 'ElementTypeInput', key: 'repos.elementTypeInput' },
  { value: 'ElementTypeInputNumber', key: 'repos.elementTypeInputNumber' },
  { value: 'ElementTypeTextArea', key: 'repos.elementTypeTextArea' },
  { value: 'ElementTypeRadio', key: 'repos.elementTypeRadio' },
  { value: 'ElementTypeNumberRadio', key: 'repos.elementTypeNumberRadio' },
  { value: 'ElementTypeSelect', key: 'repos.elementTypeSelect' },
  { value: 'ElementTypeNumberSelect', key: 'repos.elementTypeNumberSelect' },
  { value: 'ElementTypeSwitch', key: 'repos.elementTypeSwitch' },
] as const

const labelCls = 'text-[12px] font-medium text-mute'

/** 拖拽负载：记录拖源（卡片=元素 / 分区标题=分区），drop 时据此重建扁平数组 */
type DragPayload =
  | { kind: 'element'; section: string; index: number }
  | { kind: 'group'; gid: string }

/**
 * 类型切换后按新类型归一化 default，杜绝「显示值 ≠ 存储值」错位与过期默认值：
 * - switch：只认 'true'/'false'，其它值（空串/文本遗留值）兜底 'false'（控件本就按非 'true' 显 false）
 * - InputNumber：清掉非数字旧值，否则浏览器数字框显空、存储却留着垃圾值，保存还照过非空校验
 * - 选择类型（Select/Radio 等）：default 必须是 selectValues 选项之一，否则清空重选，
 *   避免保存撞「默认值必须命中选项」
 */
function normalizeDefaultForType(el: Element): Element {
  const { type, default: d, selectValues } = el
  if (type === 'ElementTypeSwitch' && d !== 'true' && d !== 'false') {
    return { ...el, default: 'false' }
  }
  if (
    type === 'ElementTypeInputNumber' &&
    d !== '' &&
    // 数字输入只接受 HTML 浮点语法（含符号/小数/指数）：hex '0x10' 等 Number() 能解析但
    // input[type=number] 显示为空，同样「显示空、存储有值」错位；再叠加有限数挡住 '1e999' 溢出
    (!/^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?$/.test(d) ||
      !Number.isFinite(Number(d)))
  ) {
    return { ...el, default: '' }
  }
  if (SELECTIVE_TYPES.has(type) && d !== '' && !selectValues.includes(d)) {
    return { ...el, default: '' }
  }
  return el
}

/**
 * 自定义配置编辑器（还原旧版 DynamicElement + 分区/折叠/拖拽排序，对齐设计 demo）：
 * 分区模式下元素渲染在各自分区容器内，分区标题可拖拽排序（写回 group_settings.order），
 * 元素卡片可在组内拖拽重排、跨分区拖拽（写回 element.group）。折叠分两层：
 * - 标题点击/箭头 = 临时折叠（仅视图，便于编辑时看布局，不写配置）；
 * - 标题旁「默认折叠」开关 = 写回 group_settings.collapsed（TabEdit 部署页进入时按此收起）。
 * 未分组区恒排最后、不可拖拽。
 * 全部元素未分组 → 无分区，退回旧版平铺（无分区头/无折叠，仅字段列表 + 添加按钮）。
 * 拖拽用原生 HTML5 DnD（跨容器拖拽语义直观，无需 dnd-kit 复杂跨容器协调）。
 */
export function DynamicElement({
  value,
  groups,
  openGroups,
  onChange,
  onGroupsChange,
  onToggleOpen,
}: {
  value: Element[]
  /** 分区配置（order/collapsed）：由 RepoFormModal 持有、随 marsConfig 持久化；与 elements 联动 */
  groups: GroupSetting[]
  /** 编辑器视图折叠态（Record<分区名, 是否展开>）：由父级持有，与配置解耦——「折叠所有组」按钮是纯视图操作 */
  openGroups: Record<string, boolean>
  onChange: (next: Element[]) => void
  onGroupsChange: (next: GroupSetting[]) => void
  /** 编辑时临时折叠/展开分区（仅视图，不改配置） */
  onToggleOpen: (name: string) => void
}) {
  const { t } = useTranslation()
  // 拖拽负载存 ref（dragstart 与 drop 在不同子节点，走闭包引用同一份）
  const dragRef = useRef<DragPayload | null>(null)
  const [dragEl, setDragEl] = useState<string | null>(null) // 正在拖拽的元素卡片（视觉降透明度）
  const [hoverEl, setHoverEl] = useState<{ key: string; before: boolean } | null>(null) // 卡片落点（前/后半）
  const [hoverSec, setHoverSec] = useState<string | null>(null) // 分区体落点（虚线框提示）
  const [dragGrp, setDragGrp] = useState<string | null>(null) // 正在拖拽的分区标题
  const [hoverGrp, setHoverGrp] = useState<string | null>(null) // 分区标题落点

  const sections = useMemo(() => buildSections(value, groups), [value, groups])
  // 全部现有分区名（「所属分区」下拉候选；未分组不在此列）
  const groupOptions = useMemo(
    () => sections.filter((s) => s.name !== '').map((s) => s.name),
    [sections],
  )

  // ---------- 分区折叠：视图展开态（openGroups）与默认折叠配置（group_settings.collapsed）彻底解耦 ----------
  // 标题/箭头/「折叠所有组」只切 openGroups（纯视图，便于编辑看布局/拖排序）；
  // 「默认折叠」开关只写 group_settings.collapsed（配置，TabEdit 部署页进入时按此收起），互不影响。
  /** 分区视图展开态：未显式记录默认展开 */
  const isOpen = (sec: GroupSection) => openGroups[sec.name] ?? true

  // ---------- 扁平数组定位工具：分区内位置 ↔ 扁平下标 ----------
  /** 定位 (section, 分区内下标) 对应的扁平下标；找不到返回 -1 */
  const flatIndexOf = (section: string, local: number) => {
    let n = 0
    for (let i = 0; i < value.length; i++) {
      if (groupOf(value[i]) === section) {
        if (n === local) return i
        n++
      }
    }
    return -1
  }

  // ---------- 元素增删改 ----------
  /** 更新某分区内第 index 个元素（patch.group 统一去空白；switch/类型切换触发 default 归一化） */
  const updateElement = (section: string, index: number, patch: Partial<Element>) => {
    const flat = flatIndexOf(section, index)
    if (flat < 0) return
    const next = [...value]
    const merged = {
      ...next[flat],
      ...patch,
      ...(patch.group !== undefined ? { group: patch.group.trim() } : {}),
    }
    next[flat] =
      merged.type === 'ElementTypeSwitch' ||
      (patch.type && merged.type !== next[flat].type)
        ? normalizeDefaultForType(merged)
        : merged
    onChange(next)
  }

  /** 追加一个元素到指定分区末尾（扁平末尾 = 分区内末尾：分区顺序由扁平数组过滤派生） */
  const addElement = (section: string) => {
    onChange([
      ...value,
      {
        path: '',
        type: 'ElementTypeUnknown' as Element['type'],
        default: '',
        description: '',
        selectValues: [],
        order: value.length,
        textareaLanguage: '',
        // 无分区平铺态新增：默认落入兜底分区（default）；有分区时跟随当前分区
        group: section === '' ? DEFAULT_GROUP : section,
      },
    ])
  }

  const removeElement = (section: string, index: number) => {
    const flat = flatIndexOf(section, index)
    if (flat < 0) return
    onChange(value.filter((_, i) => i !== flat))
  }

  /** 移动元素：分区内重排 / 跨分区移动（改 group）。toIndex 为目标分区内插入位置（0=区首，len=区尾） */
  const moveElement = (
    fromSection: string,
    fromIndex: number,
    toSection: string,
    toIndex: number,
  ) => {
    const src = flatIndexOf(fromSection, fromIndex)
    if (src < 0) return
    const next = [...value]
    const [moved] = next.splice(src, 1)
    // 同分区且目标在源之后：抽离后目标位置前移一位（HTML5 卡片 drop 传的是含源的下标）
    let adjusted = toIndex
    if (toSection === fromSection && fromIndex < toIndex) adjusted -= 1
    // 目标分区内插入位置 → 扁平下标（adjusted 越界 = 追加到分区末尾 = 扁平末尾）
    let insert = next.length
    let n = 0
    for (let i = 0; i < next.length; i++) {
      if (groupOf(next[i]) === toSection) {
        if (n === adjusted) {
          insert = i
          break
        }
        n++
      }
    }
    next.splice(insert, 0, { ...moved, group: toSection })
    // 排序后刷新 order，保持与列表顺序一致
    onChange(next.map((el, i) => ({ ...el, order: i })))
  }

  /** 「所属分区」下拉换组：与 demo 一致，挪到目标分组末尾 */
  const changeGroup = (fromSection: string, fromIndex: number, toGroup: string) => {
    if (toGroup === fromSection) return
    const src = flatIndexOf(fromSection, fromIndex)
    if (src < 0) return
    const next = [...value]
    const [moved] = next.splice(src, 1)
    next.push({ ...moved, group: toGroup })
    onChange(next.map((el, i) => ({ ...el, order: i })))
  }

  // ---------- 分区操作 ----------
  /** 分区拖拽重排：按当前展示顺序写回完整 group_settings（order = 位置，保留 collapsed） */
  const reorderSection = (fromName: string, toName: string) => {
    const names = sections.filter((s) => s.name !== '').map((s) => s.name)
    const from = names.indexOf(fromName)
    let to = names.indexOf(toName)
    if (from < 0 || to < 0 || from === to) return
    const [moved] = names.splice(from, 1)
    // 抽离后目标下标可能位移一位
    to = names.indexOf(toName)
    names.splice(to, 0, moved)
    const settings = new Map(groups.map((g) => [g.name, g]))
    onGroupsChange(
      names.map((name, i) => ({
        name,
        order: i,
        collapsed: settings.get(name)?.collapsed ?? false,
      })),
    )
  }

  /** 写入单组「默认折叠」配置：只写 group_settings.collapsed，不动视图展开态（与折叠展开解耦） */
  const setGroupCollapsed = (name: string, collapsed: boolean) => {
    const settings = new Map(groups.map((g) => [g.name, g]))
    onGroupsChange(
      sections
        .filter((s) => s.name !== '')
        .map((s, i) => ({
          name: s.name,
          order: i,
          collapsed:
            s.name === name
              ? collapsed
              : (settings.get(s.name)?.collapsed ?? false),
        })),
    )
  }

  // ---------- 原生 HTML5 DnD：元素卡片 ----------
  /** 交互控件内不触发拖拽（拖文本框/下拉会误拖卡片）；阻止冒泡让 drop 只落到卡片 */
  const onCardDragStart = (e: DragEvent, section: string, index: number) => {
    if (
      (e.target as HTMLElement).closest(
        'input, textarea, select, button, [contenteditable]',
      )
    ) {
      e.preventDefault()
      return
    }
    e.dataTransfer.setData('text/plain', 'card')
    e.dataTransfer.effectAllowed = 'move'
    dragRef.current = { kind: 'element', section, index }
    setDragEl(`${section}:${index}`)
  }

  const onCardDragOver = (e: DragEvent, section: string, index: number) => {
    const d = dragRef.current
    if (!d || d.kind !== 'element') return
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    // 上下半区判定落点（对齐 demo 的 drop-before/drop-after 指示线）
    const rect = e.currentTarget.getBoundingClientRect()
    const before = e.clientY - rect.top < rect.height / 2
    const key = `${section}:${index}`
    setHoverEl((prev) =>
      prev && prev.key === key && prev.before === before ? prev : { key, before },
    )
    // 悬在卡片上时清除分区体的虚线提示，避免两种落点指示叠加
    setHoverSec((prev) => (prev === null ? prev : null))
  }

  const onCardDrop = (e: DragEvent, section: string, index: number) => {
    e.preventDefault()
    e.stopPropagation()
    const d = dragRef.current
    if (!d || d.kind !== 'element') return
    const rect = e.currentTarget.getBoundingClientRect()
    const before = e.clientY - rect.top < rect.height / 2
    moveElement(d.section, d.index, section, before ? index : index + 1)
    endDrag()
  }

  /** 分区体 dragover/drop：鼠标悬空/落到分区体空白处 = 追加到该分区末尾（跨分区挪组的快捷方式） */
  const onBodyDragOver = (e: DragEvent, section: string) => {
    const d = dragRef.current
    if (!d || d.kind !== 'element') return
    // 悬在卡片上时交给卡片自己的 dragover（避免重复 preventDefault 抢落点）
    if ((e.target as HTMLElement).closest('[data-card]')) return
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setHoverSec((prev) => (prev === section ? prev : section))
    // 悬在分区体空白处时清除卡片落点指示，避免与虚线提示叠加
    setHoverEl((prev) => (prev === null ? prev : null))
  }

  const onBodyDrop = (e: DragEvent, section: string) => {
    const d = dragRef.current
    if (!d || d.kind !== 'element') return
    if ((e.target as HTMLElement).closest('[data-card]')) return
    e.preventDefault()
    const sec = sections.find((s) => s.name === section)
    moveElement(d.section, d.index, section, sec?.elements.length ?? 0)
    endDrag()
  }

  // ---------- 原生 HTML5 DnD：分区标题 ----------
  const onGroupDragStart = (e: DragEvent, gid: string) => {
    e.dataTransfer.setData('text/plain', 'group')
    e.dataTransfer.effectAllowed = 'move'
    dragRef.current = { kind: 'group', gid }
    setDragGrp(gid)
  }

  const onHeaderDragOver = (e: DragEvent, gid: string) => {
    const d = dragRef.current
    if (!d || d.kind !== 'group' || d.gid === gid) return
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setHoverGrp((prev) => (prev === gid ? prev : gid))
  }

  const onHeaderDrop = (e: DragEvent, gid: string) => {
    e.preventDefault()
    e.stopPropagation()
    const d = dragRef.current
    if (!d || d.kind !== 'group' || d.gid === gid) return
    reorderSection(d.gid, gid)
    endDrag()
  }

  /** 拖拽收尾：清负载与全部视觉状态（HTML5 dragend 必在拖源上触发，卡片/标题都挂） */
  const endDrag = () => {
    dragRef.current = null
    setDragEl(null)
    setHoverEl(null)
    setHoverSec(null)
    setDragGrp(null)
    setHoverGrp(null)
  }

  /** 元素卡片（扁平/分区共用）：拖拽 + 字段编辑 + 「所属分区」归组 */
  const renderCard = (el: Element, section: string, index: number) => {
    const key = `${section}:${index}`
    const hover = hoverEl?.key === key && key !== dragEl ? hoverEl : null
    return (
      <ElementCard
        key={key}
        element={el}
        section={section}
        index={index}
        groupOptions={groupOptions}
        isDragging={dragEl === key}
        isHoverBefore={hover?.before ?? false}
        isHoverAfter={hover ? !hover.before : false}
        onDragStart={onCardDragStart}
        onDragEnd={endDrag}
        onDragOver={onCardDragOver}
        onDrop={onCardDrop}
        onChange={(patch) => updateElement(section, index, patch)}
        onRemove={() => removeElement(section, index)}
        onGroupChange={(g) => changeGroup(section, index, g)}
      />
    )
  }

  return (
    <div className="flex flex-col gap-2">
      {sections.length === 0 ? (
        // 兼容旧版：无分区 → 平铺全部字段，无分区头/无折叠；卡片仍带「所属分区」供进入分组
        // 布局：一行 2 张卡片（lg+ 两列，窄屏回退单列），与分区模式一致
        <div className="grid grid-cols-1 gap-2 rounded-lg border border-line bg-surface p-3 lg:grid-cols-3">
          {value.map((el, i) => renderCard(el, '', i))}
          <Button variant="dashed" className="lg:col-span-3" onClick={() => addElement('')}>
            <Icon name="plus" />
            {t('repos.addElement')}
          </Button>
        </div>
      ) : (
        sections.map((sec) => {
          const isUngrouped = sec.name === ''
          return (
            <section
              key={sec.name}
              data-section={sec.name}
              className="overflow-hidden rounded-lg border border-line bg-surface"
            >
              {/* 分区头：非未分组可拖拽排序——仅 grip 图标发起拖拽，其余区域点击=临时折叠/展开（视图）。
                  默认折叠开关=纯配置（写回 group_settings.collapsed），点击不触发折叠，与视图解耦 */}
              <div
                onClick={() => !isUngrouped && onToggleOpen(sec.name)}
                onDragOver={
                  !isUngrouped ? (e) => onHeaderDragOver(e, sec.name) : undefined
                }
                onDrop={
                  !isUngrouped ? (e) => onHeaderDrop(e, sec.name) : undefined
                }
                style={
                  hoverGrp === sec.name
                    ? { boxShadow: '0 -2px 0 0 var(--primary)' }
                    : undefined
                }
                className={cn(
                  'flex items-center gap-2 border-b border-line px-3 py-2',
                  !isUngrouped &&
                    cn(
                      'cursor-pointer select-none',
                      dragGrp === sec.name && 'opacity-60',
                    ),
                )}
              >
                {!isUngrouped && (
                  <span
                    draggable
                    onDragStart={(e) => onGroupDragStart(e, sec.name)}
                    onDragEnd={endDrag}
                    onClick={(e) => e.stopPropagation()}
                    title={t('repos.groupsDragTip')}
                    className="shrink-0 cursor-grab active:cursor-grabbing text-faint transition-colors hover:text-ink"
                  >
                    <Icon name="grip-vertical" className="size-4" />
                  </span>
                )}
                {isUngrouped ? (
                  <span className="min-w-0 flex-1 truncate text-[13px] font-medium text-mute">
                    {t('repos.groupUngrouped')}
                  </span>
                ) : (
                  <span className="flex min-w-0 flex-1 items-center gap-1.5">
                    <span className="min-w-0 truncate text-[13px] font-medium text-ink">
                      {groupLabel(sec.name, t)}
                    </span>
                    {/* 默认折叠：纯配置项，紧跟组名；只写 group_settings.collapsed，与当前折叠展开解耦 */}
                    <span
                      className="flex shrink-0 items-center gap-1 text-[11px] text-mute"
                      title={t('repos.groupDefaultCollapsedTip')}
                      onClick={(e) => e.stopPropagation()}
                    >
                      <Switch
                        size="sm"
                        checked={sec.collapsed}
                        onCheckedChange={(v) => setGroupCollapsed(sec.name, v)}
                      />
                      {t('repos.groupDefaultCollapsed')}
                    </span>
                  </span>
                )}
                <span className="shrink-0 rounded-full border border-line px-2 py-0.5 text-[11px] text-mute">
                  {t('repos.groupsItemCount', { count: sec.elements.length })}
                </span>
                {!isUngrouped && (
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation()
                      onToggleOpen(sec.name)
                    }}
                    aria-label={t('repos.groupsCollapseTip')}
                    className="shrink-0 text-faint transition-transform hover:text-ink"
                  >
                    <Icon
                      name="chevron-down"
                      className={cn(
                        'size-4 transition-transform',
                        !isOpen(sec) && '-rotate-90',
                      )}
                    />
                  </button>
                )}
              </div>
              {isOpen(sec) && (
                // 分区内容一行 2 张卡片（lg+ 两列，窄屏回退单列）。拖拽落点判定仍按卡片
                // 上下半区（前/后）对应扁平顺序，2 列下指示线语义不变
                <div
                  className="grid grid-cols-1 gap-2 p-3 lg:grid-cols-3"
                  onDragOver={(e) => onBodyDragOver(e, sec.name)}
                  onDrop={(e) => onBodyDrop(e, sec.name)}
                  style={
                    hoverSec === sec.name
                      ? {
                          outline: '2px dashed var(--primary)',
                          outlineOffset: '-2px',
                        }
                      : undefined
                  }
                >
                  {sec.elements.map((el, i) => renderCard(el, sec.name, i))}
                  <Button
                    variant="dashed"
                    className="lg:col-span-3"
                    onClick={() => addElement(sec.name)}
                  >
                    <Icon name="plus" />
                    {t('repos.addElement')}
                  </Button>
                </div>
              )}
            </section>
          )
        })
      )}
    </div>
  )
}

/** 单个自定义字段卡片：拖拽手柄 + 字段编辑，可组内排序/跨组移动（原生 HTML5 DnD） */
function ElementCard({
  element,
  section,
  index,
  groupOptions,
  isDragging,
  isHoverBefore,
  isHoverAfter,
  onDragStart,
  onDragEnd,
  onDragOver,
  onDrop,
  onChange,
  onRemove,
  onGroupChange,
}: {
  element: Element
  section: string
  index: number
  /** 现有全部分区名（「所属分区」下拉候选） */
  groupOptions: string[]
  isDragging: boolean
  isHoverBefore: boolean
  isHoverAfter: boolean
  onDragStart: (e: DragEvent, section: string, index: number) => void
  onDragEnd: () => void
  onDragOver: (e: DragEvent, section: string, index: number) => void
  onDrop: (e: DragEvent, section: string, index: number) => void
  onChange: (patch: Partial<Element>) => void
  onRemove: () => void
  onGroupChange: (group: string) => void
}) {
  const { t } = useTranslation()
  const selective = SELECTIVE_TYPES.has(element.type)
  const defaultRequired = DEFAULT_REQUIRED_TYPES.has(element.type)

  return (
    <div
      data-card
      draggable
      onDragStart={(e) => onDragStart(e, section, index)}
      onDragEnd={onDragEnd}
      onDragOver={(e) => onDragOver(e, section, index)}
      onDrop={(e) => onDrop(e, section, index)}
      // 落点指示线：上半区 = 插到该卡片前，下半区 = 插到该卡片后（对齐 demo 的 box-shadow 指示）
      style={
        isHoverBefore
          ? { boxShadow: '0 -2px 0 0 var(--primary)' }
          : isHoverAfter
            ? { boxShadow: '0 2px 0 0 var(--primary)' }
            : undefined
      }
      className={cn(
        'rounded-lg border border-line bg-surface p-3 shadow-sm',
        isDragging && 'opacity-60',
      )}
    >
      <div className="flex flex-col gap-2.5">
        <div className="flex items-start gap-2">
          {/* 拖拽手柄（整卡 draggable，拖手柄最顺）：span 非交互元素，不触发拖拽取消逻辑 */}
          <span
            className="mt-2.5 cursor-grab text-faint transition-colors hover:text-ink active:cursor-grabbing"
            title={t('repos.elementDrag')}
          >
            <Icon name="grip-vertical" className="size-4" />
          </span>
          <label className="flex min-w-0 flex-1 flex-col gap-1.5">
            <span className={labelCls}>
              {t('repos.elementPath')} <span className="text-err">*</span>
            </span>
            <Input
              value={element.path}
              onChange={(e) => onChange({ path: e.target.value })}
              placeholder={t('repos.elementPathPlaceholder')}
            />
          </label>
          <label className="flex w-[150px] shrink-0 flex-col gap-1.5">
            <span className={labelCls}>
              {t('repos.elementType')} <span className="text-err">*</span>
            </span>
            <Select
              value={element.type}
              onValueChange={(v) => onChange({ type: v as Element['type'] })}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder={t('repos.elementTypePlaceholder')} />
              </SelectTrigger>
              <SelectContent>
                {ELEMENT_TYPE_KEYS.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {t(o.key)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </label>
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={onRemove}
            aria-label={t('common.delete')}
            className="mt-2 shrink-0"
          >
            <Icon name="close" />
          </Button>
        </div>

        <div className="grid grid-cols-1 gap-2.5 pl-6 sm:grid-cols-2">
          <label className="flex flex-col gap-1.5">
            <span className={labelCls}>
              {t('repos.elementDescription')} <span className="text-err">*</span>
            </span>
            <Input
              value={element.description}
              onChange={(e) => onChange({ description: e.target.value })}
              placeholder={t('repos.elementDescriptionPlaceholder')}
            />
          </label>
          {element.type === 'ElementTypeTextArea' ? (
            // TextArea：编辑器语言与字段描述同行（右半列），下方 CodeMirror 默认值独占一行
            <label className="flex flex-col gap-1.5">
              <span className={labelCls}>{t('repos.elementTextAreaLanguage')}</span>
              {/* 复用配置文件类型选择器：55 种语言候选 + 可搜索 + 自由值兜底 */}
              <SelectFileType
                value={element.textareaLanguage}
                onChange={(v) => onChange({ textareaLanguage: v })}
                placeholder={t('repos.elementTextAreaLanguagePlaceholder')}
              />
            </label>
          ) : (
            <label className="flex flex-col gap-1.5">
              <span className={labelCls}>
                {t('repos.elementDefault')}
                {defaultRequired && <span className="text-err"> *</span>}
              </span>
              <DefaultValueInput
                type={element.type}
                value={element.default}
                onChange={(v) => onChange({ default: v })}
              />
            </label>
          )}
        </div>

        {/* 所属分区：creatable 单选出分区名，输入新值回车即创建分区（进入分区模式）；首项「未分组」解除归组 */}
        <label className="flex flex-col gap-1.5 pl-6">
          <span className={labelCls}>{t('repos.elementGroup')}</span>
          <SearchableSelect
            value={element.group ?? ''}
            options={[
              { value: '', label: t('repos.groupUngrouped') },
              // 兜底分区（default）展示名走 groupLabel 国际化，其余分区名原样
              ...groupOptions.map((g) => ({ value: g, label: groupLabel(g, t) })),
            ]}
            creatable
            onChange={(v) => onGroupChange(v as string)}
            placeholder={t('repos.elementGroupPlaceholder')}
            searchPlaceholder={t('common.search')}
            emptyText={t('common.empty')}
            createText={(q) => t('repos.elementGroupCreate', { name: q })}
          />
        </label>

        {/* TextArea 默认值 CodeMirror：独占一行（独立于上方 grid），语言联动——language 直接取
            textareaLanguage，切换语言选择器即重新解析高亮；空串/未知值回退 textile（与部署表单
            Elements.tsx 的 FILE_TYPES 判定一致） */}
        {element.type === 'ElementTypeTextArea' && (
          <label className="flex flex-col gap-1.5 pl-6">
            <span className={labelCls}>{t('repos.elementDefault')}</span>
            <CodeEditor
              value={element.default}
              onChange={(v) => onChange({ default: v })}
              language={
                (FILE_TYPES as readonly string[]).includes(element.textareaLanguage)
                  ? element.textareaLanguage
                  : 'textile'
              }
              minHeight="100px"
              className="text-[12px]"
            />
          </label>
        )}

        {selective && (
          <label className="flex flex-col gap-1.5 pl-6">
            <span className={labelCls}>{t('repos.elementSelectValues')}</span>
            {/* 可自定义多选（对齐旧版 tags Select）：options 即当前值本身（自引用），
                输入新值回车直接加入 chip，也支持从列表勾选/取消 */}
            <SearchableSelect
              multiple
              creatable
              value={element.selectValues}
              options={element.selectValues.map((v) => ({ value: v, label: v }))}
              onChange={(v) =>
                onChange({ selectValues: Array.isArray(v) ? v : [v] })
              }
              placeholder={t('repos.elementSelectValuesPlaceholder')}
              searchPlaceholder={t('repos.elementSelectValuesSearchPlaceholder')}
              emptyText={t('repos.elementSelectValuesEmpty')}
              createText={(q) => t('repos.elementSelectValuesCreate', { name: q })}
            />
          </label>
        )}
      </div>
    </div>
  )
}

/** 按类型渲染默认值控件：Switch → true/false，InputNumber → 数字框，TextArea → 多行 */
function DefaultValueInput({
  type,
  value,
  onChange,
}: {
  type: Element['type']
  value: string
  onChange: (v: string) => void
}) {
  if (type === 'ElementTypeInputNumber') {
    return (
      <Input
        type="number"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    )
  }
  if (type === 'ElementTypeSwitch') {
    return (
      <Select value={value === 'true' ? 'true' : 'false'} onValueChange={onChange}>
        <SelectTrigger className="w-full">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="false">false</SelectItem>
          <SelectItem value="true">true</SelectItem>
        </SelectContent>
      </Select>
    )
  }
  return <Input value={value} onChange={(e) => onChange(e.target.value)} />
}
