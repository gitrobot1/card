import gsap from 'gsap'
import { suitColor, suitSymbol } from '../../constants/games'
import { getYzsBeamStyle } from '../../constants/yzsBeamStyles'
import { equipDisplayName, weaponMetaForKind } from '../../constants/yzsWeapons'
import { beamEndpointsForSeats } from './useYzsBeamGeometry'
import type { YzsCard, YzsEvent } from '../../types/yuzhousha'

const SHA_BOLT_LENGTH = 216
const SHA_BOLT_HEIGHT = 2

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function seatSelector(index: number) {
  return `.ddz__seat-anchor[data-seat="${index}"]`
}

function createFlyCard(card: YzsCard): HTMLElement {
  const el = document.createElement('div')
  const pip = card.suit ? suitSymbol(card.suit) : ''
  const weaponMeta = weaponMetaForKind(card.kind)
  const displayName = equipDisplayName(card)
  const rangeHtml = weaponMeta ? `<span class="yzs-card__range">距离 ${weaponMeta.range}</span>` : ''
  el.className = `yzs-fly-card yzs-card yzs-card--${card.kind}${card.suit ? ` yzs-card--${suitColor(card.suit)}` : ''}${weaponMeta ? ' yzs-card--equip' : ''}`
  el.innerHTML = `<span class="yzs-card__corner"><span class="yzs-card__label">${card.label ?? '?'}</span>${
    pip ? `<span class="yzs-card__pip">${pip}</span>` : ''
  }</span>${rangeHtml}<span class="yzs-card__kind">${displayName}</span>`
  return el
}

function createBackCard(): HTMLElement {
  const el = document.createElement('div')
  el.className = 'yzs-fly-card yzs-card yzs-card--back'
  el.innerHTML = '<span class="yzs-card__back-mark">牌</span>'
  return el
}

function drawTarget(seatIndex: number, viewerSeat: number, handArea: HTMLElement | null) {
  if (seatIndex === viewerSeat && handArea) return handArea
  return document.querySelector<HTMLElement>(seatSelector(seatIndex))
}

/** 批量摸牌：先飞到手牌区，落点后由视图插入真实手牌 */
export async function animateYzsDrawBatch(
  drawArea: HTMLElement | null,
  seatIndex: number,
  count: number,
  viewerSeat: number,
  handArea: HTMLElement | null,
  onLand?: () => void,
) {
  if (!drawArea || count <= 0) {
    onLand?.()
    return
  }

  const toEl = drawTarget(seatIndex, viewerSeat, handArea)
  if (!toEl) {
    onLand?.()
    return
  }

  const fromRect = drawArea.getBoundingClientRect()
  const toRect = toEl.getBoundingClientRect()
  const el = createBackCard()
  if (count > 1) {
    el.innerHTML = `<span class="yzs-card__back-mark">+${count}</span>`
  }
  document.body.appendChild(el)

  gsap.set(el, {
    position: 'fixed',
    left: fromRect.left + fromRect.width / 2 - 28,
    top: fromRect.top + fromRect.height / 2 - 40,
    width: 56,
    height: 80,
    zIndex: 9999,
    opacity: 0.9,
    scale: 0.85,
  })

  await new Promise<void>((resolve) => {
    gsap.to(el, {
      left: toRect.left + toRect.width / 2 - 28,
      top: toRect.top + toRect.height * 0.2 - 40,
      scale: 0.92,
      duration: seatIndex === viewerSeat ? 0.34 : 0.22,
      ease: 'power2.inOut',
      onComplete: () => resolve(),
    })
  })

  onLand?.()
  await new Promise<void>((resolve) => {
    gsap.to(el, {
      opacity: 0,
      y: -10,
      duration: 0.08,
      ease: 'power1.out',
      onComplete: () => resolve(),
    })
  })
  el.remove()
  await sleep(60)
}

/** 开局发牌：牌堆飞向座位（保留兼容，内部走批量） */
export async function animateYzsDealToSeat(
  drawArea: HTMLElement,
  seatIndex: number,
  viewerSeat = 0,
  handArea: HTMLElement | null = null,
) {
  await animateYzsDrawBatch(drawArea, seatIndex, 1, viewerSeat, handArea)
}

