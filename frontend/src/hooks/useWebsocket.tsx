import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { websocket } from '@/api/websocket'
import { getToken } from '@/api/token'
import type { components } from '@/api/schema'

export type WsMetadata = websocket.Metadata

/** 集群信息（ClusterInfoSync 帧 payload 的类型） */
export type ClusterInfo = components['schemas']['websocket.ClusterInfo']

/** 帧处理回调：metadata + 原始帧字节（具体 payload 由订阅方按 type 二次解码） */
export type FrameHandler = (meta: WsMetadata, raw: Uint8Array) => void

interface WsContextValue {
  ws: WebSocket | null
  ready: boolean
  /** 发送 protobuf 编码后的帧 */
  send: (bytes: Uint8Array) => void
  /** 按 slug（流标识）订阅 */
  subscribe: (slug: string, handler: FrameHandler) => () => void
  /** 按消息类型订阅（集群信息/项目重载等全局事件） */
  subscribeType: (type: number, handler: FrameHandler) => () => void
  /** 集群信息实时值（ClusterInfoSync 帧最新一次推送），未收到时为 null */
  clusterInfo: ClusterInfo | null
  /** 项目重载次数：收到 ReloadProjects 后 debounce 500ms 递增，Workbench 用作刷新信号 */
  reloadProjectsRev: number
  /** 本次 ReloadProjects 批次受影响的空间 id 集合（与 reloadProjectsRev 同批更新）。
   *  debounce 窗口内累积多个空间并发变更，不互相覆盖；null = 窗口内有解不出 nsId 的坏帧 → 消费端整页刷新。 */
  reloadNsIds: number[] | null
  /** 注册 ProjectPodEvent 回调：pod 重启等事件时按 projectId 分发 */
  subscribeProjectPodEvent: (projectId: number, handler: () => void) => () => void
  /** 加入/退出某项目的 pod 事件订阅（弹窗打开 join、关闭 leave，后端随后向该项目推送 pod 事件） */
  joinProjectPodEvent: (namespaceId: number, projectId: number, join: boolean) => void
}

const WsContext = createContext<WsContextValue | null>(null)

/**
 * WebSocket 实时通道 Provider：连接 /ws → 鉴权 → 分发 protobuf 帧。
 * 帧格式：外层一律为 WsMetadataResponse（含 Metadata），
 * 具体 payload 由订阅方按 metadata.type 二次解码。
 * 断线自动重连（3s 退避），鉴权失败后停止重试。
 */
