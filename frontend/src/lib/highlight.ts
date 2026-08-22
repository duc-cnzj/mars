import Prism from 'prismjs'
import 'prismjs/components/prism-markup-templating'
import 'prismjs/components/prism-yaml'
import 'prismjs/components/prism-json'
import 'prismjs/components/prism-bash'
import 'prismjs/components/prism-python'
import 'prismjs/components/prism-go'
import 'prismjs/components/prism-php'
import 'prismjs/components/prism-sql'
import 'prismjs/components/prism-java'
import 'prismjs/components/prism-c'
import 'prismjs/components/prism-cpp'
import 'prismjs/components/prism-csharp'
import 'prismjs/components/prism-ruby'
import 'prismjs/components/prism-rust'
import 'prismjs/components/prism-ini'
import 'prismjs/components/prism-properties'
import 'prismjs/components/prism-toml'
import 'prismjs/components/prism-markdown'
import 'prismjs/components/prism-jsx'
import 'prismjs/components/prism-typescript'
import 'prismjs/components/prism-tsx'

/** 文件类型 → Prism 语法名（缺失回退纯文本转义） */
const FILE_TYPE_TO_PRISM: Record<string, string> = {
  yaml: 'yaml',
  json: 'json',
  env: 'properties',
  properties: 'properties',
  shell: 'bash',
  powershell: 'bash',
  python: 'python',
  py: 'python',
  go: 'go',
  golang: 'go',
  php: 'php',
  javascript: 'javascript',
  js: 'javascript',
  typescript: 'typescript',
  ts: 'typescript',
  jsx: 'jsx',
  tsx: 'tsx',
  sql: 'sql',
  mysql: 'sql',
  pgsql: 'sql',
  java: 'java',
  c: 'c',
  cpp: 'cpp',
  csharp: 'csharp',
  ruby: 'ruby',
  rust: 'rust',
  ini: 'ini',
  toml: 'toml',
  nginx: 'ini',
  dockerfile: 'ini',
  html: 'markup',
  xml: 'markup',
  markdown: 'markdown',
  css: 'css',
}

/** HTML 转义（未知语言时的兜底渲染）；null/undefined 按空串处理，避免 .replace 崩溃 */
function escapeHtml(s: string | null | undefined): string {
  return (s ?? '').replace(
    /[&<>"']/g,
    (c) =>
      (({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }) as Record<string, string>)[c],
  )
}

/** 按文件类型高亮代码，返回可直接注入的 HTML；未知类型转义原样输出 */
export function getHighlightSyntax(str: string | null | undefined, fileType?: string): string {
  // fileType 空串（后端 changelog 的 configType 可能为空）也要回退默认 yaml：
  // ?? 只兜 null/undefined，空串会走映射失败 → 纯文本无高亮，diff 和外面配置观感脱节。
  const lang = FILE_TYPE_TO_PRISM[(fileType || 'yaml').toLowerCase()]
  if (!lang || !Prism.languages[lang as keyof typeof Prism.languages]) return escapeHtml(str)
  try {
    return Prism.highlight(str ?? '', Prism.languages[lang as keyof typeof Prism.languages], lang)
  } catch {
    return escapeHtml(str)
  }
}
