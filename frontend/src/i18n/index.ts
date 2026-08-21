import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import zhCN from './locales/zh-CN'
import en from './locales/en'

/** 让 t() 的 key 获得类型校验（基于中文词条推导） */
declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'translation'
    resources: { translation: typeof zhCN }
  }
}

export const LOCALE_KEY = 'mars.locale'

export type Locale = 'zh-CN' | 'en'

/** 读取用户语言偏好（localStorage，缺省中文） */
export const getLocale = (): Locale =>
  (localStorage.getItem(LOCALE_KEY) as Locale | null) ?? 'zh-CN'

/** 设置语言并写回偏好 */
export const setLocale = (l: Locale): void => {
  localStorage.setItem(LOCALE_KEY, l)
  void i18n.changeLanguage(l)
}

void i18n.use(initReactI18next).init({
  resources: {
    'zh-CN': { translation: zhCN },
    en: { translation: en },
  },
  lng: getLocale(),
  fallbackLng: 'zh-CN',
  interpolation: { escapeValue: false },
})

export default i18n
