import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { AISparklesIcon } from '@/components/Icons'
import { SearchableSelect, type SearchableSelectOption } from '@/components/SearchableSelect'
import { aiBindingKey, type AIBinding } from './AIContext'

type NamespaceModel = components['schemas']['types.NamespaceModel']

/** 一次拉全量用于选择：绑定选择器只在弹层打开时调用，不做分页交互。 */
const BINDING_LIST_PAGE_SIZE = 500

/** 把选择器的字符串数组还原成绑定对象（保留用户选择顺序）：
 *  单段 `空间` = 绑定整个空间，两段 `空间/项目` = 只绑定该项目（与后端 bindings 结构一致）。 */
function toBindings(keys: string[]): AIBinding[] {
  return keys.map((key) => {
    const at = key.indexOf('/')
    if (at < 0) return { namespace: key }
    return { namespace: key.slice(0, at), project: key.slice(at + 1) }
  })
}

/** 比较两组 `空间/项目` 是否等价（顺序无关），避免仅改顺序就重建终端会话。 */
function sameSelection(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false
  const set = new Set(b)
  return a.every((key) => set.has(key))
}

/**
 * AI 绑定条：显示当前默认针对的「空间」或「空间/项目」（可多选），并可增删或清空绑定。
 *
 * 复用全局 `SearchableSelect`（multiple）：搜索、chips、弹层 z-index 与滚动锁都由
 * 组件内部处理，本组件只负责提供选项与提交时机。绑定随每次对话请求下发，
 * 只影响 system prompt 里的默认关注范围，不会代发任何问题。具体 Pod 不参与绑定：
 * 滚动更新后 Pod 名会变，按空间/项目绑定、由 AI 用 MCP 工具现查更稳。
 *
 * 提交时机：弹层打开期间只改草稿（勾选多个项目不会每点一次就触发一次对话状态变更），
 * 关闭弹层时一次性提交；关闭态直接点 chip 上的 X 则立即提交。
 * 换绑会重置对话，因此生成中/等待确认时整条置灰禁用（disabled），避免打断进行中的轮次。
 */
export function AIBindingBar({
  bindings,
  onChange,
  disabled = false,
}: {
  /** 当前绑定列表；为空表示未绑定（AI 按用户提问里明确写出的范围作答） */
  bindings: AIBinding[]
  /** 用户确认新的绑定列表或清空时回调 */
  onChange: (bindings: AIBinding[]) => void
  /** 禁用修改（生成中或等待写工具确认） */
  disabled?: boolean
}) {
  const { t } = useTranslation()
  const [namespaces, setNamespaces] = useState<NamespaceModel[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [failed, setFailed] = useState(false)
  // 弹层打开期间的草稿（`空间` / `空间/项目` 列表）；null = 弹层关闭，此时直接展示 props。
  const [draft, setDraft] = useState<string[] | null>(null)

  const selected = useMemo(() => bindings.map(aiBindingKey), [bindings])

  /** 打开弹层时按需拉一次可见空间/项目列表；已有数据直接复用。 */
  const load = () => {
    if (namespaces || loading) return
    setLoading(true)
    setFailed(false)
    api
      .GET(API.namespaces, {
        params: { query: { page: 1, pageSize: BINDING_LIST_PAGE_SIZE } },
      })
      .then(({ data, error }) => {
        if (error) throw new Error(error.message ?? String(error))
        setNamespaces(data?.items ?? [])
      })
      .catch(() => setFailed(true))
      .finally(() => setLoading(false))
  }

  /** 选项 = 每个可见空间（绑定整个空间）+ 空间下的项目；
   *  已绑定但不在列表里的旧项（项目已删/权限收紧）由 SearchableSelect 回退显示原始值，
   *  不会凭空消失。
   *  label 用 `空间/项目` 全名：项目名在不同空间可能重名，只显示项目名无法分辨
   *  绑定到了哪个空间；空间项则标注「全部项目」，与单项目选项区分开。
   *  没有任何可见项目的空间不提供空间级选项，绑定一个没有可查询项目的空空间没有意义。
   *  chip 端靠 maxTagCount + 宽度截断控制长度。
   *  列表按需加载：未打开过弹层时用当前绑定兜底建选项，保证 chip 从首帧起就有 label。 */
  const options = useMemo<SearchableSelectOption[]>(() => {
    const map = new Map<string, SearchableSelectOption>()
    for (const ns of namespaces ?? []) {
      if (ns.projects.length === 0) continue
      map.set(ns.name, {
        value: ns.name,
        label: `${ns.name} · ${t('ai.bindingAllProjects')}`,
      })
      for (const p of ns.projects) {
        map.set(`${ns.name}/${p.name}`, {
          value: `${ns.name}/${p.name}`,
          label: `${ns.name}/${p.name}`,
        })
      }
    }
    for (const b of bindings) {
      const key = aiBindingKey(b)
      if (map.has(key)) continue
      map.set(key, { value: key, label: key })
    }
    return [...map.values()]
  }, [namespaces, bindings, t])

  /** 提交一组选择；与当前绑定等价时不重复通知父级。 */
  const commit = (keys: string[]) => {
    if (!sameSelection(keys, selected)) onChange(toBindings(keys))
  }

  /** 弹层开合：打开时开草稿并拉选项，关闭时把草稿一次性提交。 */
  const handleOpenChange = (open: boolean) => {
    if (disabled) return
    if (open) {
      setDraft(selected)
      load()
      return
    }
    if (draft !== null) commit(draft)
    setDraft(null)
  }

  /** 选择变化：弹层内只更新草稿；关闭态（chip X）立即提交。 */
  const handleChange = (value: string | string[]) => {
    if (disabled) return
    const keys = Array.isArray(value) ? value : [value]
    if (draft !== null) {
      setDraft(keys)
      return
    }
    commit(keys)
  }

  const emptyText = loading
    ? t('common.loading')
    : failed
      ? t('ai.bindingLoadFailed')
      : t('common.empty')

  return (
    <div
      aria-disabled={disabled}
      className={`flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] ${
        disabled ? 'pointer-events-none opacity-60' : ''
      }`}
    >
      <span className="inline-flex items-center gap-1.5 font-medium text-primary">
        <AISparklesIcon className="text-[14px]" />
        {t('ai.bindingLabel')}
      </span>
      <SearchableSelect
        multiple
        value={draft ?? selected}
        options={options}
        onChange={handleChange}
        onOpenChange={handleOpenChange}
        placeholder={t('ai.bindingNone')}
        searchPlaceholder={t('common.search')}
        emptyText={emptyText}
        // 无边框幽灵触发器：只渲染 chips，不再套一层输入框（卡片本身已是容器）。
        // 紧凑标签：最多展示 2 个绑定项，其余折叠成 +N（hover 看全量）；
        // 单个 chip 超宽截断，避免长空间名把触发器撑成多行。
        maxTagCount={2}
        tagClassName="max-w-[10rem] rounded-md bg-primary/10 px-1.5 text-[11px] text-foreground"
        className="min-h-0 w-fit max-w-full rounded-lg border-0 bg-transparent px-1 py-0.5 text-[12px] shadow-none transition-colors hover:bg-primary/5"
        // 触发器按 chips 收缩（w-fit），下拉若跟随 trigger 宽度会把 `空间/项目` 截断；
        // 弹层给独立宽度，保证选项与搜索框可读。
        contentClassName="w-[22rem] max-w-[calc(100vw-2rem)]"
      />
      <span className="text-faint">{t('ai.bindingHint')}</span>
    </div>
  )
}
