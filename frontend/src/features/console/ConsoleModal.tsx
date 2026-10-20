import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { selectAllOnDoubleClick } from '@/lib/selection'
import { useDraggableDialog } from '@/hooks/useDraggableDialog'
import { useWheelRedirect } from '@/hooks/useWheelRedirect'
import { Icon } from '@/components/Icons'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/shadcn/tabs'

/**
 * 控制台弹窗（**纯 UI 原型**，尚未对接任何接口）。
 *
 * 壳（非模态 / 无遮罩 / 只能 X 关 / 拖拽 / 双击最大化 / 四边缩放 / 滚轮重定向）逐字对齐
 * ProjectDetailModal，只把 Tab 收成「命令行 + 日志」两项——视觉与项目详情弹窗构造级一致，
 * 避免两处各写一套迟早漂移。等真接后端时：命令行换 TabShell（/ws pty）、日志换 TabLog。
 *
 * ⚠️ 两处待定（本原型用假数据顶替，见文件末尾的说明）：
 *   1. 标题里写死的 `svc` / `team-a` 是占位项目——从顶栏打开时到底作用于哪个项目还没拍板；
 *   2. 命令行/日志里的内容是 mock，任何交互都不会真正发给后端。
 */

/** Tab 键 → i18n 文案键（顺序即展示序） */
const TABS = [
  { key: 'shell', label: 'console.tabCommand' },
  { key: 'logs', label: 'console.tabLogs' },
] as const

type ConsoleTab = (typeof TABS)[number]['key']

// 原型占位：真实实现应由调用方传入 project / namespaceName（对齐 ProjectDetailModal 入参）
const MOCK_PROJECT = 'svc'
const MOCK_NAMESPACE = 'team-a'

