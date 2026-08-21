import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import type { components } from '../../api/schema'
import { Input } from '@/components/ui/shadcn/input'
import { Switch } from '@/components/ui/shadcn/switch'
import { CodeEditor, FILE_TYPES } from '@/components/CodeEditor'
import { SearchableSelect } from '@/components/SearchableSelect'
import { nextZIndex } from '@/hooks/useDraggableDialog'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'
import { ChevronDown } from 'lucide-react'

type Element = components['schemas']['mars.Element']
type ExtraValue = components['schemas']['websocket.ExtraValue']

/** 与旧版 Elements.tsx 一致的布尔判定：1/true/"1"/"True"/"true" 视为开 */
const isTrue = (v: unknown): boolean =>
  v === 1 || v === true || v === '1' || v === 'True' || v === 'true'

/**
 * 动态部署参数表单：按 mars.Element 类型渲染对应控件（Input/CodeEditor/InputNumber/
 * Select/Radio/Switch），值写入 websocket.ExtraValue[]（path 键控，保留其他项）。
 * 布局：响应式 3 列网格；除多行（label 换行、全宽 CodeMirror，label 后可折叠，
 * 语言由 element.textareaLanguage 指定、不匹配支持类型时回退 textile）外，其余控件
 * label 与控件同行（inline）；行内 label 超长收缩省略号、hover 完整 tooltip。
 */
export function Elements({
  elements,
  value,
  onChange,
}: {
  elements: Element[]
  value: ExtraValue[]
  onChange: (value: ExtraValue[]) => void
}) {
  const rows = useMemo(() => {
    const map = new Map<string, string>()
    for (const v of value) map.set(v.path, v.value)
    return elements.map((element): { element: Element; display: string } => ({
      element,
      display: map.get(element.path) ?? element.default ?? '',
    }))
  }, [elements, value])

  /** 更新指定 path 的取值（保留其余项），统一转字符串存储 */
  const update = (path: string, raw: unknown) => {
    const next = value.filter((v) => v.path !== path)
    next.push({ path, value: String(raw) })
    onChange(next)
  }

  if (elements.length === 0) return null

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-3">
      {rows.map(({ element, display }) => (
        <ElementField
          key={element.path}
          element={element}
          display={display}
          update={update}
        />
      ))}
    </div>
  )
}

/** 单个字段：按类型分发到 shadcn 控件 */
function ElementField({
  element,
  display,
  update,
}: {
  element: Element
  display: string
  update: (path: string, raw: unknown) => void
}) {
  const { t } = useTranslation()
  const label = element.description || element.path
  // 每个字段用 path 作 id，FieldLabel 以 htmlFor 关联控件（label 可点击聚焦）
  const fieldId = element.path
  // 多行（CodeMirror）折叠开关：key=path 的实例独立持态，切换时值保留在 value 数组里
  const [collapsed, setCollapsed] = useState(false)
  // element.type 为 schema 里的 string 枚举（仅类型声明，无运行时产物），故按字符串比较
  switch (element.type as string) {
    case 'ElementTypeTextArea':
      return (
        <div className="md:col-span-3">
          <div className="mb-1 flex items-center gap-1.5">
            <FieldLabel label={label} truncate className="min-w-0 shrink" />
            <button
              type="button"
              onClick={() => setCollapsed((c) => !c)}
              aria-expanded={!collapsed}
              title={collapsed ? '展开' : '折叠'}
              className="shrink-0 rounded text-faint transition-colors hover:text-ink focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
            >
              <ChevronDown
                size={14}
                className={cn('transition-transform', collapsed && '-rotate-90')}
              />
            </button>
          </div>
          {!collapsed && (
            <CodeEditor
              value={display}
              onChange={(v) => update(element.path, v)}
              className="text-[12px]"
              // 多行编辑器语言由后端 textarea_language 字段指定；不在 CodeEditor 支持的类型
              // 集里（含空串/未知自由值）回退 textile（旧版 env/dotenv 使用的 mode）
              language={
                (FILE_TYPES as readonly string[]).includes(element.textareaLanguage)
                  ? element.textareaLanguage
                  : 'textile'
              }
              minHeight="72px"
            />
          )}
        </div>
      )
    case 'ElementTypeInputNumber':
      return (
        <div className="flex items-center gap-2">
          <FieldLabel label={label} htmlFor={fieldId} truncate className="max-w-[45%]" />
          <Input
            id={fieldId}
            type="number"
            value={display}
            onChange={(e) => update(element.path, e.target.value)}
            className="h-8 min-w-0 flex-1 text-[12px]!"
          />
        </div>
      )
    case 'ElementTypeInput':
      return (
        <div className="flex items-center gap-2">
          <FieldLabel label={label} htmlFor={fieldId} truncate className="max-w-[45%]" />
          <Input
            id={fieldId}
            value={display}
            onChange={(e) => update(element.path, e.target.value)}
            className="h-8 min-w-0 flex-1 text-[12px]!"
          />
        </div>
      )
    case 'ElementTypeSelect':
    case 'ElementTypeNumberSelect':
      return (
        <div className="flex items-center gap-2">
          <FieldLabel label={label} truncate className="max-w-[45%]" />
          {/* 走项目级可搜索 Select：选项按关键词过滤 + 最多渲染 MAX_RENDERED，弹层 z-index 打开时自愈盖过宿主弹窗 */}
          <SearchableSelect
            value={display}
            options={element.selectValues.map((sv) => ({ value: sv, label: sv }))}
            onChange={(v) => update(element.path, v as string)}
            placeholder={label}
            searchPlaceholder={t('common.search')}
            emptyText={t('common.empty')}
            className="min-w-0 min-h-8 flex-1 px-2 py-1 text-[12px]"
          />
        </div>
      )
    case 'ElementTypeRadio':
    case 'ElementTypeNumberRadio':
      return (
        <div className="flex items-center gap-2">
          <FieldLabel label={label} truncate className="max-w-[45%]" />
          <div className="flex min-w-0 flex-wrap gap-3">
            {element.selectValues.map((sv) => (
              <label
                key={sv}
                className="flex cursor-pointer items-center gap-1.5 text-[12px] text-ink"
              >
                <input
                  type="radio"
                  name={element.path}
                  value={sv}
                  checked={display === sv}
                  onChange={() => update(element.path, sv)}
                />
                {sv}
              </label>
            ))}
          </div>
        </div>
      )
    case 'ElementTypeSwitch':
      return (
        <div className="flex items-center gap-2">
          <FieldLabel label={label} htmlFor={fieldId} truncate className="max-w-[45%]" />
          <Switch
            id={fieldId}
            checked={isTrue(display)}
            onCheckedChange={(v) => update(element.path, v)}
          />
        </div>
      )
    default:
      return null
  }
}

