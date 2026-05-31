import gsap from 'gsap'
import type { Card, GameEvent } from '../types/doudizhu'

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function seatSelector(index: number) {
  return `.ddz__seat-anchor[data-seat="${index}"]`
}

export async function animatePlayEvent(
  event: GameEvent,
  playArea: HTMLElement,
  onLand?: (event: GameEvent) => void,
) {
  const fromEl = document.querySelector<HTMLElement>(seatSelector(event.player_index))
  if (!fromEl || !event.cards?.length) {
    onLand?.(event)
    return
  }

  const cards = event.cards
  const fromRect = fromEl.getBoundingClientRect()
  const toRect = playArea.getBoundingClientRect()
  const targetX = toRect.left + toRect.width / 2
  const targetY = toRect.top + toRect.height / 2

  const flyers: HTMLElement[] = []
  const tweens: Promise<void>[] = []

  cards.forEach((card, index) => {
    const el = document.createElement('div')
    el.className = `ddz-fly-card ddz-fly-card--${card.suit === 'H' || card.suit === 'D' ? 'red' : 'black'}`
    el.innerHTML = `<span>${card.label}</span>`
    document.body.appendChild(el)
    flyers.push(el)

    const offsetX = (index - (cards.length - 1) / 2) * 28
    const startX = fromRect.left + fromRect.width / 2 - 41 + offsetX
    const startY = fromRect.top + fromRect.height / 2 - 59

    gsap.set(el, {
      position: 'fixed',
      left: startX,
      top: startY,
      width: 82,
      height: 118,
      zIndex: 9999,
      opacity: 0.95,
      scale: 0.85,
      rotation: event.player_index === 0 ? 0 : event.player_index === 1 ? -12 : 12,
    })

    tweens.push(
      new Promise((resolve) => {
        gsap.to(el, {
          left: targetX + offsetX - 41,
          top: targetY - 59,
          scale: 1,
          rotation: 0,
          duration: 0.28,
          delay: index * 0.03,
          ease: 'power2.out',
          onComplete: () => resolve(),
        })
      }),
    )
  })

  await Promise.all(tweens)
  flyers.forEach((el) => el.remove())
  onLand?.(event)
  await sleep(100)
}

export async function showPassBubble(event: GameEvent) {
  const seat = document.querySelector<HTMLElement>(seatSelector(event.player_index))
  if (!seat) return

  const bubble = document.createElement('div')
  bubble.className = 'ddz-pass-bubble'
  bubble.textContent = '不出'
  seat.appendChild(bubble)

  gsap.fromTo(
    bubble,
    { opacity: 0, scale: 0.7, y: 8 },
    { opacity: 1, scale: 1, y: 0, duration: 0.25, ease: 'back.out(2)' },
  )

  await sleep(700)

  gsap.to(bubble, {
    opacity: 0,
    y: -8,
    duration: 0.2,
    onComplete: () => bubble.remove(),
  })
  await sleep(200)
}

export async function showCallBubble(event: GameEvent) {
  const seat = document.querySelector<HTMLElement>(seatSelector(event.player_index))
  if (!seat) return

  const bubble = document.createElement('div')
  bubble.className = 'ddz-pass-bubble'
  bubble.textContent = event.call ? '抢地主' : '不抢'
  seat.appendChild(bubble)

  gsap.fromTo(
    bubble,
    { opacity: 0, scale: 0.7 },
    { opacity: 1, scale: 1, duration: 0.25, ease: 'back.out(2)' },
  )
  await sleep(600)
  bubble.remove()
}

export function removeCardsFromHand(hand: Card[], playedIds: string[]) {
  const idSet = new Set(playedIds)
  return hand.filter((c) => !idSet.has(c.id))
}
