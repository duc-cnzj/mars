import { useEffect, useRef, useState } from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { materialDark, yamlScalarHighlight } from '@/lib/prism-material-dark'
import { color } from '@uiw/codemirror-extensions-color'
import { yaml } from '@codemirror/lang-yaml'
import { json, jsonParseLinter } from '@codemirror/lang-json'
import { LanguageDescription } from '@codemirror/language'
import { languages } from '@codemirror/language-data'
import { Prec, type Extension } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import {
  autocompletion,
  startCompletion,
  type Completion,
  type CompletionContext,
  type CompletionResult,
} from '@codemirror/autocomplete'
import { linter, type Diagnostic } from '@codemirror/lint'
import * as YAML from 'yaml'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'

import { cn } from '@/lib/utils'

/** 文件类型别名 → 规范语言名（对齐旧版 Player/MyCodeMirror 的 getMode 映射） */
export function getMode(mode: string): string {
  switch (mode) {
    case 'dotenv':
    case 'env':
    case '.env':
      return 'textile'
    case 'js':
      return 'javascript'
    case 'ini':
      return 'properties'
    case 'py':
      return 'python'
    default:
      return mode
  }
}

/**
 * 配置文件类型候选（marsConfig.configFileType 为自由字符串）。
 * 还原旧版 SelectFileType.tsx 的语言列表（约 55 种），供 repo 表单下拉复用。
 */
export const FILE_TYPES = [
  'env',
  'php',
  'json',
  'yaml',
  'go',
  'c',
  'csharp',
  'scala',
  'kotlin',
  'objectiveC',
  'objectiveCpp',
  'dart',
  'cmake',
  'groovy',
  'haskell',
  'dockerfile',
  'http',
  'jinja2',
  'properties',
  'protobuf',
  'puppet',
  'sass',
  'textile',
  'javascript',
  'jsx',
  'typescript',
  'tsx',
  'html',
  'css',
  'python',
  'markdown',
  'xml',
  'sql',
  'mysql',
  'pgsql',
  'java',
  'rust',
  'cpp',
  'shell',
  'lua',
  'swift',
  'vb',
  'powershell',
  'stylus',
  'ruby',
  'erlang',
  'nginx',
  'perl',
  'less',
  'toml',
  'vbscript',
  'coffeescript',
  'julia',
] as const

/** 文件类型 → @codemirror/language-data 里的语言名（缺失时回退 YAML） */
const FILE_TYPE_TO_LANG: Record<string, string> = {
  env: 'Properties files',
  php: 'PHP',
  json: 'JSON',
  yaml: 'YAML',
  go: 'Go',
  c: 'C',
  csharp: 'C#',
  scala: 'Scala',
  kotlin: 'Kotlin',
  objectiveC: 'Objective-C',
  objectiveCpp: 'Objective-C++',
  dart: 'Dart',
  cmake: 'CMake',
  groovy: 'Groovy',
  haskell: 'Haskell',
  dockerfile: 'Dockerfile',
  http: 'HTTP',
  jinja2: 'Jinja',
  properties: 'Properties files',
  protobuf: 'ProtoBuf',
  puppet: 'Puppet',
  sass: 'Sass',
  textile: 'Textile',
  javascript: 'JavaScript',
  jsx: 'JSX',
  typescript: 'TypeScript',
  tsx: 'TSX',
  html: 'HTML',
  css: 'CSS',
  python: 'Python',
  markdown: 'Markdown',
  xml: 'XML',
  sql: 'SQL',
  mysql: 'MySQL',
  pgsql: 'PostgreSQL',
  java: 'Java',
  rust: 'Rust',
  cpp: 'C++',
  shell: 'Shell',
  lua: 'Lua',
  swift: 'Swift',
  vb: 'VBScript',
  powershell: 'PowerShell',
  stylus: 'Stylus',
  ruby: 'Ruby',
  erlang: 'Erlang',
  nginx: 'Nginx',
  perl: 'Perl',
  less: 'LESS',
  toml: 'TOML',
  vbscript: 'VBScript',
  coffeescript: 'CoffeeScript',
  julia: 'Julia',
}

