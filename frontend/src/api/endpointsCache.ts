import type { components } from './schema'
import { api } from './client'
import { API } from './endpoints'

type ServiceEndpoint = components['schemas']['types.ServiceEndpoint']

/**
 * 项目访问地址的内存缓存（TTL 60s + 并发去重）。
 *
 * 背景：工作台每张卡片把命名空间下所有项目渲染成行，每行挂载即拉
 * /api/endpoints/projects/{projectId}。SPA 切页离开工作台再回来时 Workbench 整体重挂，
 * 所有行全部重拉一遍——项目一多就成了请求风暴。
 *
 * 这里把成功返回的 endpoints 按 projectId 缓存 TTL 秒，切页回来直接复用；
 * 并发去重让 StrictMode 双挂 / 同项目多消费方共享同一次请求。
 * URL 仅在部署时变化，60s 过期窗口在实测中不可感知（且行本身部署后也不重拉，见 ProjectRow）。
 */
const cache = new Map<number, { items: ServiceEndpoint[]; until: number }>()
const inflight = new Map<number, Promise<ServiceEndpoint[]>>()
const TTL_MS = 60_000

/** 取项目访问地址：命中缓存（TTL 内）直接返回，否则拉取并缓存，成功才缓存 */
export function getEndpoints(projectId: number): Promise<ServiceEndpoint[]> {
  const hit = cache.get(projectId)
  if (hit) {
    if (Date.now() <= hit.until) return Promise.resolve(hit.items)
    cache.delete(projectId)
  }
  const pending = inflight.get(projectId)
  if (pending) return pending
  const p = api
    .GET(API.endpointsProject, { params: { path: { projectId } } })
    .then(({ data }) => {
      const items = data?.items ?? []
      cache.set(projectId, { items, until: Date.now() + TTL_MS })
      return items
    })
    .catch(() => []) // 失败静默降级为空（与原组件行为一致），不缓存
    .finally(() => {
      inflight.delete(projectId)
    })
  inflight.set(projectId, p)
  return p
}
