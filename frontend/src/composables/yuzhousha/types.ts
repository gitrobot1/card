import type { ComputedRef, Ref } from 'vue'
import type { YuzhoushaState, YzsCard, YzsPlayer } from '../../types/yuzhousha'

export type YzsSeatAt = (seat: number) => YzsPlayer | undefined

export interface YzsTargetingDeps {
  state: Ref<YuzhoushaState | null>
  mySeat: ComputedRef<number>
  opponentSeat: ComputedRef<number>
  myPlayer: ComputedRef<YzsPlayer | undefined>
  myHand: ComputedRef<YzsCard[]>
  shaTarget: Ref<number | null>
  selectedTargetZone: Ref<string>
  selectedTargetCardId: Ref<string>
  selectedQilinZone: Ref<string>
  hitFlashSeat: Ref<number | null>
  blockFlashSeat: Ref<number | null>
  seatAt: YzsSeatAt
  isMyPlay: ComputedRef<boolean>
  isFinished: ComputedRef<boolean>
  isResponse: ComputedRef<boolean>
  isFankui: ComputedRef<boolean>
  isTuxiTake: ComputedRef<boolean>
  isQixiTake: ComputedRef<boolean>
  isPojun: ComputedRef<boolean>
  isPojunDiscard: ComputedRef<boolean>
  selectedCard: ComputedRef<YzsCard | null>
  canPlaySha: ComputedRef<boolean>
  cardPlaysAsSha: (card: YzsCard | null | undefined) => boolean
  needsOpponentTarget: (card: YzsCard | null | undefined) => boolean
  equipTagLabel: (card: YzsCard) => string
  isKongchengProtected: (player?: YzsPlayer) => boolean
  attackRangeOf: (player?: YzsPlayer) => number
}
