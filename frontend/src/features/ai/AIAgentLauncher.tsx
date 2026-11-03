import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/shadcn/button'
import { toast } from '@/lib/toast'
import {
  AI_LAST_USED_STORAGE_KEY,
  AIHTTPError,
  enableAIAgent,
  fetchAIStatus,
  reportAIUsage,
  touchAIActivity,
  type AIStatus,
} from './aiApi'
import type { TabKey } from '../projects/tabVisibility'
import { AIBindingBar } from './AIBindingBar'
import { AIQuickPrompts } from './AIQuickPrompts'
import { aiBindingKey, type AIBinding } from './AIContext'

// 复用工作台的项目详情弹窗（懒加载，与 Workbench 同策略）：AI 项目就是普通 project，
// 打开后无需任何定制 UI——多出来的只有「先唤醒再打开」这一层编排。
const ProjectDetailModal = lazy(() =>
  import('../projects/ProjectDetailModal').then((m) => ({ default: m.ProjectDetailModal })),
)

type ProjectModel = components['schemas']['types.ProjectModel']

// 沙箱弹窗只保留命令行：AI 页面就是与 mai 对话，日志由 AI 自己查，
// 配置更新/拓扑/详细信息/日志 Tab 对 AI 场景没有意义（限制只作用于 AI 编排这一层）。
const AI_SANDBOX_TABS: readonly TabKey[] = ['shell']

// 沙箱就绪轮询：sandbox=running 表示 Deployment 的 readinessProbe（opencode /health）
// 已通过，此时进入容器 `mai` attach 才不会连不上常驻 server。首次启动可能包含镜像
// 拉取，预算给到 3 分钟；轮询间隔 1.5s 对后端只是廉价的 Deployment 读。
const AI_READY_POLL_INTERVAL_MS = 1500
const AI_READY_TIMEOUT_MS = 180_000

// 沙箱活跃度心跳：弹窗打开且页面可见时每分钟续一次，避免用户正在看日志/终端
// 却被空闲缩容任务误关；后台标签页不续活，资源让给真正在用的人。
const AI_ACTIVITY_HEARTBEAT_MS = 60_000

/** 等待 ms 毫秒；用于轮询间隔，不依赖外部库。 */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * 把任意文本包成 POSIX 单引号字面量，供 `mai '<问题>'` 安全嵌入终端命令。
 * 单引号内除 `'` 外都不解释；`'` 本身用 `'\''` 拼接，避免问题里的引号/空格
 * 被 shell 拆成多个参数。
 */
function shellQuote(text: string): string {
  return `'${text.replace(/'/g, "'\\''")}'`
}

/**
 * 把绑定压成 `mai --bind` 能表达的范围。
 *
 * mars-agent 只接受 1~2 段（`空间` 或 `空间/项目`），且任一段不能以 `-` 开头：
 * 项目名自带 `/` 时拼出的 `空间/项目` 会变成 3 段，mars-agent 判为非法后会把
 * **全部**绑定一起丢掉。这类名字（以及空白名、以 `-` 开头的名字）的项目实际也
 * 部署不起来（Helm release 名同样有约束），降级为空间级绑定——至少保留空间
 * 上下文，也不拖累其它合法绑定。
 */
function cliBindScope(binding: AIBinding): string {
  // mars-agent 会 trim 各段，这里先 trim 再判断：空白名同样无法表达。
  const project = binding.project?.trim()
  if (!project) return binding.namespace
  if (project.includes('/') || project.startsWith('-')) return binding.namespace
  return aiBindingKey({ namespace: binding.namespace, project })
}

/**
 * 组装命令行自动执行的 mai 命令：
 * `mai [--unbind] [--bind '<空间>' | --bind '<空间>/<项目>']... ['<首条问题>']`。
 * 每个绑定各占一个 `--bind`（可多选），问题单独作为位置参数；所有值经单引号转义。
 * 绑定只提供默认上下文，不代发问题。
 *
 * 绑定为空时显式下发 `--unbind`：会话是按目录复用的，之前注入的「当前绑定范围」
 * 藏在历史里，只执行裸 `mai` 不会清除它（AI 仍会按旧范围作答），必须显式覆盖。
 */
