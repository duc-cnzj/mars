import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import * as YAML from 'yaml'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { CodeEditor, getMode } from '@/components/CodeEditor'
import { Icon } from '@/components/Icons'
import { Skeleton } from '@/components/ui/shadcn/skeleton'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { SearchableSelect } from '@/components/SearchableSelect'
import { Switch } from '@/components/ui/shadcn/switch'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/shadcn/popover'
import {
  Sheet,
  SheetContent,
  SheetTitle,
} from '@/components/ui/shadcn/sheet'
import { copyText } from '@/lib/copy'
import { SelectFileType } from './SelectFileType'
import { DEFAULT_REQUIRED_TYPES, DynamicElement, SELECTIVE_TYPES } from './DynamicElement'

type RepoModel = components['schemas']['types.RepoModel']
type GitItem = components['schemas']['git.AllReposResponse_Item']
type BranchOption = components['schemas']['git.Option']
type PipelinePassRule = components['schemas']['mars.PipelinePassRule']
type Element = components['schemas']['mars.Element']

interface FormState {
  name: string
  description: string
  needGitRepo: boolean
  gitProjectId: number
  branches: string[]
  pipelinePassRules: PipelinePassRule[]
  localChartPath: string
  configField: string
  configFileType: string
  configFileValues: string
  valuesYaml: string
  elements: Element[]
}

const DEFAULTS: FormState = {
  name: '',
  description: '',
  needGitRepo: true,
  gitProjectId: 0,
  branches: ['*'],
  pipelinePassRules: [],
  localChartPath: '',
  configField: '',
  configFileType: 'yaml',
  configFileValues: '',
  valuesYaml: '',
  elements: [],
}

/** 复用表单行样式 */
const labelCls = 'text-[12px] font-medium text-mute'

/** 按 'a->b->c' 路径从对象取值（还原旧版 lodash.get 语义，缺失返回 ''） */
function deepGet(obj: unknown, parts: string[]): unknown {
  let cur: unknown = obj
  for (const p of parts) {
    if (cur && typeof cur === 'object' && p in (cur as Record<string, unknown>)) {
      cur = (cur as Record<string, unknown>)[p]
    } else {
      return ''
    }
  }
  return cur
}

/**
 * 添加 / 编辑 repo 弹窗：核心字段 + mars.Config 配置。
 * 编辑时从 editItem 回填；git 项目与分支动态拉取。
 */
