import gsap from 'gsap'

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function animateOneCardFromCenter(el: HTMLElement, originElement: HTMLElement, duration = 0.2) {
  const originRect = originElement.getBoundingClientRect()
  const originX = originRect.left + originRect.width / 2
  const originY = originRect.top + originRect.height / 2
  const rect = el.getBoundingClientRect()
  const targetX = rect.left + rect.width / 2
  const targetY = rect.top + rect.height / 2
  const deltaX = originX - targetX
  const deltaY = originY - targetY

  gsap.set(el, {
    position: 'fixed',
    left: rect.left,
    top: rect.top,
    width: rect.width,
    height: rect.height,
    margin: 0,
    zIndex: 1200,
    x: deltaX,
    y: deltaY,
    opacity: 0,
    scale: 0.55,
    rotation: -8,
  })

  await new Promise<void>((resolve) => {
    gsap.to(el, {
      x: 0,
      y: 0,
      opacity: 1,
      scale: 1,
      rotation: 0,
      duration,
      ease: 'back.out(1.5)',
      onComplete: () => {
        gsap.set(el, { clearProps: 'all' })
        resolve()
      },
    })
  })
}

export async function animateCardsFromCenter(
  cardElements: HTMLElement[],
  originElement: HTMLElement,
  staggerMs = 70,
) {
  if (cardElements.length === 0) {
    return
  }

  for (let i = 0; i < cardElements.length; i++) {
    await animateOneCardFromCenter(cardElements[i], originElement)
    if (i < cardElements.length - 1) {
      await sleep(staggerMs)
    }
  }
}

/** 多张牌同时从牌堆飞向目标（每人一次发完） */
export async function animateCardsFromCenterBatch(
  cardElements: HTMLElement[],
  originElement: HTMLElement,
  duration = 0.2,
) {
  if (cardElements.length === 0) {
    return
  }
  await Promise.all(cardElements.map((el) => animateOneCardFromCenter(el, originElement, duration)))
}

export async function animateOpponentDeal(count: number, onTick: (current: number) => void, staggerMs = 45) {
  for (let i = 1; i <= count; i++) {
    onTick(i)
    await sleep(staggerMs)
  }
}