/**
 * 字段 label：truncate 模式支持自适应——短 label 按内容宽，容器收窄时收缩省略号，
 * hover（或键盘聚焦）出完整文本 tooltip（仅截断时弹，同 ProjectRow 全名 tooltip）。
 * block 模式（多行字段的全宽 label）不加 truncate，长文本正常换行。
 */
function FieldLabel({
  label,
  htmlFor,
  className,
  truncate = false,
}: { label: string; htmlFor?: string; className?: string; truncate?: boolean }): ReactNode {
  const ref = useRef<HTMLLabelElement>(null)
  const [truncated, setTruncated] = useState(false)
  const [tipHover, setTipHover] = useState(false)
  // tooltip content 是 portal 挂 body，须盖过可拖拽宿主弹窗的动态 z-index（z-51+）：
  // 初始 50 兜底，hover 打开时取下一个共享 z（同 AreaSpark 单位提示/强杀确认框机制）
  const [tipZ, setTipZ] = useState(50)

  // 截断检测：scrollWidth > clientWidth（+1px 缓冲防亚像素抖动）；ResizeObserver 在 label/容器尺寸变化时重算。
  // truncate 模式恒为 Tooltip 分支（open 受控，见下方），ref 不随截断状态切换 DOM，
  // observer 一直挂在同一 label 上，无需重绑。
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const measure = () => setTruncated(el.scrollWidth > el.clientWidth + 1)
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(el)
    return () => ro.disconnect()
  }, [label])

  const labelCls = cn(
    'min-w-0 text-[10px] text-faint',
    truncate && 'shrink truncate',
    className,
  )

  // 非 truncate 模式直接渲染 label，不引入 TooltipProvider。
  // truncate 模式恒走 Tooltip 分支、open 受控（truncated && hover，onOpenChange 置空）：
  // 短 label 或未截断时 open 恒 false、content 不挂载，行为同裸 label；截断后 hover 弹完整文本。
  // 对齐 ProjectRow 全名 tooltip 的可用模式——不要按 truncated 在裸 label ↔ Tooltip 间切换 DOM，
  // 也不要给 onOpenChange 接真实 setState（Radix 受控开合逻辑会反过来和鼠标事件打架，工具提示一开就关）。
  if (!truncate) {
    return (
      <label ref={ref} htmlFor={htmlFor} className={labelCls}>
        {label}
      </label>
    )
  }

  return (
    <TooltipProvider delayDuration={100}>
      <Tooltip open={truncated && tipHover} onOpenChange={() => {}}>
        <TooltipTrigger asChild>
          <label
            ref={ref}
            htmlFor={htmlFor}
            onMouseEnter={() => {
              // 仅截断时 tooltip 才会真正打开才需要抬 z-index；未截断时每次 hover
              // 都 bump 全局共享 zCounter（useDraggableDialog）并触发重渲染，纯浪费
              if (truncated) setTipZ(nextZIndex())
              setTipHover(true)
            }}
            onMouseLeave={() => setTipHover(false)}
            className={labelCls}
          >
            {label}
          </label>
        </TooltipTrigger>
        <TooltipContent
          side="top"
          // inline style 覆盖基座 z-50（盖过宿主弹窗动态 z）+ 防御性去掉 text-wrap:balance。
          // 注意必须写长属性 textWrapStyle（text-wrap-style）而不是 shorthand textWrap（text-wrap:normal）：
          // text-balance 基座类写的是 longhand text-wrap-style:balance，Chrome 对「inline shorthand 压
          // class longhand」有交互怪癖，computed 仍算出 balance（无头 Chrome 实测）；写同长属性才能赢。
          style={{ zIndex: tipZ, textWrapStyle: 'auto' }}
          className="max-w-[min(320px,80vw)] break-words"
        >
          {label}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
