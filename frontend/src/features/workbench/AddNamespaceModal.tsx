import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { api } from '@/api/client'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { Icon } from '@/components/Icons'

/** 名称格式校验（与旧版 AddNamespace 一致）：小写字母/数字/中划线 */
const NAME_RE = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/

/**
 * 新建命名空间弹窗：名称 + 描述，提交后回调刷新列表。
 * 名称需满足正则（小写字母/数字/中划线，且不以中划线开头或结尾）；已存在时由后端兜底。
 */
export function AddNamespaceModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean
  onClose: () => void
  onCreated: () => void
}) {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [desc, setDesc] = useState('')
  const [saving, setSaving] = useState(false)

  // 输入非空但不匹配正则时展示行内错误
  const nameError = name.trim() !== '' && !NAME_RE.test(name.trim())

  const submit = async () => {
    const value = name.trim()
    if (!value) {
      toast.error(t('workbench.nameRequired'))
      return
    }
    if (!NAME_RE.test(value)) {
      toast.error(t('workbench.nameFormatError'))
      return
    }
    setSaving(true)
    try {
      const { error } = await api.POST('/api/namespaces', {
        body: { namespace: value, description: desc.trim() },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('workbench.createSuccess', { name: value }))
      setName('')
      setDesc('')
      onClose()
      onCreated()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('workbench.addNamespace')}</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-3">
          <label className="flex flex-col gap-1.5">
            <span className="text-[12px] font-medium text-mute">{t('common.name')}</span>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && submit()}
              placeholder={t('workbench.searchPlaceholder')}
              autoFocus
              className={nameError ? 'border-err focus-visible:ring-err/40' : ''}
            />
            {nameError && (
              <span className="text-[11px] text-err">{t('workbench.nameFormatError')}</span>
            )}
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-[12px] font-medium text-mute">{t('common.description')}</span>
            <Input value={desc} onChange={(e) => setDesc(e.target.value)} />
          </label>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button variant="default" disabled={!name.trim() || nameError || saving} onClick={submit}>
            {saving && <Icon name="loader" className="size-4 animate-spin" />}
            {t('common.create')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
