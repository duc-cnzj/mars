import { Component, type ErrorInfo, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

/**
 * 全局错误边界：子树渲染异常时不白屏，展示可恢复的兜底态。
 * 通过 key 触发重置（路由切换时由调用方 key={location.pathname} 重建）。
 */
export class ErrorBoundary extends Component<
  { children: ReactNode; fallback?: ReactNode },
  { error: Error | null }
> {
  state = { error: null as Error | null }

  static getDerivedStateFromError(error: Error) {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // 生产兜底：记录到 console，避免整页崩溃无痕
    console.error('[ErrorBoundary]', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      return this.props.fallback ?? <ErrorFallback error={this.state.error} />
    }
    return this.props.children
  }
}

/** 兜底 UI：错误信息 + 重新加载 */
function ErrorFallback({ error }: { error: Error }) {
  const { t } = useTranslation()
  return (
    <div className="flex min-h-[40vh] flex-col items-center justify-center gap-3 p-8 text-center">
      <div className="text-[13px] font-medium text-err">{t('common.error')}</div>
      <pre className="max-w-full overflow-auto rounded-lg bg-raised px-4 py-3 font-mono text-[12px] text-mute">
        {error.message || String(error)}
      </pre>
      <button
        type="button"
        className="rounded-md border border-line bg-surface px-4 py-1.5 text-[12px] text-ink transition-colors hover:bg-raised"
        onClick={() => window.location.reload()}
      >
        {t('common.refresh')}
      </button>
    </div>
  )
}
