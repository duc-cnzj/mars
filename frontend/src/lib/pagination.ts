/** 分页渲染项：页码数字或省略号折叠占位 */
type PageItem = number | '...'

/**
 * 经典省略号分页：首尾固定两页 + 当前页±1，中间用省略号折叠（如 1 2 … 5 6 7 … 99 100）。
 * 总页数 ≤ 5 时全量展示——省略号折叠不省空间，直接列全。
 */
export function buildPages(page: number, total: number): PageItem[] {
  if (total <= 5) return Array.from({ length: total }, (_, i) => i + 1)
  const items: PageItem[] = []
  const start = Math.max(3, page - 1)
  const end = Math.min(total - 2, page + 1)
  items.push(1, 2)
  if (start > 3) items.push('...')
  for (let i = start; i <= end; i += 1) items.push(i)
  if (end < total - 2) items.push('...')
  items.push(total - 1, total)
  return items
}
