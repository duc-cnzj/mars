import {
  HighlightStyle,
  syntaxHighlighting,
  syntaxTree,
} from '@codemirror/language'
import {
  Decoration,
  EditorView,
  ViewPlugin,
  type DecorationSet,
  type ViewUpdate,
} from '@codemirror/view'
import { RangeSetBuilder } from '@codemirror/state'
import { tags as t } from '@lezer/highlight'

/**
 * CodeMirror 主题：直接使用官方 prism-material-dark.css（node_modules/prism-themes，
 * 已在 main.tsx 全局引入，DiffViewer 同款）。
 *
 * 做法：把 CodeMirror 高亮标签映射成 Prism 的 .token.<class> DOM 类名，让官方 CSS 上色，
 * 与旧版 Prism 渲染逐 token 一致（不在此处抄色板，避免漂移）。
 *
 * 注意：不要用 @uiw/codemirror-themes 的 createTheme —— 它的 @codemirror/language
 * 预打包与主应用分叉，syntaxHighlighting 注册进另一份 highlighterFacet，不生效。
 * 这里直接 import @codemirror/language 才能注册成功。
 */
const chrome = EditorView.theme(
  {
    // 对齐 react-diff-viewer 暗色 chrome（与 DiffViewer 同色系，配色直接取 styles.js 暗色调板）：
    //   diffViewerBackground #2e303c / gutterBackground #2c2f3a / gutterColor #464c67 / gutterBackgroundDark #262933
    // 让编辑器与 DiffViewer 并排时底色、gutter 两色调、行号色一致，不再是一块偏灰的 #2f2f2f。
    // （注意：编辑器和 diff 同底后，TabEdit 两块面板靠接缝的边框/间隔区分，不靠底色差异。）
    '&': { backgroundColor: '#2e303c', color: '#eee' },
    '.cm-gutters': { backgroundColor: '#2c2f3a', color: '#464c67' },
    '.cm-activeLineGutter': { backgroundColor: '#262933', color: '#8f98bb' },
    '.cm-content': { caretColor: '#eee' },
    '.cm-cursor, .cm-dropCursor': { borderLeftColor: '#eee' },
    // 选中高亮：官方 prism 的 #363636 与编辑器底 #2e303c 几乎同深（选中看不见）。
    // 先用 VS Code #264f78 用户仍觉不明显，升级为 GitHub 暗色选中蓝 #2f81f7——
    // 更亮的饱和蓝，对 #eee 白字与 #2e303c 底都清晰可辨
    '&.cm-focused .cm-selectionBackground, & .cm-line::selection, & .cm-selectionLayer .cm-selectionBackground, .cm-content ::selection':
      { backgroundColor: '#2f81f7 !important' },
    // 命中匹配高亮：比主选中暗一档的蓝灰，与 #363636 同级的旧值也近不可见，一并提亮
    '& .cm-selectionMatch': { backgroundColor: '#3a4a6b' },
    '&.cm-editor.cm-focused': { outline: 'none' },
    // 搜索面板保持 @codemirror/search 默认：sticky 横条贴在编辑器底部（top:false 的面板容器
    // .cm-panels.cm-panels-bottom），不自定义定位/卡片样式，避免和编辑器内容抢视觉。
    // 只加宽查找/替换输入框：基座 .cm-textfield 无 width，走浏览器默认（~177px）太窄。
    // min(220px, 50vw) 兼顾窄编辑器不横向溢出面板
    '.cm-panel.cm-search .cm-textfield': {
      width: 'min(220px, 50vw)',
    },
  },
  { dark: true },
)

