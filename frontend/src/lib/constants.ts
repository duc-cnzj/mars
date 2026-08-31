/** 前端通用常量表：跨文件复用的时间/数量常量集中于此，消灭重复簇。
 *  单文件私有的魔法数字留在文件顶部命名常量，这里只放「跨文件共享」的值。 */

/** 搜索输入防抖（ms）：停顿后触发列表重拉。统一 300（原 300/400 两套口径归一为多数派）。 */
export const SEARCH_DEBOUNCE_MS = 300

/** pod 事件 → 重拉容器列表防抖（ms）：WS 事件风暴合并为一次重拉。统一 500（对齐 TopologyTab 已命名口径）。 */
export const POD_DEBOUNCE_MS = 500

/** 下载/导出后延迟 revokeObjectURL（ms）：留足下载发起，避免浏览器提前回收中断下载。 */
export const REVOKE_OBJECT_URL_MS = 1000
