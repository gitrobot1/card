import type { Card } from '../types/doudizhu'
import type { DouNiuHandLayout } from '../types/douniu'

export function cardsFromLayout(hand: Card[], layout?: DouNiuHandLayout | null): {
  head: Card[]
  niu: Card[]
} {
  const map = new Map(hand.map((c) => [c.id, c]))
  const head = (layout?.head_ids ?? []).map((id) => map.get(id)).filter(Boolean) as Card[]
  const niu = (layout?.niu_ids ?? []).map((id) => map.get(id)).filter(Boolean) as Card[]
  if (head.length + niu.length !== hand.length) {
    const sorted = sortHandByRank(hand)
    return { head: sorted, niu: [] }
  }
  return { head, niu }
}

export function sortHandByRank(hand: Card[]): Card[] {
  return [...hand].sort((a, b) => rankOrder(b.rank) - rankOrder(a.rank))
}

function rankOrder(rank: number): number {
  if (rank === 15) return 14
  if (rank === 14) return 1
  if (rank >= 11 && rank <= 13) return rank
  if (rank === 10) return 10
  return rank
}

export function isSpecialHandType(type?: string): boolean {
  return type === 'five_small' || type === 'bomb' || type === 'five_flower' || type === 'none'
}