export function ConsoleModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useTranslation()
  const dnd = useDraggableDialog()
  // 整块弹窗区域滚轮重定向：指针停在标题/Tab 栏等不可滚动区也滚内容区，不穿透滚主页面
  const { dialogRef, contentRef } = useWheelRedirect()
  const [tab, setTab] = useState<ConsoleTab>('shell')

  return (
    <Dialog open={open} modal={false} onOpenChange={(o) => !o && onClose()}>
      <DialogContent
        ref={dialogRef}
        {...dnd.contentProps}
        // 与 ProjectDetailModal 同款：无遮罩、三条 dismiss 路径全拦，只能点 X 关闭
        showOverlay={false}
        onPointerDownOutside={(e) => e.preventDefault()}
        onFocusOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={(e) => e.preventDefault()}
        style={dnd.contentStyle}
        className="sm:max-w-5xl h-[70vh] flex flex-col"
      >
        <DialogHeader>
          <DialogTitle
            {...dnd.dragHandleProps}
            className="relative flex cursor-move items-center justify-center px-8"
            title={t('project.dragTitle')}
          >
            {/* 项目名居中 + 命名空间贴右上（不参与居中）：与项目详情弹窗同一套排版 */}
            <span className="relative min-w-0 cursor-text" data-no-drag>
              <span
                className="block truncate text-[18px] font-semibold text-ink"
                onDoubleClick={selectAllOnDoubleClick}
              >
                {MOCK_PROJECT}
              </span>
              <span
                className="absolute left-full top-0 ml-0.5 -mt-1 whitespace-nowrap text-[10px] leading-none text-primary"
                style={{ fontFamily: '"dank mono", ui-monospace, monospace' }}
                onDoubleClick={selectAllOnDoubleClick}
              >
                {MOCK_NAMESPACE}
              </span>
            </span>
          </DialogTitle>
        </DialogHeader>

        <Tabs value={tab} onValueChange={(v) => setTab(v as ConsoleTab)}>
          <TabsList variant="line" className="w-full border-b border-line">
            {TABS.map((it) => (
              <TabsTrigger
                key={it.key}
                value={it.key}
                className="focus-visible:!border-transparent focus-visible:!ring-0 focus-visible:!outline-none"
              >
                {t(it.label)}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>

        {/* overflow-hidden：两个面板都是 h-full 自管滚动，外层不参与滚动 */}
        <div ref={contentRef} className="mt-1 min-h-0 flex-1 overflow-hidden">
          {tab === 'shell' ? <CommandPane /> : <LogPane />}
        </div>

        {/* 四边交互分层：逐字复制 ProjectDetailModal 的边条 / 缩放把手（手势行为必须一致） */}
        <div aria-hidden title={t('project.dragTitle')} {...dnd.dragHandleProps} className="pointer-events-auto absolute inset-x-0 top-0 h-6 cursor-move" />
        <div aria-hidden title={t('project.dragTitle')} {...dnd.dragHandleProps} className="pointer-events-auto absolute inset-x-0 bottom-0 h-6 cursor-move" />
        <div aria-hidden title={t('project.dragTitle')} {...dnd.dragHandleProps} className="pointer-events-auto absolute inset-y-0 left-0 w-6 cursor-move" />
        <div aria-hidden title={t('project.dragTitle')} {...dnd.dragHandleProps} className="pointer-events-auto absolute inset-y-0 right-0 w-6 cursor-move" />
        {!dnd.isMaximized && (
          <>
            <div aria-hidden title={t('project.resizeEdge')} {...dnd.getResizeHandleProps('n')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute inset-x-0 top-0 z-10 h-1.5 cursor-ns-resize" />
            <div aria-hidden title={t('project.resizeEdge')} {...dnd.getResizeHandleProps('s')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute inset-x-0 bottom-0 z-10 h-1.5 cursor-ns-resize" />
            <div aria-hidden title={t('project.resizeEdge')} {...dnd.getResizeHandleProps('w')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute inset-y-0 left-0 z-10 w-1.5 cursor-ew-resize" />
            <div aria-hidden title={t('project.resizeEdge')} {...dnd.getResizeHandleProps('e')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute inset-y-0 right-0 z-10 w-1.5 cursor-ew-resize" />
            <div aria-hidden title={t('project.resizeCorner')} {...dnd.getResizeHandleProps('nw')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute top-0 left-0 z-10 size-4 cursor-nwse-resize" />
            <div aria-hidden title={t('project.resizeCorner')} {...dnd.getResizeHandleProps('ne')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute top-0 right-0 z-10 size-4 cursor-nesw-resize" />
            <div aria-hidden title={t('project.resizeCorner')} {...dnd.getResizeHandleProps('sw')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute bottom-0 left-0 z-10 size-4 cursor-nesw-resize" />
            <div aria-hidden title={t('project.resizeCorner')} {...dnd.getResizeHandleProps('se')} onDoubleClick={dnd.dragHandleProps.onDoubleClick} className="pointer-events-auto absolute bottom-0 right-0 z-10 size-4 cursor-nwse-resize" />
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

// ---------------------------------------------------------------- 命令行

/** 终端里的一行：in = 我们敲进去的命令，out = 终端回吐的内容 */
type TermLine = { kind: 'in' | 'out'; text: string }

/** 原型用的假回显：只为让输入有反馈。真接后端后由 PTY 输出替换 */
const CANNED_OUTPUT: Record<string, string[]> = {
  ls: ['app  config  logs  package.json  README.md'],
  pwd: ['/app'],
  whoami: ['root'],
  'uname -a': ['Linux svc-5f8c9d-7b4c 5.15.0-119-generic #129-Ubuntu SMP x86_64 GNU/Linux'],
}

const INITIAL_TERMINAL: TermLine[] = [
  { kind: 'in', text: 'kubectl get pods -n team-a' },
  { kind: 'out', text: 'NAME                    READY   STATUS    RESTARTS   AGE' },
  { kind: 'out', text: 'svc-5f8c9d-7b4c-t8q2m   1/1     Running   0          12m' },
  { kind: 'in', text: 'tail -n 3 /app/logs/app.log' },
  { kind: 'out', text: '2026-10-16 12:01:03 INFO  server listening on :8080' },
  { kind: 'out', text: '2026-10-16 12:01:07 WARN  cache miss ratio 0.42' },
  { kind: 'out', text: '2026-10-16 12:01:11 INFO  healthz ok' },
]

/**
 * 命令行面板：终端观感（深底等宽）+ 底部输入行，回车把命令追加进滚动区。
 * 纯前端状态，不发任何请求——「回车有反馈」只是为了让原型可点。
 */
function CommandPane() {
  const { t } = useTranslation()
  const [lines, setLines] = useState<TermLine[]>(INITIAL_TERMINAL)
  const [cmd, setCmd] = useState('')
  const bodyRef = useRef<HTMLDivElement>(null)

  // 新行追加后贴底（真实终端的行为）；用户往上翻看历史时该行为会被后续输出打断，
  // 真接后端时需换成「贴底判定」（TabLog 的 autoScroll 同款），原型从简。
  useEffect(() => {
    const el = bodyRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [lines])

  const run = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key !== 'Enter') return
    const text = cmd.trim()
    if (!text) return
    const canned = CANNED_OUTPUT[text]
    const reply: TermLine[] = canned
      ? canned.map((line) => ({ kind: 'out', text: line }))
      : [{ kind: 'out', text: `${t('console.mockNotice')} — sh: ${text}: 命令未执行` }]
    setLines((prev) => [...prev, { kind: 'in', text }, ...reply])
    setCmd('')
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-9 shrink-0 items-center gap-2 border-b border-line px-3">
        <span className="flex items-center gap-1.5 rounded-md bg-raised px-2 py-1 font-mono text-[11px] text-ink">
          <span className="size-1.5 rounded-full bg-ok" />
          svc-5f8c9d-7b4c-t8q2m / app
        </span>
        <button
          type="button"
          onClick={() => setLines([])}
          className="ml-auto flex h-6 items-center gap-1 rounded-md px-2 text-xs text-mute transition-colors hover:bg-raised hover:text-ink"
        >
          <Icon name="minus" className="text-[13px]" />
          {t('console.clear')}
        </button>
      </div>

      <div ref={bodyRef} className="min-h-0 flex-1 overflow-auto bg-[#0b1220] px-3 py-2">
        {lines.length === 0 ? (
          <div className="font-mono text-[12px] text-slate-500">{t('console.commandHint')}</div>
        ) : (
          lines.map((line, i) =>
            line.kind === 'in' ? (
              <div key={i} className="flex whitespace-pre font-mono text-[12.5px] leading-[1.65]">
                <span className="mr-2 shrink-0 select-none text-emerald-400">$</span>
                <span className="text-slate-100">{line.text}</span>
              </div>
            ) : (
              <div key={i} className="whitespace-pre pl-[18px] font-mono text-[12.5px] leading-[1.65] text-slate-400">
                {line.text}
              </div>
            ),
          )
        )}
      </div>

      <div className="flex h-10 shrink-0 items-center gap-2 border-t border-line bg-[#0b1220] px-3">
        <span className="shrink-0 select-none font-mono text-[12.5px] text-emerald-400">$</span>
        <input
          value={cmd}
          onChange={(e) => setCmd(e.target.value)}
          onKeyDown={run}
          placeholder={t('console.commandPlaceholder')}
          autoComplete="off"
          spellCheck={false}
          className="min-w-0 flex-1 bg-transparent font-mono text-[12.5px] text-slate-100 outline-none placeholder:text-slate-600"
        />
      </div>
    </div>
  )
}

// ---------------------------------------------------------------- 日志

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'

/** 级别筛选顺序（ALL 在最前） */
const LEVEL_FILTERS: readonly (LogLevel | 'ALL')[] = ['ALL', 'DEBUG', 'INFO', 'WARN', 'ERROR']

/** 级别 → 文字色（复用平台语义 token，跟随主题） */
const LEVEL_CLASS: Record<LogLevel, string> = {
  DEBUG: 'text-faint',
  INFO: 'text-info',
  WARN: 'text-warn',
  ERROR: 'text-err',
}

const MOCK_LOGS: { ts: string; level: LogLevel; msg: string }[] = [
  { ts: '12:00:41.002', level: 'INFO', msg: 'starting svc v2.14.0 (commit deadbeef)' },
  { ts: '12:00:41.118', level: 'DEBUG', msg: 'loading config from /app/config/app.yaml' },
  { ts: '12:00:41.204', level: 'DEBUG', msg: 'dial postgres 10.42.3.17:5432' },
  { ts: '12:00:41.376', level: 'INFO', msg: 'database pool ready (max=16 idle=4)' },
  { ts: '12:00:41.501', level: 'DEBUG', msg: 'registering routes: /healthz /metrics /api' },
  { ts: '12:00:41.590', level: 'INFO', msg: 'server listening on :8080' },
  { ts: '12:00:52.011', level: 'INFO', msg: 'GET /healthz 200 1.2ms' },
  { ts: '12:01:03.442', level: 'WARN', msg: 'cache miss ratio 0.42 above threshold 0.30' },
  { ts: '12:01:07.880', level: 'DEBUG', msg: 'reloading redis topology (1 node changed)' },
  { ts: '12:01:09.115', level: 'INFO', msg: 'cache miss ratio recovered 0.11' },
  { ts: '12:01:14.702', level: 'INFO', msg: 'GET /api/v1/items 200 42.7ms' },
  { ts: '12:01:22.339', level: 'ERROR', msg: 'upstream timeout calling billing-svc (10s exceeded)' },
  { ts: '12:01:22.341', level: 'WARN', msg: 'retrying billing-svc (attempt 1/3)' },
  { ts: '12:01:25.918', level: 'INFO', msg: 'billing-svc recovered after 3.5s' },
  { ts: '12:01:31.044', level: 'DEBUG', msg: 'gc pause 1.8ms (heap 214MB / 512MB)' },
  { ts: '12:01:40.276', level: 'INFO', msg: 'POST /api/v1/orders 201 88.3ms' },
  { ts: '12:01:48.601', level: 'INFO', msg: 'healthz ok' },
]

/**
 * 日志面板：工具条（级别筛选 + 关键词搜索 + 自动滚动）+ 行号/时间/级别/正文的日志流。
 * 筛选为纯前端本地过滤（真接后端后改由 TabLog 的 /api 流式接口供给）。
 */
function LogPane() {
  const { t } = useTranslation()
  const [q, setQ] = useState('')
  const [level, setLevel] = useState<LogLevel | 'ALL'>('ALL')
  const [autoScroll, setAutoScroll] = useState(true)
  const bodyRef = useRef<HTMLDivElement>(null)

  const filtered = useMemo(() => {
    const kw = q.trim().toLowerCase()
    return MOCK_LOGS.filter(
      (l) => (level === 'ALL' || l.level === level) && (!kw || l.msg.toLowerCase().includes(kw)),
    )
  }, [q, level])

  useEffect(() => {
    if (!autoScroll) return
    const el = bodyRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [filtered, autoScroll])

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-9 shrink-0 items-center gap-2 border-b border-line px-3">
        <div className="flex items-center gap-1">
          {LEVEL_FILTERS.map((lv) => (
            <button
              key={lv}
              type="button"
              onClick={() => setLevel(lv)}
              className={
                'h-6 rounded-md px-2 font-mono text-[11px] transition-colors ' +
                (level === lv ? 'bg-primary/20 text-primary' : 'text-mute hover:bg-raised hover:text-ink')
              }
            >
              {lv === 'ALL' ? t('console.logLevelAll') : lv}
            </button>
          ))}
        </div>

        <div className="ml-auto flex items-center gap-2">
          <div className="flex h-6 items-center gap-1.5 rounded-md border border-line px-2">
            <Icon name="search" className="text-[12px] text-faint" />
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder={t('console.logSearch')}
              autoComplete="off"
              spellCheck={false}
              className="w-40 bg-transparent text-[11px] text-ink outline-none placeholder:text-faint"
            />
          </div>
          <button
            type="button"
            onClick={() => setAutoScroll((v) => !v)}
            className={
              'h-6 rounded-md px-2 text-[11px] transition-colors ' +
              (autoScroll ? 'bg-primary/20 text-primary' : 'text-mute hover:bg-raised hover:text-ink')
            }
          >
            {t('console.logAutoScroll')}
          </button>
        </div>
      </div>

      <div ref={bodyRef} className="min-h-0 flex-1 overflow-auto bg-[#0b1220] py-1">
        {filtered.length === 0 ? (
          <div className="px-3 py-6 text-center font-mono text-[12px] text-slate-500">
            {t('console.logEmpty')}
          </div>
        ) : (
          filtered.map((l, i) => (
            <div
              key={i}
              className="flex whitespace-pre px-3 font-mono text-[12.5px] leading-[1.65] hover:bg-white/5"
              style={{ contentVisibility: 'auto' }}
            >
              <span className="w-10 shrink-0 select-none pr-3 text-right tabular-nums text-slate-600">
                {i + 1}
              </span>
              <span className="shrink-0 pr-3 text-slate-500">{l.ts}</span>
              <span className={'w-14 shrink-0 ' + LEVEL_CLASS[l.level]}>{l.level}</span>
              <span className="text-slate-300">{l.msg}</span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
