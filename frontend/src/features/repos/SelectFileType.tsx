import { useTranslation } from 'react-i18next'
import { FILE_TYPES } from '@/components/CodeEditor'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/shadcn/select'

/** 把文件类型值转成下拉展示文案（env 显示为 .env） */
const fileTypeLabel = (v: string): string => (v === 'env' ? '.env' : v)

/**
 * 配置文件类型选择器：列出 CodeEditor 的 FILE_TYPES 全部 55 种。
 * 还原旧版 SelectFileType.tsx 的候选集 + antd Select showSearch 可搜索能力
 * （shadcn SelectContent 支持 searchPlaceholder，弹层顶部渲染搜索框）。
 * value 可为自由字符串：不在候选里时自动追加为兜底项，避免回填值显示为空
 * （空串视为未选择，不追加，触发框显示 placeholder）。
 */
export function SelectFileType({
  value,
  onChange,
  className,
  placeholder,
}: {
  value: string
  onChange: (value: string) => void
  className?: string
  placeholder?: string
}) {
  const { t } = useTranslation()
  const options =
    value && !(FILE_TYPES as readonly string[]).includes(value)
      ? ([value, ...FILE_TYPES] as const)
      : FILE_TYPES

  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger className={className ?? 'w-full'}>
        <SelectValue placeholder={placeholder ?? t('repos.configFileType')} />
      </SelectTrigger>
      <SelectContent
        searchPlaceholder={t('common.search')}
        emptyText={t('common.empty')}
      >
        {options.map((ft) => (
          <SelectItem key={ft} value={ft}>
            {fileTypeLabel(ft)}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
