import type { components } from './schema'
import { api } from './client'
import { API } from './endpoints'

type HibernateWindow = components['schemas']['hibernate.HibernateWindow']

/**
 * 全局休眠时间窗的内存缓存（TTL 60s + 并发去重）。
 *
 * 背景：项目卡片的休眠弹层与详情「定时休眠」段都要按 window_name 展示「窗名（时段）」，
 * 而时段只存在于全局时间窗列表。该查询已放宽为登录可读（后端 Authorize 豁免到用户级），
 * 但没有缓存时每次打开弹层/详情都会重拉一次，切页反复打开即成小请求风暴。
 *
 * 时间窗由管理员在后台维护、变更频率极低，60s 过期窗口足够；并发去重让同屏多个
 * 消费方共享同一次请求。失败静默降级为空列表（调用方回退为只显示窗名），不缓存。
 */
const TTL_MS = 60_000
let cache: { items: HibernateWindow[]; until: number } | null = null
let inflight: Promise<HibernateWindow[]> | null = null

/** 取全部全局休眠时间窗：命中缓存（TTL 内）直接返回，否则拉取并缓存，成功才缓存 */
export function getHibernateWindows(): Promise<HibernateWindow[]> {
  if (cache) {
    if (Date.now() <= cache.until) return Promise.resolve(cache.items)
    cache = null
  }
  if (inflight) return inflight
  inflight = api
    .GET(API.hibernateWindows)
    .then(({ data }) => {
      const items = data?.items ?? []
      cache = { items, until: Date.now() + TTL_MS }
      return items
    })
    .catch(() => []) // 失败静默降级为空（调用方回退只显示窗名），不缓存
    .finally(() => {
      inflight = null
    })
  return inflight
}

/** 使缓存立即失效，下次 getHibernateWindows 重新拉取。后台时间窗增删改成功后必须调用，
 *  否则各消费方（项目卡片弹层 / 详情 / 配置页）最多 60s 内仍显示旧时段或已删除的窗。 */
export function invalidateHibernateWindows(): void {
  cache = null
}