/**
 * 大文档阈值：值超过 1MB 不再挂 CodeMirror。实测 CodeMirror 处理 8-20MB 单次操作在
 * 100ms 内，但 12MB 文档的 Text 结构构建、@uiw/react-codemirror 每次 render 的
 * doc.toString() + 12MB 字符串比较、以及 GC 压力会随渲染次数叠加（打开弹窗时左侧
 * charts 预览同步挂 3 份编辑器），表现为打开/取消弹窗卡顿。
 * 只读场景回退为 <pre>（完整展示，无高亮），可编辑场景回退原生 <textarea>（保留全量编辑）。
 */
const LARGE_DOC_LIMIT = 1024 * 1024 // 1MB

/** 模块级缓存：同一语言只 resolve 一次（language-data 的 load 是异步的） */
const langCache = new Map<string, Promise<Extension>>()

/** 按文件类型解析 CodeMirror 扩展；未知类型回退 YAML */
function resolveLang(fileType: string): Promise<Extension> {
  const cached = langCache.get(fileType)
  if (cached) return cached
  const p = (async () => {
    const name = FILE_TYPE_TO_LANG[fileType] ?? 'YAML'
    if (name === 'JSON') return json()
    if (name === 'YAML') return yaml()
    const desc = LanguageDescription.matchLanguageName(languages, name, true)
    if (!desc) return yaml()
    try {
      const support = await desc.load()
      return support.language
    } catch {
      return yaml()
    }
  })()
  langCache.set(fileType, p)
  return p
}

/** 是否按 YAML 语法处理该文件类型（决定补全与 linter） */
const isYamlType = (fileType: string): boolean =>
  fileType === 'yaml' || FILE_TYPE_TO_LANG[fileType] === 'YAML'

/** 是否按 JSON 语法处理该文件类型（决定 linter） */
const isJsonType = (fileType: string): boolean =>
  fileType === 'json' || FILE_TYPE_TO_LANG[fileType] === 'JSON'

/** 大文档 lint 阈值：超过则跳过，避免 parseDocument 全量解析卡死主线程（yaml 包 10MB 约 16s） */
const YAML_LINT_LIMIT = 100 * 1024 // 100KB

/** yaml 语法校验 linter：用 yaml 包的 parseDocument 收集错误诊断 */
const yamlLinter = linter((view: EditorView): Diagnostic[] => {
  if (view.state.doc.length > YAML_LINT_LIMIT) return []
  try {
    const doc = YAML.parseDocument(String(view.state.doc))
    const err = doc.errors[0]
    if (!err) return []
    const len = view.state.doc.length
    const from = Math.min(err.pos?.[0] ?? 0, len)
    const to = Math.min(err.pos?.[1] ?? from + 1, len)
    return [{ from, to, message: err.message, severity: 'error' }]
  } catch {
    return []
  }
})

