import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { Slot } from "radix-ui"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap transition-[background-color,border-color,box-shadow,color,scale] duration-150 outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 active:scale-98 active:duration-0 disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        /* 主操作：品牌渐变（135deg primary→primary-strong，对齐登录按钮）+ primary 色投影辉光，
           hover 上浮 1px 且投影加深（「浮起」），active 回缩复位（「按下」） */
        default:
          "bg-[linear-gradient(135deg,var(--primary),var(--primary-strong))] text-primary-foreground shadow-sm shadow-primary/25 hover:shadow-md hover:shadow-primary/40 active:shadow-sm active:shadow-primary/20",
        destructive:
          "bg-destructive text-white shadow-sm shadow-destructive/25 hover:bg-destructive/90 hover:shadow-md hover:shadow-destructive/35 focus-visible:ring-destructive/20 dark:bg-destructive/60 dark:focus-visible:ring-destructive/40 active:shadow-sm active:bg-destructive/85",
        /* 描边按钮：透明底 + 带主色相的描边（border-primary/40，比 border-strong 更明显且与 hover 同色相过渡） */
        outline:
          "border border-primary/40 bg-transparent text-ink hover:border-primary hover:bg-primary-soft hover:text-primary active:border-primary-strong active:bg-primary-soft active:text-primary-strong",
        /* 虚线按钮：outline 的虚线版（border-dashed），语义=「次级/浏览/添加」入口，
           与 outline 共享「hover 点亮 primary」交互，替代散落的 className="border-dashed" 覆盖 */
        dashed:
          "border border-dashed border-primary/40 bg-transparent text-ink hover:border-primary hover:bg-primary-soft hover:text-primary active:border-primary-strong active:bg-primary-soft active:text-primary-strong",
        /* 实心次级：hover 点亮为 primary（灰底 → primary-soft 填充 + primary 文字），
           与 outline/dashed 统一的「次级按钮 hover 点亮」交互语言 */
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-primary-soft hover:text-primary active:bg-primary-soft active:text-primary-strong",
        ghost:
          "text-mute hover:bg-raised/60 hover:text-ink active:bg-raised/80 active:text-ink",
        link: "text-primary underline-offset-4 hover:text-primary-strong active:text-primary-strong",
      },
      size: {
        default: "h-9 px-4 py-2 has-[>svg]:px-3",
        xs: "h-6 gap-1 rounded-md px-2 text-xs has-[>svg]:px-1.5 [&_svg:not([class*='size-'])]:size-3",
        sm: "h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5",
        lg: "h-10 rounded-md px-6 has-[>svg]:px-4",
        icon: "size-9",
        "icon-xs": "size-6 rounded-md [&_svg:not([class*='size-'])]:size-3",
        "icon-sm": "size-8",
        "icon-lg": "size-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

const Button = React.forwardRef<
  HTMLButtonElement,
  React.ComponentProps<"button"> &
    VariantProps<typeof buttonVariants> & {
      asChild?: boolean
    }
>(({ className, variant = "default", size = "default", asChild = false, ...props }, ref) => {
  const Comp = asChild ? Slot.Root : "button"

  return (
    <Comp
      ref={ref}
      data-slot="button"
      data-variant={variant}
      data-size={size}
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  )
})
Button.displayName = "Button"

export { Button, buttonVariants }
