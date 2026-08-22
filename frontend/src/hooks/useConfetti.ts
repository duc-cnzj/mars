import { useCallback } from 'react'
import confetti from 'canvas-confetti'

/**
 * 部署成功彩蛋：从顶部洒落南瓜/圣诞树/爱心三种形状（canvas-confetti）。
 * 移植自旧版 mars/frontend/src/contexts/useCheers.ts。
 */
export function useConfetti() {
  const fireConfetti = useCallback(() => {
    // 形状通过 shapeFromPath 定义；运行时重复构建有性能损耗，这里沿用旧版写法
    //（旧版注释：can only use a path once in development and save the result）

    // pumpkin shape from https://thenounproject.com/icon/pumpkin-5253388/
    const pumpkin = confetti.shapeFromPath({
      path: 'M449.4 142c-5 0-10 .3-15 1a183 183 0 0 0-66.9-19.1V87.5a17.5 17.5 0 1 0-35 0v36.4a183 183 0 0 0-67 19c-4.9-.6-9.9-1-14.8-1C170.3 142 105 219.6 105 315s65.3 173 145.7 173c5 0 10-.3 14.8-1a184.7 184.7 0 0 0 169 0c4.9.7 9.9 1 14.9 1 80.3 0 145.6-77.6 145.6-173s-65.3-173-145.7-173zm-220 138 27.4-40.4a11.6 11.6 0 0 1 16.4-2.7l54.7 40.3a11.3 11.3 0 0 1-7 20.3H239a11.3 11.3 0 0 1-9.6-17.5zM444 383.8l-43.7 17.5a17.7 17.7 0 0 1-13 0l-37.3-15-37.2 15a17.8 17.8 0 0 1-13 0L256 383.8a17.5 17.5 0 0 1 13-32.6l37.3 15 37.2-15c4.2-1.6 8.8-1.6 13 0l37.3 15 37.2-15a17.5 17.5 0 0 1 13 32.6zm17-86.3h-82a11.3 11.3 0 0 1-6.9-20.4l54.7-40.3a11.6 11.6 0 0 1 16.4 2.8l27.4 40.4a11.3 11.3 0 0 1-9.6 17.5z',
      matrix: [
        0.020491803278688523, 0, 0, 0.020491803278688523, -7.172131147540983,
        -5.9016393442622945,
      ] as unknown as DOMMatrix,
    })
    // tree shape from https://thenounproject.com/icon/pine-tree-1471679/
    const tree = confetti.shapeFromPath({
      path: 'M120 240c-41,14 -91,18 -120,1 29,-10 57,-22 81,-40 -18,2 -37,3 -55,-3 25,-14 48,-30 66,-51 -11,5 -26,8 -45,7 20,-14 40,-30 57,-49 -13,1 -26,2 -38,-1 18,-11 35,-25 51,-43 -13,3 -24,5 -35,6 21,-19 40,-41 53,-67 14,26 32,48 54,67 -11,-1 -23,-3 -35,-6 15,18 32,32 51,43 -13,3 -26,2 -38,1 17,19 36,35 56,49 -19,1 -33,-2 -45,-7 19,21 42,37 67,51 -19,6 -37,5 -56,3 25,18 53,30 82,40 -30,17 -79,13 -120,-1l0 41 -31 0 0 -41z',
      matrix: [
        0.03597122302158273, 0, 0, 0.03597122302158273, -4.856115107913669,
        -5.071942446043165,
      ] as unknown as DOMMatrix,
    })
    // heart shape from https://thenounproject.com/icon/heart-1545381/
    const heart = confetti.shapeFromPath({
      path: 'M167 72c19,-38 37,-56 75,-56 42,0 76,33 76,75 0,76 -76,151 -151,227 -76,-76 -151,-151 -151,-227 0,-42 33,-75 75,-75 38,0 57,18 76,56z',
      matrix: [
        0.03333333333333333, 0, 0, 0.03333333333333333, -5.566666666666666,
        -5.533333333333333,
      ] as unknown as DOMMatrix,
    })

    const defaults = {
      scalar: 2,
      spread: 180,
      particleCount: 30,
      origin: { y: -0.1 },
      startVelocity: -35,
    }

    confetti({
      ...defaults,
      shapes: [pumpkin],
      colors: ['#ff9a00', '#ff7400', '#ff4d00'],
      zIndex: 9999,
    })
    confetti({
      ...defaults,
      shapes: [tree],
      colors: ['#8d960f', '#be0f10', '#445404'],
      zIndex: 9999,
    })
    confetti({
      ...defaults,
      shapes: [heart],
      colors: ['#f93963', '#a10864', '#ee0b93'],
      zIndex: 9999,
    })
  }, [])

  return fireConfetti
}
