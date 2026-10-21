import { lazy, Suspense, useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/shadcn/button'
import { toast } from '@/lib/toast'
import { AIHTTPError, enableAIAgent, fetchAIStatus, type AIStatus } from './aiApi'
import type { TabKey } from '../projects/tabVisibility'

// 复用工作台的项目详情弹窗（懒加载，与 Workbench 同策略）：AI 项目就是普通 project，
// 打开后无需任何定制 UI——多出来的只有「先唤醒再打开」这一层编排。
const ProjectDetailModal = lazy(() =>
  import('../projects/ProjectDetailModal').then((m) => ({ default: m.ProjectDetailModal })),
)

type ProjectModel = components['schemas']['types.ProjectModel']

// 沙箱弹窗只保留「容器日志 + 命令行」：AI 入口不需要配置更新/拓扑/详细信息，
// 普通项目从工作台进入时仍是完整 Tab（限制只作用于 AI 编排这一层）。
const AI_SANDBOX_TABS: readonly TabKey[] = ['logs', 'shell']

// 沙箱就绪轮询：sandbox=running 表示 Deployment 的 readinessProbe（opencode /health）
// 已通过，此时进入容器 `mai` attach 才不会连不上常驻 server。首次启动可能包含镜像
// 拉取，预算给到 3 分钟；轮询间隔 1.5s 对后端只是廉价的 Deployment 读。
const AI_READY_POLL_INTERVAL_MS = 1500
const AI_READY_TIMEOUT_MS = 180_000

/** 等待 ms 毫秒；用于轮询间隔，不依赖外部库。 */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * 顶栏 AI 入口的编排层：点击 AI 图标 → 确保沙箱处于运行态 → 打开项目详情弹窗。
 *
 * 流程（对应后端 /api/ai/*）：
 *  1. GET  /api/ai/status：仓库未配置（enabled=false）直接提示并关闭；
 *  2. 沙箱 sleeping/absent 或尚未部署 → POST /api/ai/agent/enable 唤醒（后端幂等，
 *     内部会扩容到 1 或重建丢失的 Deployment）；409 = 部署进行中，项目已存在时继续等待；
 *  3. 轮询 GET /api/ai/status 直到 sandbox=running（readinessProbe 通过），超时 3 分钟；
 *  4. GET /api/projects/{id} 拉项目详情，交给 ProjectDetailModal 渲染。
 *
 * 弹窗的生命周期完全由本组件持有：关闭 = 卸载 ProjectDetailModal；重新点击 = 重新编排。
 */
export function AIAgentLauncher({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useTranslation()
  const [project, setProject] = useState<ProjectModel | null>(null)
  const [loading, setLoading] = useState(false)
  // 等待阶段：waking=正在触发唤醒；starting=已扩容、等容器就绪（首次启动可能较久）。
  const [phase, setPhase] = useState<'waking' | 'starting'>('waking')
  // 每次打开递增：关闭后仍在飞行的请求结果直接丢弃，防止旧结果把新一次的弹窗顶掉。
  const runIdRef = useRef(0)

  /** 按项目 ID 拉取最新详情（ProjectDetailModal 同款接口）。 */
  const loadDetail = useCallback(
    async (projectId: number): Promise<ProjectModel> => {
      const { data, error } = await api.GET(API.projectsDetail, {
        params: { path: { id: projectId } },
      })
      if (error) throw new Error(error.message ?? String(error))
      if (!data?.item) throw new Error(t('ai.projectMissing'))
      return data.item
    },
    [t],
  )

  /**
   * 轮询直到沙箱 running：running 由后端按 Deployment 就绪副本数派生，
   * 对应 readinessProbe（opencode /health）已通过——此时再打开弹窗，
   * 命令行 Tab 的 `mai` attach 才能连上常驻 server，不会白等。
   * 超时抛错（镜像拉取/启动失败都会一直停在 starting）；alive 失效时提前返回。
   */
  const waitUntilRunning = useCallback(
    async (initial: AIStatus, alive: () => boolean): Promise<AIStatus> => {
      let status = initial
      let waited = 0
      while (status.agent?.sandbox !== 'running') {
        if (!alive()) return status
        if (!status.agent) throw new Error(t('ai.notDeployed'))
        if (waited >= AI_READY_TIMEOUT_MS) throw new Error(t('ai.startTimeout'))
        await sleep(AI_READY_POLL_INTERVAL_MS)
        waited += AI_READY_POLL_INTERVAL_MS
        if (!alive()) return status
        status = await fetchAIStatus()
        // 等待期间管理员禁用仓库：继续等到超时没有意义，直接按未启用处理。
        if (!status.enabled) throw new Error(t('ai.disabled'))
      }
      return status
    },
    [t],
  )

  /** 一次完整的「唤醒 → 等就绪 → 打开」编排。 */
  const start = useCallback(async () => {
    const runId = ++runIdRef.current
    const alive = () => runIdRef.current === runId
    setProject(null)
    setPhase('waking')
    setLoading(true)
    try {
      const initial = await fetchAIStatus()
      if (!alive()) return
      if (!initial.enabled) {
        toast.error(t('ai.disabled'))
        onClose()
        return
      }

      let status: AIStatus = initial
      // running 无需唤醒；starting 说明已有部署在扩容，直接进入等待。
      if (!status.agent || status.agent.sandbox === 'sleeping' || status.agent.sandbox === 'absent') {
        toast.info(t('ai.waking'))
        try {
          status = await enableAIAgent()
        } catch (e) {
          // 409 = 同一项目的部署进行中。项目行已存在则继续等它就绪，否则没有可打开的对象。
          if (e instanceof AIHTTPError && e.status === 409 && status.agent) {
            toast.warning(t('ai.deployInProgress'))
          } else {
            throw e
          }
        }
        if (!alive()) return
      }

      if (!status.agent) throw new Error(t('ai.notDeployed'))
      // 容器启动需要时间：等 readinessProbe 通过（running）后再拉详情弹窗，
      // 避免用户打开后对着「暂无容器/连不上」干等。
      if (status.agent.sandbox !== 'running') {
        setPhase('starting')
        status = await waitUntilRunning(status, alive)
        if (!alive()) return
      }
      if (!status.agent) throw new Error(t('ai.notDeployed'))

      const detail = await loadDetail(status.agent.project_id)
      if (!alive()) return
      setProject(detail)
    } catch (e) {
      if (!alive()) return
      toast.error(e instanceof Error ? e.message : String(e))
      onClose()
    } finally {
      if (alive()) setLoading(false)
    }
  }, [loadDetail, onClose, t, waitUntilRunning])

  useEffect(() => {
    if (!open) {
      // 关闭：作废在途请求并清空上一次的项目，下一次打开从干净状态开始。
      runIdRef.current++
      setProject(null)
      setLoading(false)
      return
    }
    void start()
    // 只在开关切换时编排；start 内部依赖的 t/onClose 在会话内稳定。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  if (!open) return null

  return (
    <>
      {/* 唤醒/等就绪期间的即时反馈：沙箱冷启动可能持续数分钟，避免点击后无响应。
          「取消」是等待期的逃生口：轮询最长 3 分钟，不能把用户锁在叠加层里。 */}
      {loading && (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/20">
          <div className="flex items-center gap-3 rounded-xl border border-line bg-surface px-4 py-3 text-[13px] text-ink shadow-lg">
            <Spinner />
            <span>{phase === 'starting' ? t('ai.starting') : t('ai.waking')}</span>
            <Button size="xs" variant="outline" onClick={onClose}>
              {t('common.cancel')}
            </Button>
          </div>
        </div>
      )}
      <Suspense fallback={null}>
        {project && (
          <ProjectDetailModal
            project={project}
            namespaceName={project.namespace?.name ?? ''}
            open
            // AI 入口默认落到命令行 Tab：用户点 AI 图标就是为了直接对话
            defaultTab="shell"
            // 沙箱只需要看日志和进命令行，其余 Tab 对 AI 场景没有意义
            allowedTabs={AI_SANDBOX_TABS}
            // 进命令行自动执行 mai：用户不用再手动敲一遍进交互
            autoShellCommand="mai"
            onClose={onClose}
            onDeleted={() => {
              toast.warning(t('ai.projectDeleted'))
              onClose()
            }}
            onChanged={() => {
              // 配置更新后刷新弹窗内详情；失败由 ProjectDetailModal 自身重拉兜底。
              void loadDetail(project.id)
                .then(setProject)
                .catch(() => undefined)
            }}
          />
        )}
      </Suspense>
    </>
  )
}
