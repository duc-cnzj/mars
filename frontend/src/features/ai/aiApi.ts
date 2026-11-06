import { API } from '@/api/endpoints'
import { getToken } from '@/api/token'

/**
 * AI 使用漏斗事件名，与服务端 isAIUsageEvent 白名单保持一致。
 * 只记录入口点击与对话生命周期，其他事件一律被服务端拒绝。
 */
export type AIUsageEvent = 'entry_click' | 'chat_first_token' | 'chat_done' | 'chat_error'

/**
 * 上报一个 AI 使用漏斗事件（POST /api/ai/usage）。
 * 埋点是尽力而为：网络失败静默吞掉，绝不能影响对话主流程。
 */
export function reportAIUsage(event: AIUsageEvent): void {
  void postAIEvent(event)
}

/** 发送无响应体的使用上报；所有失败（网络/HTTP）都静默。 */
async function postAIEvent(event: AIUsageEvent): Promise<void> {
  try {
    await fetch(API.aiUsage, {
      method: 'POST',
      headers: {
        Authorization: getToken(),
        'X-Requested-With': 'XMLHttpRequest',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ event }),
    })
  } catch {
    // 尽力而为的埋点，不让网络错误冒泡。
  }
}