/** 出牌飞向弃牌堆 */
export async function animateYzsPlayEvent(
  event: YzsEvent,
  discardArea: HTMLElement | null,
  mySeat: number,
  onLand?: () => void,
) {
  const card = event.card
  if (!card || !discardArea || event.player_index == null) {
    onLand?.()
    return
  }

  const fromEl = document.querySelector<HTMLElement>(seatSelector(event.player_index))
  const cardEl =
    event.player_index === mySeat
      ? document.querySelector<HTMLElement>(`[data-card-id="${card.id}"]`)
      : null
  const anchor = cardEl ?? fromEl
  if (!anchor) {
    onLand?.()
    return
  }

  if (cardEl) cardEl.style.opacity = '0'

  const fromRect = anchor.getBoundingClientRect()
  const toRect = discardArea.getBoundingClientRect()
  const el = createFlyCard(card)
  document.body.appendChild(el)

  gsap.set(el, {
    position: 'fixed',
    left: fromRect.left + fromRect.width / 2 - 32,
    top: fromRect.top + fromRect.height / 2 - 45,
    width: 64,
    height: 90,
    zIndex: 9999,
    opacity: 0.96,
    scale: event.player_index === mySeat ? 1 : 0.88,
    rotation: event.player_index === mySeat ? 0 : -8,
  })

  await new Promise<void>((resolve) => {
    gsap.to(el, {
      left: toRect.left + toRect.width / 2 - 32,
      top: toRect.top + toRect.height / 2 - 45,
      scale: 0.92,
      rotation: 0,
      duration: 0.3,
      ease: 'power2.out',
      onComplete: () => resolve(),
    })
  })

  el.remove()
  if (cardEl) cardEl.style.opacity = ''
  onLand?.()
  await sleep(70)
}

const DISCARD_STACK_GAP_X = 16
const DISCARD_STACK_GAP_Y = -12

async function flyOneDiscardCard(
  event: YzsEvent,
  playArea: HTMLElement,
  mySeat: number,
  handArea: HTMLElement | null,
  stackIndex: number,
  total: number,
): Promise<void> {
  const card = event.card
  if (!card || event.player_index == null) return

  const cardEl =
    event.player_index === mySeat
      ? document.querySelector<HTMLElement>(`[data-card-id="${card.id}"]`)
      : null
  const handAnchor = event.player_index === mySeat ? handArea : null
  const seatEl = document.querySelector<HTMLElement>(seatSelector(event.player_index))
  const anchor = cardEl ?? handAnchor ?? seatEl
  if (!anchor) return

  if (cardEl) cardEl.style.opacity = '0'

  const fromRect = anchor.getBoundingClientRect()
  const toRect = playArea.getBoundingClientRect()
  const centerOffset = ((total - 1) * DISCARD_STACK_GAP_X) / 2
  const stackOffsetX = stackIndex * DISCARD_STACK_GAP_X - centerOffset
  const stackOffsetY = stackIndex * DISCARD_STACK_GAP_Y
  const el = createFlyCard(card)
  document.body.appendChild(el)

  gsap.set(el, {
    position: 'fixed',
    left: fromRect.left + fromRect.width / 2 - 32,
    top: fromRect.top + fromRect.height / 2 - 45,
    width: 64,
    height: 90,
    zIndex: 9999 + stackIndex,
    opacity: 0.96,
    scale: event.player_index === mySeat ? 1 : 0.86,
    rotation: event.player_index === mySeat ? 0 : -6,
  })

  await new Promise<void>((resolve) => {
    gsap.to(el, {
      left: toRect.left + toRect.width / 2 - 32 + stackOffsetX,
      top: toRect.top + toRect.height / 2 - 45 + stackOffsetY,
      scale: 1,
      rotation: stackIndex * 3 - (total - 1) * 1.5,
      duration: 0.34,
      ease: 'power2.out',
      onComplete: () => resolve(),
    })
  })

  el.remove()
  if (cardEl) cardEl.style.opacity = ''
}

/** 批量弃牌：多张同时飞向牌桌中央并错开叠放 */
export async function animateYzsDiscardBatch(
  events: YzsEvent[],
  playArea: HTMLElement | null,
  mySeat: number,
  handArea: HTMLElement | null,
  onLand?: () => void,
) {
  const batch = events.filter((ev) => ev.card && ev.player_index != null)
  if (!playArea || batch.length === 0) {
    onLand?.()
    return
  }

  await Promise.all(
    batch.map((event, index) =>
      flyOneDiscardCard(event, playArea, mySeat, handArea, index, batch.length),
    ),
  )

  onLand?.()
  await sleep(80)
}

