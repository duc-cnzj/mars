import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { Empty, RefreshFade, SkeletonList, Tag } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { copyText } from '@/lib/copy'
import { toast } from '@/lib/toast'
import { api } from '@/api/client'
import type { components } from '@/api/schema'
import type { TKey } from '@/i18n/keys'

type ConfigGroup = components['schemas']['settings.ConfigGroup']

/** 敏感值统一掩码（与源码值隔离，避免密钥类明文常显） */
const MASK = '••••••'

/** 配置分组 ID → i18n 词条键（后端固定返回六组：服务/运行/插件/数据库/集群/认证） */
const GROUP_TITLE_KEY: Record<string, TKey> = {
  server: 'settings.groupServer',
  runtime: 'settings.groupRuntime',
  plugins: 'settings.groupPlugins',
  database: 'settings.groupDatabase',
  cluster: 'settings.groupCluster',
  auth: 'settings.groupAuth',
}

/** 标量配置 key → i18n 词条键；插件参数/凭证/OIDC 等扁平化子项不在映射内，标签回退原始 key */
const SETTING_LABEL_KEY: Record<string, TKey> = {
  app_port: 'settings.appPort',
  grpc_port: 'settings.grpcPort',
  debug: 'settings.debug',
  external_ip: 'settings.externalIp',
  log_channel: 'settings.logChannel',
  tracing_endpoint: 'settings.tracingEndpoint',
  cache_driver: 'settings.cacheDriver',
  db_auto_migrate: 'settings.dbAutoMigrate',
  git_server_cached: 'settings.gitServerCached',
  upload_max_size: 'settings.uploadMaxSize',
  upload_dir: 'settings.uploadDir',
  db_driver: 'settings.dbDriver',
  db_database: 'settings.dbDatabase',
  db_host: 'settings.dbHost',
  db_port: 'settings.dbPort',
  db_username: 'settings.dbUsername',
  db_password: 'settings.dbPassword',
  db_slow_log_enabled: 'settings.dbSlowLogEnabled',
  db_slow_log_threshold: 'settings.dbSlowLogThreshold',
  kubeconfig: 'settings.kubeconfig',
  ns_prefix: 'settings.nsPrefix',
  install_timeout: 'settings.installTimeout',
  s3_enabled: 'settings.s3Enabled',
  s3_endpoint: 'settings.s3Endpoint',
  s3_use_ssl: 'settings.s3UseSsl',
  s3_bucket: 'settings.s3Bucket',
  s3_access_key_id: 'settings.s3AccessKey',
  s3_secret_access_key: 'settings.s3SecretKey',
  admin_password: 'settings.adminPassword',
}

/**
 * 系统设置（管理员后台 · 只读配置展示）
 *
 * 只读视图：展示服务端已加载配置（config.yaml）的有效项，不提供任何编辑入口——
 * 配置的修改在服务端完成（编辑配置文件 + 重启），前端仅做展示。
 * - 数据由 /api/admin/settings 提供：按六组返回扁平 key/value 条目
 * - 敏感项（密码/凭证/token 等 masked=true）默认掩码，点眼睛查看明文、点复制复制明文
 * - 标签优先取词条映射，插件参数/镜像仓库凭证/OIDC 等扁平化子项回退展示原始配置 key
 */
export function SystemSettings() {
  const { t } = useTranslation()
  const [groups, setGroups] = useState<ConfigGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  // 明文可见的敏感字段（key 集合）：默认全隐藏，点眼睛逐个展开
  const [revealed, setRevealed] = useState<Record<string, boolean>>({})

  // 拉取配置分组视图（只读聚合，纯内存读取）
  useEffect(() => {
    let ignore = false
    setLoading(true)
    void api
      .GET('/api/admin/settings')
      .then(({ data, error: err }) => {
        if (ignore) return
        if (err) {
          setError(err.message ?? String(err))
          setGroups([])
          return
        }
        setError('')
        if (!data) return
        setGroups(data.groups)
      })
      .finally(() => {
        if (!ignore) setLoading(false)
      })
    return () => {
      ignore = true
    }
  }, [])

  /** 复制配置值：敏感项需先「眼睛」揭示明文才可复制（未揭示禁用，复制走剪贴板即显式授权，与掩码展示状态同步） */
  const copyValue = async (value: string) => {
    const ok = await copyText(value)
    if (ok) toast.success(t('settings.copied'))
    else toast.error(t('common.retry'))
  }

  /** 切换敏感字段明文可见性 */
  const toggleReveal = (key: string) => setRevealed((prev) => ({ ...prev, [key]: !prev[key] }))

  return (
    <div className="flex flex-col gap-4">
      {/* 页头：标题 + 只读标记 + 配置来源说明 */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <h2 className="text-[16px] font-semibold text-ink">{t('settings.title')}</h2>
          <Tag tone="mute">{t('settings.readonlyTag')}</Tag>
        </div>
        <p className="flex items-center gap-1.5 text-[12px] text-faint">
          <Icon name="info" className="size-3.5 shrink-0" />
          {t('settings.sourceNote')}
        </p>
      </div>

      {loading && groups.length === 0 ? (
        <section className="overflow-hidden rounded-lg border border-line bg-surface">
          <SkeletonList count={8} bare />
        </section>
      ) : error ? (
        <section className="overflow-hidden rounded-lg border border-line bg-surface">
          <Empty icon="gear" text={error} />
        </section>
      ) : (
        // 配置分组卡片：每组标题 + 字段（标签左列 / 值左列对齐，敏感项掩码 + 眼睛/复制）
        // 单次取数无重取 → version 静态 0，分组行挂载即播一次进入渐入（错峰对齐其他页）
        <RefreshFade version={0} className="flex flex-col gap-4">
          {groups.map((g) => {
            const titleKey = GROUP_TITLE_KEY[g.id]
            return (
              <section key={g.id} className="overflow-hidden rounded-lg border border-line bg-surface">
                <div className="border-b border-line px-4 py-2.5 text-[13px] font-medium text-ink">
                  {titleKey ? t(titleKey) : g.id}
                </div>
                <div className="divide-y divide-line">
                  {g.items.map((item) => {
                    // 字段级明文可见性：驱动该敏感项掩码展开
                    const visible = !item.masked || revealed[item.key]
                    const labelKey = SETTING_LABEL_KEY[item.key]
                    return (
                      <div
                        key={item.key}
                        className="grid grid-cols-[10rem_minmax(0,1fr)_auto] items-center gap-3 px-4 py-2.5"
                      >
                        <span className="truncate text-[13px] text-mute" title={labelKey ? undefined : item.key}>
                          {labelKey ? t(labelKey) : item.key}
                        </span>

                        <span
                          className={`truncate font-mono text-[13px] ${
                            item.masked && !visible ? 'text-mute' : 'text-ink'
                          }`}
                          title={item.masked && !visible ? undefined : item.value}
                        >
                          {item.masked && !visible ? MASK : item.value}
                        </span>

                        <div className="flex items-center gap-1">
                          {item.masked && (
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              className="size-7"
                              aria-label={visible ? t('settings.hide') : t('settings.reveal')}
                              onClick={() => toggleReveal(item.key)}
                            >
                              <Icon name={visible ? 'eye-off' : 'eye'} className="size-3.5" />
                            </Button>
                          )}
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-7"
                            aria-label={t('common.copy')}
                            disabled={!visible}
                            onClick={() => copyValue(item.value)}
                          >
                            <Icon name="copy" className="size-3.5" />
                          </Button>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </section>
            )
          })}
        </RefreshFade>
      )}
    </div>
  )
}
