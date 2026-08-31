import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { nextZIndex } from '@/lib/zIndex'
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
import { diffLines } from '@/lib/diffLines'
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
  // 历史弹窗（portal 挂 body）须盖过可拖拽宿主弹窗 z-51+：打开时取下一个共享 z-index
  const [z, setZ] = useState(() => nextZIndex())
  // 触发源是普通 Button（非 Radix DialogTrigger），onOpenChange(true) 不会触发 → 由 effect 在打开时置顶
  useEffect(() => {
    if (open) setZ(nextZIndex())
  }, [open])
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
        <DialogContent className="sm:max-w-3xl" style={{ zIndex: z }}>
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
              const prev = idx + 1 < items.length ? items[idx + 1].config : ''
              const isExpanded = expanded === item.version
              return (
                <div
                  key={item.version}
                  className="rounded-lg border border-line bg-surface"
                >
                  <button
                    type="button"
                    onClick={() => setExpanded(isExpanded ? null : item.version)}
                    className="flex w-full flex-wrap items-center gap-2 rounded-lg px-3 py-2 text-left transition-colors hover:bg-raised"
                  >
                    <Tag tone="accent" dot={false}>
                      v{item.version}
                    </Tag>
                    <span className="text-[13px] font-medium text-ink">{item.username}</span>
                    <span className="text-[12px] text-faint">{item.date}</span>
                    {item.configChanged && (
                      <Tag tone="warn">{t('project.configChanged')}</Tag>
                    )}
                    {item.gitCommitWebUrl && (
                      // 仅链接文本可跳转（新标签页）：不加 flex-1 撑满整行，
                      // 链接只包住文字宽度，行内其余区域点击走外层按钮的展开/收缩。
                      // min-w-0 是关键：flex 子项默认 min-width:auto 按 nowrap 内容撑开，
                      // 吞掉 truncate 的省略效果导致长 commit title 换行——显式允许收缩后
                      // truncate（nowrap + overflow-hidden + text-ellipsis）才真正单行省略。
                      // tooltip 兜底：省略部分悬停显示全文（项目工具提示统一走 shadcn Tooltip）。
                      <TooltipProvider delayDuration={100}>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <a
                              href={item.gitCommitWebUrl}
                              target="_blank"
                              rel="noreferrer"
                              className="ml-auto min-w-0 max-w-[50%] truncate text-[12px] text-primary hover:underline"
                              onClick={(e) => e.stopPropagation()}
                            >
                              {item.gitCommitTitle}
                            </a>
                          </TooltipTrigger>
                          {/* tooltip portal 到 body，默认 z-50 会被可拖拽宿主弹窗（nextZIndex 从 51 起
                              动态置顶，本弹窗 z 已在 52+）盖住——显式抬到宿主弹窗之上才可见 */}
                          <TooltipContent style={{ zIndex: z + 1 }}>{item.gitCommitTitle}</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    )}
                    <Icon
                      name="chevron-down"
                      className={`text-[12px] text-faint transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                    />
                  </button>
                  {isExpanded && (
                    <div className="border-t border-line px-3 py-2">
                      <DiffLines
                        oldText={prev}
                        newText={item.config}
                        lang={item.configType || configFileType || 'yaml'}
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

/** 版本 diff：无变更显示占位，有变更走增强 DiffViewer（分屏/高亮/复制） */
function DiffLines({ oldText, newText, lang }: { oldText: string; newText: string; lang: string }) {
  const { t } = useTranslation()
  const lines = diffLines(oldText, newText)
  if (lines.every((l) => l.type === 'same')) {
    return <div className="text-[12px] text-faint">{t('project.noConfigChanged')}</div>
  }
  return <DiffViewer oldValue={oldText} newValue={newText} language={lang} />
}
