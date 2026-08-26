import type { TFunction } from 'i18next'
import type { components } from '@/api/schema'

type Element = components['schemas']['mars.Element']
type GroupSetting = components['schemas']['mars.GroupSetting']

/** 兜底分区名（数据值）：未显式归组的元素（新增默认 / 后端旧数据无 group 字段）统一落入该分区，
    作为真实组名持久化到 group_settings（表单提交的就是 'default'）。展示层用 groupLabel 国际化 */
export const DEFAULT_GROUP = 'default'

/** 分区展示名：兜底分区（default）按 i18n 显示（zh: 基础配置 / en: Default），
    其余用户自定义分区名按数据原样输出（group 名是持久化数据，不做翻译）。
    t 用 i18next TFunction（字面量 key 约束），不做泛化签名以免 strict 模式不兼容 */
export const groupLabel = (name: string, t: TFunction) =>
  name === DEFAULT_GROUP ? t('repos.groupDefault') : name

/** 分区展示模型（由 elements 派生，不是独立实体）：name='' 表示未分组（恒排最后） */
export type GroupSection = { name: string; collapsed: boolean; elements: Element[] }

/** 元素所属分区名（去首尾空白归一，避免「Redis 」与「Redis」被当成两个分区） */
export const groupOf = (el: Element) => el.group?.trim() ?? ''

/**
 * 分区派生（编辑页 DynamicElement 与部署表单 Elements 共用同一份，保证两页分组一致）：
 * 分区不是独立实体，而是由 element.group 推导。
 * 展示顺序 = 已配置分区按 group_settings.order 升序 → 未配置的新分区按首字段出现顺序 →
 * 未分组（空 group）恒排最后。全部元素都未分组时返回空数组（无分区 = 兼容旧版平铺）。
 */
export function buildSections(value: Element[], groups: GroupSetting[]): GroupSection[] {
  const used = new Set(value.map(groupOf).filter((g) => g !== ''))
  if (used.size === 0) return []
  const settings = new Map(groups.map((g) => [g.name, g]))
  const names: string[] = []
  // 已配置分区：按 settings.order 升序
  groups
    .filter((g) => used.has(g.name))
    .sort((a, b) => a.order - b.order)
    .forEach((g) => names.push(g.name))
  // 未配置的新分区（用户刚在卡片上输入）：按首字段出现顺序补在已配置分区之后
  const seen = new Set(names)
  value.forEach((el) => {
    const g = groupOf(el)
    if (g && !seen.has(g)) {
      names.push(g)
      seen.add(g)
    }
  })
  // 未分组恒排最后（存在未分组元素才渲染该区）
  if (value.some((el) => groupOf(el) === '')) names.push('')
  return names.map((name) => ({
    name,
    collapsed: settings.get(name)?.collapsed ?? false,
    elements: value.filter((el) => groupOf(el) === name),
  }))
}
