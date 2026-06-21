import type { Ref } from 'vue'
import type { YuzhoushaState, YzsCard, YzsEvent } from '../../../types/yuzhousha'

export interface EventReplayContext {
  event: YzsEvent
  state: Ref<YuzhoushaState | null>
  mySeat: number
  centerMessage: Ref<string>
  tableActionHint: Ref<string>
  displayedHand: Ref<YzsCard[]>
  displayedTableCards: Ref<YzsCard[]>
  drawAreaRef: Ref<HTMLElement | null>
  playAreaRef: Ref<HTMLElement | null>
  handAreaRef: Ref<HTMLElement | null>
  tableWrapRef: Ref<HTMLElement | null>
  appendDrawnCards: (cards: YzsCard[]) => void
  setTableCard: (card: YzsCard) => void
  sleep: (ms: number) => Promise<void>
  flashSeatHit: (seat: number) => Promise<void>
  flashSeatBlocked: (seat: number) => Promise<void>
  runShaFlyBolt: (source: number, target: number) => Promise<void>
  runHitSlash: (seat: number) => Promise<void>
  nextTick: () => Promise<void>
}
