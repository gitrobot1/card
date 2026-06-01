import { onUnmounted, ref } from 'vue'

const FACE_ROTATIONS: Record<number, { x: number; y: number }> = {
  1: { x: 0, y: 0 },
  2: { x: 0, y: -90 },
  3: { x: -90, y: 0 },
  4: { x: 90, y: 0 },
  5: { x: 0, y: 90 },
  6: { x: 0, y: 180 },
}

export interface DiceRotation {
  x: number
  y: number
  z: number
}

export interface DiceRollOptions {
  value?: number
  duration?: number
  /** 为 false 时不显示 DiceRoller 浮层（供双骰居中动画使用） */
  show?: boolean
}

const DEFAULT_DURATION = 1800
const QUICK_ROLL_DURATION = 720

function createDiceAnimator(defaultDuration = DEFAULT_DURATION) {
  const visible = ref(false)
  const rolling = ref(false)
  const value = ref(1)
  const rotation = ref<DiceRotation>({ x: 0, y: 0, z: 0 })

  let rafId = 0

  function cancelAnimation() {
    if (rafId) {
      cancelAnimationFrame(rafId)
      rafId = 0
    }
  }

  function hide() {
    cancelAnimation()
    visible.value = false
    rolling.value = false
  }

  function roll(options: DiceRollOptions = {}): Promise<number> {
    cancelAnimation()

    const final = clampValue(options.value ?? randomFace())
    const duration = options.duration ?? defaultDuration
    const end = FACE_ROTATIONS[final]
    const start = { ...rotation.value }
    const spinTurns = duration <= QUICK_ROLL_DURATION ? 1 : 2 + Math.floor(Math.random() * 2)
    const spinX = 360 * spinTurns
    const spinY = 360 * spinTurns
    const spinZ = duration <= QUICK_ROLL_DURATION ? 0 : 360 * Math.floor(Math.random() * 2)
    const wobbleScale = duration <= QUICK_ROLL_DURATION ? 0.45 : 1

    const targetRotation = {
      x: settleAxisFrom(start.x, end.x, spinX),
      y: settleAxisFrom(start.y, end.y, spinY),
      z: settleAxisFrom(start.z, 0, spinZ * 0.35),
    }

    value.value = final
    if (options.show !== false) {
      visible.value = true
    }
    rolling.value = true

    return new Promise((resolve) => {
      const startAt = performance.now()

      const tick = (now: number) => {
        const t = Math.min(1, (now - startAt) / duration)
        const eased = 1 - (1 - t) ** 3
        const wobble = (1 - t) ** 2

        if (t < 1) {
          rotation.value = {
            x: start.x + (targetRotation.x - start.x) * eased + Math.sin(now / 70) * 28 * wobble * wobbleScale,
            y: start.y + (targetRotation.y - start.y) * eased + Math.cos(now / 55) * 28 * wobble * wobbleScale,
            z: start.z + (targetRotation.z - start.z) * eased + Math.sin(now / 90) * 12 * wobble * wobbleScale,
          }
          rafId = requestAnimationFrame(tick)
          return
        }

        rotation.value = { ...targetRotation }
        rolling.value = false
        rafId = 0
        resolve(final)
      }

      rafId = requestAnimationFrame(tick)
    })
  }

  return { visible, rolling, value, rotation, roll, hide, cancelAnimation }
}

export interface DoubleDiceRollOptions {
  d1?: number
  d2?: number
  duration?: number
  playerName?: string
}

export function useDoubleDiceRoll(quick = false) {
  const duration = quick ? QUICK_ROLL_DURATION : DEFAULT_DURATION
  const diceA = createDiceAnimator(duration)
  const diceB = createDiceAnimator(duration)

  async function rollPair(options: DoubleDiceRollOptions = {}) {
    const d1 = clampValue(options.d1 ?? randomFace())
    const d2 = clampValue(options.d2 ?? randomFace())
    const rollDuration = options.duration ?? duration

    diceA.hide()
    diceB.hide()

    await Promise.all([
      diceA.roll({ value: d1, duration: rollDuration, show: false }),
      diceB.roll({ value: d2, duration: rollDuration, show: false }),
    ])

    return { d1, d2, sum: d1 + d2 }
  }

  onUnmounted(() => {
    diceA.cancelAnimation()
    diceB.cancelAnimation()
  })

  return {
    diceA,
    diceB,
    rollPair,
  }
}

export function useDiceRoll() {
  const dice = createDiceAnimator()
  onUnmounted(dice.cancelAnimation)
  return dice
}

function clampValue(n: number) {
  return Math.min(6, Math.max(1, Math.round(n)))
}

function settleAxisFrom(current: number, faceDeg: number, minExtraSpin: number) {
  const base = current + minExtraSpin
  const remainder = ((base % 360) + 360) % 360
  let delta = faceDeg - remainder
  if (delta <= 0) delta += 360
  return base + delta
}

function randomFace() {
  return Math.floor(Math.random() * 6) + 1
}
