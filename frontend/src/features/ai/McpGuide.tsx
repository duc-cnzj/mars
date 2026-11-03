import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { Button } from '@/components/ui/shadcn/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { copyText } from '@/lib/copy'
import { toast } from '@/lib/toast'

/** 接入引导展示的客户端：文案键与配置片段一一对应。 */
type Client = 'claude' | 'opencode' | 'cursor'

const CLIENTS: readonly Client[] = ['claude', 'opencode', 'cursor']

const CLIENT_KEY = {
  claude: 'mcpGuide.clientClaude',
  opencode: 'mcpGuide.clientOpencode',
  cursor: 'mcpGuide.clientCursor',
} as const

/**
 * 生成各客户端的 MCP 配置片段。
 * 地址取当前站点 origin + `/mcp`；鉴权统一用环境变量 MARS_TOKEN，
 * 避免把用户令牌写进可被分享的配置文本。
 */
function snippetOf(client: Client, endpoint: string): string {
  switch (client) {
    case 'claude':
      return `claude mcp add --transport http mars ${endpoint} \\
  --header "Authorization: Bearer $MARS_TOKEN"`
    case 'opencode':
      return JSON.stringify(
        {
          mcp: {
            mars: {
              type: 'remote',
              url: endpoint,
              enabled: true,
              headers: { Authorization: 'Bearer {env:MARS_TOKEN}' },
            },
          },
        },
        null,
        2,
      )
    case 'cursor':
      return JSON.stringify(
        {
          mcpServers: {
            mars: {
              url: endpoint,
              headers: { Authorization: 'Bearer ${env:MARS_TOKEN}' },
            },
          },
        },
        null,
        2,
      )
  }
}

/**
 * MCP 接入引导：令牌管理页的入口弹窗。
 * 只提供可复制的客户端配置与地址说明，不读取/展示任何具体令牌——
 * 令牌由用户在本页创建后自行放进环境变量，避免引导页变成凭据泄露面。
 */
export function McpGuideButton() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [client, setClient] = useState<Client>('claude')
  const endpoint = useMemo(() => `${window.location.origin}/mcp`, [])
  const snippet = snippetOf(client, endpoint)

  const copy = async () => {
    const ok = await copyText(snippet)
    if (ok) toast.success(t('common.copied'))
    else toast.error(t('common.copyFailed'))
  }

  return (
    <>
      <Button size="sm" variant="outline" onClick={() => setOpen(true)}>
        <Icon name="terminal" className="size-4" />
        {t('mcpGuide.open')}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t('mcpGuide.title')}</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-3">
            <p className="text-[12px] text-mute">{t('mcpGuide.desc')}</p>
            <div className="flex flex-wrap items-center gap-2 rounded-md bg-raised px-3 py-2">
              <span className="text-[12px] text-mute">{t('mcpGuide.endpoint')}</span>
              <code className="min-w-0 flex-1 truncate font-mono text-[12px] text-primary">
                {endpoint}
              </code>
            </div>
            <p className="text-[12px] text-mute">{t('mcpGuide.tokenHint')}</p>

            <div className="flex flex-wrap items-center gap-2">
              {CLIENTS.map((c) => (
                <Button
                  key={c}
                  size="xs"
                  variant={client === c ? 'default' : 'outline'}
                  onClick={() => setClient(c)}
                >
                  {t(CLIENT_KEY[c])}
                </Button>
              ))}
            </div>

            <pre className="max-h-72 overflow-auto rounded-md border border-line bg-raised p-3 font-mono text-[12px] text-ink">
              {snippet}
            </pre>

            <div className="flex justify-end">
              <Button size="sm" variant="default" onClick={copy}>
                <Icon name="copy" className="size-4" />
                {t('mcpGuide.copy')}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
