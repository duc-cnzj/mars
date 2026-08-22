import { useCallback, useEffect, useState } from 'react'
import { websocket } from '@/api/websocket'
import { useWebsocket } from '@/hooks/useWebsocket'

// 部署日志行数上限：长跑/异常刷屏的部署会话（如卡在某步反复推日志）日志不被无限累积
// 拖垮内存，超出丢弃最旧行。容器日志 TabLog 独立用 5000，部署面板会话短、留 1000 足够。
const MAX_DEPLOY_LOG_LINES = 1000

export type DeployStreamStatus = 'idle' | 'deploying' | 'deployed' | 'failed' | 'canceled'

export interface DeployLogLine {
  msg: string
  result: number
  containers: websocket.Container[]
}

export interface DeployCreateParams {
  repoId: number
  gitBranch?: string
  gitCommit?: string
  config: string
  extraValues?: websocket.ExtraValue[]
  atomic?: boolean
}

export interface DeployUpdateParams {
  projectId: number
  version: number
  gitBranch?: string
  gitCommit?: string
  config: string
  extraValues?: websocket.ExtraValue[]
  atomic?: boolean
}

/** 部署会话 slug：与后端 deploy.GetSlugName 对齐（sha256("{namespaceId}-{name}") hex） */
async function toSlug(namespaceId: number, name: string): Promise<string> {
  const buf = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(`${namespaceId}-${name}`))
  return Array.from(new Uint8Array(buf))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

/**
 * 实时部署流：按 slug 订阅 Create/UpdateProject 帧，累积日志行 + 进度 + 终态。
 * 触发 create/update 后后端在 WS 通道回投进度帧（metadata.result/percent/end）。
 */
export function useDeployStream(namespaceId: number, name: string) {
  const { ready, send, subscribe } = useWebsocket()
  const [slug, setSlug] = useState('')
  const [status, setStatus] = useState<DeployStreamStatus>('idle')
  const [percent, setPercent] = useState(0)
  const [logs, setLogs] = useState<DeployLogLine[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    let alive = true
    void toSlug(namespaceId, name).then((s) => {
      if (alive) setSlug(s)
    })
    return () => {
      alive = false
    }
  }, [namespaceId, name])

  useEffect(() => {
    if (!slug) return
    return subscribe(slug, (meta, raw) => {
      // 独立进度帧
      if (meta.type === websocket.Type.ProcessPercent) {
        if (typeof meta.percent === 'number') setPercent(meta.percent)
        return
      }
      // 部署日志帧（Create/UpdateProject）
      if (meta.type !== websocket.Type.CreateProject && meta.type !== websocket.Type.UpdateProject) return

      let containers: websocket.Container[] = []
      if (meta.result === websocket.ResultType.LogWithContainers) {
        try {
          containers = websocket.WsWithContainerMessageResponse.decode(raw).containers ?? []
        } catch {
          containers = []
        }
      }
      // 超上限丢最旧：纯函数式更新，StrictMode 双调用幂等（slice 只读 prev）
      setLogs((prev) => {
        const next = [...prev, { msg: meta.message ?? '', result: meta.result ?? 0, containers }]
        return next.length > MAX_DEPLOY_LOG_LINES ? next.slice(next.length - MAX_DEPLOY_LOG_LINES) : next
      })
      // 进度只认 ProcessPercent 帧；日志帧不带 percent（protobuf 默认解成 0），
      // 不能把它打回 0 —— 否则每条日志一来进度就跳 0
      if (typeof meta.percent === 'number' && meta.percent > 0) setPercent(meta.percent)

      if (meta.end) {
        setLoading(false)
        switch (meta.result) {
          case websocket.ResultType.Deployed:
            setStatus('deployed')
            setPercent(100)
            break
          case websocket.ResultType.DeployedCanceled:
            setStatus('canceled')
            break
          default:
            setStatus('failed')
        }
      }
    })
  }, [slug, subscribe])

  const create = useCallback(
    (p: DeployCreateParams) => {
      if (!ready) return
      setStatus('deploying')
      setLoading(true)
      setLogs([])
      setPercent(0)
      send(
        websocket.CreateProjectInput.encode({
          type: websocket.Type.CreateProject,
          namespaceId,
          name,
          repoId: p.repoId,
          gitBranch: p.gitBranch ?? '',
          gitCommit: p.gitCommit ?? '',
          config: p.config,
          extraValues: p.extraValues ?? [],
          atomic: p.atomic,
        }).finish(),
      )
    },
    [ready, send, namespaceId, name],
  )

  const update = useCallback(
    (p: DeployUpdateParams) => {
      if (!ready) return
      setStatus('deploying')
      setLoading(true)
      setLogs([])
      setPercent(0)
      send(
        websocket.UpdateProjectInput.encode({
          type: websocket.Type.UpdateProject,
          projectId: p.projectId,
          gitBranch: p.gitBranch ?? '',
          gitCommit: p.gitCommit ?? '',
          config: p.config,
          extraValues: p.extraValues ?? [],
          version: p.version,
          atomic: p.atomic,
        }).finish(),
      )
    },
    [ready, send],
  )

  const cancel = useCallback(() => {
    if (!ready) return
    send(
      websocket.CancelInput.encode({
        type: websocket.Type.CancelProject,
        namespaceId,
        name,
      }).finish(),
    )
  }, [ready, send, namespaceId, name])

  return { ready, status, percent, logs, loading, create, update, cancel }
}
