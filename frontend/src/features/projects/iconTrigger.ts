import { cn } from '@/lib/utils'

/**
 * 弹层触发图标基类（几何 / 命中区 / 焦点环）：24px 命中区、可聚焦、点击不冒泡到父 button。
 * 列表行工具簇（CPU / 端点）与详情弹窗标题栏的休眠入口共用同一套几何，抽到本模块避免两处漂移。
 */
export const iconTriggerBase =
  'flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-md transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50'

/** 未激活态：淡色图标，hover 转主色 */
export const iconTriggerCls = cn(iconTriggerBase, 'text-faint hover:bg-raised hover:text-primary')

/** 休眠开启态：不着底色块，只把月亮本体填实成主色（图标切 moon-fill），与未激活的淡色描边月亮一眼可辨；
    hover 仅垫一层 raised 底提示可点，不改变月亮的实心主色 */
export const hibernateIconActive = cn(iconTriggerBase, 'text-primary hover:bg-raised')

/** 列表行休眠中 zzz 信号：与 CPU / 端点同为点击弹层触发（同一套 24px 命中区几何、可聚焦、点击不冒泡），
    仅保留专属蓝——不与工具簇的淡色图标混同，一眼可辨「已入睡」 */
export const hibernateSleepIcon = cn(iconTriggerBase, 'text-hibernate hover:bg-raised')
