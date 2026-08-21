import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/shadcn/button'

/** 404 页面：未知路由兜底 */
export function NotFound() {
  const { t } = useTranslation()
  return (
    <div className="flex h-screen flex-col items-center justify-center gap-4 bg-bg text-ink">
      <div className="font-mono text-[64px] font-bold leading-none text-primary">404</div>
      <p className="text-[14px] text-mute">{t('common.pageNotFound')}</p>
      <Link to="/">
        <Button variant="default">{t('common.back')}</Button>
      </Link>
    </div>
  )
}
