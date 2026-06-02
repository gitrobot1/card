export const GAME_ROUTES: Record<string, string | null> = {
  doudizhu: '/games/doudizhu',
  zhajinhua: '/games/zhajinhua',
  douniu: '/games/douniu',
  sanguosha: null,
  uno: '/games/uno',
}

export function suitColor(suit: string) {
  if (suit === 'H' || suit === 'D') return 'red'
  if (suit === 'J') return 'joker'
  return 'black'
}

export function suitSymbol(suit: string) {
  switch (suit) {
    case 'S': return '♠'
    case 'H': return '♥'
    case 'C': return '♣'
    case 'D': return '♦'
    default: return ''
  }
}
