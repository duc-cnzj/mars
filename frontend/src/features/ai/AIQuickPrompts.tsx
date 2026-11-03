import { useTranslation } from 'react-i18next'
import { AISparklesIcon } from '@/components/Icons'
import { copyText } from '@/lib/copy'
import { toast } from '@/lib/toast'
import type { TKey } from '@/i18n/keys'

/**
 * 固定常见问题：覆盖 AI 在沙箱里最高频的三件事——部署最新代码、容器未就绪排查、
 * 容器日志错误分析。label 是按钮文案，prompt 是点击后复制到剪贴板的完整问题。
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
 * 常见问题条：渲染在绑定条下方，点击只把完整问题复制到剪贴板，由用户自己粘贴到
 * `mai` 输入框发送——不重建终端会话、不模拟按键，也不发起任何网络请求。
 */
export function AIQuickPrompts() {
  const { t } = useTranslation()
  /** 点击问题：复制完整问题文本，成功/失败用 toast 反馈。 */
  const handlePick = (prompt: string) => {
    void copyText(prompt).then((ok) => {
      if (ok) {
        toast.success(t('common.copied'))
        return
      }
      toast.error(t('common.copyFailed'))
    })
  }
  return (
    <div className="flex flex-wrap items-center gap-1.5 text-[12px]">
      <span className="inline-flex items-center gap-1.5 font-medium text-primary">
        <AISparklesIcon className="text-[14px]" />
        {t('ai.quickPromptsLabel')}
      </span>
      {QUICK_PROMPTS.map((item) => (
        <button
          key={item.label}
          type="button"
          onClick={() => handlePick(t(item.prompt))}
          className="rounded-full border border-primary/30 px-2 py-0.5 text-[11px] text-primary transition-colors hover:bg-primary/10"
        >
          {t(item.label)}
        </button>
      ))}
    </div>
  )
}
