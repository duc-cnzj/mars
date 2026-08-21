/**
 * ANSI 转义序列解析（port 自旧前端 pkg/lazylog/ansiparse.js）。
 * 支持 30-37/90-97 前景色、40-47/100-107 背景色、38;5;n / 48;5;n 256 色、
 * 1 bold / 3 italic / 4 underline，以及 22/23/24 取消样式、39/49 重置前景/背景。
 * 保留 \b 退格擦除语义。38;2;r;g;b 真彩色当前不支持（解析时跳过参数）。
 * 渲染产物：AnsiText（React 组件，见 components/AnsiText.tsx）。
 */

const foregroundColors: Record<string, string> = {
  30: 'black',
  31: 'red',
  32: 'green',
  33: 'yellow',
  34: 'blue',
  35: 'magenta',
  36: 'cyan',
  37: 'white',
  90: 'grey',
}

const backgroundColors: Record<string, string> = {
  40: 'black',
  41: 'red',
  42: 'green',
  43: 'yellow',
  44: 'blue',
  45: 'magenta',
  46: 'cyan',
  47: 'white',
}

const styles: Record<string, 'bold' | 'italic' | 'underline'> = {
  1: 'bold',
  3: 'italic',
  4: 'underline',
}

/** 一段带样式的文本 */
export interface AnsiSegment {
  text: string
  foreground?: string
  background?: string
  bold?: boolean
  italic?: boolean
  underline?: boolean
}

/** 解析过程中的可变片段：text 在收尾时才赋值，因此显式标记为可选 */
type MutableSegment = Omit<AnsiSegment, 'text'> & { text?: string }

/** 退格 \b：删除已累积字符（末尾）或最后一个结果段的一个字符 */
function eraseChar(
  matchingText: string,
  result: AnsiSegment[],
): [string, AnsiSegment[]] {
  if (matchingText.length) {
    return [matchingText.substr(0, matchingText.length - 1), result]
  }
  if (result.length) {
    const index = result.length - 1
    const { text } = result[index]
    const newResult =
      text.length === 1
        ? result.slice(0, result.length - 1)
        : result.map((item, i) =>
            index === i ? { ...item, text: text.substr(0, text.length - 1) } : item,
          )
    return [matchingText, newResult]
  }
  return [matchingText, result]
}

/** 解析 ANSI 字符串为带样式片段数组 */
export function ansiparse(str: string): AnsiSegment[] {
  let matchingControl: string | null = null
  let matchingData: string | null = null
  let matchingText = ''
  let ansiState: string[] = []
  let result: AnsiSegment[] = []
  let state: MutableSegment = {}

  for (let i = 0; i < str.length; i += 1) {
    if (matchingControl !== null) {
      if (matchingControl === '\x1b' && str[i] === '[') {
        if (matchingText) {
          state.text = matchingText
          result.push(state as AnsiSegment)
          state = {}
          matchingText = ''
        }
        matchingControl = null
        matchingData = ''
      } else {
        matchingText += matchingControl + str[i]
        matchingControl = null
      }
      continue
    }

    if (matchingData !== null) {
      if (str[i] === ';') {
        ansiState.push(matchingData)
        matchingData = ''
      } else if (str[i] === 'm') {
        ansiState.push(matchingData)
        matchingData = null
        matchingText = ''

        for (let c = 0; c < ansiState.length; c += 1) {
          const ansiCode = ansiState[c]
          // 扩展色 38/48;5;n（256 色）与 38/48;2;r;g;b（真彩，跳过其参数）
          if (ansiCode === '38' || ansiCode === '48') {
            const isFg = ansiCode === '38'
            if (ansiState[c + 1] === '5') {
              const idx = parseInt(ansiState[c + 2], 10)
              if (!isNaN(idx) && idx >= 0 && idx <= 255) {
                const hex = ansi256(idx)
                if (isFg) state.foreground = hex
                else state.background = hex
              }
              c += 2
            } else if (ansiState[c + 1] === '2') {
              c += 3 // truecolor r,g,b：当前不支持，跳过
            }
            continue
          }
          if (foregroundColors[ansiCode]) {
            state.foreground = foregroundColors[ansiCode]
          } else if (backgroundColors[ansiCode]) {
            state.background = backgroundColors[ansiCode]
          } else if (ansiCode >= '90' && ansiCode <= '97') {
            // 亮前景 90-97：原仅实现 90(grey)，补齐其余（复用 256 色 8-15 同色）
            state.foreground = ansi256(8 + (parseInt(ansiCode, 10) - 90))
          } else if (ansiCode >= '100' && ansiCode <= '107') {
            // 亮背景 100-107：补齐
            state.background = ansi256(8 + (parseInt(ansiCode, 10) - 100))
          } else if (ansiCode === '39') {
            delete state.foreground
          } else if (ansiCode === '49') {
            delete state.background
          } else if (styles[ansiCode]) {
            state[styles[ansiCode]] = true
          } else if (ansiCode === '22') {
            state.bold = false
          } else if (ansiCode === '23') {
            state.italic = false
          } else if (ansiCode === '24') {
            state.underline = false
          }
        }
        ansiState = []
      } else {
        matchingData += str[i]
      }
      continue
    }

    if (str[i] === '\x1b') {
      matchingControl = str[i]
    } else if (str[i] === '\u0008') {
      ;[matchingText, result] = eraseChar(matchingText, result)
    } else {
      matchingText += str[i]
    }
  }

  if (matchingText) {
    state.text = matchingText + (matchingControl || '')
    result.push(state as AnsiSegment)
  }

  return result
}

/** 终端配色 → CSS 颜色（固定 16 色，保持跨主题一致） */
export const PALETTE: Record<string, string> = {
  black: '#181818',
  red: '#cd3131',
  green: '#0dbc79',
  yellow: '#e5e510',
  blue: '#2472c8',
  magenta: '#bc3fbc',
  cyan: '#11a8cd',
  white: '#e5e5e5',
  grey: '#767676',
}

/** 256 色索引 → hex。0-15 对齐现有 16 色 PALETTE（保证 31 与 38;5;1 同色），
 *  16-231 标准 6×6×6 色立方（levels 0,95,135,175,215,255），232-255 灰度渐变。
 *  扩展色直接产出 hex：segmentStyle 里 PALETTE[name] ?? hex 兼容命名色与 hex 两种存储 */
function ansi256(index: number): string {
  const toHex = (v: number) => v.toString(16).padStart(2, '0')
  if (index < 16) {
    const base = ['black', 'red', 'green', 'yellow', 'blue', 'magenta', 'cyan', 'white']
    if (index < 8) return PALETTE[base[index]]
    if (index === 8) return PALETTE.grey // 明亮黑对齐现有 90(grey)
    // 9-15 明亮变体（对齐 VS Code dark+ 明亮色）
    return ['#f14c4c', '#23d18b', '#f5f543', '#3b8eea', '#d670d6', '#29b8db', '#f5f5f5'][index - 9]
  }
  if (index < 232) {
    const n = index - 16
    const levels = [0, 95, 135, 175, 215, 255]
    const r = levels[Math.floor(n / 36) % 6]
    const g = levels[Math.floor(n / 6) % 6]
    const b = levels[n % 6]
    return `#${toHex(r)}${toHex(g)}${toHex(b)}`
  }
  const v = 8 + (index - 232) * 10
  return `#${toHex(v)}${toHex(v)}${toHex(v)}`
}