export function ProvideWebsocket({ children }: { children: ReactNode }) {
  const connRef = useRef<WebSocket | null>(null)
  const authFailedRef = useRef(false)
  const retryTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reloadTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  /** Provider 卸载后置位：conn.close() 异步触发 onclose，会赶在 cleanup 之后才执行并再排重连定时器，
   *  在已卸载组件上无限重连泄漏连接。disposed 守卫阻断后续 connect/onclose 的重连路径。 */
  const disposedRef = useRef(false)
  // ReloadProjects debounce 窗口内的累积器：受影响空间 id 集合 + 是否出现坏帧。
  // 多空间并发变更不再互相覆盖，窗口结束后整批下发一次。
  const pendingNsIdsRef = useRef<Set<number>>(new Set())
  const pendingFullReloadRef = useRef(false)
  const slugHandlers = useRef(new Map<string, Set<FrameHandler>>())
  const typeHandlers = useRef(new Map<number, Set<FrameHandler>>())
  const podEventHandlers = useRef(new Map<number, Set<() => void>>())
  const [tick, setTick] = useState(0)
  const [clusterInfo, setClusterInfo] = useState<ClusterInfo | null>(null)
  const [reloadProjectsRev, setReloadProjectsRev] = useState(0)
  const [reloadNsIds, setReloadNsIds] = useState<number[] | null>(null)

  const subscribe = useCallback((slug: string, handler: FrameHandler) => {
    const set = slugHandlers.current.get(slug) ?? new Set<FrameHandler>()
    set.add(handler)
    slugHandlers.current.set(slug, set)
    return () => {
      set.delete(handler)
      if (set.size === 0) slugHandlers.current.delete(slug)
    }
  }, [])

  const subscribeType = useCallback((type: number, handler: FrameHandler) => {
    const set = typeHandlers.current.get(type) ?? new Set<FrameHandler>()
    set.add(handler)
    typeHandlers.current.set(type, set)
    return () => {
      set.delete(handler)
      if (set.size === 0) typeHandlers.current.delete(type)
    }
  }, [])

  /** 按 projectId 注册 pod 事件回调（同一项目可多个订阅方，各自独立计数） */
  const subscribeProjectPodEvent = useCallback((projectId: number, handler: () => void) => {
    const set = podEventHandlers.current.get(projectId) ?? new Set<() => void>()
    set.add(handler)
    podEventHandlers.current.set(projectId, set)
    return () => {
      set.delete(handler)
      if (set.size === 0) podEventHandlers.current.delete(projectId)
    }
  }, [])

  const connect = useCallback(() => {
    if (disposedRef.current) return
    if (authFailedRef.current) return
    if (retryTimer.current) {
      clearTimeout(retryTimer.current)
      retryTimer.current = null
    }
    const token = getToken()
    if (!token) return
    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${window.location.host}/ws`
    const conn = new WebSocket(url)
    conn.binaryType = 'arraybuffer'
    conn.onopen = () => {
      conn.send(
        websocket.AuthorizeTokenInput.encode({
          type: websocket.Type.HandleAuthorize,
          token,
        }).finish(),
      )
      setTick((n) => n + 1)
    }
    conn.onmessage = (evt) => {
      const raw = new Uint8Array(evt.data as ArrayBuffer)
      let meta: WsMetadata
      try {
        meta = websocket.WsMetadataResponse.decode(raw).metadata ?? ({} as WsMetadata)
      } catch {
        return
      }
      const slug = meta.slug ?? ''
      const type = Number(meta.type ?? 0)
      if (slug && slugHandlers.current.has(slug)) {
        for (const h of slugHandlers.current.get(slug) ?? []) h(meta, raw)
      }
      if (type && typeHandlers.current.has(type)) {
        for (const h of typeHandlers.current.get(type) ?? []) h(meta, raw)
      }
      if (type === websocket.Type.SetUid && meta.message) {
        localStorage.setItem('uid', meta.message)
      }
      if (type === websocket.Type.InternalError) {
        authFailedRef.current = true
      }

      // ClusterInfoSync：实时集群状态 → 覆盖 context.clusterInfo（ClusterStatus 用）
      if (type === websocket.Type.ClusterInfoSync) {
        try {
          const res = websocket.WsHandleClusterResponse.decode(raw)
          if (res.info) setClusterInfo(res.info)
        } catch {
          /* 坏帧忽略 */
        }
      }

      // ProjectPodEvent：pod 重启/删除等 → 按 projectId 通知订阅方刷新容器列表/日志
      if (type === websocket.Type.ProjectPodEvent) {
        try {
          const ev = websocket.WsProjectPodEventResponse.decode(raw)
          const set = podEventHandlers.current.get(ev.projectId)
          if (set) for (const h of [...set]) h()
        } catch {
          /* 坏帧忽略 */
        }
      }

      // ReloadProjects：命名空间/项目变更 → debounce 后递增刷新信号并把受影响空间 id 集合下发。
      // 窗口内累积 Set：多个空间并发变更不再互相覆盖（旧实现单值最后一条胜出会丢变更）。
      // 坏帧（解不出 nsId）置 pendingFullReload，本次批次整体退化为整页刷新。
      // 集合供 Workbench 精确给被更新空间卡显示 loading（对齐旧版 setNamespaceReload(true, nsID)）。
      if (type === websocket.Type.ReloadProjects) {
        let nsId: number | null = null
        try {
          nsId = websocket.WsReloadProjectsResponse.decode(raw).namespaceId ?? null
        } catch {
          /* 坏帧忽略，nsId=null */
        }
        // proto3 未设置的 int32 解出 0 而非 null：0/负/坏帧一律视为「未知空间」→ 整页刷新兜底。
        // 只判 null 会让缺省帧被 add(0) 当成 namespace 0 做一次注定 404 的单卡刷新（no-op），
        // 「坏帧退化整页刷新」这条兜底路径永远不生效。
        if (nsId === null || nsId <= 0) pendingFullReloadRef.current = true
        else pendingNsIdsRef.current.add(nsId)
        if (reloadTimer.current) clearTimeout(reloadTimer.current)
        reloadTimer.current = setTimeout(() => {
          reloadTimer.current = null
          if (pendingFullReloadRef.current) {
            setReloadNsIds(null)
          } else {
            setReloadNsIds([...pendingNsIdsRef.current])
          }
          pendingNsIdsRef.current.clear()
          pendingFullReloadRef.current = false
          setReloadProjectsRev((r) => r + 1)
        }, 500)
      }
    }
    conn.onclose = () => {
      if (disposedRef.current) return
      // 过期连接（StrictMode 双挂载残留 / 被新连接顶替）关闭：忽略，重连由当前连接负责。
      // 没有这条守卫，StrictMode 下旧 conn 的 onclose 会多排一个 3s 重连定时器，
      // 触发时再开一条新连接——与当前已建立连接并存，之后每条旧连接关闭又排新的，无限churn。
      // disposedRef 只堵卸载路径，堵不住「重挂后旧连接才关闭」这条。
      if (connRef.current !== conn) return
      connRef.current = null
      if (authFailedRef.current) return
      retryTimer.current = setTimeout(() => {
        retryTimer.current = null
        connect()
      }, 3000)
      setTick((n) => n + 1)
    }
    connRef.current = conn
    setTick((n) => n + 1)
  }, [])

  useEffect(() => {
    disposedRef.current = false
    connect()
    return () => {
      disposedRef.current = true
      if (retryTimer.current) clearTimeout(retryTimer.current)
      if (reloadTimer.current) clearTimeout(reloadTimer.current)
      connRef.current?.close()
      connRef.current = null
    }
  }, [connect])

  const send = useCallback((bytes: Uint8Array) => {
    const c = connRef.current
    if (c && c.readyState === WebSocket.OPEN) c.send(bytes)
  }, [])

  /** 加入/退出项目 pod 事件订阅：编码为 ProjectPodEventJoinInput 帧（与旧前端 useProjectRoom 一致） */
  const joinProjectPodEvent = useCallback(
    (namespaceId: number, projectId: number, join: boolean) => {
      send(
        websocket.ProjectPodEventJoinInput.encode({
          type: websocket.Type.ProjectPodEvent,
          join,
          namespaceId,
          projectId,
        }).finish(),
      )
    },
    [send],
  )

  const value = useMemo<WsContextValue>(
    () => ({
      ws: connRef.current,
      ready: connRef.current?.readyState === WebSocket.OPEN,
      send,
      subscribe,
      subscribeType,
      clusterInfo,
      reloadProjectsRev,
      reloadNsIds,
      subscribeProjectPodEvent,
      joinProjectPodEvent,
    }),
    // tick 驱动连接状态变更时重算 ready/ws；clusterInfo / reloadProjectsRev / reloadNsIds 变化时重发最新值
    //（send/subscribe/subscribeProjectPodEvent/joinProjectPodEvent 走 ref，稳定不变）
    [
      send,
      subscribe,
      subscribeType,
      subscribeProjectPodEvent,
      joinProjectPodEvent,
      clusterInfo,
      reloadProjectsRev,
      reloadNsIds,
      tick,
    ],
  )

  return <WsContext.Provider value={value}>{children}</WsContext.Provider>
}

export function useWebsocket(): WsContextValue {
  const ctx = useContext(WsContext)
  if (!ctx) throw new Error('useWebsocket 必须在 ProvideWebsocket 内使用')
  return ctx
}