/** yaml 模板补全（还原旧版 MyCodeMirror 的 30+ 个 <.Branch>/<.Commit> 变量） */
/** 模板补全列表：label/apply 里的 `<.*>` 变量与英文标识符固定，仅 3 条注释文案走 t 随语言切换 */
function buildYamlCompletions(t: TFunction): Completion[] {
  return [
  {
    apply: '<.ImagePullSecrets>',
    label: '<.ImagePullSecrets>',
    type: 'text',
    info: () => {
      const div = document.createElement('div')
      div.textContent = `- name: secret1\n- name: secret2\n- name: secret3\n`
      return div
    },
  },
  {
    apply: '<.ImagePullSecretsNoName>',
    label: '<.ImagePullSecretsNoName>',
    type: 'text',
    info: () => {
      const div = document.createElement('div')
      div.textContent = `- secret1\n- secret2\n- secret3\n`
      return div
    },
  },
  { apply: '<.Branch>', label: '<.Branch>', type: 'text' },
  { apply: '<.Commit>', label: '<.Commit>', type: 'text' },
  { apply: '<.Pipeline>', label: '<.Pipeline>', type: 'text' },
  { apply: '<.ClusterIssuer>', label: '<.ClusterIssuer>', type: 'text' },
  { apply: '<.Namespace>', label: '<.Namespace>', type: 'text' },
  {
    apply: '<.LongCommit>',
    label: '<.LongCommit>',
    type: 'text',
    detail: `# ${t('repos.yamlCompleteLongCommit')}`,
  },
  { apply: '<.Host1>', label: '<.Host1>', type: 'text' },
  { apply: '<.Host2>', label: '<.Host2>', type: 'text' },
  { apply: '<.Host3>', label: '<.Host3>', type: 'text' },
  { apply: '<.Host4>', label: '<.Host4>', type: 'text' },
  { apply: '<.Host5>', label: '<.Host5>', type: 'text' },
  { apply: '<.Host6>', label: '<.Host6>', type: 'text' },
  { apply: '<.Host7>', label: '<.Host7>', type: 'text' },
  { apply: '<.Host8>', label: '<.Host8>', type: 'text' },
  { apply: '<.Host9>', label: '<.Host9>', type: 'text' },
  { apply: '<.Host10>', label: '<.Host10>', type: 'text' },
  { apply: '<.TlsSecret1>', label: '<.TlsSecret1>', type: 'text' },
  { apply: '<.TlsSecret2>', label: '<.TlsSecret2>', type: 'text' },
  { apply: '<.TlsSecret3>', label: '<.TlsSecret3>', type: 'text' },
  { apply: '<.TlsSecret4>', label: '<.TlsSecret4>', type: 'text' },
  { apply: '<.TlsSecret5>', label: '<.TlsSecret5>', type: 'text' },
  { apply: '<.TlsSecret6>', label: '<.TlsSecret6>', type: 'text' },
  { apply: '<.TlsSecret7>', label: '<.TlsSecret7>', type: 'text' },
  { apply: '<.TlsSecret8>', label: '<.TlsSecret8>', type: 'text' },
  { apply: '<.TlsSecret9>', label: '<.TlsSecret9>', type: 'text' },
  { apply: '<.TlsSecret10>', label: '<.TlsSecret10>', type: 'text' },
  {
    apply: 'cert-manager.io/cluster-issuer: "<.ClusterIssuer>"',
    label: 'certManager',
  },
  {
    apply: `"<.Branch>-<.Pipeline>"`,
    label: 'imageTag',
    detail: '<.Branch>-<.Pipeline>',
  },
  {
    apply: `mars.duc-cnzj.github.io/ignore-containers: "app1,app2" # ${t('repos.yamlCompleteIgnoreContainersApply')}`,
    label: 'annotationIgnoreContainerNames',
    detail: `# ${t('repos.yamlCompleteIgnoreContainers')}`,
    info: () => {
      const div = document.createElement('div')
      div.textContent = `mars.duc-cnzj.github.io/ignore-containers: "app1,app2"`
      return div
    },
  },
  {
    apply: `mars.duc-cnzj.github.io/order-index: "10" # ${t('repos.yamlCompleteOrderIndexApply')}`,
    label: 'annotationPodOrderIndex',
    detail: `# ${t('repos.yamlCompleteOrderIndex')}`,
    info: () => {
      const div = document.createElement('div')
      div.textContent = `mars.duc-cnzj.github.io/order-index: "10"`
      return div
    },
  },
  ]
}

/** 对齐旧版 MyCodeMirror 的编辑器外观微调：行高、gutter 间距、去 focus 外框 */
const editorChrome = EditorView.theme(
  {
    '&': {
      outline: 'none',
      height: '100%',
    },
    '.cm-content': {
      paddingTop: 0,
    },
    '&.cm-editor .cm-scroller .cm-gutters': {
      marginRight: '5px',
    },
    '&.cm-editor.cm-focused': {
      outline: 'none',
    },
    '.cm-completionIcon-text': {
      '&:after': { content: "''", fontSize: '50%', verticalAlign: 'middle' },
    },
    '.cm-line': {
      padding: '1px 0',
    },
  },
  {},
)

