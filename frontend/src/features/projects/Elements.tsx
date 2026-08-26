import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import type { components } from '@/api/schema'
import { Input } from '@/components/ui/shadcn/input'
import { Switch } from '@/components/ui/shadcn/switch'
import { CodeEditor, FILE_TYPES } from '@/components/CodeEditor'
import { SearchableSelect } from '@/components/SearchableSelect'
import { nextZIndex } from '@/lib/zIndex'
import { buildSections, groupLabel } from '@/lib/groupSections'
import type { GroupSection } from '@/lib/groupSections'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'
import { Icon } from '@/components/Icons'

type Element = components['schemas']['mars.Element']
type GroupSetting = components['schemas']['mars.GroupSetting']
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
 * variant 过滤：'compact' 排除 TextArea（部署参数网格用，长文本块下移底部 tab）、
 * 'all' 不过滤（CreateProjectModal 等）。底部「自定义配置」TextArea 面板不再经本组件渲染——
 * 由 TabEdit 直接以配置文件面板同一套 grid 定高结构渲染（见 TabEdit 底部 tab）。
 */
export function Elements({
  elements,
  value,
  onChange,
  active,
  variant = 'all',
  groupSettings = [],
}: {
  elements: Element[]
  value: ExtraValue[]
  onChange: (value: ExtraValue[]) => void
  /** 当前是否为激活 Tab（keep-alive 隐藏时为 false）：仅 select 型字段随 active 翻转重建，
   *  强制关闭可能开着的 SearchableSelect 弹层——portal 挂 body，display:none 裁不掉，否则
   *  切走 tab 残留成幽灵下拉。其余字段（textarea CodeEditor/折叠态等）保持挂载不丢状态 */
  active?: boolean
  /** 字段过滤：'compact' 不含 TextArea / 'all' 全部（默认） */
  variant?: 'all' | 'compact'
  /** 分区配置（order/collapsed）：决定分区展示顺序与默认折叠；空 = 无分区平铺（兼容旧版） */
  groupSettings?: GroupSetting[]
}) {
  // variant 过滤：'compact' 排除 TextArea / 'all' 全部（默认）
  const visible = useMemo(
    () =>
      elements.filter((element) =>
        variant === 'compact' ? element.type !== 'ElementTypeTextArea' : true,
      ),
    [elements, variant],
  )
  // 分区派生：编辑页 DynamicElement 与部署表单共用同一份 buildSections，保证两页分组/顺序一致
  const sections = useMemo(
    () => buildSections(visible, groupSettings),
    [visible, groupSettings],
  )
  // path → 取值映射（命中值为空时回退默认值）
  const map = useMemo(() => {
    const m = new Map<string, string>()
    for (const v of value) m.set(v.path, v.value)
    return m
  }, [value])

  /** 更新指定 path 的取值（保留其余项），统一转字符串存储 */
  const update = (path: string, raw: unknown) => {
    const next = value.filter((v) => v.path !== path)
    next.push({ path, value: String(raw) })
    onChange(next)
  }

  /** 单个字段 key：select 型拼 active（keep-alive 失活重建弹层、复位 open），其余用 path 保
   *  实例（textarea CodeEditor/折叠态/滚动位置）；select 型弹层是点击打开的 portal（挂 body），
   *  display:none 裁不掉，切走不重建会残留幽灵下拉 */
  const fieldKey = (element: Element) =>
    element.type === 'ElementTypeSelect' || element.type === 'ElementTypeNumberSelect'
      ? `${element.path}-${active ? 'on' : 'off'}`
      : element.path

  if (elements.length === 0) return null

  // 无分区：兼容旧版平铺（无分区头/无折叠）
  if (sections.length === 0) {
    return (
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-3">
        {visible.map((element) => (
          <ElementField
            key={fieldKey(element)}
            element={element}
            display={map.get(element.path) ?? element.default ?? ''}
            update={update}
          />
        ))}
      </div>
    )
  }

  // 分区模式：所有分区收敛进单个面板容器，分区间以 border-top 分隔（.block + .block），
  // 头栏对齐设计稿 .block-title.collapsible：主题色 accent 竖条 + 分区名 + 计数 badge + 右侧 chevron。
  // 未分组恒排最后、不可折叠（无 accent/chevron）。容器无边框（用户要求 borderless，去掉卡片感）
  return (
    <div className="rounded-lg bg-surface px-5 py-2">
      {sections.map((section, i) => (
        <ElementSection
          key={section.name}
          section={section}
          map={map}
          update={update}
          fieldKey={fieldKey}
          isFirst={i === 0}
        />
      ))}
    </div>
  )
}

