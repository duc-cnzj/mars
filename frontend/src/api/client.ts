import createClient, { type Middleware } from 'openapi-fetch'
import type { paths } from './schema'
import { getToken, removeToken, getLogoutUrl, removeLogoutUrl } from './token'
import { toast } from '@/lib/toast'
import i18n from '../i18n'

/** 401 提示去抖：同一时间窗口内只弹一次，避免并发请求刷屏 */
let last401Toast = 0

function loginExpiredToast() {
  const now = Date.now()
  if (now - last401Toast < 2000) return
  last401Toast = now
  toast.error(i18n.t('auth.loginExpired'))
}

/**
 * 全局 API 客户端：基于后端 openapi 生成类型，零手写接口定义。
 * - 请求统一挂 Authorization 头
 * - 401 统一清除 token 并跳 SSO 登出地址（getLogoutUrl）或 /login，
 *   同时弹"登录过期"提示（复用旧前端语义）
 */
export const api = createClient<paths>({
  baseUrl: '',
  headers: {
    'X-Requested-With': 'XMLHttpRequest',
    'Accept-Language': 'zh',
  },
})

const authMiddleware: Middleware = {
  async onRequest({ request }) {
    request.headers.set('Authorization', getToken())
    return request
  },
  async onResponse({ response }) {
    if (response.status === 401 && getToken()) {
      removeToken()
      loginExpiredToast()
      // 延迟跳转，避免在 /login 本身报 401 时死循环
      if (window.location.pathname !== '/login') {
        const href = getLogoutUrl() || '/login'
        removeLogoutUrl()
        window.location.href = href
      }
    }
    return response
  },
}

api.use(authMiddleware)
