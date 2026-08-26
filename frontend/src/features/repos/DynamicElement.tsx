import { useTranslation } from 'react-i18next'
import {
  DndContext,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import type { DragEndEvent } from '@dnd-kit/core'
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { components } from '@/api/schema'
import { SearchableSelect } from '@/components/SearchableSelect'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
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

type Element = components['schemas']['mars.Element']

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
 * 自定义配置编辑器（还原旧版 DynamicElement）：mars.Config.elements 的增删/拖拽排序。
 * 每个元素定义部署表单里的一个自定义参数：path / type / description / default / selectValues / order。
 * 拖拽用 dnd-kit（旧版 react-beautiful-dnd 已停止维护）。
 */
export function DynamicElement({
  value,
  onChange,
}: {
  value: Element[]
  onChange: (next: Element[]) => void
}) {
  const { t } = useTranslation()
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
  )

  const update = (i: number, patch: Partial<Element>) => {
    const next = [...value]
    const merged = { ...next[i], ...patch }
    // switch 归一化幂等、任意编辑都安全（default 只会是 'true'/'false'，命中即原样）；
    // InputNumber/选择类型的清理只在类型真正切换时触发，避免打断用户正在输入默认值
    next[i] =
      merged.type === 'ElementTypeSwitch' ||
      (patch.type && merged.type !== next[i].type)
        ? normalizeDefaultForType(merged)
        : merged
    onChange(next)
  }

  const add = () =>
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
      },
    ])

  const remove = (i: number) =>
    onChange(value.filter((_, idx) => idx !== i))

  const onDragEnd = ({ active, over }: DragEndEvent) => {
    if (!over || active.id === over.id) return
    const from = Number(String(active.id).replace('el-', ''))
    const to = Number(String(over.id).replace('el-', ''))
    if (Number.isNaN(from) || Number.isNaN(to)) return
    const next = [...value]
    const [moved] = next.splice(from, 1)
    next.splice(to, 0, moved)
    // 排序后刷新 order，保持与列表顺序一致
    onChange(next.map((el, i) => ({ ...el, order: i })))
  }

  return (
    <div className="flex flex-col gap-2">
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={onDragEnd}
      >
        <SortableContext
          items={value.map((_, i) => `el-${i}`)}
          strategy={verticalListSortingStrategy}
        >
          {value.map((el, i) => (
            <SortableElementItem
              key={`el-${i}`}
              id={`el-${i}`}
              element={el}
              onChange={(patch) => update(i, patch)}
              onRemove={() => remove(i)}
            />
          ))}
        </SortableContext>
      </DndContext>
      <Button variant="outline" className="border-dashed" onClick={add}>
        <Icon name="plus" />
        {t('repos.addElement')}
      </Button>
    </div>
  )
}

/** 单个自定义字段卡片：拖拽手柄 + 字段编辑，可拖拽排序 */
function SortableElementItem({
  id,
  element,
  onChange,
  onRemove,
}: {
  id: string
  element: Element
  onChange: (patch: Partial<Element>) => void
  onRemove: () => void
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id })
  const { t } = useTranslation()
  const selective = SELECTIVE_TYPES.has(element.type)
  const defaultRequired = DEFAULT_REQUIRED_TYPES.has(element.type)

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
        opacity: isDragging ? 0.7 : 1,
        zIndex: isDragging ? 10 : undefined,
      }}
      className={`rounded-lg border p-3 shadow-sm ${
        isDragging ? 'border-primary' : 'border-line'
      } bg-surface`}
    >
      <div className="flex flex-col gap-2.5">
        <div className="flex items-start gap-2">
          <button
            type="button"
            {...attributes}
            {...listeners}
            aria-label={t('repos.elementDrag')}
            className="mt-2.5 cursor-grab text-faint transition-colors hover:text-ink active:cursor-grabbing"
          >
            <Icon name="grip-vertical" className="size-4" />
          </button>
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
