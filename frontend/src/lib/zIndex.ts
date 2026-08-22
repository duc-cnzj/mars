/** 跨实例共享的 z-index 计数器：从 shadcn 的 z-50 之上开始累加，实现可拖拽弹窗置顶。
 *  计数器在模块级持有，全局唯一；弹窗/浮层每次打开/置顶分配一次。 */
let zCounter = 51

/** 分配下一个 z-index。供可拖拽弹窗及需盖过它们的浮层（如强杀确认框）使用 */
export function nextZIndex() {
  return ++zCounter
}
