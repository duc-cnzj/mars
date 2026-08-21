import type { SVGProps } from 'react'

/**
 * iconfont 精灵图图标（旧版 Mars 品牌图标）。
 * 符号由 public/iconfont.js 注入到文档的隐藏 <svg> sprite，这里用 <use> 引用。
 * 用法：<IconFont name="#icon-naicha" className="size-6" />
 */
export function IconFont({
  name,
  className,
  ...rest
}: { name: string; className?: string } & Omit<SVGProps<SVGSVGElement>, 'name' | 'children'>) {
  return (
    <svg
      className={className}
      width="1em"
      height="1em"
      fill="currentColor"
      aria-hidden
      {...rest}
    >
      <use xlinkHref={name} />
    </svg>
  )
}
