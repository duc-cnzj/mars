import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

/** 合并 Tailwind 类名：clsx 组合 + tailwind-merge 去冲突 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
