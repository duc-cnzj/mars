/**
 * token / 会话辅助存储。
 * 与旧前端共用 localStorage 的 `token` 键（值带 `Bearer ` 前缀），
 * 保证同源下已登录会话无缝衔接。
 */

const TOKEN_KEY = 'token'
const LOGOUT_URL_KEY = 'logout_url'
const STATE_KEY = 'state'

/** 写入 token（自动补 `Bearer ` 前缀，与后端 Authorization 头约定一致） */
export const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, `Bearer ${token}`)
}

/** 读取完整 Authorization 值（无 token 返回空串） */
export const getToken = (): string => localStorage.getItem(TOKEN_KEY) ?? ''

/** 清除 token */
export const removeToken = (): void => {
  localStorage.removeItem(TOKEN_KEY)
}

/** 记住 SSO 登出地址 */
export const setLogoutUrl = (url: string): void => {
  localStorage.setItem(LOGOUT_URL_KEY, url)
}

export const getLogoutUrl = (): string => localStorage.getItem(LOGOUT_URL_KEY) ?? ''

export const removeLogoutUrl = (): void => {
  localStorage.removeItem(LOGOUT_URL_KEY)
}

/** 登录页是否使用随机壁纸（沿用旧前端 `random_bg` 键） */
export const isRandomBg = (): boolean => localStorage.getItem('random_bg') === '1'

/** 切换随机壁纸开关，返回切换后的状态 */
export const toggleRandomBg = (): boolean => {
  const next = !isRandomBg()
  localStorage.setItem('random_bg', next ? '1' : '0')
  return next
}

/** OIDC state（防 CSRF） */
export const setState = (state: string): void => {
  localStorage.setItem(STATE_KEY, state)
}

export const getState = (): string => localStorage.getItem(STATE_KEY) ?? ''

export const removeState = (): void => {
  localStorage.removeItem(STATE_KEY)
}