/** yaml 模式补全 source：匹配当前词并返回模板项（注释文案随 t 走当前语言） */
function yamlCompletions(t: TFunction) {
  return (context: CompletionContext): CompletionResult | null => {
    const word = context.matchBefore(/\w*/)
    if (!word || (word.from === word.to && !context.explicit)) return null
    return { from: word.from, options: buildYamlCompletions(t) }
  }
}

/** Alt-Enter / Mod-Enter 触发补全（提高优先级覆盖默认 Mod-Enter 换行） */
const completionKeymap = Prec.high(
  keymap.of([
    { key: 'Alt-Enter', run: startCompletion },
    { key: 'Mod-Enter', run: startCompletion },
  ]),
)

/**
 * 解析当前文件类型后追加能力：yaml 补全+linter、json linter、补全快捷键（暗色主题经 theme 属性注入）。
 * 只读预览只挂语法高亮（scalar 高亮 + 语言），不挂 linter/autocomplete/color ——
 * 编辑弹窗里只读大 values.yaml 同步挂载三份编辑器，去掉交互扩展能省掉整棵树的解析与视图抖动。
 * yaml 模板补全（YAML_COMPLETIONS 的 <.*> 变量）仅当 enableYamlTemplateCompletion 显式开启时挂载，
 * 用于 repos 表单的 values.yaml「自动补全: alt+enter」场景；其余 yaml 编辑只保留基础 autocompletion。
 */
function buildExtensions(
  fileType: string,
  langExt: Extension,
  readOnly = false,
  enableYamlTemplateCompletion = false,
  t: TFunction,
): Extension[] {
  const ext: Extension[] = [langExt, editorChrome]
  if (isYamlType(fileType)) ext.push(yamlScalarHighlight)
  if (readOnly) return ext
  ext.push(color, completionKeymap)
  if (isYamlType(fileType)) {
    ext.push(
      yamlLinter,
      enableYamlTemplateCompletion
        ? autocompletion({ override: [yamlCompletions(t)] })
        : autocompletion(),
    )
  } else if (isJsonType(fileType)) {
    ext.push(linter(jsonParseLinter()), autocompletion())
  } else {
    ext.push(autocompletion())
  }
  return ext
}

/**
 * 轻量代码编辑器（CodeMirror 6）：多语言语法高亮 + 行号 + 折叠 + 暗色主题。
 * - yaml：yaml 语法校验；30+ 个 <.*> 模板补全仅当 yamlTemplateCompletion 开启时挂载（Alt-Enter/Mod-Enter 触发）
 * - json：JSON.parse 语法校验
 * 用于项目配置 / repo 的 values.yaml / config 等代码编辑场景。
 * language 传入 marsConfig.configFileType 对应的文件类型（如 php/yaml/json）。
 */
