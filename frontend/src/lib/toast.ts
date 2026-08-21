import { toast as sonnerToast } from 'sonner'

/**
 * 全局 toast 适配层：sonner 风格 API（底层 sonner，见 App.tsx 挂载的 <Toaster/>）。
 * - 裸调用 `toast(msg)` → info（对齐旧版 NotifyX 默认态语义，sonner 裸调用是 data-type=default
 *   中性态，会失去主题色相映射，因此显式映射到 info）
 * - `toast.success / error / warning / info` 一一对应
 * 明暗主题由 App 挂载的 <Toaster theme={...}/> 按当前主题同步。
 */
export interface ToastFn {
  (message: string): string | number
  success: (message: string) => string | number
  error: (message: string) => string | number
  warning: (message: string) => string | number
  info: (message: string) => string | number
}

export const toast: ToastFn = Object.assign(
  (message: string) => sonnerToast.info(message),
  {
    success: (message: string) => sonnerToast.success(message),
    error: (message: string) => sonnerToast.error(message),
    warning: (message: string) => sonnerToast.warning(message),
    info: (message: string) => sonnerToast.info(message),
  },
)
