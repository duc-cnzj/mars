import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { components } from '@/api/schema'
import { SearchableSelect, type SearchableSelectOption } from '@/components/SearchableSelect'
import { cn } from '@/lib/utils'
import { hibernateWindowTimeRange } from './hibernate'

type HibernateWindow = components['schemas']['hibernate.HibernateWindow']

/**
 * 项目休眠时间窗选择器（自包含字段）：字段名 + 全宽多选 + 生效时段一览。
 *
 * 项目从所属仓库声明的窗列表里多选一个子集——该子集即项目生效窗，「空选 = 未选」（启用时必须
 * 至少选一个，不再继承仓库全量）。选项限定在仓库的 windowNames 内——后端校验子集归属，越界即
 * 400；标签为窗名、说明为该窗「HH:MM-HH:MM」（全局列表未加载或窗悬空时说明缺省，仅显示窗名）。
 *
 * 下方一行「生效时段」给出当前真正生效的并集时段（多选合一），随选择实时变化——补齐 trigger
 * 只列窗名、看不出时段的空白，用户无需开下拉即可确认生效范围。
 *
 * 宿主为配置更新页 TabEdit 的月亮弹层：标签与语义提示收进组件，宿主不再自己拼一行 label + tip。
 */
export function HibernateWindowSelect({
  value,
  repoWindowNames,
  windows,
  onChange,
  disabled,
  className,
}: {
  /** 项目当前选定的窗名子集（空数组 = 未选，启用时不允许） */
  value: string[]
  /** 所属仓库声明的可用窗名列表（选择范围） */
  repoWindowNames: string[]
  /** 全局时间窗列表（拼时段说明；未加载时选项仅显示窗名、生效时段行不渲染） */
  windows: HibernateWindow[]
  onChange: (value: string[]) => void
  /** 保存进行中：封住多选交互。受控组件下父级守卫会吞掉 onChange，若不封交互，勾选会「回弹」误导用户 */
  disabled?: boolean
  className?: string
}) {
  const { t } = useTranslation()
  const noRepo = repoWindowNames.length === 0
  const options = useMemo<SearchableSelectOption[]>(
    () =>
      repoWindowNames.map((name) => {
        const w = windows.find((it) => it.name === name)
        return { value: name, label: name, description: w ? `${w.start}-${w.end}` : undefined }
      }),
    [repoWindowNames, windows],
  )
  // 生效窗 = 项目选定子集本身；时段取并集（未命中/未加载的窗跳过，可能为空串）
  const ranges = hibernateWindowTimeRange(value, windows)

  return (
    <div className={cn('flex flex-col gap-2', className)}>
      {/* 字段头：左字段名、右语义提示（必选，至少一个）——两端对齐，不再挤成一行 */}
      <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-0.5">
        <span className="text-[13px] font-medium text-ink">{t('project.hibernateWindow')}</span>
        {!noRepo && (
          <span className="text-[11px] text-faint">{t('project.hibernateWindowSelectTip')}</span>
        )}
      </div>

      {/* 保存中封住多选：SearchableSelect 由 value 受控，父级守卫吞掉 onChange 时选中态不会落地，
          若不禁用则用户点了没反应（勾选回弹）。用禁用层封鼠标/点击，与同页开关的禁用态对齐 */}
      <div className={cn(disabled && 'pointer-events-none opacity-60')} aria-disabled={disabled}>
        <SearchableSelect
          multiple
          value={value}
          options={options}
          onChange={(v) => onChange(Array.isArray(v) ? v : [v])}
          placeholder={
            noRepo ? t('project.hibernateWindowNoneInRepo') : t('project.hibernateWindowSelectTip')
          }
          searchPlaceholder={t('common.search')}
          emptyText={t('project.hibernateWindowNoneInRepo')}
        />
      </div>

      {/* 生效时段：多窗并集「HH:MM-HH:MM」逗号连接；无可用时段（仓库未声明/全悬空/列表未加载）不占位 */}
      {ranges && (
        <div className="flex items-baseline gap-1.5 text-[12px] text-mute">
          <span className="shrink-0 text-faint">{t('project.hibernateEffective')}</span>
          <span className="font-mono">{ranges}</span>
        </div>
      )}
    </div>
  )
}