export function CodeEditor({
  value,
  onChange,
  minHeight = '160px',
  readOnly = false,
  language = 'yaml',
  yamlTemplateCompletion = false,
  className,
}: {
  value: string
  onChange: (value: string) => void
  minHeight?: string
  readOnly?: boolean
  language?: string
  /** 仅 repos 表单 values.yaml 场景开启：挂载 YAML_COMPLETIONS 模板补全（对应「自动补全: alt+enter」提示） */
  yamlTemplateCompletion?: boolean
  className?: string
}) {
  // 记录当前已应用到编辑器的语言与语言包。初始即挂载，避免挂载时对同一扩展再 setExtensions
  // 一次触发 CodeMirror 全量 reconfigure（重新解析 + 重新高亮 + 重渲染 DOM，
  // 打开编辑弹窗时 3 份编辑器一起卡顿）。locale 变化也要重建（补全注释文案随语言切换）。
  const { t, i18n } = useTranslation()
  const lastLang = useRef(language)
  const lastLocale = useRef(i18n.language)

  const [extensions, setExtensions] = useState<Extension[]>(() => [
    // 初始扩展按语言取同步扩展：yaml/json 用对应 lang 扩展，其他语言先占位 yaml 等异步 resolve。
    // 否则 configFileType=json 的仓库打开编辑时，编辑器会一直用 yaml 语法。
    ...buildExtensions(
      language,
      isJsonType(language) ? json() : yaml(),
      readOnly,
      yamlTemplateCompletion,
      t,
    ),
  ])

  // yaml/json 扩展同步；language 变化（含 yaml↔json 互切）都要重建扩展，
  // 否则编辑器语言/高亮/lint 不跟随 configFileType。其他语言异步 resolveLang 后替换。
  useEffect(() => {
    const name = FILE_TYPE_TO_LANG[language]
    if (name === 'YAML' || name === 'JSON') {
      if (lastLang.current === language && lastLocale.current === i18n.language) return // 已应用，跳过，避免挂载时重复 reconfigure
      lastLang.current = language
      lastLocale.current = i18n.language
      setExtensions(
        buildExtensions(
          language,
          name === 'JSON' ? json() : yaml(),
          readOnly,
          yamlTemplateCompletion,
          t,
        ),
      )
      return
    }
    let alive = true
    void resolveLang(language).then((ext) => {
      if (!alive) return
      lastLang.current = language
      lastLocale.current = i18n.language
      setExtensions(buildExtensions(language, ext, readOnly, yamlTemplateCompletion, t))
    })
    return () => {
      alive = false
    }
  }, [language, readOnly, i18n.language])

  // 大文档回退：1MB 以上的值不进 CodeMirror（见 LARGE_DOC_LIMIT 注释）。
  // 放在常规渲染之前：大文档立即渲染，避免走 CodeMirror 挂载路径。
  const isLarge = value.length > LARGE_DOC_LIMIT
  if (isLarge) {
    if (readOnly) {
      // 只读大文档：<pre> 完整展示（不截断），无语法高亮
      return (
        <div
          className={cn(
            'flex flex-col overflow-hidden rounded-md border border-line bg-surface text-[12px]',
            className,
          )}
          style={{ height: '100%', minHeight }}
        >
          <pre className="min-h-0 flex-1 overflow-auto whitespace-pre p-2 font-mono leading-relaxed text-ink">
            {value}
          </pre>
        </div>
      )
    }
    // 可编辑（仓库保存的大 values.yaml / config）：回退原生 textarea，保留全量编辑。
    return (
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        spellCheck={false}
        className={cn(
          'w-full resize-none overflow-auto rounded-md border border-line bg-surface p-2 font-mono text-[12px] leading-relaxed text-ink outline-none focus-visible:ring-2 focus-visible:ring-ring/50',
          className,
        )}
        style={{ height: '100%', minHeight }}
      />
    )
  }

  return (
    <CodeMirror
      value={value}
      onChange={onChange}
      extensions={extensions}
      theme={materialDark}
      minHeight={minHeight}
      style={{ height: '100%' }}
      readOnly={readOnly}
      basicSetup={{
        lineNumbers: true,
        highlightActiveLineGutter: false,
        foldGutter: true,
        dropCursor: true,
        allowMultipleSelections: true,
        indentOnInput: true,
        bracketMatching: true,
        closeBrackets: true,
        autocompletion: true,
        rectangularSelection: true,
        crosshairCursor: true,
        highlightActiveLine: false,
        highlightSelectionMatches: true,
        closeBracketsKeymap: true,
        searchKeymap: true,
        foldKeymap: true,
        completionKeymap: true,
        lintKeymap: true,
      }}
      className={cn(
        'overflow-hidden rounded-md border border-line text-[12px]',
        className,
      )}
    />
  )
}
