import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { nextZIndex } from '@/lib/zIndex'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { useOverlayZ } from '@/hooks/useOverlayZ'
import { Icon } from '@/components/Icons'
import { Empty, Tag } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { DiffViewer } from '@/components/DiffViewer'

type ChangelogModel = components['schemas']['types.ChangelogModel']

/**
 * 配置历史：拉取项目配置改动日志，逐条展示「版本 + 更新人 + 提交 + 配置变更」，
 * 展开后显示与上一版本的逐行 diff（LCS）。
 */
export function ConfigHistory({
  projectId,
  configFileType,
}: {
  projectId: number
  /** 项目当前配置文件格式（外面配置编辑器的 language，来自 marsConfig.configFileType）。
   *   changelog 的 configType 可能为空串，导致 diff 无高亮、与外面配置观感脱节；此处兜底用项目真实格式。 */
  configFileType?: string
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  // 历史弹窗（portal 挂 body）须盖过可拖拽宿主弹窗 z-51+。
  // 触发源是普通 Button（非 Radix DialogTrigger），onOpenChange(true) 不会触发 → 只能由 effect 在打开时置顶
  const z = useOverlayZ(open)
  const [items, setItems] = useState<ChangelogModel[]>([])
  const [loading, setLoading] = useState(false)
  const [expanded, setExpanded] = useState<number | null>(null)

  const fetchHistory = async () => {
    setLoading(true)
    try {
      const { data, error } = await api.POST(API.changelogsFindLast, {
        body: { projectId, onlyChanged: true },
      })
      if (error) throw new Error(error.message ?? String(error))
      setItems(data?.items ?? [])
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (open) void fetchHistory()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  return (
    <>
      {/* 虚线按钮：outline 底 + dashed 边框，同 RepoFormModal/DynamicElement 的「添加」入口样式，
          与上方实心主按钮/ghost 次按钮区分，暗示「浏览记录」为次级入口 */}
      <Button variant="dashed" size="xs" onClick={() => setOpen(true)}>
        <Icon name="pulse" className="text-[13px]" />
        {t('project.configHistory')}
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        {/* 宽度加到 5xl（768→1024px）：展开版本后是左右分屏 diff，3xl 下每列仅约 350px，
            长行反复折行；5xl 与宿主项目详情弹窗（ProjectDetailModal 同用 5xl）等宽，观感连续 */}
        <DialogContent className="sm:max-w-5xl" style={{ zIndex: z }}>
          <DialogHeader>
            <DialogTitle>{t('project.configHistory')}</DialogTitle>
          </DialogHeader>
          {loading ? (
            <div className="py-10 text-center text-[13px] text-faint">{t('common.loading')}</div>
          ) : items.length === 0 ? (
            <Empty text={t('common.empty')} icon="clock" />
          ) : (
          <div className="flex max-h-[60vh] flex-col gap-1.5 overflow-auto overscroll-contain">
            {items.map((item, idx) => {
              // 上一版：列表按版本号降序（后端 OrderByVersionDesc），下一条即更早的版本。
              // 最后一条没有可比对象（onlyChanged 过滤 + limit 截断，它的更早版本可能压根没返回）：
              // prevItem 置空 → 分支/提交只显示当前值，配置按「全新增」处理（与展开区一致）。
              // 此时绝不能拿空串去比——那必然得出「已变更」，是缺数据而非真变更。
              const prevItem = idx + 1 < items.length ? items[idx + 1] : null
              const configChanged = prevItem !== null && isConfigChanged(prevItem.config, item.config)
              const isExpanded = expanded === item.version
              return (
                <div
                  key={item.version}
                  className="rounded-lg border border-line bg-surface"
                >
                  <button
                    type="button"
                    onClick={() => setExpanded(isExpanded ? null : item.version)}
                    className="flex w-full flex-col gap-1.5 rounded-lg px-3 py-2 text-left transition-colors hover:bg-raised"
                  >
                    {/* 行 1：版本 / 更新人 / 时间 / 配置是否变更 / 展开箭头 */}
                    <div className="flex w-full flex-wrap items-center gap-2">
                      <Tag tone="accent" dot={false}>
                        v{item.version}
                      </Tag>
                      <span className="text-[13px] font-medium text-ink">{item.username}</span>
                      <span className="text-[12px] text-faint">{item.date}</span>
                      {/* 配置文件变更与否都给 tag（未变更用 mute 弱化），不让「没变更」的版本空着一块。
                          最老一条没有更早版本可比 → 中性态「无对比版本」，不能谎报未变更/已变更 */}
                      {prevItem === null ? (
                        <Tag tone="mute" className="ml-auto">
                          {t('project.configNoBaseline')}
                        </Tag>
                      ) : (
                        <Tag tone={configChanged ? 'warn' : 'mute'} className="ml-auto">
                          {configChanged ? t('project.configChanged') : t('project.configUnchanged')}
                        </Tag>
                      )}
                      <Icon
                        name="chevron-down"
                        className={`text-[12px] text-faint transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                      />
                    </div>
                    {/* 行 2-3：分支、提交「从旧改成新」。标签列 auto 定宽对齐，值列 1fr 兜住长内容 */}
                    <div className="grid w-full grid-cols-[auto_minmax(0,1fr)] items-center gap-x-2 gap-y-1 text-[12px]">
                      <span className="text-faint">{t('project.branch')}</span>
                      <FromTo from={prevItem?.gitBranch} to={item.gitBranch} />
                      <span className="text-faint">{t('project.commit')}</span>
                      <div className="flex min-w-0 items-center gap-2">
                        <FromTo
                          from={prevItem?.gitCommit}
                          to={item.gitCommit}
                          render={(v) => <span className="font-mono">{shortSha(v)}</span>}
                        />
                        {item.gitCommitWebUrl && item.gitCommitTitle && (
                          <CommitTitleLink
                            href={item.gitCommitWebUrl}
                            title={item.gitCommitTitle}
                          />
                        )}
                      </div>
                    </div>
                  </button>
                  {isExpanded && (
                    <div className="border-t border-line px-3 py-2">
                      <DiffLines
                        oldText={prevItem?.config ?? ''}
                        newText={item.config}
                        lang={item.configType || configFileType || 'yaml'}
                        baseline={prevItem === null}
                      />
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}
        </DialogContent>
      </Dialog>
    </>
  )
}

/**
 * 配置文本相对上一版是否有变更（只比 config，不含 commit）。
 *
 * 不能用后端的 `item.configChanged`：它的判定是「配置或提交有一个变了」——
 * 后端 eventhandler/event_coordinator.go 里 `configChanged = lastChange.Config != proj.Config
 * || lastChange.GitCommit != proj.GitCommit`，onlyChanged 过滤用的也是这个布尔。照搬过来就会
 * 出现「只换了 commit、配置一字未动」的版本被打上「配置已变更」，展开后 diff 却说「无配置变更」
 * ——tag 与展开区互相打脸（重新部署同配置正是常见场景）。列表行的 tag 与展开区共用本判定，
 * 两者永远一致，不会各说各话。
 *
 * 直接比字符串即可：split('\n')/join('\n') 可逆，所以逐行 LCS「全为 same」⟺ 两串全等，
 * 而这里只问「有没有一行不同」——LCS 的 O(n·m) 是白花的。列表最多 15 行（后端 Limit 写死 15）、
 * 配置动辄上千行，且每次展开/收起重渲染都要逐行全量跑一遍：实测 1000 行 × 5 行，
 * 逐行 LCS 16.9ms/行（一次渲染 84ms），全等相比 0.0007ms/行。
 */
function isConfigChanged(oldText: string, newText: string): boolean {
  return oldText !== newText
}

/** 提交 sha 缩略成 7 位（GitHub 惯例）；后端偶尔给短 sha，够短就原样显示 */
const shortSha = (v: string): string => (v.length > 7 ? v.slice(0, 7) : v)

/**
 * 「从旧改成新」的值对比：有变化显示 `旧 → 新`，无变化（或没有上一版可比）只显示当前值。
 * 颜色沿用 diff 的增删约定（旧值 err 红=被替换掉、新值 ok 绿=现在的值），
 * 展开后的 diff 就是这套色，一眼能对上；中间的 `→` 保证不单靠颜色传递语义（色盲可辨）。
 * 旧值再叠一道删除线：红只是「被替换」的弱信号，删除线把「这个值没了」说白，
 * 一眼分清哪边是过去、哪边是现在。删除线取 currentColor，自动跟红字同色。
 */
function FromTo({
  from,
  to,
  render,
}: {
  /** 上一版的同名值；列表最老一条没有上一版 → undefined */
  from?: string
  to: string
  /** 值渲染器：提交 sha 需要缩略 + 等宽字体 */
  render?: (v: string) => ReactNode
}) {
  const fmt = render ?? ((v: string) => v)
  if (!to) return <span className="text-faint">-</span>
  if (!from || from === to) return <span className="text-mute">{fmt(to)}</span>
  return (
    <span className="flex min-w-0 items-center gap-1.5">
      <span className="min-w-0 truncate text-err line-through">{fmt(from)}</span>
      <span className="shrink-0 text-faint">→</span>
      <span className="min-w-0 truncate text-ok">{fmt(to)}</span>
    </span>
  )
}

/**
 * 提交标题链接：只被省略时才挂 tooltip 显示全文。
 *
 * 不省略就别弹——正文已经一字不差地摊在那，再浮一个同样的框是纯噪音（还得为它常驻一棵
 * Tooltip 组件树）。判定与写法全部沿用项目既有做法（见 components/EndpointUrl.tsx）：
 * 外层 span 恒带 `min-w-0 flex-1 truncate`，两个分支都在——它才是真正裁掉文字的元素，
 * 所以尺寸必须量在它身上（量在外层之外的地方恒不溢出、永远判定为「没省略」）。
 * 外层元素跨分支类型/位置不变，ref 与 ResizeObserver 的目标也就跟着稳定，不会因切换分支重挂。
 * `scrollWidth > clientWidth + 1` 的 +1 是缓冲亚像素抖动；ResizeObserver 让容器变宽后能改判。
 *
 * 受控 open：Radix Tooltip 默认 hover 即开，且它靠 trigger 上的 pointerenter 起效——
 * Tooltip 若在指针已经悬停后才挂载，那个 enter 事件早发过了，Radix 收不到，tooltip 就永远不开。
 * 故自己接管 onMouseEnter/onMouseLeave，用 open={truncated && hover} 直接驱动，
 * onOpenChange 给空函数挡掉 Radix 自己的开关逻辑。
 *
 * z-index：tooltip portal 到 body，默认 z-50 会被宿主弹窗（useOverlayZ 分配的 52+）盖住，
 * 故悬浮时现取一个 nextZIndex()，保证压在所有已分配浮层之上（同 EndpointUrl）。
 */
function CommitTitleLink({ href, title }: { href: string; title: string }) {
  const ref = useRef<HTMLSpanElement>(null)
  const [truncated, setTruncated] = useState(false)
  const [hover, setHover] = useState(false)
  const [tipZ, setTipZ] = useState(50)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const measure = () => setTruncated(el.scrollWidth > el.clientWidth + 1)
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(el)
    return () => ro.disconnect()
  }, [title])

  // 仅链接文本可跳转（新标签页）：onClick 阻止冒泡，行内其余区域点击走外层按钮的展开/收缩。
  // 截断交给外层 span：这里不再自带 truncate，避免量测量到的是「没被裁」的内层元素。
  const link = (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className="text-primary hover:underline"
      onClick={(e) => e.stopPropagation()}
      onMouseEnter={() => {
        setTipZ(nextZIndex())
        setHover(true)
      }}
      onMouseLeave={() => setHover(false)}
    >
      {title}
    </a>
  )

  // 未截断 + 未悬浮：裸渲染，省掉 TooltipProvider/Tooltip 组件树
  if (!truncated && !hover) {
    return <span ref={ref} className="min-w-0 flex-1 truncate">{link}</span>
  }

  return (
    <span ref={ref} className="min-w-0 flex-1 truncate">
      <TooltipProvider delayDuration={100}>
        <Tooltip open={truncated && hover} onOpenChange={() => {}}>
          <TooltipTrigger asChild>{link}</TooltipTrigger>
          <TooltipContent style={{ zIndex: tipZ }}>{title}</TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </span>
  )
}

/** 版本 diff：无变更显示占位，有变更走增强 DiffViewer（分屏/高亮/复制） */
function DiffLines({
  oldText,
  newText,
  lang,
  baseline,
}: {
  oldText: string
  newText: string
  lang: string
  /** 最老一条没有更早版本可比：整块按新增展示属于数据缺失的必然结果，先说明再看内容 */
  baseline?: boolean
}) {
  const { t } = useTranslation()
  if (baseline) {
    return (
      <div className="flex flex-col gap-1.5">
        <div className="text-[12px] text-faint">{t('project.noBaselineConfig')}</div>
        <DiffViewer oldValue="" newValue={newText} language={lang} />
      </div>
    )
  }
  if (!isConfigChanged(oldText, newText)) {
    return <div className="text-[12px] text-faint">{t('project.noConfigChanged')}</div>
  }
  return <DiffViewer oldValue={oldText} newValue={newText} language={lang} />
}