export function RepoFormModal({
  open,
  editItem,
  onClose,
  onSaved,
}: {
  open: boolean
  editItem?: RepoModel
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()

  const [form, setForm] = useState<FormState>(DEFAULTS)
  const [projects, setProjects] = useState<GitItem[]>([])
  const [branches, setBranches] = useState<BranchOption[]>([])
  const [saving, setSaving] = useState(false)
  // 左侧「charts 默认值」：get_chart_values_yaml 拉取的 chart 模板（只读引用，独立于表单值）。
  // 旧版 AddRepoModal 用独立 state 承载它，从不写进表单字段——右侧 values.yaml 是库里保存的配置。
  const [chartDefaults, setChartDefaults] = useState('')
  // values.yaml 自动拉取 loading
  const [valuesLoading, setValuesLoading] = useState(false)
  // 从 valuesYaml + configField 自动推导出的可用全局配置
  const [configFileContent, setConfigFileContent] = useState('')
  const [configDetected, setConfigDetected] = useState(false)
  const [configPopoverOpen, setConfigPopoverOpen] = useState(false)
  // pipeline 通过规则下拉选项（项目级最近 pipeline 的 stage/job）
  const [pipelineOptions, setPipelineOptions] = useState<{
    stages: string[]
    jobs: string[]
  }>({ stages: [], jobs: [] })
  /** 已填充/已选中的 git 项目 id：识别「用户切换了 git 项目」而非初始回填，切换时清空已选分支与通过规则 */
  const gitSelectionRef = useRef(0)

  const set = <K extends keyof FormState>(k: K, v: FormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  /** 用 RepoModel 回填表单（marsConfig 可能为 null/缺失，全部可选链兜底） */
  const fillForm = (item: RepoModel) => {
    const c = item.marsConfig
    setForm({
      name: item.name,
      description: item.description,
      needGitRepo: item.needGitRepo,
      gitProjectId: item.gitProjectId ?? 0,
      // 未配置具体分支（空数组 = 全部）时回填为 '*'，默认「全部」
      branches: c?.branches?.length ? c.branches : ['*'],
      pipelinePassRules: c?.pipelinePassRules ?? [],
      localChartPath: c?.localChartPath ?? '',
      configField: c?.configField ?? '',
      configFileType: c?.configFileType ?? 'yaml',
      configFileValues: c?.configFileValues ?? '',
      valuesYaml: c?.valuesYaml ?? '',
      elements: c?.elements ?? [],
    })
  }

  /** 详情补拉只回填配置区（marsConfig 派生字段），不动 name/description/needGitRepo/gitProjectId：
      基础字段列表项已完整，且详情返回前用户可能已在编辑——整体覆盖会丢掉输入 */
  const fillConfig = (item: RepoModel) => {
    const c = item.marsConfig
    // 详情请求返回前用户已切换 git 项目：branches/pipelinePassRules 属于旧项目，
    // 已被分支 effect 清空并重新拉取——这里不能再拿旧项目的配置覆盖回来
    //（分支 effect deps 不含 branches 不会重跑，stale 分支会被保存到错误项目上）
    const projectSwitched = gitSelectionRef.current !== (item.gitProjectId ?? 0)
    setForm((f) => ({
      ...f,
      branches: projectSwitched ? f.branches : c?.branches?.length ? c.branches : ['*'],
      pipelinePassRules: projectSwitched ? f.pipelinePassRules : c?.pipelinePassRules ?? [],
      localChartPath: c?.localChartPath ?? '',
      configField: c?.configField ?? '',
      configFileType: c?.configFileType ?? 'yaml',
      configFileValues: c?.configFileValues ?? '',
      valuesYaml: c?.valuesYaml ?? '',
      elements: c?.elements ?? [],
    }))
  }

  // 打开时回填：列表项即时回填（先看到名称/描述，配置缺失也不白屏），
  // 再后台补拉详情拿权威 marsConfig 覆盖（UAT 列表接口实测不携带完整配置，
  // 单靠列表项编辑时配置区全空）。回填前重置为 DEFAULTS，避免残留上次数据。
  useEffect(() => {
    if (!open) return
    setChartDefaults('') // 左侧 chart 模板重置，等待重新拉取
    if (!editItem) {
      setForm(DEFAULTS)
      gitSelectionRef.current = 0
      return
    }
    fillForm(editItem)
    gitSelectionRef.current = editItem.gitProjectId ?? 0
  }, [open, editItem]) // eslint-disable-line react-hooks/exhaustive-deps

  // 后台补拉详情：列表项可能不含 marsConfig，详情响应才是权威配置。
  // 只用 fillConfig 回填配置区，避免覆盖用户已在编辑的基础字段。
  useEffect(() => {
    if (!open || !editItem) return
    let alive = true
    void api
      .GET('/api/repos/{id}', { params: { path: { id: editItem.id } } })
      .then(({ data, error }) => {
        if (!alive || error || !data) return
        fillConfig(data.item)
      })
    return () => {
      alive = false
    }
  }, [open, editItem]) // eslint-disable-line react-hooks/exhaustive-deps

  // git 项目下拉数据
  useEffect(() => {
    if (!open) return
    void api.GET('/api/git/all_repos').then(({ data }) => {
      if (data) setProjects(data.items)
    })
  }, [open])

  // 选中 git 项目后拉取其分支；编辑时切换 git 项目会清空已选分支与通过规则
  useEffect(() => {
    if (!open || !form.gitProjectId) {
      setBranches([])
      return
    }
    if (form.gitProjectId !== gitSelectionRef.current) {
      // git 项目切换（含创建模式；旧实现只对比 editItem.gitProjectId，创建态会残留上个项目的分支）：
      // 清空已选分支与通过规则，它们属于上一个项目
      gitSelectionRef.current = form.gitProjectId
      setForm((f) => ({ ...f, branches: [], pipelinePassRules: [] }))
    }
    void api
      .GET('/api/git/projects/{gitProjectId}/branch_options', {
        params: { path: { gitProjectId: form.gitProjectId } },
      })
      .then(({ data }) => {
        if (data) setBranches(data.items)
      })
  }, [open, form.gitProjectId])

  // 选中 git 项目后拉取该项目的 stage/job 下拉选项（pipeline 通过规则用）
  useEffect(() => {
    if (!open || !form.gitProjectId) {
      setPipelineOptions({ stages: [], jobs: [] })
      return
    }
    void api
      .GET('/api/git/projects/{gitProjectId}/pipeline_job_options', {
        params: { path: { gitProjectId: form.gitProjectId } },
      })
      .then(({ data }) => {
        if (data)
          setPipelineOptions({ stages: data.stages ?? [], jobs: data.jobs ?? [] })
      })
  }, [open, form.gitProjectId])

  // 左侧「charts 默认值」自动拉取：关联 git 且有 localChartPath 时拉取。打开弹窗立即拉（不等待），
  // 之后 localChartPath 变化走 2s 防抖，避免输入时狂打接口。
  // 用 ref 记录「本次打开是否已首次拉取」：首次 0ms，后续 2000ms。
  // 注意：结果只进 chartDefaults（只读引用），不覆盖表单 valuesYaml——右侧可编辑值来自库里的保存配置。
  const initialValuesFetch = useRef(false)
  useEffect(() => {
    if (!open) {
      initialValuesFetch.current = false
      setValuesLoading(false) // 关闭时复位，避免上次在途请求因 alive 失效而卡住 loading 态
      return
    }
    if (!form.needGitRepo || !form.localChartPath.trim()) {
      setValuesLoading(false) // 路径被清空时在途请求被 cleanup 弃用（alive=false），这里兜底复位 loading
      return
    }
    const first = !initialValuesFetch.current
    if (first) initialValuesFetch.current = true
    let alive = true
    const timer = setTimeout(() => {
      setValuesLoading(true)
      api
        .POST('/api/git/get_chart_values_yaml', {
          body: { input: form.localChartPath.trim() },
        })
        .then(({ data, error }) => {
          if (!alive || error || !data) return
          setChartDefaults(data.values)
        })
        .finally(() => {
          if (alive) setValuesLoading(false)
        })
    }, first ? 0 : 2000)
    return () => {
      clearTimeout(timer)
      alive = false // 请求在途时 effect 重跑，丢弃过期响应，避免旧数据覆盖新输入
    }
  }, [open, form.needGitRepo, form.localChartPath])

  // 配置自动推导：chartDefaults（拉取的 chart 模板）+ configField 就绪时，防抖 1s 从模板里取
  // configField 指向的全局配置。对齐旧版：从左侧只读模板推导，不读右侧表单值。
  // 序列化跟随 configFileType：json 出 JSON 文本，其余出 YAML —— 否则切换类型后编辑器拿到的是旧格式，
  // 语言/lint 已切到新类型而内容不匹配。
  // 模板过大时跳过推导：yaml 包全量解析极慢（10MB 约 12s），同步解析会卡死主线程几秒。
  const CONFIG_PARSE_LIMIT = 100 * 1024 // 100KB
  useEffect(() => {
    if (!open) return
    const timer = setTimeout(() => {
      if (form.configField && chartDefaults) {
        if (chartDefaults.length > CONFIG_PARSE_LIMIT) {
          setConfigFileContent('')
          setConfigDetected(false)
          return
        }
        let parsed: unknown
        try {
          parsed = YAML.parse(chartDefaults)
        } catch {
          parsed = undefined
        }
        let data: unknown = deepGet(parsed, form.configField.split('->'))
        if (data && typeof data === 'object') {
          if (Array.isArray(data)) {
            const parts = form.configField.split('->')
            const key = parts[parts.length - 1] ?? ''
            data = { [String(key)]: data }
          }
          if (getMode(form.configFileType) === 'json') {
            setConfigFileContent(JSON.stringify(data, null, 2))
          } else {
            setConfigFileContent(YAML.stringify(data))
          }
        } else if (data) {
          setConfigFileContent(String(data))
        } else {
          setConfigFileContent('')
        }
        setConfigDetected(Boolean(data))
      } else {
        setConfigFileContent('')
        setConfigDetected(false)
      }
    }, 1000)
    return () => clearTimeout(timer)
  }, [open, form.configField, chartDefaults, form.configFileType])

  // 配置被"使用"后或打开时重置推导态
  useEffect(() => {
    if (!open) return
    setConfigFileContent('')
    setConfigDetected(false)
    setConfigPopoverOpen(false)
  }, [open])

  const applyDetectedConfig = () => {
    setForm((f) => ({ ...f, configFileValues: configFileContent }))
    setConfigDetected(false)
    setConfigPopoverOpen(false)
    setConfigFileContent('')
  }

  /** 复制左侧 charts 默认值（拉取的 chart 模板） */
  const copyValues = async () => {
    const ok = await copyText(chartDefaults)
    if (ok) toast.success(t('common.copied'))
    else toast.error(t('common.copyFailed'))
  }

  const projectOptions = useMemo(
    () => [
      { value: '0', label: t('repos.noProject'), description: '' },
      ...projects.map((p) => ({
        value: String(p.id),
        label: p.name,
        description: p.description,
      })),
    ],
    [projects, t],
  )

  /** 分支下拉选项：首项 '*' 固定表示「全部分支」，其余用服务端返回的 label */
  const branchOptions = useMemo(
    () => [
      { value: '*', label: t('repos.allBranches') },
      ...branches
        .filter((b) => b.branch && b.branch !== '*')
        .map((b) => ({ value: b.branch, label: b.label || b.branch })),
    ],
    [branches, t],
  )

  /** pipeline 通过规则的 stage/job 下拉选项（服务端 pipeline_job_options 返回） */
  const stageOptions = useMemo(
    () => pipelineOptions.stages.map((s) => ({ value: s, label: s })),
    [pipelineOptions.stages],
  )
  const jobOptions = useMemo(
    () => pipelineOptions.jobs.map((j) => ({ value: j, label: j })),
    [pipelineOptions.jobs],
  )

  const updateRule = (i: number, key: 'stageName' | 'jobName', v: string) => {
    setForm((f) => {
      const next = [...f.pipelinePassRules]
      next[i] = { ...next[i], [key]: v }
      return { ...f, pipelinePassRules: next }
    })
  }

  const addRule = () =>
    setForm((f) => ({
      ...f,
      pipelinePassRules: [...f.pipelinePassRules, { stageName: '', jobName: '' }],
    }))

  const removeRule = (i: number) =>
    setForm((f) => ({
      ...f,
      pipelinePassRules: f.pipelinePassRules.filter((_, idx) => idx !== i),
    }))

  const buildConfig = () => ({
    configFile: '',
    configFileValues: form.configFileValues,
    configField: form.configField,
    isSimpleEnv: false,
    configFileType: form.configFileType,
    localChartPath: form.localChartPath,
    branches: form.branches,
    pipelinePassRules: form.pipelinePassRules,
    valuesYaml: form.valuesYaml,
    elements: form.elements,
    displayName: '',
  })

  /** 自定义字段校验：对齐旧版 DynamicElement 的 antd rules */
  const validateElements = (): string | null => {
    for (const el of form.elements) {
      if (!el.path.trim()) return t('repos.elementPathRequired')
      if (!el.type || el.type === 'ElementTypeUnknown')
        return t('repos.elementTypeRequired')
      if (!el.description.trim()) return t('repos.elementDescriptionRequired')
      if (DEFAULT_REQUIRED_TYPES.has(el.type) && el.default === '')
        return t('repos.elementDefaultRequired')
      if (SELECTIVE_TYPES.has(el.type) && el.default !== '' && !el.selectValues.includes(el.default))
        return t('repos.elementDefaultInValues')
    }
    return null
  }

  const submit = async () => {
    if (!form.name.trim()) {
      toast.error(t('repos.nameRequired'))
      return
    }
    if (
      form.needGitRepo &&
      form.pipelinePassRules.some(
        (r) => !r.stageName.trim() || !r.jobName.trim(),
      )
    ) {
      toast.error(t('repos.ruleRequired'))
      return
    }
    const elErr = validateElements()
    if (elErr) {
      toast.error(elErr)
      return
    }
    setSaving(true)
    const body = {
      name: form.name.trim(),
      description: form.description.trim(),
      needGitRepo: form.needGitRepo,
      gitProjectId: form.needGitRepo ? form.gitProjectId || undefined : undefined,
      marsConfig: buildConfig(),
    }
    try {
      if (editItem) {
        const { error } = await api.PUT('/api/repos/{id}', {
          body: { ...body, id: editItem.id },
          params: { path: { id: editItem.id } },
        })
        if (error) throw new Error(error.message ?? String(error))
        toast.success(t('repos.updateSuccess'))
      } else {
        const { error } = await api.POST('/api/repos', { body })
        if (error) throw new Error(error.message ?? String(error))
        toast.success(t('repos.createSuccess'))
      }
      onClose()
      onSaved()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent side="right" className="!w-full !max-w-full p-0">
        <div className="flex h-full flex-col">
          {/* 标题栏：标题 + 取消/保存（还原旧版 Drawer title 栏放主按钮） */}
          <div className="flex shrink-0 items-center justify-between border-b border-line px-5 py-4">
            <SheetTitle className="text-[15px]">
              {editItem ? `${t('repos.update')}: ${editItem.name}` : t('repos.add')}
            </SheetTitle>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={onClose}>
                {t('common.cancel')}
              </Button>
              <Button variant="default" size="sm" disabled={saving} onClick={submit}>
                {saving && <Icon name="loader" className="size-4 animate-spin" />}
                {editItem ? t('common.save') : t('repos.create')}
              </Button>
            </div>
          </div>

          {/* 左右分屏：左 = charts 默认值只读预览，右 = 表单 */}
          <div className="grid h-full min-h-0 flex-1 grid-cols-1 gap-4 overflow-hidden p-5 md:grid-cols-2">
            {/* 左侧：values.yaml 只读预览 + 复制 */}
            <div className="flex min-h-0 flex-col overflow-hidden rounded-lg border border-line bg-surface">
              <div className="flex shrink-0 items-center justify-between border-b border-line px-3 py-2">
                <span className="flex items-center gap-1.5">
                  <span className="rounded bg-primary-soft px-2 py-0.5 text-[12px] font-medium text-primary">
                    {t('repos.chartsDefaults')}
                  </span>
                  {valuesLoading && <Icon name="loader" className="size-3 animate-spin text-mute" />}
                </span>
                <button
                  type="button"
                  onClick={copyValues}
                  className="flex items-center gap-1 text-[12px] text-mute transition-colors hover:text-primary"
                >
                  <Icon name="copy" className="text-[12px]" />
                  {t('common.copy')}
                </button>
              </div>
              <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain">
                {valuesLoading && !chartDefaults ? (
                  // 拉取图表默认值在途：给加载骨架，避免空白暗盒被误认为卡死
                  <div className="space-y-1.5 p-2" aria-busy>
                    {Array.from({ length: 12 }, (_, i) => (
                      <Skeleton
                        key={i}
                        className={`h-3 ${i % 3 === 0 ? 'w-3/4' : i % 3 === 1 ? 'w-1/2' : 'w-2/3'}`}
                      />
                    ))}
                  </div>
                ) : (
                  <CodeEditor
                    value={chartDefaults}
                    onChange={() => {}}
                    readOnly
                    className="h-full [&>.cm-editor]:h-full"
                  />
                )}
              </div>
            </div>

            {/* 右侧：表单 */}
            <div className="flex min-h-0 flex-col gap-5 overflow-y-auto overscroll-contain p-1">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5">
              <span className={labelCls}>
                {t('repos.appName')} <span className="text-err">*</span>
              </span>
              <Input value={form.name} onChange={(e) => set('name', e.target.value)} />
            </label>
            <label className="flex flex-col gap-1.5">
              <span className={labelCls}>{t('repos.appDescription')}</span>
              <Input
                value={form.description}
                onChange={(e) => set('description', e.target.value)}
              />
            </label>
          </div>

          <label className="flex items-center gap-3">
            <span className={labelCls}>{t('repos.needGitRepo')}</span>
            <Switch checked={form.needGitRepo} onCheckedChange={(v) => set('needGitRepo', v)} />
          </label>

          {form.needGitRepo && (
            <>
              <div className="flex flex-col gap-3">
                <label className="flex flex-col gap-1.5">
                  <span className={labelCls}>{t('repos.gitRepo')}</span>
                  <SearchableSelect
                    value={String(form.gitProjectId)}
                    options={projectOptions}
                    onChange={(v) => set('gitProjectId', Number(v))}
                    placeholder={t('repos.noProject')}
                    searchPlaceholder={t('common.search')}
                    emptyText={t('common.empty')}
                  />
                </label>
                <label className="flex flex-col gap-1.5">
                  <span className={labelCls}>{t('repos.enabledBranches')}</span>
                  {/* 多选：form.branches 即后端 mars.Config.branches（string[]）。'*' 表示全部并默认选中；
                      两者互斥：在「全部」上点具体分支 → 取消 '*' 保留该分支；点 '*' → 清掉具体分支 */}
                  <SearchableSelect
                    multiple
                    creatable
                    value={form.branches}
                    options={branchOptions}
                    onChange={(v) => {
                      const next = Array.isArray(v) ? v : [v]
                      // 之前选中「全部」，现在新增了具体分支 → 取消 '*'，保留该分支
                      if (
                        form.branches.includes('*') &&
                        next.some((b) => b !== '*')
                      ) {
                        set('branches', next.filter((b) => b !== '*'))
                        return
                      }
                      // 选中 '*'（全部）：清掉其它分支
                      if (next.includes('*')) set('branches', ['*'])
                      else set('branches', next)
                    }}
                    placeholder={t('repos.enabledBranches')}
                    searchPlaceholder={t('repos.branchSearchPlaceholder')}
                    emptyText={t('repos.branchSearchEmpty')}
                    limitText={(shown, total) =>
                      t('repos.branchSearchLimit', { shown, total })
                    }
                    createText={(q) => t('repos.branchSearchCreate', { name: q })}
                  />
                </label>
              </div>

              <div className="flex flex-col gap-2">
                <span
                  className={`${labelCls} flex items-center gap-1`}
                  title={t('repos.pipelineRulesTip')}
                >
                  {t('repos.pipelinePassRules')}
                </span>
                {form.pipelinePassRules.length > 0 && (
                  <div className="grid grid-cols-[1fr_1fr_auto] items-center gap-2 text-[12px] text-mute">
                    <span>{t('repos.stageName')}</span>
                    <span>{t('repos.jobName')}</span>
                    <span />
                  </div>
                )}
                {form.pipelinePassRules.map((r, i) => (
                  <div
                    key={i}
                    className="grid grid-cols-[1fr_1fr_auto] items-center gap-2"
                  >
                    <SearchableSelect
                      value={r.stageName}
                      options={stageOptions}
                      creatable
                      onChange={(v) => updateRule(i, 'stageName', v as string)}
                      placeholder={t('repos.stagePlaceholder')}
                      searchPlaceholder={t('common.search')}
                      emptyText={t('common.empty')}
                      createText={(q) => t('repos.stageCreate', { name: q })}
                    />
                    <SearchableSelect
                      value={r.jobName}
                      options={jobOptions}
                      creatable
                      onChange={(v) => updateRule(i, 'jobName', v as string)}
                      placeholder={t('repos.jobPlaceholder')}
                      searchPlaceholder={t('common.search')}
                      emptyText={t('common.empty')}
                      createText={(q) => t('repos.jobCreate', { name: q })}
                    />
                    <Button
                      variant="ghost"
                      size="icon-xs"
                      onClick={() => removeRule(i)}
                      aria-label={t('common.delete')}
                    >
                      <Icon name="close" />
                    </Button>
                  </div>
                ))}
                <Button
                  variant="outline"
                  className="border-dashed"
                  onClick={addRule}
                >
                  <Icon name="plus" />
                  {t('repos.addRule')}
                </Button>
              </div>
            </>
          )}

          <label className="flex flex-col gap-1.5">
            <span className={labelCls}>{t('repos.chartsPath')}</span>
            <Input
              value={form.localChartPath}
              onChange={(e) => set('localChartPath', e.target.value)}
              placeholder="pid|branch|path"
            />
          </label>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5">
              <span className={labelCls} title={t('repos.configFieldTip')}>
                {t('repos.configField')}
              </span>
              <Input
                value={form.configField}
                onChange={(e) => set('configField', e.target.value)}
                placeholder="conf->config"
              />
            </label>
            <label className="flex flex-col gap-1.5">
              <span className={labelCls}>{t('repos.configFileType')}</span>
              <SelectFileType
                value={form.configFileType}
                onChange={(v) => set('configFileType', v)}
              />
            </label>
          </div>

          <label className="flex flex-col gap-1.5">
            <span className="flex items-center justify-between">
              <span className={labelCls} title={t('repos.configFileValuesTip')}>
                {t('repos.configFileValues')}
              </span>
              {configDetected && !form.configFileValues && configFileContent && (
                <Popover open={configPopoverOpen} onOpenChange={setConfigPopoverOpen}>
                  <PopoverTrigger asChild>
                    <button
                      type="button"
                      className="rounded-md border border-primary/40 bg-primary-soft px-2 py-0.5 text-[11px] text-primary transition-colors hover:bg-primary/20"
                    >
                      {t('repos.configDetected')} →
                    </button>
                  </PopoverTrigger>
                  <PopoverContent align="end" side="top" className="w-[min(480px,90vw)] p-2">
                    <div className="mb-1.5 flex items-center justify-between">
                      <span className="text-[12px] font-medium">{t('repos.configDetected')}</span>
                      <Button size="sm" onClick={applyDetectedConfig}>
                        {t('repos.useConfig')}
                      </Button>
                    </div>
                    <pre className="max-h-56 overflow-auto overscroll-contain rounded-md bg-black/80 p-2 font-mono text-[11px] leading-relaxed text-green-300">
                      {configFileContent}
                    </pre>
                  </PopoverContent>
                </Popover>
              )}
            </span>
            <CodeEditor
              value={form.configFileValues}
              onChange={(v) => set('configFileValues', v)}
              minHeight="120px"
              language={getMode(form.configFileType)}
            />
          </label>

          <div className="flex flex-col gap-2">
            <span className={labelCls} title={t('repos.dynamicElementsTip')}>
              {t('repos.dynamicElements')}
            </span>
            <DynamicElement
              value={form.elements}
              onChange={(v) => set('elements', v)}
            />
          </div>

          <label className="flex flex-col gap-1.5">
            <span className="flex items-center gap-1.5">
              <span className={labelCls} title={t('repos.valuesYamlTip')}>
                {t('repos.valuesYaml')}
              </span>
              <span className="text-[11px] text-faint">
                {t('repos.valuesYamlAutoComplete')}
              </span>
            </span>
            <CodeEditor
              value={form.valuesYaml}
              onChange={(v) => set('valuesYaml', v)}
              minHeight="200px"
              yamlTemplateCompletion
            />
          </label>
            </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
