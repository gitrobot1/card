import { Application, Container, Graphics, Text } from 'pixi.js'
import gsap from 'gsap'

export class CardStage {
  private app: Application | null = null
  private cards: Container[] = []

  async mount(host: HTMLElement) {
    const app = new Application()
    await app.init({
      width: host.clientWidth || 960,
      height: host.clientHeight || 540,
      background: '#0b1220',
      antialias: true,
      resizeTo: host,
    })

    host.replaceChildren(app.canvas)
    this.app = app

    const title = new Text({
      text: 'Card Stage Ready',
      style: {
        fill: '#e2e8f0',
        fontFamily: 'Avenir, Helvetica, Arial, sans-serif',
        fontSize: 28,
        fontWeight: '700',
      },
    })
    title.position.set(24, 20)
    app.stage.addChild(title)

    this.spawnCards()
  }

  private spawnCards() {
    if (!this.app) {
      return
    }

    const colors = [0xf97316, 0x22c55e, 0x3b82f6, 0xa855f7, 0xef4444]
    const startX = this.app.screen.width / 2 - 220

    colors.forEach((color, index) => {
      const card = this.createCard(color)
      card.position.set(startX + index * 110, this.app!.screen.height + 120)
      card.alpha = 0
      this.app!.stage.addChild(card)
      this.cards.push(card)

      gsap.to(card, {
        y: this.app!.screen.height / 2 - 70,
        alpha: 1,
        duration: 0.8,
        delay: index * 0.12,
        ease: 'back.out(1.7)',
      })

      gsap.to(card.scale, {
        x: 1.05,
        y: 1.05,
        duration: 1.2,
        delay: 0.8 + index * 0.12,
        yoyo: true,
        repeat: -1,
        ease: 'sine.inOut',
      })
    })
  }

  private createCard(color: number) {
    const card = new Container()

    const body = new Graphics()
      .roundRect(0, 0, 90, 130, 12)
      .fill({ color })
      .stroke({ color: 0xffffff, width: 2, alpha: 0.35 })

    const shine = new Graphics()
      .roundRect(8, 8, 24, 48, 8)
      .fill({ color: 0xffffff, alpha: 0.18 })

    card.addChild(body, shine)
    card.pivot.set(45, 65)
    return card
  }

  destroy() {
    gsap.killTweensOf(this.cards)
    this.cards.forEach((card) => gsap.killTweensOf(card.scale))
    this.cards = []
    this.app?.destroy(true, { children: true })
    this.app = null
  }
}