/**
 * 分区块（部署表单面板内）：头栏 = 主题色竖条（accent）+ 分区名 + 计数 badge + 右侧 chevron，
 * 整个标题行可点击折叠，对齐 oms-deploy-config-redesign.html 的 .block-title.collapsible——
 * accent 3px/12px 主题色、badge 灰底圆角、chevron margin-left:auto 贴右、折叠时 rotate(-90deg)。
 * 块间以 border-top 分隔（.block + .block），不再各自成独立卡片。默认折叠取 group_settings.collapsed
 * （进入页面时该分区收起），之后为本地瞬态——部署表单的折叠只是浏览态、不写回配置（与编辑页
 * DynamicElement 的持久化折叠语义不同）。未分组区恒排最后、不可折叠。
 */
function ElementSection({
  section,
  map,
  update,
  fieldKey,
  isFirst,
}: {
  section: GroupSection
  map: Map<string, string>
  update: (path: string, raw: unknown) => void
  fieldKey: (element: Element) => string
  /** 面板内首块不加分隔线（.block + .block 语义：首块不画顶边） */
  isFirst?: boolean
}) {
  const { t } = useTranslation()
  const isUngrouped = section.name === ''
  // 默认折叠 = 初始收起（useState 惰性初始化，仅首次挂载生效）；keep-alive 切走/切回保留用户瞬态
  const [open, setOpen] = useState(() => !section.collapsed)
  // 头栏公共内容：accent 竖条（未分组无）+ 分区名 + 计数 badge
  const headerInner = (
    <>
      {!isUngrouped && <span className="h-3 w-[3px] shrink-0 rounded-[2px] bg-primary" />}
      <span
        className={cn(
          'min-w-0 truncate text-[13px] font-semibold group-hover:text-primary',
          isUngrouped ? 'text-mute' : 'text-ink',
        )}
      >
        {isUngrouped ? t('repos.groupUngrouped') : groupLabel(section.name, t)}
      </span>
      <span className="shrink-0 rounded-full bg-secondary px-2 py-0.5 text-[11px] font-normal text-mute">
        {t('repos.groupsItemCount', { count: section.elements.length })}
      </span>
    </>
  )
  return (
    <div className={cn(!isFirst && 'border-t border-line')}>
      {isUngrouped ? (
        // 未分组：纯展示头（无折叠），不加可点击/焦点交互
        <div className="flex items-center gap-2 py-3">{headerInner}</div>
      ) : (
        // 可折叠块：整个标题行即按钮（对齐设计稿 .block-title.collapsible 整体点击语义）
        <button
          type="button"
          onClick={() => setOpen((o) => !o)}
          aria-expanded={open}
          title={open ? t('project.collapse') : t('project.expand')}
          className="group flex w-full cursor-pointer items-center gap-2 rounded py-3 text-left select-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
        >
          {headerInner}
          <Icon
            name="chevron-down"
            className={cn(
              'ml-auto size-4 shrink-0 text-ink transition-transform duration-200 group-hover:text-primary',
              !open && '-rotate-90',
            )}
          />
        </button>
      )}
      {open && (
        // pb-3 对齐头栏 py-3：面板 py-2 对称后，上距=pt-2(8)+头栏顶(12)=20、下距=内容pb-3(12)+pb-2(8)=20，两侧一致
        <div className="pb-3">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-3">
            {section.elements.map((element) => (
              <ElementField
                key={fieldKey(element)}
                element={element}
                display={map.get(element.path) ?? element.default ?? ''}
                update={update}
              />
            ))}
          </div>
        </div>
      )}
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
              title={collapsed ? t('project.expand') : t('project.collapse')}
              className="shrink-0 rounded text-faint transition-colors hover:text-ink focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
            >
              <Icon
                name="chevron-down"
                className={cn('size-3.5 transition-transform', collapsed && '-rotate-90')}
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
              // 限高对齐底部配置栏（TabEdit 配置 CodeEditor minHeight=500px）：内容超长时内部滚动，不再无限撑高
              maxHeight="500px"
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
            className="h-8 min-w-0 flex-1 text-[13px]!"
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
            className="h-8 min-w-0 flex-1 text-[13px]!"
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
            className="min-w-0 min-h-8 flex-1 px-2 py-1 text-[13px]"
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
                className="flex cursor-pointer items-center gap-1.5 text-[13px] text-ink"
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

  // label 可读性：10px + faint 太细看不清，提为 12px + medium 加粗 + mute 加深对比；
  // 与 13px 的控件值保持「值略大、label 略粗」的双重层级，一眼能分 label/value
  const labelCls = cn(
    'min-w-0 text-[12px] font-medium text-mute',
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
