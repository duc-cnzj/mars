import { memo, type CSSProperties, type ReactNode } from 'react'
import { ansiparse, PALETTE, type AnsiSegment } from '../utils/ansi'

function segmentStyle(seg: AnsiSegment): CSSProperties {
  const style: CSSProperties = {}
  if (seg.foreground) style.color = PALETTE[seg.foreground] ?? seg.foreground
  if (seg.background) style.background = PALETTE[seg.background] ?? seg.background
  if (seg.bold) style.fontWeight = 700
  if (seg.italic) style.fontStyle = 'italic'
  if (seg.underline) style.textDecoration = 'underline'
  return style
}

/**
 * ANSI 着色文本组件：解析后逐段渲染为带样式的 <span>。
 * 契约导出名（TabLog 会 import 此组件）。
 * 传入 highlight 时：段内命中关键字（不区分大小写）的片段用黄底黑字 <mark> 高亮，
 * 命中片段继承所在段前景色以外的样式被覆盖为黑字，其余文字保持原 ANSI 颜色不变。
 */
export const AnsiText = memo(function AnsiText({
  text,
  highlight,
}: {
  text: string
  highlight?: string
}) {
  const kw = highlight?.trim().toLowerCase() || ''
  return (
    <>
      {ansiparse(text).map((seg, i) => {
        if (!kw || !seg.text) {
          return (
            <span key={i} style={segmentStyle(seg)}>
              {seg.text}
            </span>
          )
        }
        const lower = seg.text.toLowerCase()
        const nodes: ReactNode[] = []
        let pos = 0
        let idx = lower.indexOf(kw, pos)
        let k = 0
        while (idx !== -1) {
          if (idx > pos) nodes.push(seg.text.slice(pos, idx))
          const end = idx + kw.length
          nodes.push(
            <mark key={k++} className="bg-[#fde047] text-black">
              {seg.text.slice(idx, end)}
            </mark>,
          )
          pos = end
          idx = lower.indexOf(kw, pos)
        }
        if (pos < seg.text.length) nodes.push(seg.text.slice(pos))
        return (
          <span key={i} style={segmentStyle(seg)}>
            {nodes}
          </span>
        )
      })}
    </>
  )
})

AnsiText.displayName = 'AnsiText'
