import type zhCN from '@/i18n/locales/zh-CN'

/** 将嵌套词条对象递归展开为点号路径的联合类型 */
type FlatKeys<T, P extends string = ''> = {
  [K in keyof T & string]: T[K] extends Record<string, unknown>
    ? FlatKeys<T[K], `${P}${K}.`>
    : `${P}${K}`
}[keyof T & string]

/**
 * 词条扁平 key 联合类型（i18next strict 字面量 key）：
 * t() 只接受来自 zh-CN 词条的真实 key，编译期保证导航/字段 key 不悬空。
 * 字段/配置里存放 key 的字符串必须标注为 TKey，否则 t(key) 报 TS2345。
 */
export type TKey = FlatKeys<typeof zhCN>
