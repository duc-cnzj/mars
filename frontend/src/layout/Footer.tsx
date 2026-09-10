import { formatDate } from '@/lib/format'
import { useTranslation } from 'react-i18next'
import { useVersion } from '@/hooks/useVersion'
import { useGrayChannel } from '@/hooks/useGrayChannel'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui'
import { Coffee as CoffeeCard } from './Coffee'
import { IconFont } from '@/components/IconFont'
import { barGradient, type ThemeId } from '@/themes'

/**
 * 渐变底栏（还原旧版 AppFooter）：版权 + 版本号 + 构建时间 + 请喝咖啡赞助入口。
 * 沿用旧版：dank mono 品牌字体 + 两行结构（line1 created by，line2 version/buildAt）居中。
 * 背景渐变与 header 保持一致（barGradient 按明暗分支），文字白。
 * 处于灰度通道时额外在版本号右侧挂一枚「抢先体验」标记（见下方注释）。
 */
export function Footer({ theme }: { theme: ThemeId }) {
  const { t } = useTranslation()
  const version = useVersion()
  const year = new Date().getFullYear()
  // 灰度通道标记：紧挨版本号——版本号回答「我跑的是哪个构建」，这枚标记回答同一问题的
  // 后半句「以及它在哪条发布通道上」，两者天然同源。
  // 判据是灰度路由 cookie 而非后端 isGray：标记要反映「此刻真正生效的通道」，而 cookie
  // 才是 ingress 分流时实际读的那个东西（详见 useGrayChannel 的注释）。
  // 颜色全部由 currentColor 派生（color-mix），不引入硬编码色值：底栏前景色已按主题调好，
  // 标记跟着走即可，明暗主题都不用另配一套。
  const isGray = useGrayChannel()

  return (
    <footer
      className="flex shrink-0 flex-col items-center justify-center gap-0 px-4 py-1.5 text-center leading-none text-primary-foreground/90 sm:px-10"
      style={
        {
          fontFamily: '"dank mono", -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
          // 与 header 一致：深色主题压暗、浅色主题亮渐变；前景取各主题自带 primary-foreground，不推白
          background: barGradient(theme),
        }
      }
    >
      <span style={{ fontSize: 13 }}>{t('footer.copyright', { year })}</span>
      {version && (
        <span className="flex items-center justify-center leading-none" style={{ fontSize: 11 }}>
          {t('footer.version', { version: version.version })}
          <span className="mx-1">,</span>
          {t('footer.buildAt', {
            date: formatDate(version.buildDate),
          })}
          {isGray && (
            <span
              className="ml-2 inline-flex items-center gap-1 rounded-full px-1.5 leading-none"
              style={{
                fontSize: 10,
                paddingTop: 1,
                paddingBottom: 1,
                border: '1px solid color-mix(in srgb, currentColor 40%, transparent)',
                background: 'color-mix(in srgb, currentColor 14%, transparent)',
              }}
              title={t('footer.grayTip')}
            >
              <span
                className="inline-block size-[5px] rounded-full"
                style={{ background: 'currentColor' }}
              />
              {t('footer.grayChannel')}
            </span>
          )}
          <Popover>
            <PopoverTrigger asChild>
              <button
                type="button"
                aria-label={t('coffee.title')}
                className="ml-2.5 grid cursor-pointer place-items-center rounded transition-transform duration-150 hover:scale-110"
              >
                <IconFont name="#icon-naicha" className="size-[22px] scale-x-[-1]" />
              </button>
            </PopoverTrigger>
            {/* 打赏卡片 z-[10000] 压过一切弹层：下拉/主题切换（9998/9999）、confetti（9999）、
                可拖拽弹窗共享计数器（51 起累加）——赞助入口任何时候不被遮挡 */}
            <PopoverContent className="z-[10000] w-[230px] p-2" align="center" sideOffset={8}>
              <CoffeeCard />
            </PopoverContent>
          </Popover>
        </span>
      )}
    </footer>
  )
}
