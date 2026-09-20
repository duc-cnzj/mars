import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import ReactDiffViewer, {
  type ReactDiffViewerStylesOverride,
} from 'react-diff-viewer'
import { toast } from '@/lib/toast'
import { cn } from '@/lib/utils'
import { getHighlightSyntax } from '@/lib/highlight'
import { copyText } from '@/lib/copy'
import { Icon } from './Icons'
import { Button } from '@/components/ui/shadcn/button'

type ViewMode = 'split' | 'unified'

/** 行内代码基座：fontFamily/lineHeight 对齐外面配置编辑器（CodeMirror monospace + 1.4）。
 *  库默认是 25px 行高 + Tailwind Preflight 的 ui-monospace 栈，与配置区观感不一致。
 *  写成 as const 的窄字面量：它会 spread 进 emotion 的 CSSObject（pre 键），
 *  标成 React.CSSProperties 那种宽类型时 TS 会判定该嵌套对象不是 Interpolation */
const codeBase = { lineHeight: 1.4, fontFamily: 'monospace' } as const

/** 容器基座：table 高度填满外层滚动容器（min-h-0 grow overflow-auto 有定高时生效）；
 *  内容超高的部分从 table 溢出、由外层容器接管垂直滚动，内容不足时 table 撑满不留空。
 *  同样用 as const：它会 spread 进 emotion 的 CSSObject，宽类型会被判定为不可赋值 */
const containerBase = { display: 'block', width: '100%', height: '100%', overflowX: 'auto' } as const

/** 选中高亮与 CodeMirror 对齐（同一套 #2f81f7，见 lib/prism-material-dark.ts 的选中蓝）：
 *  全局 ::selection 是品牌色 30% 淡底，落在此处暗底 #2e303c 上偏淡，还把选中文字染成
 *  var(--text)（默认浅色主题下近黑）——而本组件恒为暗底白字，就成了黑字压蓝底。
 *  覆写按组件作用域（emotion 编成 `.css-哈希 ::selection`），只命中 diff 内容。
 *  文字色必须写死暗色主题的 diffViewerColor #FFF：不能用 color: inherit——Blink 的高亮
 *  伪元素继承自根元素而非宿主元素，inherit 取回的是 body 的 var(--text)，等于没改。
 *  写死白字也正合编辑器行为：只换底色，文字保持内容本身颜色。 */
const selectionRule = { '& ::selection': { backgroundColor: '#2f81f7', color: '#FFF' } }

/** 统一视图样式（旧版 defaultStyle，还原旧版视觉）。
 *  单列，长行不折行：table 占满宽度、超宽内容由 table 自身横向滚动（display:block 才有
 *  overflow），行高恒定。
 *  库默认 pre{white-space:pre-wrap} 会把窄列里的长 diff 行折成两行，行高变高、看起来"多出一行" */
const defaultStyle: ReactDiffViewerStylesOverride = {
  gutter: { padding: '0 5px', minWidth: 25 },
  marker: { padding: '0 6px' },
  diffContainer: {
    ...containerBase,
    ...selectionRule,
    pre: { whiteSpace: 'pre', ...codeBase },
  },
  line: { fontSize: 12 },
}

/** 分屏视图样式：左右严格 5:5，超长行就地折行（以此换取两列恒等宽）。
 *
 *  auto 布局下两个内容列按各自 max-content 抢宽（实测 900px 容器里左 1496px / 右 152px，
 *  即 91%:9%）——左边一句长 yaml 就能把右半屏挤没。把长行折掉之后，内容列的 min-content
 *  降到一个 token 的宽度，库自带的 `.splitView .content{width:50%}` 才真正生效，两列恒为
 *  1:1（实测 900px 容器里 406:406）。不折行那版只能反过来给表格写死总宽、把横向滚动推给
 *  外层容器（表格固定布局 + 2×最长行的总宽），代价是长行看不全，故不采用。
 *
 *  折行必须挂在容器 pre 上：内容 pre 里还有一层 renderContent 自带的 `<pre style="display:
 *  inline">`，它不是 contentText，只有 `容器 pre` 这条规则能命中它——挂在 contentText 上时
 *  内层 pre 仍是 white-space:pre，单元格 min-content 照样被撑到 1496px，白改。
 *  库默认 line 无 word-break，长 token（URL/超长注解）不折，故显式 overflow-wrap:anywhere。
 *  行号列被这条容器规则一并放开折行，要单独用 lineNumber 压回去——实测 320px 窄容器下不压的
 *  话三位数行号会被折成两行。压的手段是嵌套 `&&`：emotion 会把类名写两遍编成 `.css-hash.css-hash`
 *  （0,2,0），比容器的 `.css-hash pre`（0,1,1）高，无需 `!important`——csstype 不接受带
 *  `!important` 的 whiteSpace 值，直接写会过不了类型检查。 */
const splitStyle: ReactDiffViewerStylesOverride = {
  ...defaultStyle,
  diffContainer: {
    ...containerBase,
    ...selectionRule,
    pre: { whiteSpace: 'pre-wrap', overflowWrap: 'anywhere', ...codeBase },
  },
  lineNumber: { '&&': { whiteSpace: 'pre' } },
}

