import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import * as YAML from 'yaml'
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
type ExtraValue = components['schemas']['websocket.ExtraValue']

/**
 * 配置历史：拉取项目配置改动日志，逐条展示「版本 + 更新人 + 提交 + 配置变更」，
 * 展开后显示与上一版本的配置文件 diff + 自定义配置（extraValues）取值变更。
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
              // 最老一条没有更早版本可比（onlyChanged 过滤 + limit 截断，它的更早版本可能压根
              // 没返回）：中性态「无对比版本」，配置按全新增处理
              const baseline = prevItem === null
              const configChanged = prevItem !== null && isConfigChanged(prevItem.config, item.config)
              // 自定义配置（extraValues）是否变更：与展开区共用 changedExtraPaths 这一个判定，
              // 行 tag 与展开区永远一致，不会 tag 说有变更、展开区说没有
              const extraChanged =
                prevItem !== null &&
                changedExtraPaths(prevItem.extraValues, item.extraValues).length > 0
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
                          最老一条没有更早版本可比 → 中性态「无对比版本」，不能谎报未变更/已变更。
                          自定义配置变更再挂一个 tag（info 色，与配置的 warn 一眼区分）：配置文件一字未动、
                          只有部署参数改了也是「这版有变更」，这个信号不能只藏在展开区里。
                          未变更时不挂——这条 tag 是给「config 没变但参数变了」的版本发的，没变就保持安静 */}
                      {baseline ? (
                        <Tag tone="mute" className="ml-auto">
                          {t('project.configNoBaseline')}
                        </Tag>
                      ) : (
                        <>
                          <Tag tone={configChanged ? 'warn' : 'mute'} className="ml-auto">
                            {configChanged
                              ? t('project.configChanged')
                              : t('project.configUnchanged')}
                          </Tag>
                          {extraChanged && <Tag tone="info">{t('project.extraValuesChanged')}</Tag>}
                        </>
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
                    <div className="flex flex-col gap-3 border-t border-line px-3 py-2">
                      {/* 两块各自独立陈述变更与否、互不代言（配置文件一字未动、只有部署参数改了，
                          这里照样把改了什么说清楚），也各自独立折叠（见 DiffSection） */}
                      <DiffSection
                        title={t('project.configSection')}
                        defaultOpen={baseline || configChanged}
                      >
                        <DiffLines
                          oldText={prevItem?.config ?? ''}
                          newText={item.config}
                          lang={item.configType || configFileType || 'yaml'}
                          baseline={baseline}
                        />
                      </DiffSection>
                      <DiffSection
                        title={t('project.extraValues')}
                        defaultOpen={baseline || extraChanged}
                      >
                        <ExtraValuesDelta
                          prev={prevItem?.extraValues ?? []}
                          curr={item.extraValues}
                          baseline={baseline}
                        />
                      </DiffSection>
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
 * 不能用后端的 `item.configChanged`：它的判定比配置文本宽得多——后端 biz.ProjectConfigChanged
 * 比的是「配置 / 分支 / 提交 / 自定义配置项」四样有没有一样变了，onlyChanged 过滤用的也是这个布尔。
 * 照搬过来就会出现「只换了 commit」或「只改了自定义配置」的版本被打上「配置已变更」，展开后
 * diff 却说「无配置变更」——tag 与展开区互相打脸（重新部署同配置正是常见场景）。列表行的 tag
 * 与展开区共用本判定，两者永远一致，不会各说各话。
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

/**
 * 展开区里的可折叠小节：表头独占一行、点整行折起/展开内容。
 *
 * 为什么两块各自折叠、不共用一个开关：项目配置动辄上千行，它一个就能把下面的自定义配置挤出
 * 屏幕，可这两块又是独立的事实（配置没变、部署参数变了很常见）——得能单独折起上面那块去看
 * 下面那块，反之亦然。
 *
 * 默认开合跟「这块有没有改动」走（defaultOpen，调用方按各自的变更判定算）：有改动才摊开，
 * 没改动的只留一行表头——展开一个版本时，有内容的块直接是内容，没内容的块不再拿一行
 * 「（无…变更）」占着位置把有内容的挤下去。最老一条没有基准（baseline）是例外：它的块里是
 * 整份配置按新增展示，是实打实的内容，折起来等于把内容藏了，故跟着展开。
 * defaultOpen 只在挂载时取一次：用户手动折过之后，父组件重渲染不该把它掰回来；收起版本行再
 * 展开会整棵重挂，那时又回到「没改动就折起」的默认——每次打开都先看要点。
 *
 * 表头要显眼（13px + font-semibold + text-ink，带可点的 hover 底）：它是这一块唯一的标题，
 * 原先 12px/mute 那版和 diff 里的注释行一样轻，两块叠起来分不清哪是哪。箭头放标题前（折叠
 * 面板的惯例），折起时转 -90° 指右——不用展开内容就能看出这块是收着的。
 * 分隔线挂在节顶上、首节去掉（first:）：两块之间有线，和上面的行之间那条线由父容器给。
 */
function DiffSection({
  title,
  defaultOpen,
  children,
}: {
  title: string
  /** 挂载时的开合状态；调用方按「这块有没有改动」算（见上） */
  defaultOpen: boolean
  children: ReactNode
}) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <div className="border-t border-line pt-2 first:border-t-0 first:pt-0">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="-mx-1.5 flex items-center gap-1.5 rounded-md px-1.5 py-1 transition-colors hover:bg-raised"
      >
        <Icon
          name="chevron-down"
          className={`text-[12px] text-faint transition-transform ${open ? '' : '-rotate-90'}`}
        />
        <span className="text-[13px] font-semibold text-ink">{title}</span>
      </button>
      {open && <div className="mt-1.5">{children}</div>}
    </div>
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

/**
 * extraValues → path 键的值映射。缺失与空串同义（都是「没有值」），取值处一律
 * `map.get(path) ?? ''`，避免「这一侧没有该 path」被判成变更。
 */
function extraValuesMap(list: ExtraValue[] | undefined): Map<string, string> {
  return new Map((list ?? []).map((v) => [v.path, v.value]))
}

/**
 * 两侧有出入的 path（本版新增的、本版已没有的、以及取值变了的），按 path 升序。
 *
 * 行 tag 的「自定义配置已变更」与展开区的 diff 共用这一个判定——tag 说有变更、展开区说没有，
 * 是最难查的一类不一致（同 isConfigChanged 的用意：一处判定，几处显示）。
 */
function changedExtraPaths(
  prev: ExtraValue[] | undefined,
  curr: ExtraValue[] | undefined,
): string[] {
  const prevMap = extraValuesMap(prev)
  const currMap = extraValuesMap(curr)
  return [...new Set([...prevMap.keys(), ...currMap.keys()])]
    .filter((path) => (prevMap.get(path) ?? '') !== (currMap.get(path) ?? ''))
    .sort()
}

/**
 * 一条 extraValue 在界面上的标签：说明优先、为空回退 path。
 *
 * 与部署表单是同一条口径（Elements.tsx 的 `element.description || element.path`）——说明文案才是
 * 表单上那个字段名，path（resources.limits.cpu 这类）是元素定义里的原始字符串，本质是内部键。
 * 表单上看到的什么，历史里就该看到什么。说明由后端在部署落库时按当时的元素定义固化（仓库配置早被
 * 改过也读得回来），历史数据可能整片没有，故这条回退必须有，不能渲染成空键。
 */
function extraLabel(v: ExtraValue): string {
  return v.description || v.path
}

/**
 * 把一版的 extraValues 合并成一份 yaml 文本（说明作键、取值作值，按 path 升序）。
 *
 * 键用说明而不是 path：path 在界面上是内部键，直接当键等于让人对着 `resources.limits.cpu` 猜
 * 这是哪个字段；说明才是部署表单上那个字段名（见 extraLabel）。
 *
 * 说明会撞车——两个不同 path 写了同一句说明（「环境变量」这种），Map 会把后一个顶掉前一个，这份
 * 文档就平白少一个键：取值真变了却在 diff 里看不见，是最难查的一类漏报。故同名的那一组一律补 path
 * 后缀消歧，且只加在撞车的组上，不撞车时仍是干净的一句说明。两侧各算各的：某一项在本版才撞名时，
 * 它会在本版一侧带上后缀，diff 里表现为这个键「改名」了——这正是「说明开始有歧义」该有的提示。
 *
 * 排序按 path、不是按显示出来的说明：path 唯一且稳定，说明文案改一个字，整份文档的键序不该跟着
 * 全部挪位。排序交给 Map（不用普通对象）：Map 保插入序，而 path 万一取成纯数字（「10」/「2」），
 * 普通对象的键会被按数值大小重排，两份文档的键序就不再一致，diff 会整片错位成「每个键都动了」。
 *
 * 值交给 yaml 库序列化，而不是手拼 `${label}: ${value}`：值可能是多行片段（TextArea 型字段），
 * 也可能是「1」/「true」这种在 yaml 里类型有歧义的标量，库该加引号加引号、该出块标量出块标量；
 * 手拼等于把值原样塞进文档，读回来已经不是同一个值。
 */
function extraValuesYaml(list: ExtraValue[] | undefined): string {
  const entries = [...(list ?? [])]
    .sort((a, b) => (a.path < b.path ? -1 : a.path > b.path ? 1 : 0))
    .map((v) => ({ label: extraLabel(v), path: v.path, value: v.value }))
  // 同名标签计数：>1 的那组整组补 path 后缀消歧（见上），否则 Map 会把同名键悄悄并成一个
  const dup = new Map<string, number>()
  for (const e of entries) dup.set(e.label, (dup.get(e.label) ?? 0) + 1)
  const rows = entries.map(
    (e): [string, string] => [
      (dup.get(e.label) ?? 0) > 1 ? `${e.label} (${e.path})` : e.label,
      e.value,
    ],
  )
  // 空表会 stringify 成 "{}\n"，diff 里平白多一行 "{}"；空就是空串
  return rows.length === 0 ? '' : YAML.stringify(new Map(rows))
}

/**
 * 自定义配置变更块：把两侧的 extraValues 各自合并成一份 yaml，走一个 diff。
 *
 * 为什么看 extraValues 而不是 finalExtraValues：两者都是「这次部署的额外配置项」，差别在
 * finalExtraValues 是后端 ElementsLoader 拿元素定义把缺的 path 用默认值补齐后的最终生效值
 * （typedValue 转换 + 兜底），每个版本都会带上一整片压根没动过的默认项；extraValues 是这次
 * 实际提交上来的那一份，更贴近「用户改了什么」，噪音也小。
 * 本块与配置文件的 diff 是两个独立事实，不共用判定（见上方 isConfigChanged 注释）。
 *
 * 为什么合并成一份 yaml 走单个 diff，而不是每个 path 各挂一个：按 path 各挂一个时，一屏里会
 * 竖排 N 个各带工具栏的 diff 卡片，改动项一多整块就散成一片，看不出「这几个键是在同一处一起
 * 变的」；合并成一份文档后，变更行按 path 顺序聚在一起，两侧相邻的未变键也在同一份文档里
 * （DiffViewer 默认「仅显示变更」，未变的行折叠成一块，不会刷屏）。
 *
 * 键是说明、判定仍按 path，两者刻意分开：说明只是给人看的标签（见 extraValuesYaml），而「变没变」
 * 只能按 path 判——说明由后端在部署时按当时的元素定义固化，仓库里改一次文案，同一项在前后两版就
 * 是不同的说明，拿说明当判定基准会把「改了个说明」误报成「部署参数变了」。
 *
 * 某侧没有该 path（早期版本没提交过这个字段、或元素定义是后来才加的）按空串参与比较，于是新增
 * 渲染成库原生的「整块纯增」、删除渲染成「整块纯删」——比画 `-` 占位更准确，也让「这一项是这版
 * 才有的 / 这版没了」一眼可辨。缺失与空串同义（都是「没有值」），故 `'' ↔ 缺失` 不算变更，
 * 不会留下两侧皆空的空 diff。
 *
 * language 固定 yaml：这份文档本身就是 extraValues 合并出来的 yaml，不存在「元素定义的真实
 * 语言无从得知」的问题（changelog 载荷只带主配置文件的 configType）；值里的多行片段也按 yaml
 * 处理，与全篇观感一致。
 */
function ExtraValuesDelta({
  prev,
  curr,
  baseline,
}: {
  /** 上一版的 extraValues；列表最老一条没有更早版本可比 → 空数组 */
  prev: ExtraValue[] | undefined
  curr: ExtraValue[] | undefined
  /** 最老一条：没有更早版本可比，整块按新增展示，先说明再看内容（同配置文件 diff） */
  baseline?: boolean
}) {
  const { t } = useTranslation()
  const changed = changedExtraPaths(prev, curr)

  if (baseline) {
    return (
      <div className="flex flex-col gap-1.5">
        <div className="text-[12px] text-faint">{t('project.noBaselineConfig')}</div>
        <DiffViewer oldValue="" newValue={extraValuesYaml(curr)} language="yaml" />
      </div>
    )
  }
  // 显式说「无变更」，不静默留白：留白与「这块没渲染出来」在观感上分不开
  if (changed.length === 0) {
    return <div className="text-[12px] text-faint">{t('project.noExtraValuesChanged')}</div>
  }
  return (
    <div className="flex flex-col gap-1.5">
      <DiffViewer oldValue={extraValuesYaml(prev)} newValue={extraValuesYaml(curr)} language="yaml" />
    </div>
  )
}
