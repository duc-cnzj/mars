import { useSyncExternalStore } from 'react'
import { api } from '@/api/client'
import type { components } from '@/api/schema'

type VersionResponse = components['schemas']['version.Response']

// 模块级单例：Topbar 与 Footer 都调用 useVersion，此前各自挂载即拉 → /api/version 每次加载打 2 笔。
// 现在全站只拉一次，多个订阅方共享同一份缓存（对齐 endpointsCache 的模块级缓存语义）。
let cache: VersionResponse | null = null
let promise: Promise<void> | null = null
const listeners = new Set<() => void>()
const emit = () => {
  for (const l of listeners) l()
}

function ensureVersion() {
  if (promise) return
  promise = api
    .GET('/api/version')
    .then(({ data }) => {
      if (data) {
        cache = data
        emit()
      }
    })
    .catch(() => {
      // 失败保持 null（与原先行为一致），单例不重试
    })
}

/**
 * 拉取后端版本信息（version/buildDate 等），供 Header/Footer 展示。
 * 模块级单例缓存：无论多少个组件调用，/api/version 全站仅请求一次。
 * 数据到达后经 useSyncExternalStore 广播给所有订阅方。
 */
export function useVersion(): VersionResponse | null {
  ensureVersion()
  return useSyncExternalStore(
    (cb) => {
      listeners.add(cb)
      return () => {
        listeners.delete(cb)
      }
    },
    () => cache,
  )
}