const highlightStyle = HighlightStyle.define([
  // 灰：#616161
  { tag: [t.comment, t.lineComment, t.blockComment, t.processingInstruction, t.documentMeta, t.meta], class: 'token comment' },
  // 紫：#c792ea（对齐 Prism：key→atrule、keyword、boolean(important)、function 等）
  { tag: [t.keyword, t.operatorKeyword, t.moduleKeyword, t.controlKeyword, t.definitionKeyword], class: 'token keyword' },
  { tag: [t.bool], class: 'token boolean important' },
  { tag: [t.atom, t.constant(t.name), t.special(t.atom)], class: 'token constant' },
  { tag: [t.definition(t.propertyName)], class: 'token atrule' },
  { tag: [t.function(t.variableName), t.function(t.definition(t.variableName)), t.labelName], class: 'token function' },
  // 绿：#a5e844
  { tag: [t.string, t.special(t.string), t.attributeValue, t.regexp], class: 'token string' },
  // 橙：#fd9170（裸标量经 yamlScalarHighlight 按内容判断）
  { tag: [t.number, t.integer, t.float, t.unit], class: 'token number' },
  // 青：#80cbc4（JSON 等语言的 key 用 property，对齐 Prism json）
  { tag: [t.propertyName], class: 'token property' },
  // 橘：#ffcb6b
  { tag: [t.attributeName, t.standard(t.name), t.standard(t.tagName)], class: 'token attr-name' },
  // 黄：#f2ff00
  { tag: [t.className, t.typeName, t.definition(t.className), t.namespace], class: 'token class-name' },
  // 红：#ff6666
  { tag: [t.variableName, t.definition(t.variableName), t.tagName, t.deleted, t.invalid, t.link], class: 'token variable' },
  // 蓝：#89ddff（冒号/逗号/标点对齐 Prism punctuation，运算符对齐 operator）
  { tag: [t.punctuation, t.separator, t.bracket, t.squareBracket, t.angleBracket], class: 'token punctuation' },
  { tag: [t.operator], class: 'token operator' },
])

export const materialDark = [chrome, syntaxHighlighting(highlightStyle)]

/* ------------------------------------------------------------------ */
/* YAML 裸值高亮：lezer-yaml 把所有未加引号的标量（数字/布尔/null/普通字符串）
 * 都打上 content 标签，无法用标签区分。这里复刻 Prism yaml 的值正则，
 * 对 Literal 节点按内容补上 number / boolean 类（普通字符串保持默认 #eee，
 * 与旧版 Prism 渲染一致）。 */
const YAML_NUM =
  /^[+-]?(?:0x[\da-f]+|0o[0-7]+|(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?|\.inf|\.nan)$/i
const YAML_BOOL = /^(?:true|false)$/i
const YAML_NULL = /^(?:null|~)$/i
const YAML_DATE =
  /^(?:\d{4}-\d\d?-\d\d?(?:[tT]|[ \t]+)\d\d?:\d{2}:\d{2}(?:\.\d*)?(?:[ \t]*(?:Z|[-+]\d\d?(?::\d{2})?))?|\d{4}-\d{2}-\d{2}|\d\d?:\d{2}(?::\d{2}(?:\.\d*)?)?)$/

function buildScalarDeco(view: EditorView): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>()
  const doc = view.state.doc
  const tree = syntaxTree(view.state)
  // 只处理可见行（与 TreeHighlighter 一致），避免大文件每次 viewport 变化都遍历整棵语法树。
  // 节点 clamp 到可见区段，保证多段 visibleRanges 下不重复 add。
  for (const { from, to } of view.visibleRanges) {
    tree.iterate({
      from,
      to,
      enter(node) {
        const name = node.type.name
        let cls: string | null = null
        if (name === 'BlockLiteralContent') {
          // 块标量（| >）内容：Prism 里 scalar 别名 string → 绿
          cls = 'token string'
        } else if (name === 'Literal') {
          const text = doc.sliceString(node.from, node.to)
          if (YAML_BOOL.test(text) || YAML_NULL.test(text)) {
            cls = 'token boolean important'
          } else if (YAML_NUM.test(text) || YAML_DATE.test(text)) {
            cls = 'token number'
          }
        }
        if (cls) {
          const f = Math.max(node.from, from)
          const t = Math.min(node.to, to)
          if (f < t) builder.add(f, t, Decoration.mark({ class: cls }))
        }
        return name === 'Literal' || name === 'BlockLiteralContent'
          ? false
          : undefined
      },
    })
  }
  return builder.finish()
}

const yamlScalarHighlight = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet
    constructor(view: EditorView) {
      this.decorations = buildScalarDeco(view)
    }
    update(update: ViewUpdate) {
      if (
        update.docChanged ||
        update.viewportChanged ||
        syntaxTree(update.startState) !== syntaxTree(update.state)
      ) {
        this.decorations = buildScalarDeco(update.view)
      }
    }
  },
  { decorations: (v) => v.decorations },
)

export { yamlScalarHighlight }
