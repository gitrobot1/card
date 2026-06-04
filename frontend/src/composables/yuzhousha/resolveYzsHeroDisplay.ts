import type { YzsCharacter } from '../../types/yuzhousha'

export interface YzsHeroDisplaySource {
  id: string
  accent_color?: string
  portrait_url?: string
  default_skin_id?: string
  skin_id?: string
  display?: {
    skin_id?: string
    portrait_url?: string
    accent_color?: string
  }
}

/** Resolve portrait/accent for catalog or in-game character (supports future skin override). */
export function resolveYzsHeroDisplay(
  hero: YzsHeroDisplaySource,
  skinId?: string,
) {
  const resolvedSkinId =
    skinId ||
    hero.skin_id ||
    hero.display?.skin_id ||
    hero.default_skin_id ||
    `${hero.id}:default`

  return {
    skin_id: resolvedSkinId,
    portrait_url: hero.display?.portrait_url ?? hero.portrait_url ?? '',
    accent_color: hero.display?.accent_color ?? hero.accent_color ?? '#6b7280',
  }
}

export function heroAccentColor(hero: YzsCharacter, skinId?: string) {
  return resolveYzsHeroDisplay(hero, skinId).accent_color
}

export function heroPortraitUrl(hero: YzsCharacter, skinId?: string) {
  return resolveYzsHeroDisplay(hero, skinId).portrait_url
}
