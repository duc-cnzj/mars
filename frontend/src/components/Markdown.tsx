import { isValidElement, type ReactNode } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { getHighlightSyntax } from '@/lib/highlight'

/**
 * AI 对话消息的 Markdown 渲染：react-markdown + remark-gfm（表格/任务列表/删除线）。
 *
 * 刻意不启用 rehype-raw：模型输出属于不可信内容，禁止原始 HTML 进 DOM（XSS）。
 * 代码块复用仓库已有的 Prism 高亮（getHighlightSyntax 对输入做 HTML 转义），
 * 未识别的语言回退纯文本转义。
 */
export function Markdown({ children }: { children: string }) {
  return (
    <div className="min-w-0 break-words text-[13px] leading-relaxed [&_blockquote]:my-2 [&_blockquote]:border-l-2 [&_blockquote]:border-primary/40 [&_blockquote]:pl-3 [&_blockquote]:text-mute [&_h1]:my-3 [&_h1]:text-base [&_h1]:font-semibold [&_h2]:my-2 [&_h2]:text-[14px] [&_h2]:font-semibold [&_h3]:my-2 [&_h3]:font-semibold [&_hr]:my-3 [&_hr]:border-line [&_li]:my-0.5 [&_ol]:my-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-2 [&_p:first-child]:mt-0 [&_p:last-child]:mb-0 [&_table]:my-2 [&_table]:w-full [&_table]:border-collapse [&_td]:border [&_td]:border-line [&_td]:px-2 [&_td]:py-1 [&_th]:border [&_th]:border-line [&_th]:px-2 [&_th]:py-1 [&_th]:text-left [&_ul]:my-2 [&_ul]:list-disc [&_ul]:pl-5">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          // 代码块整体渲染（含高亮）：不再走默认 <pre><code>，避免二次包裹与样式冲突。
          pre: ({ children }) => {
            const { text, lang } = extractCode(children)
            return (
              // 底色固定暗色，正文字色必须同时固定：否则继承容器 text-ink，
              // 在浅色主题下就是深字压深底（Prism 只为 .token 上色，纯文本会看不见）。
              <pre className="my-3 overflow-x-auto rounded-lg border border-line bg-[#1e1e1e] px-3 py-2 text-[12px] leading-relaxed text-[#eee]">
                <code dangerouslySetInnerHTML={{ __html: getHighlightSyntax(text, lang) }} />
              </pre>
            )
          },
          // 行内代码；块级代码已在上面的 pre 里整体处理。
          code: ({ children }) => (
            <code className="rounded bg-raised px-1 py-0.5 text-[12px] text-ink">{children}</code>
          ),
          a: ({ children, href }) => (
            <a
              href={href}
              target="_blank"
              rel="noreferrer noopener"
              className="text-primary underline underline-offset-2 hover:text-primary-strong"
            >
              {children}
            </a>
          ),
          img: ({ alt, src }) => (
            // 模型可能引用外部图片：懒加载 + 不带 Referer，避免把当前页面地址泄露给第三方图床。
            <img
              src={src}
              alt={alt}
              loading="lazy"
              referrerPolicy="no-referrer"
              className="my-2 max-w-full rounded-lg border border-line"
            />
          ),
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  )
}

/** 从 pre 的子节点里取出代码原文与语言名（react-markdown 把语言放在 code 的 className）。 */
function extractCode(node: ReactNode): { text: string; lang: string } {
  if (isValidElement<{ className?: string; children?: ReactNode }>(node)) {
    const match = /language-([\w-]+)/.exec(node.props.className ?? '')
    return { text: textOf(node.props.children), lang: match?.[1] ?? 'text' }
  }
  return { text: textOf(node), lang: 'text' }
}

/** 递归收集 React 节点里的纯文本。 */
function textOf(node: ReactNode): string {
  if (node == null || typeof node === 'boolean') return ''
  if (typeof node === 'string' || typeof node === 'number') return String(node)
  if (Array.isArray(node)) return node.map(textOf).join('')
  if (isValidElement<{ children?: ReactNode }>(node)) return textOf(node.props.children)
  return ''
}
