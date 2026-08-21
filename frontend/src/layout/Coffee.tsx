import { useState } from 'react'
import { useTranslation } from 'react-i18next'

/**
 * 请喝咖啡赞助卡：正面为支付宝收款码，点击翻面显示微信收款码。
 * 还原旧版 Coffee.tsx 的翻转卡交互（旧版 hover，新版点击翻面）。
 */
export function Coffee() {
  const { t } = useTranslation()
  const [flipped, setFlipped] = useState(false)

  const cardSize = { width: 200, height: 273 }

  return (
    <div className="flex w-full flex-col items-center gap-2">
      <div className="coffee-title">{t('coffee.title')}</div>
      <button
        type="button"
        aria-label={t('coffee.flipHint')}
        onClick={() => setFlipped((v) => !v)}
        className="cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        style={{ perspective: '700px' }}
      >
        <div
          className="relative"
          style={{
            transformStyle: 'preserve-3d',
            transform: flipped ? 'rotateY(180deg)' : 'rotateY(0deg)',
            transition: 'transform 0.5s',
          }}
        >
          <img
            src="/alipay.jpg"
            alt={t('coffee.alipay')}
            className="block rounded-lg border border-line bg-surface shadow-md"
            style={{
              ...cardSize,
              backfaceVisibility: 'hidden',
            }}
          />
          <img
            src="/wechatpay.jpg"
            alt={t('coffee.wechat')}
            className="absolute inset-0 rounded-lg border border-line bg-surface shadow-md"
            style={{
              ...cardSize,
              backfaceVisibility: 'hidden',
              transform: 'rotateY(180deg)',
            }}
          />
        </div>
      </button>
      <p className="text-center text-[11px] leading-relaxed text-muted-foreground">
        {t('coffee.flipHint')}
      </p>
    </div>
  )
}