/** 出杀：光段从攻击方飞向目标，到达后消失（不含受击反馈） */
export async function animateYzsShaFlyBolt(
  sourceIndex: number,
  targetIndex: number,
  humanSeat: number,
  handArea: HTMLElement | null,
  styleId = 'dash-flow',
  root?: ParentNode | null,
) {
  const pts = beamEndpointsForSeats(sourceIndex, targetIndex, humanSeat, handArea, root)
  if (!pts) return

  const style = getYzsBeamStyle(styleId)
  const dx = pts.x2 - pts.x1
  const dy = pts.y2 - pts.y1
  const dist = Math.hypot(dx, dy)
  if (dist < 12) return

  const ux = dx / dist
  const uy = dy / dist
  const angle = (Math.atan2(dy, dx) * 180) / Math.PI
  const boltLen = Math.min(SHA_BOLT_LENGTH, dist)
  const travel = Math.max(0, dist - boltLen)
  const startX = pts.x1
  const startY = pts.y1
  const endX = startX + ux * travel
  const endY = startY + uy * travel

  const bolt = document.createElement('div')
  bolt.className = `yzs-fly-bolt ${style.flyClassName}`
  document.body.appendChild(bolt)

  gsap.set(bolt, {
    position: 'fixed',
    left: 0,
    top: 0,
    width: boltLen,
    height: SHA_BOLT_HEIGHT,
    x: startX,
    y: startY - SHA_BOLT_HEIGHT / 2,
    rotation: angle,
    transformOrigin: '0px 50%',
    opacity: 1,
    zIndex: 9998,
    pointerEvents: 'none',
  })

  await new Promise<void>((resolve) => {
    gsap.to(bolt, {
      x: endX,
      y: endY - SHA_BOLT_HEIGHT / 2,
      duration: 0.38,
      ease: 'power2.in',
      onComplete: () => resolve(),
    })
  })

  await new Promise<void>((resolve) => {
    gsap.to(bolt, {
      opacity: 0,
      scale: 0.55,
      duration: 0.12,
      ease: 'power1.in',
      onComplete: () => {
        bolt.remove()
        resolve()
      },
    })
  })

  await sleep(40)
}

/** 八卦阵判定：从牌堆翻牌到场心，展示红/黑判定结果 */
export async function animateYzsBaguaJudge(
  drawArea: HTMLElement | null,
  playArea: HTMLElement | null,
  card: YzsCard,
  success: boolean,
) {
  const stage =
    playArea?.closest('.yzs__center-stage') ??
    document.querySelector<HTMLElement>('.yzs__center-stage')
  if (!stage) return

  const fromRect =
    drawArea?.getBoundingClientRect() ??
    stage.getBoundingClientRect()
  const toRect = stage.getBoundingClientRect()

  const wrap = document.createElement('div')
  wrap.className = 'yzs-judge-wrap'
  document.body.appendChild(wrap)

  const cardEl = createFlyCard(card)
  cardEl.classList.add('yzs-judge-card')
  wrap.appendChild(cardEl)

  const badge = document.createElement('div')
  badge.className = `yzs-judge-badge yzs-judge-badge--${success ? 'success' : 'fail'}`
  badge.setAttribute('aria-hidden', 'true')
  badge.textContent = success ? '✓' : '✗'
  wrap.appendChild(badge)

  const label = document.createElement('p')
  label.className = 'yzs-judge-label'
  label.textContent = success ? '八卦阵 · 判定成功' : '八卦阵 · 判定失败'
  wrap.appendChild(label)

  const cardW = 76
  const cardH = 108
  const centerX = toRect.left + toRect.width / 2
  const centerY = toRect.top + toRect.height / 2 - 24

  gsap.set(wrap, {
    position: 'fixed',
    left: centerX,
    top: centerY,
    xPercent: -50,
    yPercent: -50,
    zIndex: 10000,
    pointerEvents: 'none',
  })

  gsap.set(cardEl, { width: cardW, height: cardH, opacity: 0.12, scale: 0.5 })
  gsap.set(badge, { opacity: 0, scale: 0.35 })
  gsap.set(label, { opacity: 0, y: 8 })

  const fromX = fromRect.left + fromRect.width / 2
  const fromY = fromRect.top + fromRect.height / 2
  gsap.set(wrap, { x: fromX - centerX, y: fromY - centerY, xPercent: -50, yPercent: -50 })

  await new Promise<void>((resolve) => {
    gsap.to(wrap, {
      x: 0,
      y: 0,
      duration: 0.44,
      ease: 'power2.out',
    })
    gsap.to(cardEl, {
      opacity: 1,
      scale: 1,
      duration: 0.44,
      ease: 'power2.out',
      onComplete: () => resolve(),
    })
  })

  await new Promise<void>((resolve) => {
    gsap.to(badge, {
      opacity: 1,
      scale: 1,
      duration: 0.28,
      ease: 'back.out(2)',
      onComplete: () => resolve(),
    })
    gsap.to(label, { opacity: 1, duration: 0.28, delay: 0.06 })
  })

  await sleep(720)

  await new Promise<void>((resolve) => {
    gsap.to(wrap, {
      opacity: 0,
      scale: 0.92,
      duration: 0.22,
      ease: 'power1.in',
      onComplete: () => {
        wrap.remove()
        resolve()
      },
    })
  })
}
