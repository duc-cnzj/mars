/**
 * 双击整选辅助。
 *
 * 空间名 / 容器镜像这类标识符含 `-` `.` `/` `:`，浏览器默认双击按「词」断选，
 * 只选中其中一段（mars-demo 只选到 demo、registry.uco.com/mars/demo:v1.0.0 只选到一段），
 * 想整段复制得手动拖选。统一改成双击选中整个文本节点，截断显示下同样能拿到完整内容。
 *
 * 高亮色由全局 ::selection（品牌色 30% 淡底）承担，此处不涉及样式。
 */
import type { MouseEvent } from 'react'

/** 双击整选：选中 currentTarget 内的全部文本（挂到承载文本的元素上） */
export function selectAllOnDoubleClick(e: MouseEvent<HTMLElement>): void {
  e.preventDefault()
  const range = document.createRange()
  range.selectNodeContents(e.currentTarget)
  const sel = window.getSelection()
  sel?.removeAllRanges()
  sel?.addRange(range)
}
