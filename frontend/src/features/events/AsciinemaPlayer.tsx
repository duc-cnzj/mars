import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import * as AsciinemaPlayerLib from 'asciinema-player'
import 'asciinema-player/dist/bundle/asciinema-player.css'
import type { Options } from 'asciinema-player'
import { RadioGroup, RadioGroupItem } from '@/components/ui/shadcn/radio-group'
import { Button } from '@/components/ui/shadcn/button'
import { Icon } from '@/components/icons'

/** 倍速档位（对齐旧版 Player 的 8 档，默认 1.5） */
const SPEEDS = [0.5, 0.75, 1, 1.5, 2, 2.5, 3, 6]

/**
 * 交互式终端录像回放器：渲染 asciinema 播放器 + 倍速切换（对齐旧版 Player.tsx），
 * 顶部栏与倍速同排提供播放/暂停按钮（对齐旧版 Player 顶部的播放控制）。
 * src 传录像 JSON 字符串（asciinema-player 3.x 的 DataSource 直接支持）。
 * asciinema-player 不支持运行时改倍速，因此倍速/起始时间/播放状态任一变化都会重建播放器，
 * 重建时保留当前位置与播放/暂停状态（同旧版逻辑）。
 */
export function AsciinemaPlayer({ src }: { src: string }) {
  const { t } = useTranslation()
  const rootRef = useRef<HTMLDivElement>(null)
  const playerRef = useRef<AsciinemaPlayerLib.Player | null>(null)
  const pausedRef = useRef(true)
  const [playing, setPlaying] = useState(false)
  const [speed, setSpeed] = useState(1.5)
  const [startAt, setStartAt] = useState(0)
  const [autoPlay, setAutoPlay] = useState(false)

  useEffect(() => {
    const el = rootRef.current
    if (!el) return
    const opts: Options = {
      autoPlay,
      preload: true,
      startAt,
      speed,
      idleTimeLimit: 3,
      theme: 'tango',
      // 宽度自适应：按容器宽等比缩放（超宽缩小、偏窄放大），始终完整可见、无横向滚动
      fit: 'width',
      terminalLineHeight: 1.5,
    }
    let player: AsciinemaPlayerLib.Player | undefined
    try {
      player = AsciinemaPlayerLib.create({ data: src }, el, opts)
    } catch {
      playerRef.current = null
      return
    }
    playerRef.current = player
    player.addEventListener('play', () => {
      pausedRef.current = false
      setPlaying(true)
    })
    player.addEventListener('pause', () => {
      pausedRef.current = true
      setPlaying(false)
    })
    player.addEventListener('ended', () => {
      pausedRef.current = true
      setPlaying(false)
    })
    return () => {
      try {
        player?.dispose()
      } catch {
        /* ignore */
      }
      el.replaceChildren()
    }
  }, [src, speed, startAt, autoPlay])

  /** 切倍速：保留当前播放位置与播放状态（对齐旧版 Player.onChange） */
  const changeSpeed = (next: number) => {
    const p = playerRef.current
    const cur = p ? p.getCurrentTime() : 0
    setStartAt(cur > 0 ? cur : 0)
    setAutoPlay(cur > 0 ? !pausedRef.current : false)
    setSpeed(next)
  }

  const togglePlay = () => {
    const p = playerRef.current
    if (!p) return
    if (playing) void p.pause()
    else void p.play()
  }

  return (
    <div className="w-full">
      {/* 顶部栏：速度在左，播放/暂停按钮在右，同一行（对齐旧版 Player 顶部控制栏） */}
      <div className="mb-1.5 flex flex-wrap items-center gap-3">
        <span className="text-[12px] text-mute">{t('events.speed')}</span>
        <RadioGroup
          value={String(speed)}
          onValueChange={(v) => changeSpeed(Number(v))}
          className="flex flex-wrap items-center gap-x-4 gap-y-1"
        >
          {SPEEDS.map((s) => (
            <label
              key={s}
              className="flex cursor-pointer items-center gap-1.5 text-[12px] text-mute transition-colors hover:text-primary"
            >
              <RadioGroupItem value={String(s)} className="size-3.5" />
              <span>{s}x</span>
            </label>
          ))}
        </RadioGroup>
        <Button
          size="sm"
          variant="outline"
          onClick={togglePlay}
          className="ml-auto gap-1"
        >
          <Icon name={playing ? 'pause' : 'play'} className="size-3.5" />
          {playing ? t('events.pause') : t('events.play')}
        </Button>
      </div>
      <div ref={rootRef} className="w-full" />
    </div>
  )
}
