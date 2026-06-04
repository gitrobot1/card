import type { ComputedRef, Ref } from 'vue'
import type { YuzhoushaState, YzsCard, YzsSkillMeta } from '../../../types/yuzhousha'

export interface TakeTargetOption {
  zone: string
  cardId: string
}

export interface PendingContext {
  state: YuzhoushaState
  loading: boolean
  isAnimating: boolean
  mySeat: number
  opponentSeat: number
  isMyDraw: boolean
  isMyResponse: boolean
  canUsePeekDeckUI: boolean
  selectedId: Ref<string>
  selectedTargetZone: Ref<string>
  selectedTargetCardId: Ref<string>
  selectedQilinZone: Ref<string>
  shaTarget: Ref<number | null>
  liuliSelectedId: Ref<string>
  ganglieDiscardIds: Ref<string[]>
  ddzCancelDiscardIds: Ref<string[]>
  yijiSelectedIds: Ref<string[]>
  peekDeckTopIds: Ref<string[]>
  peekDeckBottomIds: Ref<string[]>
  fankuiTargetOptions: ComputedRef<TakeTargetOption[]>
  tuxiTargetOptions: ComputedRef<TakeTargetOption[]>
  qixiTargetOptions: ComputedRef<TakeTargetOption[]>
  myCharacterSkills: ComputedRef<YzsSkillMeta[]>
  peekDeckSkillId: ComputedRef<string>
  yijiGiveRemaining: ComputedRef<number>
  centerMessage: Ref<string>
  selectedCard: ComputedRef<YzsCard | null>
  responseRequiredKind: ComputedRef<string>
  canPlayWuxiek: ComputedRef<boolean>
  cardPlaysAsSha: (card: YzsCard | null | undefined) => boolean
  cardPlaysAsTao: (card: YzsCard | null | undefined) => boolean
  cardPlaysAsShan: (card: YzsCard | null | undefined) => boolean
  isRedCard: (card: YzsCard | null | undefined) => boolean
  isBlackCard: (card: YzsCard | null | undefined) => boolean
  act: (fn: () => Promise<YuzhoushaState>) => Promise<void>
}
