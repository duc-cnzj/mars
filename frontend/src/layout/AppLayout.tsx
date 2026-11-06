import { Suspense, lazy, useCallback, useEffect, useMemo, useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Topbar } from './Topbar'
import { Footer } from './Footer'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { fetchAIChatStatus } from '@/features/ai/aiChatApi'
import { reportAIUsage } from '@/features/ai/aiApi'
import { AIEntryProvider, type AIBinding, type AIEntry } from '@/features/ai/AIContext'
import { useAuth } from '@/features/auth/AuthProvider'
import type { ThemeId } from '@/themes'

// 顶栏 AI 入口：服务端对话面板，按需加载（react-markdown 等依赖不进入首屏）。
const AIChatLauncher = lazy(() =>
  import('@/features/ai/AIChatLauncher').then((m) => ({ default: m.AIChatLauncher })),
)

/**
 * 应用整体布局（旧版风格）：全宽渐变顶栏 + 居中内容区 + 渐变底栏。
 * 无侧边栏，导航收在顶栏用户下拉，品牌字标点击回工作台。
 *
 * 页面随文档整体滚动，footer 位于内容末尾（不固定到底部）——对齐旧版 antd Layout 行为。
 * 需要容器内滚动的页面（events/tokens）自足高度 calc(100dvh - 顶栏 - padding)，
 * 不依赖布局提供有界高度；仓库管理已迁入管理后台（/admin/repos），由 AdminLayout 提供有界滚动容器。
 * 管理后台（/admin/*）为控制台形态：隐藏底栏，文档高度锁定 100dvh，侧栏导航与面包屑常驻固顶。
 * 侧边距响应式收窄（50px 固定值在小屏会吃掉近半屏）。
 */
export function AppLayout({
  theme,
  onSelectTheme,
}: {
  theme: ThemeId
  onSelectTheme: (t: ThemeId) => void
}) {
  const location = useLocation()
  const { t } = useTranslation()
  const { user } = useAuth()
  // AI 助手弹窗标题：对用户显示「<用户名> 的 AI 助手」；用户信息尚未恢复时退回通用标题。
  const assistantTitle = user?.name
    ? t('ai.assistantTitle', { name: user.name })
    : t('ai.open')
  // 管理后台（/admin/*）是控制台形态：隐藏营销底栏，文档高度恰好锁死 100dvh 不产生页面滚动，
  // AdminLayout 的侧栏导航 + 面包屑才能真正常驻固顶（否则 footer 会把文档撑出视口高，导航随页漂移）。
  const isAdmin = location.pathname.startsWith('/admin')
  // AI 助手开关：面板挂布局层，顶栏与页面内场景入口共用同一份对话状态。
  const [aiOpen, setAiOpen] = useState(false)
  // 入口上下文（场景问题 + 项目卡片绑定）与点击序号合并成一个状态：
  // nonce 每次点击递增，编排层据此识别「面板已打开时又点了一次入口」。
  const [aiLaunch, setAiLaunch] = useState<{
    nonce: number
    prompt?: string
    binding?: AIBinding
  }>({ nonce: 0 })
  // null = 尚未查询（按可用处理，避免网络失败把入口永久藏掉）；false = 明确未启用。
  const [aiEnabled, setAiEnabled] = useState<boolean | null>(null)

  // 布局挂载时查询一次对话可用性：未配置时隐藏顶栏入口，避免用户点了才看到 503；
  // 查询失败保持可见，点击后由对话请求给出明确报错。
  useEffect(() => {
    let alive = true
    void fetchAIChatStatus()
      .then((status) => {
        if (alive) setAiEnabled(status.enabled)
      })
      .catch(() => {
        if (alive) setAiEnabled(true)
      })
    return () => {
      alive = false
    }
  }, [])

  /** 打开 AI 入口：先记漏斗事件，再交给对话面板处理上下文。 */
  const openAI = useCallback<AIEntry['open']>((prompt, binding) => {
    reportAIUsage('entry_click')
    // nonce 递增：面板已打开时再次点击入口，编排层据此只处理一次新上下文。
    setAiLaunch((prev) => ({ nonce: prev.nonce + 1, prompt, binding }))
    setAiOpen(true)
  }, [])

  /** 关闭并清空上下文：下次从顶栏或场景入口打开时不残留上一次的问句与绑定。 */
  const closeAI = useCallback(() => {
    setAiOpen(false)
    setAiLaunch({ nonce: 0 })
  }, [])

  // AI 入口对页面内场景（项目详情里的「AI 诊断」）可见：与顶栏同一套启用判定。
  const aiEntry = useMemo<AIEntry>(
    () => ({
      available: aiEnabled !== false,
      open: openAI,
    }),
    [aiEnabled, openAI],
  )

  return (
    // 页面随文档整体滚动，footer 位于内容末尾（不固定到底部）——对齐旧版 antd Layout 行为
    <div className="flex min-h-screen flex-col bg-bg text-ink">
      <Topbar
        theme={theme}
        onSelectTheme={onSelectTheme}
        onOpenAI={() => openAI()}
        aiEnabled={aiEnabled !== false}
      />
      <main className="flex-1 px-4 pt-6 pb-3 sm:px-6 lg:px-10">
        <div className="h-full w-full">
          {/* 错误边界按页面出口重建（key=pathname）：页面崩溃切页即恢复，
              但不卸载应用壳——Topbar/ClusterStatus/websocket 常驻，切页不重新拉取 */}
          <AIEntryProvider value={aiEntry}>
            <ErrorBoundary key={location.pathname}>
              <Outlet />
            </ErrorBoundary>
          </AIEntryProvider>
        </div>
      </main>
      {/* 管理后台无底栏（控制台形态）；工作台/事件/令牌保留营销底栏 */}
      {!isAdmin && <Footer />}

      {/* AI 对话面板（顶栏/场景入口）：服务端编排 */}
      <Suspense fallback={null}>
        {aiOpen && (
          <AIChatLauncher
            open
            onClose={closeAI}
            initialPrompt={aiLaunch.prompt}
            initialBinding={aiLaunch.binding}
            entryNonce={aiLaunch.nonce}
            assistantTitle={assistantTitle}
          />
        )}
      </Suspense>
    </div>
  )
}
