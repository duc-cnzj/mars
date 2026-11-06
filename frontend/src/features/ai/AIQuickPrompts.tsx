import { useTranslation } from 'react-i18next'
import { AISparklesIcon } from '@/components/Icons'
import type { TKey } from '@/i18n/keys'

/**
 * 固定常见问题：覆盖 AI 最高频的三件事——部署最新代码、容器未就绪排查、
 * 容器日志错误分析。label 是按钮文案，prompt 是点击后填入输入框的完整问题。
 *
 * 文案固定、不做网络请求、不读取用户数据：行为可预期，也不会因为模型/事件
 * 数据波动而给出无关建议。
 */
const QUICK_PROMPTS: readonly { label: TKey; prompt: TKey }[] = [
  { label: 'ai.quickPromptDeploy', prompt: 'ai.quickPromptDeployPrompt' },
  { label: 'ai.quickPromptNotReady', prompt: 'ai.quickPromptNotReadyPrompt' },
  { label: 'ai.quickPromptLogs', prompt: 'ai.quickPromptLogsPrompt' },
]

/**
 * 常见问题条：空白对话时展示在消息列表中间，点击只把完整问题填入输入框，
 * 由用户确认/补充后再发送——避免误触直接消耗一轮模型调用。
 * 不在生成中/等待确认时展示，避免打断正在进行的对话。
 */
export function AIQuickPrompts({ onFill }: { onFill: (prompt: string) => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col items-center gap-2">
      <span className="inline-flex items-center gap-1.5 text-[12px] font-medium text-primary">
        <AISparklesIcon className="text-[14px]" />
        {t('ai.quickPromptsLabel')}
      </span>
      <div className="flex flex-wrap justify-center gap-1.5 text-[12px]">
        {QUICK_PROMPTS.map((item) => (
          <button
            key={item.label}
            type="button"
            onClick={() => onFill(t(item.prompt))}
            className="rounded-full border border-primary/30 px-2.5 py-1 text-[11px] text-primary transition-colors hover:bg-primary/10"
          >
            {t(item.label)}
          </button>
        ))}
      </div>
    </div>
  )
}
