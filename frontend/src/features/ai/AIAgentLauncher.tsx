import { lazy, Suspense, useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { Spinner } from '@/components/ui/Spinner'
import { toast } from '@/lib/toast'
import { AIHTTPError, enableAIAgent, fetchAIStatus, type AIStatus } from './aiApi'

// 复用工作台的项目详情弹窗（懒加载，与 Workbench 同策略）：AI 项目就是普通 project，
// 打开后无需任何定制 UI——多出来的只有「先唤醒再打开」这一层编排。
const ProjectDetailModal = lazy(() =>
  import('../projects/ProjectDetailModal').then((m) => ({ default: m.ProjectDetailModal })),
)

type ProjectModel = components['schemas']['types.ProjectModel']

/**
 * 顶栏 AI 入口的编排层：点击 AI 图标 → 确保沙箱处于运行态 → 打开项目详情弹窗。
 *
 * 流程（对应后端 /api/ai/*）：
 *  1. GET  /api/ai/status：仓库未配置（enabled=false）直接提示并关闭；
 *  2. 沙箱 sleeping/absent 或尚未部署 → POST /api/ai/agent/enable 唤醒（后端幂等，
 *     内部会缩容到 1 或重建丢失的 Deployment）；409 = 部署进行中，项目已存在时照常打开；
 *  3. GET  /api/projects/{id} 拉项目详情，交给 ProjectDetailModal 渲染。
 *
 * 弹窗的生命周期完全由本组件持有：关闭 = 卸载 ProjectDetailModal；重新点击 = 重新编排。
 */
export function AIAgentLauncher({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useTranslation()
  const [project, setProject] = useState<ProjectModel | null>(null)
  const [loading, setLoading] = useState(false)
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

  /** 一次完整的「唤醒 + 打开」编排。 */
  const start = useCallback(async () => {
    const runId = ++runIdRef.current
    const alive = () => runIdRef.current === runId
    setProject(null)
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
      // running/starting 视为已在运行（starting = 正在扩容未就绪），无需重复唤醒。
      if (!status.agent || status.agent.sandbox === 'sleeping' || status.agent.sandbox === 'absent') {
        toast.info(t('ai.waking'))
        try {
          status = await enableAIAgent()
        } catch (e) {
          // 409 = 同一项目的部署进行中。项目行已存在则照常打开详情（用户能看到部署进度），
          // 否则没有可打开的对象，按失败处理。
          if (e instanceof AIHTTPError && e.status === 409 && status.agent) {
            toast.warning(t('ai.deployInProgress'))
          } else {
            throw e
          }
        }
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
  }, [loadDetail, onClose, t])

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
      {/* 唤醒/拉详情期间的即时反馈：沙箱冷启动可能持续数秒，避免点击后无响应。 */}
      {loading && (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/20">
          <div className="flex items-center gap-2 rounded-xl border border-line bg-surface px-4 py-3 text-[13px] text-ink shadow-lg">
            <Spinner />
            <span>{t('ai.waking')}</span>
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