/**
 * 通用 diff 查看器（沿用旧版 react-diff-viewer 组件）：
 * split/unified 切换 + Prism 语法高亮 + copy old/new + 仅显示变更开关。
 * 明暗跟随当前主题（旧版固定 useDarkTheme，这里按 theme.mode 自适应）。
 * 对外 API 与旧包装一致：{ oldValue, newValue, language }。
 */
export function DiffViewer({
  oldValue,
  newValue,
  language,
  className,
  initialView = 'unified',
  hideToolbar = false,
  viewportClassName,
}: {
  oldValue: string | null | undefined
  newValue: string | null | undefined
  language?: string
  className?: string
  /** 初始视图；打开新 diff（old/new 变化）时重置回该值。
   *  默认 unified 对齐旧版 config 历史（恒分屏关闭）；events 改动传 split、纯增/删传 unified。 */
  initialView?: ViewMode
  /** 隐藏工具栏（分屏/统一、仅变更、复制）：diff 区只显示内容。配置更新页用。 */
  hideToolbar?: boolean
  /** 内部滚动容器的额外类。默认 rounded-md（独立 diff 卡片圆角）；
   *  与编辑器无缝并排时传 rounded-l-none 把接缝侧圆角拉平（根节点 overflow-hidden 裁不掉子元素内部圆角缺口）。 */
  viewportClassName?: string
}) {
  const { t } = useTranslation()
  const [view, setView] = useState<ViewMode>(initialView)
  const [diffOnly, setDiffOnly] = useState(true)
  // 指向 react-diff-viewer 实例：手动点开折叠块后，重开「仅显示变更」时调 resetCodeBlocks
  // 清空 expandedBlocks，把未变更代码重新折叠（库把手动展开记在内部 state，toggle 不会自动收回）
  const diffRef = useRef<ReactDiffViewer>(null)

  // 复用组件打开新内容时回到初始视图，避免残留上次 split/unified 切换（old/new 仅作触发）
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    setView(initialView)
  }, [initialView, oldValue, newValue])
  // 后端某些事件 old/new 可能为 null：归一化为空串，避免 react-diff-viewer 抛 "should be strings"
  const safeOld = oldValue ?? ''
  const safeNew = newValue ?? ''

  const renderContent = useCallback(
    // react-diff-viewer 会把 undefined 行值传进来，先归一回退空串
    (str: string | null | undefined) => (
      <pre
        style={{ display: 'inline' }}
        dangerouslySetInnerHTML={{ __html: getHighlightSyntax(str, language) }}
      />
    ),
    [language],
  )

  const copy = async (text: string) => {
    const ok = await copyText(text)
    if (ok) toast.success(t('common.copied'))
    else toast.error(t('common.copyFailed'))
  }

  const toolbar = (
    <div className="mb-1.5 flex shrink-0 flex-wrap items-center gap-1.5">
      <div className="flex overflow-hidden rounded-md border border-line">
        <Button
          size="sm"
          variant={view === 'split' ? 'default' : 'ghost'}
          className="h-6 rounded-none px-2 text-[11px]"
          onClick={() => setView('split')}
        >
          {t('diff.split')}
        </Button>
        <Button
          size="sm"
          variant={view === 'unified' ? 'default' : 'ghost'}
          className="h-6 rounded-none px-2 text-[11px]"
          onClick={() => setView('unified')}
        >
          {t('diff.unified')}
        </Button>
      </div>
      <Button
        size="sm"
        variant="outline"
        className="h-6 px-2 text-[11px]"
        onClick={() => {
          const next = !diffOnly
          setDiffOnly(next)
          // 重新开启时重置手动展开的折叠块：否则之前点开的未变更行仍展开，收起不了
          if (next) diffRef.current?.resetCodeBlocks()
        }}
      >
        {t('diff.showDiffOnly')}
        {diffOnly && <Icon name="check" className="ml-1 text-[11px]" />}
      </Button>
      <div className="ml-auto flex gap-1.5">
        {safeOld && (
          <Button
            size="sm"
            variant="outline"
            className="h-6 px-2 text-[11px]"
            onClick={() => copy(safeOld)}
          >
            {t('diff.copyOld')}
          </Button>
        )}
        <Button
          size="sm"
          variant="outline"
          className="h-6 px-2 text-[11px]"
          onClick={() => copy(safeNew)}
        >
          {t('diff.copyNew')}
        </Button>
      </div>
    </div>
  )

  if (!safeOld && !safeNew) {
    return (
      <div className={className}>
        {!hideToolbar && toolbar}
        <div className="px-3 py-4 text-[12px] text-mute">{t('common.empty')}</div>
      </div>
    )
  }

  // 根节点 flex flex-col：diff 区域自包含滚动。外部限高（className 传 h-full）时
  // toolbar 固定、diff 在 grow 内滚动；外部不限高时 grow（basis auto）随内容长高，
  // 由父级滚动容器接手，行为同旧版。
  return (
    <div className={cn('flex min-h-0 flex-col', className)}>
      {!hideToolbar && toolbar}
      <div className={cn('min-h-0 grow overflow-auto', viewportClassName ?? 'rounded-md')}>
        <ReactDiffViewer
          ref={diffRef}
          oldValue={safeOld}
          newValue={safeNew}
          splitView={view === 'split'}
          useDarkTheme
          disableWordDiff
          renderContent={renderContent}
          showDiffOnly={diffOnly}
          styles={view === 'split' ? splitStyle : defaultStyle}
        />
      </div>
    </div>
  )
}
