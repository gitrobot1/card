import gsap from 'gsap'
import type { UnoEvent } from '../types/uno'
import { unoCardCenterClass, unoColorClass } from '../types/uno'

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function seatSelector(index: number) {
  return `.ddz__seat-anchor[data-seat="${index}"]`
}

export async function animateUnoPlayEvent(
  event: UnoEvent,
  discardArea: HTMLElement | null,
  mySeat: number,
  onLand?: () => void,
) {
  const card = event.card
  if (!card || !discardArea) {
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
  const targetX = toRect.left + toRect.width / 2
  const targetY = toRect.top + toRect.height / 2

  const el = document.createElement('div')
  el.className = `uno-fly-card uno-card ${unoColorClass(card.color)}`
  el.innerHTML = `<span class="${unoCardCenterClass(card.label)}">${card.label}</span>`
  document.body.appendChild(el)

  const startX = fromRect.left + fromRect.width / 2 - 41
  const startY = fromRect.top + fromRect.height / 2 - 59

  gsap.set(el, {
    position: 'fixed',
    left: startX,
    top: startY,
    width: 82,
    height: 118,
    zIndex: 9999,
    opacity: 0.96,
    scale: event.player_index === mySeat ? 1 : 0.88,
    rotation: event.player_index === mySeat ? 0 : event.player_index === 1 ? -10 : 10,
  })

  await new Promise<void>((resolve) => {
    gsap.to(el, {
      left: targetX - 41,
      top: targetY - 59,
      scale: 0.92,
      rotation: 0,
      duration: 0.32,
      ease: 'power2.out',
      onComplete: () => resolve(),
    })
  })

  el.remove()
  if (cardEl) cardEl.style.opacity = ''
  onLand?.()
  await sleep(80)
}

export async function animateUnoDrawEvent(
  event: UnoEvent,
  drawArea: HTMLElement | null,
  viewerSeat: number,
  onLand?: () => void,
) {
  if (!drawArea) {
    onLand?.()
    return
  }

  const toEl = document.querySelector<HTMLElement>(seatSelector(event.player_index))
  if (!toEl) {
    onLand?.()
    return
  }

  const faceDown = event.player_index !== viewerSeat
  const card = event.card

  const el = document.createElement('div')
  if (faceDown || !card) {
    el.className = 'uno-fly-card uno-card uno-card--back'
    el.innerHTML = '<span class="uno-card__back-mark">UNO</span>'
  } else {
    el.className = `uno-fly-card uno-card ${unoColorClass(card.color)}`
    el.innerHTML = `<span class="${unoCardCenterClass(card.label)}">${card.label}</span>`
  }
  document.body.appendChild(el)

  const fromRect = drawArea.getBoundingClientRect()
  const toRect = toEl.getBoundingClientRect()

  gsap.set(el, {
    position: 'fixed',
    left: fromRect.left + fromRect.width / 2 - 26,
    top: fromRect.top + fromRect.height / 2 - 38,
    width: 52,
    height: 76,
    zIndex: 9999,
    opacity: 0.9,
    scale: 0.85,
  })

  await new Promise<void>((resolve) => {
    gsap.to(el, {
      left: toRect.left + toRect.width / 2 - 41,
      top: toRect.top + toRect.height / 2 - 59,
      width: 82,
      height: 118,
      scale: 1,
      duration: 0.28,
      ease: 'power2.out',
      onComplete: () => resolve(),
    })
  })

  el.remove()
  onLand?.()
  await sleep(60)
}

/** 开局发牌：牌堆飞向座位（牌背） */
export async function animateUnoDealToSeat(drawArea: HTMLElement, seatIndex: number) {
  const toEl = document.querySelector<HTMLElement>(seatSelector(seatIndex))
  if (!toEl) return

  const fromRect = drawArea.getBoundingClientRect()
  const toRect = toEl.getBoundingClientRect()

  const el = document.createElement('div')
  el.className = 'uno-fly-card uno-card uno-card--back'
  el.innerHTML = '<span class="uno-card__back-mark">UNO</span>'
  document.body.appendChild(el)

  gsap.set(el, {
    position: 'fixed',
    left: fromRect.left + fromRect.width / 2 - 26,
    top: fromRect.top + fromRect.height / 2 - 38,
    width: 52,
    height: 76,
    zIndex: 9999,
    opacity: 0.95,
  })

  await new Promise<void>((resolve) => {
    gsap.to(el, {
      left: toRect.left + toRect.width / 2 - 26,
      top: toRect.top + toRect.height / 2 - 38,
      scale: 0.72,
      duration: 0.16,
      ease: 'power2.out',
      onComplete: () => resolve(),
    })
  })

  el.remove()
}

/** 翻开展示牌堆顶牌到出牌区 */
export async function animateUnoRevealTopCard(
  drawArea: HTMLElement,
  discardArea: HTMLElement,
  label: string,
  colorClass: string,
) {
  const fromRect = drawArea.getBoundingClientRect()
  const toRect = discardArea.getBoundingClientRect()

  const el = document.createElement('div')
  el.className = `uno-fly-card uno-card ${colorClass}`
  el.innerHTML = `<span class="${unoCardCenterClass(label)}">${label}</span>`
  document.body.appendChild(el)

  gsap.set(el, {
    position: 'fixed',
    left: fromRect.left + fromRect.width / 2 - 26,
    top: fromRect.top + fromRect.height / 2 - 38,
    width: 52,
    height: 76,
    zIndex: 9999,
    opacity: 0.95,
    scale: 0.85,
  })

  await new Promise<void>((resolve) => {
    gsap.to(el, {
      left: toRect.left + toRect.width / 2 - 41,
      top: toRect.top + toRect.height / 2 - 59,
      width: 82,
      height: 118,
      scale: 1,
      duration: 0.28,
      ease: 'power2.out',
      onComplete: () => resolve(),
    })
  })

  el.remove()
  await sleep(120)
}
