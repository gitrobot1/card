import gsap from 'gsap'

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export async function animateCardsFromCenter(
  cardElements: HTMLElement[],
  originElement: HTMLElement,
  staggerMs = 70,
) {
  if (cardElements.length === 0) {
    return
  }

  const originRect = originElement.getBoundingClientRect()
  const originX = originRect.left + originRect.width / 2
  const originY = originRect.top + originRect.height / 2

  for (let i = 0; i < cardElements.length; i++) {
    const el = cardElements[i]
    const rect = el.getBoundingClientRect()
    const targetX = rect.left + rect.width / 2
    const targetY = rect.top + rect.height / 2
    const deltaX = originX - targetX
    const deltaY = originY - targetY

    // 动画期间脱离文档流，避免 transform 撑大手牌区滚动容器
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
        duration: 0.22,
        ease: 'back.out(1.5)',
        onComplete: () => {
          gsap.set(el, { clearProps: 'all' })
          resolve()
        },
      })
    })

    await sleep(staggerMs)
  }
}

export async function animateOpponentDeal(count: number, onTick: (current: number) => void, staggerMs = 45) {
  for (let i = 1; i <= count; i++) {
    onTick(i)
    await sleep(staggerMs)
  }
}
