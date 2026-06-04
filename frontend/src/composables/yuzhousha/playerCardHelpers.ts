import type { YzsCard, YzsPlayer } from '../../types/yuzhousha'

export function judgeAreaCards(player?: YzsPlayer) {
  return player?.judge_area ?? []
}

export function trickStaysInJudge(kind: string) {
  return kind === 'lebu' || kind === 'bingliang' || kind === 'shandian'
}

export function removeJudgeCardFromPlayer(player: YzsPlayer, card: YzsCard | undefined) {
  if (!player || !card) return player
  const area = player.judge_area?.filter((j) => j.id !== card.id) ?? []
  const next = { ...player, judge_area: area.length ? area : undefined }
  if (card.kind === 'lebu') next.skip_play = false
  if (card.kind === 'bingliang') next.skip_draw = false
  return next
}

export function equipSlotOf(card: YzsCard) {
  if (card.kind.startsWith('weapon_')) return 'weapon'
  if (card.kind === 'armor') return 'armor'
  if (card.kind === 'plus_horse') return 'plus_horse'
  if (card.kind === 'minus_horse') return 'minus_horse'
  return ''
}

export function equippedCards(player?: YzsPlayer) {
  return [
    player?.weapon,
    player?.armor,
    player?.plus_horse,
    player?.minus_horse,
  ].filter(Boolean) as YzsCard[]
}

export function removeKnownCardFromPlayer(player: YzsPlayer, card: YzsCard | undefined) {
  if (!player || !card) return player
  const next = { ...player }
  let removedKnown = false
  if (next.weapon?.id === card.id) {
    next.weapon = undefined
    removedKnown = true
  }
  if (next.armor?.id === card.id) {
    next.armor = undefined
    removedKnown = true
  }
  if (next.plus_horse?.id === card.id) {
    next.plus_horse = undefined
    removedKnown = true
  }
  if (next.minus_horse?.id === card.id) {
    next.minus_horse = undefined
    removedKnown = true
  }
  if (player.judge_area?.some((j) => j.id === card.id)) {
    return removeJudgeCardFromPlayer(next, card)
  }
  if (!removedKnown) {
    next.hand_count = Math.max(0, next.hand_count - 1)
  }
  return next
}
