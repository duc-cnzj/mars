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
import { websocket } from '../api/websocket'
import { getToken } from '../api/token'
import type { components } from '../api/schema'

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
  /** 最近一次 ReloadProjects 触发的空间 id（与 reloadProjectsRev 同批更新，无 id 时为 null） */
  reloadNsId: number | null
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
  const slugHandlers = useRef(new Map<string, Set<FrameHandler>>())
  const typeHandlers = useRef(new Map<number, Set<FrameHandler>>())
  const podEventHandlers = useRef(new Map<number, Set<() => void>>())
  const [tick, setTick] = useState(0)
  const [clusterInfo, setClusterInfo] = useState<ClusterInfo | null>(null)
  const [reloadProjectsRev, setReloadProjectsRev] = useState(0)
  const [reloadNsId, setReloadNsId] = useState<number | null>(null)

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

      // ReloadProjects：命名空间/项目变更 → debounce 后递增刷新信号并记录触发空间 id。
      // nsId 供 Workbench 精确给被更新空间卡显示 loading（对齐旧版 setNamespaceReload(true, nsID)）。
      if (type === websocket.Type.ReloadProjects) {
        let nsId: number | null = null
        try {
          nsId = websocket.WsReloadProjectsResponse.decode(raw).namespaceId ?? null
        } catch {
          /* 坏帧忽略，nsId=null（不做精确卡 loading，仍触发全量刷新） */
        }
        if (reloadTimer.current) clearTimeout(reloadTimer.current)
        reloadTimer.current = setTimeout(() => {
          reloadTimer.current = null
          setReloadNsId(nsId)
          setReloadProjectsRev((r) => r + 1)
        }, 500)
      }
    }
    conn.onclose = () => {
      if (connRef.current === conn) connRef.current = null
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
    connect()
    return () => {
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
      reloadNsId,
      subscribeProjectPodEvent,
      joinProjectPodEvent,
    }),
    // tick 驱动连接状态变更时重算 ready/ws；clusterInfo / reloadProjectsRev / reloadNsId 变化时重发最新值
    //（send/subscribe/subscribeProjectPodEvent/joinProjectPodEvent 走 ref，稳定不变）
    [
      send,
      subscribe,
      subscribeType,
      subscribeProjectPodEvent,
      joinProjectPodEvent,
      clusterInfo,
      reloadProjectsRev,
      reloadNsId,
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