function buildMaiCommand(bindings: AIBinding[], prompt?: string): string {
  const parts = ['mai']
  // 降级可能让多个绑定收敛到同一空间：按命令里的范围去重，避免重复 --bind。
  const scopes = [...new Set(bindings.map(cliBindScope))]
  if (scopes.length === 0) {
    parts.push('--unbind')
  } else {
    for (const scope of scopes) {
      parts.push('--bind', shellQuote(scope))
    }
  }
  if (prompt) parts.push(shellQuote(prompt))
  return parts.join(' ')
}

/** 入口问题的归一化：空白折叠成单行（避免换行被 shell 提前提交），空串按「无问题」处理。 */
function normalizePrompt(prompt?: string): string | undefined {
  const value = prompt?.replace(/\s+/g, ' ').trim()
  return value ? value : undefined
}

/** 命令行启动上下文：绑定列表 + 一次性问题，两者总是原子更新。 */
interface LaunchContext {
  bindings: AIBinding[]
  /** 一次性问题：只随首次/场景入口下发，改绑后清空。 */
  prompt?: string
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
export function AIAgentLauncher({
  open,
  onClose,
  initialPrompt,
  initialBinding,
  assistantTitle,
  entryNonce = 0,
}: {
  open: boolean
  onClose: () => void
  /** 场景入口带来的首条问题：非空时命令行执行 `mai '<prompt>'`，进入交互式 TUI 并自动发送。 */
  initialPrompt?: string
  /** 项目卡片入口带来的默认绑定：命令行执行 `mai --bind '<空间>/<项目>'`，只绑定不代发问题。 */
  initialBinding?: AIBinding
  /** AI 助手定制标题（如「duc 的 AI 助手」）：沙箱项目名/空间不展示给用户 */
  assistantTitle?: string
  /** 入口点击序号：弹窗已打开时再次点击入口会递增，用于只处理一次新上下文（见下方 effect）。 */
  entryNonce?: number
}) {
  const { t } = useTranslation()
  // 场景入口的首条问题作为 `mai` 的位置参数下发：mars-agent 会把它投递到当前目录的
  // 会话并 attach 交互式 TUI（问题已发送、回复在流式输出），不再依赖浏览器端按键时序。
  // 绑定与问题合并成一个状态原子更新：避免「问题已清空、绑定还没换」的中间态。
  const [launch, setLaunch] = useState<LaunchContext>(() => ({
    bindings: initialBinding ? [initialBinding] : [],
    prompt: normalizePrompt(initialPrompt),
  }))
  const autoShellCommand = useMemo(
    () => buildMaiCommand(launch.bindings, launch.prompt),
    [launch],
  )
  const [project, setProject] = useState<ProjectModel | null>(null)
  const [loading, setLoading] = useState(false)
  // 等待阶段：waking=正在触发唤醒；starting=已扩容、等容器就绪（首次启动可能较久）。
  const [phase, setPhase] = useState<'waking' | 'starting'>('waking')
  // 每次打开递增：关闭后仍在飞行的请求结果直接丢弃，防止旧结果把新一次的弹窗顶掉。
  const runIdRef = useRef(0)

  /** 切换绑定：命令行会话随 initialCommand 变化重建，注入新的默认项目上下文；
   *  首条问题只随首次进入下发，改绑不重复发送。 */
  const handleBindingChange = useCallback((next: AIBinding[]) => {
    setLaunch({ bindings: next, prompt: undefined })
  }, [])

  /** 已处理过的入口点击序号：弹窗保持挂载时，每次点击入口都带来新的上下文，
   *  按序号只处理一次；用对象身份判断会把「重复点同一张卡片」误当成新绑定。 */
  const handledEntryRef = useRef(entryNonce)

  /**
   * 弹窗已打开时再次点击入口：把新上下文合并进来。
   * - 追加尚未绑定的项目；重复点同一张卡片时什么都不变（不打断当前会话）；
   * - 场景问题（诊断/日志分析）覆盖上一轮问题；只绑定项目时清掉旧问题，避免旧问题带新绑定重发；
   * - 无 prompt 的裸入口（顶栏）保留现有上下文，不误杀正在进行的对话。
   */
  useEffect(() => {
    if (!open || handledEntryRef.current === entryNonce) return
    handledEntryRef.current = entryNonce
    setLaunch((prev) => {
      let bindings = prev.bindings
      if (initialBinding) {
        const key = aiBindingKey(initialBinding)
        if (!prev.bindings.some((b) => aiBindingKey(b) === key)) {
          bindings = [...prev.bindings, initialBinding]
        }
      }
      const appended = bindings !== prev.bindings
      const prompt = normalizePrompt(initialPrompt)
      // 追加了绑定 = 新上下文：旧的一次性问题作废，除非本次入口同时带了新问题。
      const nextPrompt = prompt !== undefined ? prompt : appended ? undefined : prev.prompt
      if (!appended && nextPrompt === prev.prompt) return prev
      return { bindings, prompt: nextPrompt }
    })
  }, [open, entryNonce, initialPrompt, initialBinding])

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
    // 从本次点击开始计时：sandbox_ready 的耗时用于衡量冷启动体验。
    const startedAt = Date.now()
    setProject(null)
    setPhase('waking')
    setLoading(true)
    try {
      const initial = await fetchAIStatus()
      if (!alive()) return
      if (!initial.enabled) {
        reportAIUsage('status_disabled')
        toast.error(t('ai.disabled'))
        onClose()
        return
      }

      let status: AIStatus = initial
      // running 无需唤醒；starting 说明已有部署在扩容，直接进入等待。
      if (!status.agent || status.agent.sandbox === 'sleeping' || status.agent.sandbox === 'absent') {
        reportAIUsage('enable_start')
        toast.info(t('ai.waking'))
        try {
          status = await enableAIAgent()
          reportAIUsage('enable_ok')
        } catch (e) {
          reportAIUsage('enable_fail')
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
        reportAIUsage('sandbox_ready', Date.now() - startedAt)
      }
      if (!status.agent) throw new Error(t('ai.notDeployed'))

      const detail = await loadDetail(status.agent.project_id)
      if (!alive()) return
      try {
        localStorage.setItem(AI_LAST_USED_STORAGE_KEY, String(Date.now()))
      } catch {
        // 隐私模式/禁用存储：预热退化为不生效，不影响本次使用。
      }
      reportAIUsage('sandbox_open')
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

  // 弹窗打开期间的心跳：项目详情渲染出来（沙箱已就绪）后先续一次，之后仅在
  // 页面可见时按分钟续活。后台标签页不续活，避免「开着页签去吃饭」把沙箱
  // 永久钉在 1 副本；启动阶段的活跃度由 enable / 容器拉配置自然刷新。
  useEffect(() => {
    if (!open || !project) return
    touchAIActivity()
    const timer = window.setInterval(() => {
      if (document.visibilityState === 'visible') touchAIActivity()
    }, AI_ACTIVITY_HEARTBEAT_MS)
    return () => window.clearInterval(timer)
  }, [open, project])

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
            titleOverride={assistantTitle}
            open
            // AI 入口默认落到命令行 Tab：用户点 AI 图标就是为了直接对话
            defaultTab="shell"
            // 沙箱只需要命令行，其余 Tab 对 AI 场景没有意义
            allowedTabs={AI_SANDBOX_TABS}
            // 场景问题随 `mai '<问题>'` 一起下发，由 mars-agent 投递并进入交互式 TUI；
            // 无问题时执行裸 `mai`。
            autoShellCommand={autoShellCommand}
            // 绑定条 + 固定常见问题：展示在命令行上方。改绑会重建命令行会话让新的
            // 默认上下文生效；常见问题只复制文本，不影响会话。
            aiBindingBar={
              <div className="flex flex-col gap-2">
                <AIBindingBar
                  bindings={launch.bindings}
                  excludeProjectId={project.id}
                  onChange={handleBindingChange}
                />
                <AIQuickPrompts />
              </div>
            }
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
