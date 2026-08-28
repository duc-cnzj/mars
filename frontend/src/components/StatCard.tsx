import { Icon, type IconName } from '@/components/Icons'

/** 统计卡片：图标 + 指标名 + 数值（tone 驱动主题语义色）。多后台页共用（用户/命名空间等）。 */
export function StatCard({
  label,
  value,
  icon,
  tone,
}: {
  label: string
  value: number
  icon: IconName
  tone: 'mute' | 'accent' | 'ok'
}) {
  const tones: Record<string, string> = {
    mute: 'text-mute bg-raised',
    accent: 'text-primary bg-primary-soft',
    ok: 'text-ok bg-ok-soft',
  }
  return (
    <section className="rounded-lg border border-line bg-surface p-4">
      <div className="flex items-center gap-2 text-[12px] text-faint">
        <span className={`grid size-6 place-items-center rounded-md ${tones[tone]}`}>
          <Icon name={icon} className="size-3.5" />
        </span>
        {label}
      </div>
      <div className="mt-2 font-mono text-[24px] font-semibold leading-none tabular-nums text-ink">
        {value}
      </div>
    </section>
  )
}
