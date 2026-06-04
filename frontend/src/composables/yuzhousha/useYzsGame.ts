import {
  computed,
  inject,
  onMounted,
  provide,
  ref,
  watch,
  type InjectionKey,
} from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePhaseTimer } from '../usePhaseTimer'
import { useYzsTargeting } from './useYzsTargeting'
import { useYzsHints } from './useYzsHints'
import { useYzsAnimations } from './useYzsAnimations'
import {
  equipSlotOf,
  equippedCards,
  judgeAreaCards,
  removeJudgeCardFromPlayer,
  removeKnownCardFromPlayer,
  trickStaysInJudge,
} from './playerCardHelpers'
import { showToast } from '../useToast'
import {
  equipDisplaySummary,
  equipMetaForKind,
  weaponMetaForKind,
  weaponRangeForKind,
} from '../../constants/yzsWeapons'
import {
  discardYuzhoushaCards,
  endYuzhoushaPlay,
  getYuzhoushaState,
  passYuzhoushaPrepare,
  passYuzhoushaDraw,
  passYuzhoushaResponse,
  baguaYuzhoushaJudge,
  playYuzhoushaCard,
  respondYuzhoushaCard,
  useYuzhoushaSkill,
  tickYuzhoushaGame,
} from '../../api/games'
import {
  YZS_CARD_LABELS,
  type YuzhoushaState,
  type YzsCard,
  type YzsSkillMeta,
} from '../../types/yuzhousha'
import {
  pendingAllowsCancel,
  pendingCanPlayCard,
  pendingCanSubmitPlay,
  pendingCanSubmitSkill,
  pendingHint,
  pendingIsSkillOnly,
  pendingOnModeChange,
  pendingSubmitAction,
  pendingSubmitPlay,
  pendingSubmitSkill,
  pendingSuppressPlaySubmit,
  type PendingContext,
} from './pendingRegistry'

const YZS_CARD_WIDTH = 64

export const YZS_GAME_KEY: InjectionKey<ReturnType<typeof useYzsGame>> = Symbol('yzsGame')

export function provideYzsGame(ctx: ReturnType<typeof useYzsGame>) {
  provide(YZS_GAME_KEY, ctx)
}

export function useYzsGameInject() {
  const ctx = inject(YZS_GAME_KEY)
  if (!ctx) {
    throw new Error('useYzsGameInject must be used within YuzhoushaView')
  }
  return ctx
}

export function useYzsGame() {
const router = useRouter()
const route = useRoute()
const drawAreaRef = ref<HTMLElement | null>(null)
const handAreaRef = ref<HTMLElement | null>(null)
const playAreaRef = ref<HTMLElement | null>(null)
const tableWrapRef = ref<HTMLElement | null>(null)
const isBoltFlying = ref(false)
const hitFlashSeat = ref<number | null>(null)
const blockFlashSeat = ref<number | null>(null)

const state = ref<YuzhoushaState | null>(null)
const loading = ref(false)
const isDealing = ref(false)
const isAnimating = ref(false)
const selectedId = ref('')
const selectedDiscardIds = ref<string[]>([])
const shaTarget = ref<number | null>(null)
const selectedTargetZone = ref('')
const selectedTargetCardId = ref('')
const centerMessage = ref('')
const tableActionHint = ref('')
const displayedHand = ref<YzsCard[]>([])
const displayedTableCards = ref<YzsCard[]>([])
const displayedDealCounts = ref<Record<number, number>>({})
const enteringDrawCardIds = ref<string[]>([])

const mySeat = computed(() => state.value?.human_player ?? 0)
const opponentSeat = computed(() => {
  const players = state.value?.players
  const me = mySeat.value
  if (state.value?.mode === '3p_chain' && players?.length === 3) {
    return (me - 1 + 3) % 3
  }
  return 1 - me
})
const myPlayer = computed(() => state.value?.players[mySeat.value])
const opponent = computed(() => state.value?.players[opponentSeat.value])
const seatAt = (seat: number) => state.value?.players[seat]
const isFinished = computed(() => state.value?.phase === 'finished')
const isResponse = computed(() => state.value?.phase === 'response')
const isMyTurn = computed(() => state.value?.current_turn === mySeat.value)
const isMyResponse = computed(
  () =>
    isResponse.value &&
    (state.value?.pending?.target_index === mySeat.value ||
      (state.value?.pending?.response_mode === 'dying_rescue' &&
        state.value?.pending?.source_index === mySeat.value) ||
      (state.value?.pending?.response_mode === 'wugu_pick' &&
        state.value?.pending?.wugu_pick_seat === mySeat.value)),
)
const isMyPrepare = computed(
  () =>
    state.value?.phase === 'playing' &&
    isMyTurn.value &&
    state.value.turn_step === 'prepare',
)
const isMyDraw = computed(
  () =>
    state.value?.phase === 'playing' &&
    isMyTurn.value &&
    state.value.turn_step === 'draw' &&
    (state.value.activatable_skills?.some((s) => s.id === 'luoyi' || s.id === 'tuxi' || s.id === 'shuangxiong') ?? false),
)
const isPeekDeck = computed(
  () =>
    isResponse.value &&
    state.value?.pending?.response_mode === 'peek_deck' &&
    state.value?.pending?.target_index === mySeat.value,
)
const peekDeckSkillId = computed(() => state.value?.pending?.skill_id ?? '')
const isMyPlay = computed(
  () =>
    state.value?.phase === 'playing' &&
    isMyTurn.value &&
    state.value.turn_step === 'play',
)
const isMyDiscard = computed(
  () =>
    state.value?.phase === 'playing' &&
    isMyTurn.value &&
    state.value.turn_step === 'discard',
)
const canUsePeekDeckUI = computed(
  () => isPeekDeck.value && !loading.value && !isDealing.value,
)

const canInteract = computed(
  () =>
    !loading.value &&
    !isDealing.value &&
    !isAnimating.value &&
    !isFinished.value &&
    (isMyResponse.value || isMyPlay.value || isMyDiscard.value || isMyPrepare.value || isMyDraw.value || isPeekDeck.value || isJijiHeal.value),
)

const peekDeckTopIds = ref<string[]>([])
const peekDeckBottomIds = ref<string[]>([])


function opponentHasKongcheng(player = opponent.value) {
  return player?.character?.skill_ids?.includes('kongcheng') ?? false
}

function isKongchengProtected(player = opponent.value) {
  return opponentHasKongcheng(player) && (player?.hand_count ?? 0) === 0
}

const myHand = computed(() => displayedHand.value)

const discardNeeded = computed(() => {
  if (!isMyDiscard.value) return 0
  return Math.max(0, myHand.value.length - (myPlayer.value?.hp ?? 0))
})

const canPlaySha = computed(() => {
  if (isGuanYuFollow.value) return true
  if (!isMyPlay.value) return false
  if (myPlayer.value?.weapon?.kind === 'weapon_1') return true
  if (hasMySkill('paoxiao')) return true
  if (!myPlayer.value?.sha_used_this_turn) return true
  if (
    state.value?.mode === '3p_ddz' &&
    mySeat.value === (state.value.landlord_seat ?? 0) &&
    !myPlayer.value?.sha_extra_used_this_turn
  ) {
    return true
  }
  return false
})

function hasMySkill(skillId: string) {
  return myPlayer.value?.character?.skill_ids?.includes(skillId) ?? false
}

function isRedCard(card: YzsCard | null | undefined) {
  if (!card?.suit) return false
  if (card.suit === 'H' || card.suit === 'D') return true
  return hasMySkill('hongyan') && card.suit === 'S'
}

function isBlackCard(card: YzsCard | null | undefined) {
  return !!card?.suit && (card.suit === 'S' || card.suit === 'C')
}

function isDiamondCard(card: YzsCard | null | undefined) {
  return card?.suit === 'D'
}

function cardPlaysAsSha(card: YzsCard | null | undefined) {
  if (!card) return false
  if (card.kind === 'sha') return true
  if (hasMySkill('longdan') && card.kind === 'shan') return true
  if (hasMySkill('wusheng') && isRedCard(card)) {
    if (isMyPlay.value && !isMyResponse.value) {
      return wushengMode.value
    }
    return true
  }
  return false
}

function cardPlaysAsShan(card: YzsCard | null | undefined) {
  if (!card) return false
  if (card.kind === 'shan') return true
  if (hasMySkill('longdan') && card.kind === 'sha') return true
  if (hasMySkill('qingguo') && isBlackCard(card)) return true
  return false
}

function cardPlaysAsTao(card: YzsCard | null | undefined) {
  if (!card) return false
  if (card.kind === 'tao') return true
  if (hasMySkill('jiji') && isRedCard(card)) {
    if (isJijiHeal.value || isDyingRescue.value) {
      return state.value?.current_turn !== mySeat.value
    }
  }
  return false
}

const isJijiHeal = computed(
  () =>
    state.value?.phase === 'playing' &&
    state.value.turn_step === 'play' &&
    !isMyTurn.value &&
    !isResponse.value &&
    hasMySkill('jiji') &&
    (myPlayer.value?.hp ?? 0) < (myPlayer.value?.max_hp ?? 4),
)

function shuangxiongActive() {
  return (myPlayer.value?.skill_counters?.shuangxiong_active ?? 0) > 0
}

function shuangxiongRefIsRed() {
  return (myPlayer.value?.skill_counters?.shuangxiong_ref_red ?? 0) > 0
}

function cardValidForShuangxiong(card: YzsCard | null | undefined) {
  if (!card?.suit || !shuangxiongActive()) return false
  return isRedCard(card) !== shuangxiongRefIsRed()
}

const isGuanYuFollow = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'guanyu_follow',
)
const isQilinBow = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'qilin_bow',
)
const isWuguPick = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'wugu_pick',
)
const selectedQilinZone = ref('')
const rendeMode = ref(false)
const rendeSelectedIds = ref<string[]>([])
const zhihengMode = ref(false)
const zhihengSelectedIds = ref<string[]>([])
const jieyinMode = ref(false)
const jieyinSelectedIds = ref<string[]>([])
const fanjianMode = ref(false)
const fanjianSelectedId = ref('')
const qixiMode = ref(false)
const qixiSelectedId = ref('')
const guoseMode = ref(false)
const shuangxiongMode = ref(false)
const shuangxiongSelectedId = ref('')
const guoseSelectedId = ref('')
const liuliSelectedId = ref('')
const wushengMode = ref(false)
const ganglieDiscardIds = ref<string[]>([])
const ddzCancelDiscardIds = ref<string[]>([])
const yijiSelectedIds = ref<string[]>([])

const activatableSkills = computed(() => state.value?.activatable_skills ?? [])

const myCharacterSkills = computed(() => myPlayer.value?.character?.skills ?? [])

const activatableSkillIds = computed(
  () => new Set(activatableSkills.value.map((s) => s.id)),
)

const wushengSkillHint = computed(() => {
  if (!wushengMode.value) return ''
  return '【武圣】已发动：可选红色牌当【杀】。点上方「取消武圣」恢复正常出牌'
})

const canCancelWusheng = computed(
  () => wushengMode.value && (isMyPlay.value || isMyResponse.value) && canInteract.value,
)

const isJijiangRespond = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_jijiang',
)

const isJianxiong = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_jianxiong',
)
const isYijiOffer = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_yiji_offer',
)
const isYijiGive = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_yiji_give',
)
const yijiGiveRemaining = computed(() => state.value?.pending?.yiji_give_remaining ?? 0)
const isGanglieOffer = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_ganglie_offer',
)
const isGanglieChoice = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_ganglie_choice',
)
const isDdzJudgeCancel = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'ddz_judge_cancel',
)
const isFankui = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_fankui',
)
const isTuxiTake = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_tuxi',
)
const isQixiTake = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_qixi',
)
const isYinghunChoice = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_yinghun',
)
const isYinghunDiscard = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_yinghun_discard',
)
const isGuicai = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_guicai',
)
const isGuidao = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_guidao',
)
const isLeijiOffer = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_leiji_offer',
)
const isFanjianSuit = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_fanjian_suit',
)
const isTianxiangOffer = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_tianxiang',
)
const isLiuliOffer = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_liuli',
)
const isDyingRescue = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'dying_rescue',
)
const isLuanwu = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_luanwu',
)
const isSkillOnlyResponse = computed(() => pendingIsSkillOnly(state.value))

const fankuiSourceSeat = computed(() =>
  isFankui.value ? (state.value?.pending?.source_index ?? opponentSeat.value) : opponentSeat.value,
)
const tuxiSourceSeat = computed(() =>
  isTuxiTake.value ? (state.value?.pending?.source_index ?? opponentSeat.value) : opponentSeat.value,
)
const qixiSourceSeat = computed(() =>
  isQixiTake.value ? (state.value?.pending?.source_index ?? opponentSeat.value) : opponentSeat.value,
)

const qilinHorseOptions = computed(() => {
  if (!isQilinBow.value) return []
  const target = state.value?.pending?.effect_target ?? opponentSeat.value
  const player = state.value?.players[target]
  const options: { zone: string; label: string }[] = []
  if (player?.plus_horse) options.push({ zone: 'plus_horse', label: player.plus_horse.name })
  if (player?.minus_horse) options.push({ zone: 'minus_horse', label: player.minus_horse.name })
  return options
})

const selectedCard = computed(() => myHand.value.find((c) => c.id === selectedId.value) ?? null)
const selfTargetKinds = new Set([
  'tao',
  'taoyuan',
  'wuzhong',
  'wugu',
  'shandian',
  'nanman',
  'wanjian',
  'jiu',
  'weapon_1',
  'weapon_2',
  'weapon_3',
  'weapon_4',
  'weapon_5',
  'armor',
  'plus_horse',
  'minus_horse',
])

const responseRequiredKind = computed(() => state.value?.pending?.required_kind ?? 'shan')
const isWuxiekOffer = computed(
  () =>
    isResponse.value &&
    (state.value?.pending?.response_mode === 'wuxiek_trick' ||
      state.value?.pending?.response_mode === 'wuxiek_lebu' ||
      state.value?.pending?.response_mode === 'wuxiek_bingliang' ||
      state.value?.pending?.response_mode === 'wuxiek_shandian'),
)
const canPlayWuxiek = computed(
  () =>
    isMyResponse.value &&
    (isWuxiekOffer.value || state.value?.pending?.allow_wuxiek === true),
)
const canSubmitBagua = computed(
  () =>
    isMyResponse.value &&
    !isWuxiekOffer.value &&
    !isGuanYuFollow.value &&
    !isQilinBow.value &&
    responseRequiredKind.value === 'shan' &&
    !!myPlayer.value?.armor &&
    !state.value?.pending?.bagua_used &&
    !state.value?.pending?.ignore_armor &&
    !loading.value &&
    !isAnimating.value,
)

function cardLabel(kind: string | undefined) {
  if (!kind) return '牌'
  return YZS_CARD_LABELS[kind] ?? kind
}

const opponentTargetKinds = new Set(['sha', 'guohe', 'tannang', 'juedou', 'lebu', 'bingliang'])

function needsOpponentTarget(card: YzsCard | null | undefined) {
  if (!card) return false
  if (cardPlaysAsSha(card)) return true
  return opponentTargetKinds.has(card.kind)
}

function needsSelfTarget(card: YzsCard | null | undefined) {
  return !!card && selfTargetKinds.has(card.kind)
}

function weaponRange(kind: string | undefined) {
  return weaponRangeForKind(kind)
}

function attackRangeOf(player = myPlayer.value) {
  return weaponRange(player?.weapon?.kind)
}

function equipTagLabel(card: YzsCard) {
  return equipDisplaySummary(card)
}


function equipTagTitle(card: YzsCard) {
  const w = weaponMetaForKind(card.kind)
  if (w) return w.effect
  const e = equipMetaForKind(card.kind)
  if (e) return e.effect
  return card.name
}

const targeting = useYzsTargeting({
  state,
  mySeat,
  opponentSeat,
  myPlayer,
  myHand,
  shaTarget,
  selectedTargetZone,
  selectedTargetCardId,
  selectedQilinZone,
  hitFlashSeat,
  blockFlashSeat,
  seatAt,
  isMyPlay,
  isFinished,
  isResponse,
  isFankui,
  isTuxiTake,
  isQixiTake,
  selectedCard,
  canPlaySha,
  cardPlaysAsSha,
  needsOpponentTarget,
  equipTagLabel,
  isKongchengProtected,
  attackRangeOf,
  fankuiSourceSeat,
  tuxiSourceSeat,
  qixiSourceSeat,
})

const {
  hasTeamMode,
  teammateSeat,
  enemySeats,
  crossSeats,
  ringDistance,
  distanceToSeat,
  takeableOptionsForPlayer,
  takeableTargetOptions,
  fankuiTargetOptions,
  tuxiTargetOptions,
  qixiTargetOptions,
  selectedCardNeedsTargetCard,
  canTargetSeat,
  canTargetOpponentWith,
  isSeatTargetable,
  seatPanelClass,
  onTargetSeat,
  onTargetOpponent,
  pickFankuiTarget,
  pickTuxiTarget,
  pickOpponentCardTarget,
  syncWeaponSkillTargeting,
} = targeting

function makePendingContext(): PendingContext | null {
  if (!state.value) return null
  return {
    state: state.value,
    loading: loading.value,
    isAnimating: isAnimating.value,
    mySeat: mySeat.value,
    opponentSeat: opponentSeat.value,
    isMyDraw: isMyDraw.value,
    isMyResponse: isMyResponse.value,
    canUsePeekDeckUI: canUsePeekDeckUI.value,
    selectedId,
    selectedTargetZone,
    selectedTargetCardId,
    selectedQilinZone,
    shaTarget,
    liuliSelectedId,
    ganglieDiscardIds,
    ddzCancelDiscardIds,
    yijiSelectedIds,
    peekDeckTopIds,
    peekDeckBottomIds,
    fankuiTargetOptions,
    tuxiTargetOptions,
    qixiTargetOptions,
    myCharacterSkills,
    peekDeckSkillId,
    yijiGiveRemaining,
    centerMessage,
    selectedCard,
    responseRequiredKind,
    canPlayWuxiek,
    cardPlaysAsSha,
    cardPlaysAsTao,
    cardPlaysAsShan,
    isRedCard,
    isBlackCard,
    act,
  }
}

watch(
  () => state.value?.pending?.response_mode,
  (mode, prevMode) => {
    const ctx = makePendingContext()
    if (!ctx) return
    pendingOnModeChange(ctx, mode, prevMode, {
      isMyPlay: isMyPlay.value,
      selectedCardNeedsTargetCard,
    })
  },
)

function canPlayCard(card: YzsCard | null | undefined) {
  if (!card) return false
  if (isMyDiscard.value) return true
  if (isMyResponse.value) {
    const ctx = makePendingContext()
    if (ctx) {
      const handled = pendingCanPlayCard(ctx, card)
      if (handled !== undefined) return handled
      if (pendingIsSkillOnly(state.value)) return false
      if (pendingSuppressPlaySubmit(state.value)) return false
    }
    if (state.value?.pending?.allow_wuxiek && card.kind === 'wuxiek') return true
    if (responseRequiredKind.value === 'sha' && cardPlaysAsSha(card)) return true
    if (responseRequiredKind.value === 'shan' && cardPlaysAsShan(card)) return true
    return card.kind === responseRequiredKind.value
  }
  if (isMyPlay.value) {
    if (rendeMode.value) {
      return rendeSelectedIds.value.length > 0 && shaTarget.value != null
    }
    if (zhihengMode.value) {
      return zhihengSelectedIds.value.length > 0
    }
    if (jieyinMode.value) {
      return jieyinSelectedIds.value.length === 2 && shaTarget.value != null
    }
    if (fanjianMode.value) {
      return fanjianSelectedId.value !== ''
    }
    if (qixiMode.value) {
      return qixiSelectedId.value !== ''
    }
    if (guoseMode.value) {
      return guoseSelectedId.value !== ''
    }
    if (shuangxiongMode.value) {
      return shuangxiongSelectedId.value !== ''
    }
    if (cardPlaysAsSha(card)) {
      return canPlaySha.value && shaTarget.value != null
    }
    if (card.kind === 'tao') {
      return (myPlayer.value?.hp ?? 0) < (myPlayer.value?.max_hp ?? 4)
    }
    if (card.kind === 'jiu') return !myPlayer.value?.drunk
    if (needsOpponentTarget(card)) {
      if (!canTargetOpponentWith(card) || shaTarget.value == null) return false
      if (selectedCardNeedsTargetCard(card)) return selectedTargetZone.value !== ''
      return true
    }
    if (needsSelfTarget(card)) return true
    return false
  }
  if (isJijiHeal.value) {
    return cardPlaysAsTao(card)
  }
  return false
}

const showActionButton = computed(
  () =>
    !isDealing.value &&
    !isFinished.value &&
    (isMyResponse.value || isMyPlay.value || isMyDiscard.value || isMyPrepare.value || isMyDraw.value || isPeekDeck.value || isJijiHeal.value),
)

const canSubmitPeekDeck = computed(() => {
  if (!isPeekDeck.value) return false
  const ctx = makePendingContext()
  if (!ctx) return false
  return pendingCanSubmitPlay(ctx) ?? false
})

const canSubmitPlay = computed(() => {
  if (loading.value || isAnimating.value) return false
  if (isPeekDeck.value) return canSubmitPeekDeck.value
  if (isMyDiscard.value) {
    return discardNeeded.value > 0 && selectedDiscardIds.value.length === discardNeeded.value
  }
  if (isMyResponse.value) {
    const ctx = makePendingContext()
    if (ctx) {
      const handled = pendingCanSubmitPlay(ctx)
      if (handled !== undefined) return handled
      if (pendingSuppressPlaySubmit(state.value)) return false
      if (pendingIsSkillOnly(state.value)) return false
    }
  }
  return canPlayCard(selectedCard.value)
})

const canSubmitEndTurn = computed(
  () => isMyPlay.value && !loading.value && !isAnimating.value,
)

const canSubmitCancel = computed(() => {
  if (!isMyResponse.value || loading.value || isAnimating.value) return false
  const ctx = makePendingContext()
  if (!ctx) return true
  return pendingAllowsCancel(ctx) ?? true
})

function pendingSkillSubmit(skillId: string) {
  const ctx = makePendingContext()
  if (!ctx) return false
  return pendingCanSubmitSkill(ctx, skillId) ?? false
}

const canSubmitGuicai = computed(() => isGuicai.value && pendingSkillSubmit('guicai'))
const canSubmitGuidao = computed(() => isGuidao.value && pendingSkillSubmit('guidao'))
const canSubmitLeiji = computed(() => isLeijiOffer.value && pendingSkillSubmit('leiji'))
const canSubmitTianxiang = computed(() => isTianxiangOffer.value && pendingSkillSubmit('tianxiang'))

const canSubmitFankui = computed(() => {
  if (!isFankui.value) return false
  const ctx = makePendingContext()
  if (!ctx) return false
  return pendingCanSubmitSkill(ctx, 'fankui') ?? false
})

const canSubmitTuxi = computed(() => {
  if (!isTuxiTake.value) return false
  const ctx = makePendingContext()
  if (!ctx) return false
  return pendingCanSubmitSkill(ctx, 'tuxi') ?? false
})

const canSubmitQixi = computed(() => isQixiTake.value && pendingSkillSubmit('qixi'))
const canSubmitLiuli = computed(() => isLiuliOffer.value && pendingSkillSubmit('liuli'))
const canSubmitYinghunDiscard = computed(
  () => isYinghunDiscard.value && pendingSkillSubmit('yinghun'),
)
const canSubmitGanglieDiscard = computed(
  () => isGanglieChoice.value && pendingSkillSubmit('ganglie'),
)
const canSubmitDdzJudgeCancel = computed(
  () =>
    isDdzJudgeCancel.value &&
    ddzCancelDiscardIds.value.length === 2 &&
    !loading.value &&
    !isAnimating.value,
)
const canSubmitYijiGive = computed(() => isYijiGive.value && pendingSkillSubmit('yiji'))

const handLayoutStyle = computed(() => {
  const count = myHand.value.length
  if (count <= 1) {
    return {
      '--hand-step': `${YZS_CARD_WIDTH}px`,
      '--hand-card-width': `${YZS_CARD_WIDTH}px`,
    }
  }
  const maxSpan = 380
  const step = Math.min(42, Math.max(26, (maxSpan - YZS_CARD_WIDTH) / (count - 1)))
  return {
    '--hand-step': `${Math.round(step)}px`,
    '--hand-card-width': `${YZS_CARD_WIDTH}px`,
  }
})

const { centerHint } = useYzsHints({
  state,
  centerMessage,
  tableActionHint,
  selectedDiscardIds,
  selectedId,
  shaTarget,
  selectedTargetZone,
  peekDeckSkillId,
  rendeMode,
  zhihengMode,
  jieyinMode,
  fanjianMode,
  qixiMode,
  wushengMode,
  isDealing,
  isFinished,
  isMyDiscard,
  isMyPrepare,
  isMyDraw,
  isPeekDeck,
  isMyPlay,
  isMyResponse,
  hasTeamMode,
  isGuanYuFollow,
  isQilinBow,
  isJijiangRespond,
  isJianxiong,
  isYijiOffer,
  isYijiGive,
  isGanglieOffer,
  isGanglieChoice,
  isDdzJudgeCancel,
  isFankui,
  isGuicai,
  isGuidao,
  isLeijiOffer,
  isFanjianSuit,
  isTianxiangOffer,
  isQixiTake,
  isYinghunChoice,
  isYinghunDiscard,
  isWuxiekOffer,
  isWuguPick,
  discardNeeded,
  activatableSkillIds,
  myCharacterSkills,
  selectedCard,
  yijiGiveRemaining,
  responseRequiredKind,
  canSubmitBagua,
  enemySeats,
  cardPlaysAsSha,
  needsOpponentTarget,
  canTargetOpponentWith,
  canPlayCard,
  selectedCardNeedsTargetCard,
  distanceToSeat,
  attackRangeOf,
  isKongchengProtected,
  seatAt,
  cardLabel,
  resolvePendingHint: () => {
    const ctx = makePendingContext()
    return ctx ? pendingHint(ctx) : null
  },
})



function opponentHandCount() {
  return seatHandCount(opponentSeat.value)
}

function seatHandCount(seat: number) {
  if (isDealing.value) return displayedDealCounts.value[seat] ?? 0
  return seatAt(seat)?.hand_count ?? 0
}


function showSeatSkillPanels(seat: number) {
  const takeableHere =
    isMyPlay.value &&
    selectedCardNeedsTargetCard() &&
    takeableTargetOptions().length > 0 &&
    (shaTarget.value === seat ||
      (!hasTeamMode.value && seat === opponentSeat.value && shaTarget.value == null))
  return (
    (isQilinBow.value && (state.value?.pending?.effect_target ?? opponentSeat.value) === seat) ||
    (isFankui.value && fankuiSourceSeat.value === seat) ||
    (isTuxiTake.value && tuxiSourceSeat.value === seat) ||
    (isQixiTake.value && qixiSourceSeat.value === seat) ||
    takeableHere
  )
}

function showSeatTimer(seat: number) {
  if (isDealing.value || isFinished.value) return false
  if (isResponse.value) {
    if (state.value?.pending?.response_mode === 'wugu_pick') {
      return state.value?.pending?.wugu_pick_seat === seat
    }
    if (state.value?.pending?.response_mode === 'peek_deck') {
      return state.value?.pending?.target_index === seat
    }
    if (state.value?.pending?.response_mode === 'skill_ganglie_choice') {
      return state.value?.pending?.target_index === seat
    }
    return state.value?.pending?.target_index === seat
  }
  if (isMyPrepare.value && seat === mySeat.value) return true
  if (isMyDiscard.value && seat === mySeat.value) return true
  return state.value?.current_turn === seat
}

const turnDeadline = computed(() => state.value?.turn_deadline_unix)
const phase = computed(() => state.value?.phase)

const phaseTimerActive = computed(
  () =>
    !isFinished.value &&
    !isDealing.value &&
    (isMyResponse.value || isMyPlay.value || isMyDiscard.value || isMyPrepare.value || isPeekDeck.value),
)

function toastError(message: string) {
  showToast(message, 'error')
}

function clearWushengMode() {
  wushengMode.value = false
}

function clearRendeMode() {
  rendeMode.value = false
  rendeSelectedIds.value = []
}

function clearZhihengMode() {
  zhihengMode.value = false
  zhihengSelectedIds.value = []
}

function clearJieyinMode() {
  jieyinMode.value = false
  jieyinSelectedIds.value = []
}

function clearFanjianMode() {
  fanjianMode.value = false
  fanjianSelectedId.value = ''
}

function clearQixiMode() {
  qixiMode.value = false
  qixiSelectedId.value = ''
}

function clearGuoseMode() {
  guoseMode.value = false
  guoseSelectedId.value = ''
}

function clearShuangxiongMode() {
  shuangxiongMode.value = false
  shuangxiongSelectedId.value = ''
}

function clearSkillSelectModes() {
  clearRendeMode()
  clearZhihengMode()
  clearJieyinMode()
  clearFanjianMode()
  clearQixiMode()
  clearGuoseMode()
  clearShuangxiongMode()
}

function clearTargeting() {
  shaTarget.value = null
  selectedTargetZone.value = ''
  selectedTargetCardId.value = ''
  selectedQilinZone.value = ''
  hitFlashSeat.value = null
  blockFlashSeat.value = null
  clearSkillSelectModes()
  clearWushengMode()
}


function syncWushengFromState() {
  wushengMode.value = (myPlayer.value?.skill_counters?.wusheng_active ?? 0) > 0
}

const animations = useYzsAnimations({
  state,
  mySeat,
  opponentSeat,
  myPlayer,
  drawAreaRef,
  handAreaRef,
  playAreaRef,
  tableWrapRef,
  isBoltFlying,
  hitFlashSeat,
  blockFlashSeat,
  isDealing,
  isAnimating,
  displayedHand,
  displayedTableCards,
  displayedDealCounts,
  enteringDrawCardIds,
  centerMessage,
  tableActionHint,
  selectedId,
  selectedDiscardIds,
  syncWeaponSkillTargeting,
  syncWushengFromState,
  clearTargeting,
})

const {
  syncDisplayFromState,
  appendDrawnCards,
  setTableCard,
  addTableCard,
  collectDiscardBatch,
  collectDrawBatch,
  replayDrawBatch,
  replayDiscardBatch,
  replayEvent,
  sleep,
  applyState,
  runInitialDealAnimation,
  flashSeatHit,
  flashSeatBlocked,
  runShaFlyBolt,
  discardActorHint,
} = animations

let timeoutInFlight = false

async function handleTimeout() {
  if (!state.value || timeoutInFlight || !phaseTimerActive.value) return

  timeoutInFlight = true
  loading.value = true
  try {
    if (isMyResponse.value) {
      await applyState(await passYuzhoushaResponse(state.value.id))
    } else if (isMyDiscard.value) {
      await applyState(await tickYuzhoushaGame(state.value.id))
    } else if (isMyPlay.value) {
      await applyState(await endYuzhoushaPlay(state.value.id))
    }
  } catch {
    // ignore
  } finally {
    loading.value = false
    timeoutInFlight = false
  }
}

const { secondsLeft } = usePhaseTimer(turnDeadline, phase, phaseTimerActive, handleTimeout)



function toggleDiscardSelection(id: string) {
  const need = discardNeeded.value
  if (need <= 0) return
  const idx = selectedDiscardIds.value.indexOf(id)
  if (idx >= 0) {
    selectedDiscardIds.value = selectedDiscardIds.value.filter((x) => x !== id)
    return
  }
  if (selectedDiscardIds.value.length >= need) return
  selectedDiscardIds.value = [...selectedDiscardIds.value, id]
}

function selectCard(id: string) {
  if (!canInteract.value) return
  const card = myHand.value.find((c) => c.id === id)
  if (!card) return

  if (isMyDiscard.value) {
    toggleDiscardSelection(id)
    return
  }

  if (rendeMode.value && isMyPlay.value) {
    if (rendeSelectedIds.value.includes(id)) {
      rendeSelectedIds.value = rendeSelectedIds.value.filter((x) => x !== id)
    } else {
      rendeSelectedIds.value = [...rendeSelectedIds.value, id]
    }
    return
  }

  if (zhihengMode.value && isMyPlay.value) {
    if (zhihengSelectedIds.value.includes(id)) {
      zhihengSelectedIds.value = zhihengSelectedIds.value.filter((x) => x !== id)
    } else {
      zhihengSelectedIds.value = [...zhihengSelectedIds.value, id]
    }
    return
  }

  if (jieyinMode.value && isMyPlay.value) {
    if (jieyinSelectedIds.value.includes(id)) {
      jieyinSelectedIds.value = jieyinSelectedIds.value.filter((x) => x !== id)
    } else if (jieyinSelectedIds.value.length < 2) {
      jieyinSelectedIds.value = [...jieyinSelectedIds.value, id]
    }
    return
  }

  if (fanjianMode.value && isMyPlay.value) {
    fanjianSelectedId.value = fanjianSelectedId.value === id ? '' : id
    return
  }

  if (qixiMode.value && isMyPlay.value) {
    if (!isBlackCard(card)) return
    qixiSelectedId.value = qixiSelectedId.value === id ? '' : id
    return
  }

  if (guoseMode.value && isMyPlay.value) {
    if (!isDiamondCard(card)) return
    guoseSelectedId.value = guoseSelectedId.value === id ? '' : id
    return
  }

  if (shuangxiongMode.value && isMyPlay.value) {
    if (!cardValidForShuangxiong(card)) return
    shuangxiongSelectedId.value = shuangxiongSelectedId.value === id ? '' : id
    return
  }

  if (isLiuliOffer.value && isMyResponse.value) {
    liuliSelectedId.value = liuliSelectedId.value === id ? '' : id
    return
  }

  if (isGanglieChoice.value) {
    if (ganglieDiscardIds.value.includes(id)) {
      ganglieDiscardIds.value = ganglieDiscardIds.value.filter((x) => x !== id)
    } else if (ganglieDiscardIds.value.length < 2) {
      ganglieDiscardIds.value = [...ganglieDiscardIds.value, id]
    }
    return
  }

  if (isDdzJudgeCancel.value) {
    if (ddzCancelDiscardIds.value.includes(id)) {
      ddzCancelDiscardIds.value = ddzCancelDiscardIds.value.filter((x) => x !== id)
    } else if (ddzCancelDiscardIds.value.length < 2) {
      ddzCancelDiscardIds.value = [...ddzCancelDiscardIds.value, id]
    }
    return
  }

  if (isYijiGive.value && isMyResponse.value) {
    if (yijiSelectedIds.value.includes(id)) {
      yijiSelectedIds.value = yijiSelectedIds.value.filter((x) => x !== id)
    } else if (yijiSelectedIds.value.length < yijiGiveRemaining.value) {
      yijiSelectedIds.value = [...yijiSelectedIds.value, id]
      if (shaTarget.value == null) {
        shaTarget.value = opponentSeat.value
      }
    }
    return
  }

  if (selectedId.value === id) {
    selectedId.value = ''
    if (needsOpponentTarget(card)) clearTargeting()
    return
  }

  selectedId.value = id
  if (!(isMyPlay.value && canTargetOpponentWith(card))) {
    clearTargeting()
  }
}


async function act(fn: () => Promise<YuzhoushaState>) {
  if (!state.value || loading.value || isDealing.value || isAnimating.value) return
  loading.value = true
  try {
    await applyState(await fn())
  } catch (err) {
    toastError(err instanceof Error ? err.message : '操作失败')
  } finally {
    loading.value = false
  }
}

async function submitCancelWusheng() {
  if (!state.value || !canCancelWusheng.value) return
  selectedId.value = ''
  clearTargeting()
  await act(() => useYuzhoushaSkill(state.value!.id, 'wusheng'))
}

async function submitSkill(skillId: string) {
  if (!state.value || loading.value) return
  if (skillId === 'rende') {
    if (rendeSelectedIds.value.length === 0 || shaTarget.value == null) return
    await act(() =>
      useYuzhoushaSkill(state.value!.id, 'rende', {
        targetIndex: shaTarget.value!,
        cardIds: [...rendeSelectedIds.value],
      }),
    )
    clearRendeMode()
    return
  }
  if (skillId === 'zhiheng') {
    if (zhihengSelectedIds.value.length === 0) return
    await act(() =>
      useYuzhoushaSkill(state.value!.id, 'zhiheng', {
        cardIds: [...zhihengSelectedIds.value],
      }),
    )
    clearZhihengMode()
    return
  }
  if (skillId === 'jieyin') {
    if (jieyinSelectedIds.value.length !== 2 || shaTarget.value == null) return
    await act(() =>
      useYuzhoushaSkill(state.value!.id, 'jieyin', {
        targetIndex: shaTarget.value!,
        cardIds: [...jieyinSelectedIds.value],
      }),
    )
    clearJieyinMode()
    return
  }
  if (skillId === 'fanjian') {
    if (fanjianSelectedId.value === '') return
    await act(() =>
      useYuzhoushaSkill(state.value!.id, 'fanjian', {
        cardIds: [fanjianSelectedId.value],
      }),
    )
    clearFanjianMode()
    return
  }
  if (skillId === 'qixi') {
    if (isQixiTake.value) {
      const ctx = makePendingContext()
      if (ctx && (await pendingSubmitSkill(ctx, 'qixi'))) return
    } else if (qixiSelectedId.value !== '') {
      await act(() =>
        useYuzhoushaSkill(state.value!.id, 'qixi', {
          cardIds: [qixiSelectedId.value],
        }),
      )
      clearQixiMode()
    }
    return
  }
  if (skillId === 'fankui') {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'fankui'))) return
    return
  }
  if (skillId === 'tuxi') {
    if (isTuxiTake.value) {
      const ctx = makePendingContext()
      if (ctx && (await pendingSubmitSkill(ctx, 'tuxi'))) return
    } else if (isMyDraw.value) {
      await act(() =>
        useYuzhoushaSkill(state.value!.id, 'tuxi', {
          targetZone: 'skip_1',
        }),
      )
    }
    return
  }
  if (skillId === 'guicai') {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'guicai'))) return
    return
  }
  if (skillId === 'guidao') {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'guidao'))) return
    return
  }
  if (skillId === 'leiji') {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'leiji'))) return
    return
  }
  if (skillId === 'ganglie' && isGanglieChoice.value) {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'ganglie'))) return
    return
  }
  if (skillId === 'ddz_judge_cancel' && isDdzJudgeCancel.value) {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'ddz_judge_cancel'))) return
    return
  }
  if (skillId === 'yiji') {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'yiji'))) return
    return
  }
  if (skillId === 'guose') {
    if (guoseSelectedId.value === '') return
    await act(() =>
      useYuzhoushaSkill(state.value!.id, 'guose', {
        targetIndex: opponentSeat.value,
        cardIds: [guoseSelectedId.value],
      }),
    )
    clearGuoseMode()
    return
  }
  if (skillId === 'shuangxiong') {
    if (isMyDraw.value) {
      await act(() => useYuzhoushaSkill(state.value!.id, 'shuangxiong'))
      return
    }
    if (shuangxiongSelectedId.value === '') return
    await act(() =>
      useYuzhoushaSkill(state.value!.id, 'shuangxiong', {
        cardIds: [shuangxiongSelectedId.value],
      }),
    )
    clearShuangxiongMode()
    return
  }
  if (skillId === 'luanwu') {
    await act(() => useYuzhoushaSkill(state.value!.id, 'luanwu'))
    return
  }
  if (skillId === 'liuli') {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'liuli'))) return
    return
  }
  if (skillId === 'tianxiang') {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'tianxiang'))) return
    return
  }
  if (skillId === 'yinghun' && isYinghunDiscard.value) {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'yinghun'))) return
    return
  }
  if (skillId === 'yinghun') {
    await act(() =>
      useYuzhoushaSkill(state.value!.id, 'yinghun', {
        targetIndex: opponentSeat.value,
      }),
    )
    return
  }
  if (skillId === 'hunzi') {
    await act(() => useYuzhoushaSkill(state.value!.id, 'hunzi'))
    return
  }
  await act(() => useYuzhoushaSkill(state.value!.id, skillId))
}

async function submitFanjianSuit(suit: string) {
  const ctx = makePendingContext()
  if (!ctx) return
  await pendingSubmitAction(ctx, `fanjian_suit:${suit}`)
}

async function submitYinghunOption(option: 'draw_both' | 'draw_two_discard') {
  const ctx = makePendingContext()
  if (!ctx) return
  const action = option === 'draw_two_discard' ? 'yinghun_draw_two_discard' : 'yinghun_draw_both'
  await pendingSubmitAction(ctx, action)
}

async function submitYinghunDiscard() {
  if (!state.value || !canSubmitYinghunDiscard.value) return
  await submitSkill('yinghun')
}

async function submitPassYijiGive() {
  const ctx = makePendingContext()
  if (!ctx) return
  await pendingSubmitAction(ctx, 'yiji_pass_give')
}

async function submitTuxiSkip(skip: 1 | 2) {
  if (!state.value || loading.value || !isMyDraw.value) return
  await act(() =>
    useYuzhoushaSkill(state.value!.id, 'tuxi', {
      targetZone: skip === 2 ? 'skip_2' : 'skip_1',
    }),
  )
}

async function submitGanglieTakeDamage() {
  const ctx = makePendingContext()
  if (!ctx) return
  await pendingSubmitAction(ctx, 'ganglie_take_damage')
}

async function submitGanglieDiscard() {
  if (!state.value || !canSubmitGanglieDiscard.value) return
  await submitSkill('ganglie')
}

function isSkillActivatable(skill: YzsSkillMeta) {
  if (skill.inactive_in_1v1) return false
  if (skill.id === 'longdan' || skill.id === 'paoxiao' || skill.id === 'kongcheng') return false
  if (skill.id === 'wusheng' && wushengMode.value && (isMyPlay.value || isMyResponse.value)) return false
  const ctx = makePendingContext()
  if (ctx && isMyResponse.value) {
    const handled = pendingCanSubmitSkill(ctx, skill.id)
    if (handled !== undefined) return handled
    if (skill.id === 'yiji' && isYijiOffer.value) return true
  }
  return activatableSkillIds.value.has(skill.id)
}

function onCharacterSkillClick(skill: YzsSkillMeta) {
  if (skill.inactive_in_1v1) return
  if (skill.id === 'rende') {
    activateSkill('rende')
    return
  }
  if (skill.id === 'zhiheng') {
    activateSkill('zhiheng')
    return
  }
  if (skill.id === 'jieyin') {
    activateSkill('jieyin')
    return
  }
  if (skill.id === 'fanjian') {
    activateSkill('fanjian')
    return
  }
  if (skill.id === 'qixi') {
    activateSkill('qixi')
    return
  }
  if (skill.id === 'guose') {
    activateSkill('guose')
    return
  }
  if (skill.id === 'shuangxiong') {
    if (isMyDraw.value) {
      void submitSkill('shuangxiong')
      return
    }
    activateSkill('shuangxiong')
    return
  }
  if (skill.id === 'hunzi' && isMyPrepare.value) {
    void submitSkill('hunzi')
    return
  }
  if (skill.id === 'fankui' && isFankui.value) {
    void submitSkill('fankui')
    return
  }
  if (skill.id === 'yiji' && (isYijiOffer.value || isYijiGive.value)) {
    void submitSkill('yiji')
    return
  }
  if (skill.id === 'tuxi' && isTuxiTake.value) {
    void submitSkill('tuxi')
    return
  }
  if (skill.id === 'qixi' && isQixiTake.value) {
    void submitSkill('qixi')
    return
  }
  if (skill.id === 'liuli' && isLiuliOffer.value) {
    void submitSkill('liuli')
    return
  }
  if (skill.id === 'guicai' && isGuicai.value) {
    void submitSkill('guicai')
    return
  }
  if (skill.id === 'guidao' && isGuidao.value) {
    void submitSkill('guidao')
    return
  }
  if (skill.id === 'leiji' && isLeijiOffer.value) {
    void submitSkill('leiji')
    return
  }
  if (skill.id === 'tianxiang' && isTianxiangOffer.value) {
    void submitSkill('tianxiang')
    return
  }
  if (activatableSkillIds.value.has(skill.id)) {
    void submitSkill(skill.id)
  }
}

function activateSkill(skillId: string) {
  if (skillId === 'rende') {
    rendeMode.value = true
    rendeSelectedIds.value = []
    selectedId.value = ''
    clearZhihengMode()
    clearJieyinMode()
    clearFanjianMode()
    clearWushengMode()
    return
  }
  if (skillId === 'zhiheng') {
    zhihengMode.value = true
    zhihengSelectedIds.value = []
    selectedId.value = ''
    clearRendeMode()
    clearJieyinMode()
    clearFanjianMode()
    clearWushengMode()
    return
  }
  if (skillId === 'jieyin') {
    jieyinMode.value = true
    jieyinSelectedIds.value = []
    shaTarget.value = opponentSeat.value
    selectedId.value = ''
    clearRendeMode()
    clearZhihengMode()
    clearFanjianMode()
    clearWushengMode()
    return
  }
  if (skillId === 'fanjian') {
    fanjianMode.value = true
    fanjianSelectedId.value = ''
    selectedId.value = ''
    clearRendeMode()
    clearZhihengMode()
    clearJieyinMode()
    clearQixiMode()
    clearWushengMode()
    return
  }
  if (skillId === 'qixi') {
    qixiMode.value = true
    qixiSelectedId.value = ''
    selectedId.value = ''
    clearRendeMode()
    clearZhihengMode()
    clearJieyinMode()
    clearFanjianMode()
    clearGuoseMode()
    clearWushengMode()
    return
  }
  if (skillId === 'guose') {
    guoseMode.value = true
    guoseSelectedId.value = ''
    selectedId.value = ''
    clearRendeMode()
    clearZhihengMode()
    clearJieyinMode()
    clearFanjianMode()
    clearQixiMode()
    clearShuangxiongMode()
    clearWushengMode()
    return
  }
  if (skillId === 'shuangxiong' && isMyPlay.value) {
    shuangxiongMode.value = true
    shuangxiongSelectedId.value = ''
    selectedId.value = ''
    clearRendeMode()
    clearZhihengMode()
    clearJieyinMode()
    clearFanjianMode()
    clearQixiMode()
    clearGuoseMode()
    clearWushengMode()
    return
  }
  void submitSkill(skillId)
}

async function submitPassPrepare() {
  if (!state.value || !isMyPrepare.value) return
  await act(() => passYuzhoushaPrepare(state.value!.id))
}

async function submitPassDraw() {
  if (!state.value || !isMyDraw.value) return
  await act(() => passYuzhoushaDraw(state.value!.id))
}

function peekDeckCard(cardId: string) {
  return state.value?.pending?.revealed_cards?.find((c) => c.id === cardId)
}

const peekDrag = ref<{ pile: 'top' | 'bottom'; index: number } | null>(null)

function onPeekDragStart(event: DragEvent, pile: 'top' | 'bottom', index: number) {
  if (!canUsePeekDeckUI.value) {
    event.preventDefault()
    return
  }
  peekDrag.value = { pile, index }
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', `${pile}:${index}`)
  }
}

function onPeekDragOver(event: DragEvent) {
  event.preventDefault()
  if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
}

function onPeekDrop(toPile: 'top' | 'bottom', toIndex: number) {
  const drag = peekDrag.value
  if (!drag || !canUsePeekDeckUI.value) return

  const fromRef = drag.pile === 'top' ? peekDeckTopIds : peekDeckBottomIds
  const toRef = toPile === 'top' ? peekDeckTopIds : peekDeckBottomIds

  if (drag.pile === toPile) {
    const arr = [...fromRef.value]
    if (drag.index < 0 || drag.index >= arr.length) return
    const [item] = arr.splice(drag.index, 1)
    let insertAt = drag.index < toIndex ? toIndex - 1 : toIndex
    insertAt = Math.max(0, Math.min(insertAt, arr.length))
    arr.splice(insertAt, 0, item)
    fromRef.value = arr
  } else {
    const fromArr = [...fromRef.value]
    if (drag.index < 0 || drag.index >= fromArr.length) return
    const [item] = fromArr.splice(drag.index, 1)
    fromRef.value = fromArr
    const toArr = [...toRef.value]
    toArr.splice(Math.min(toIndex, toArr.length), 0, item)
    toRef.value = toArr
  }
  peekDrag.value = null
}

function onPeekZoneDrop(toPile: 'top' | 'bottom') {
  const list = toPile === 'top' ? peekDeckTopIds : peekDeckBottomIds
  onPeekDrop(toPile, list.value.length)
}

function onPeekDragEnd() {
  peekDrag.value = null
}

async function submitPeekDeck() {
  if (!state.value || !canSubmitPeekDeck.value) return
  const ctx = makePendingContext()
  if (ctx) await pendingSubmitPlay(ctx)
}

async function submitPlayCard() {
  if (!state.value || !canSubmitPlay.value) return

  if (isPeekDeck.value) {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitPlay(ctx))) return
    return
  }

  if (isMyDiscard.value) {
    await act(() => discardYuzhoushaCards(state.value!.id, [...selectedDiscardIds.value]))
    selectedDiscardIds.value = []
    return
  }

  if (isMyResponse.value) {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitPlay(ctx))) return

    const card = selectedCard.value
    if (!card) return
    if (card.kind === 'wuxiek' && canPlayWuxiek.value) {
      await act(() => respondYuzhoushaCard(state.value!.id, card.id))
      return
    }
    if (card.kind !== responseRequiredKind.value && !cardPlaysAsSha(card) && !cardPlaysAsShan(card)) {
      return
    }
    if (responseRequiredKind.value === 'shan' && !cardPlaysAsShan(card)) {
      return
    }
    if (responseRequiredKind.value === 'sha' && !cardPlaysAsSha(card)) {
      return
    }
    await act(() => respondYuzhoushaCard(state.value!.id, card.id))
    return
  }

  if (isJijiHeal.value) {
    const card = selectedCard.value
    if (!card) return
    await act(() => playYuzhoushaCard(state.value!.id, card.id, mySeat.value))
    return
  }

  if (isMyPlay.value) {
    if (rendeMode.value && rendeSelectedIds.value.length > 0 && shaTarget.value != null) {
      await submitSkill('rende')
      return
    }
    if (zhihengMode.value && zhihengSelectedIds.value.length > 0) {
      await submitSkill('zhiheng')
      return
    }
    if (jieyinMode.value && jieyinSelectedIds.value.length === 2 && shaTarget.value != null) {
      await submitSkill('jieyin')
      return
    }
    if (fanjianMode.value && fanjianSelectedId.value !== '') {
      await submitSkill('fanjian')
      return
    }
    if (qixiMode.value && qixiSelectedId.value !== '') {
      await submitSkill('qixi')
      return
    }
    if (guoseMode.value && guoseSelectedId.value !== '') {
      await submitSkill('guose')
      return
    }
    if (shuangxiongMode.value && shuangxiongSelectedId.value !== '') {
      await submitSkill('shuangxiong')
      return
    }
    const card = selectedCard.value
    if (!card) return

    if (cardPlaysAsSha(card) && shaTarget.value != null) {
      await act(() => playYuzhoushaCard(state.value!.id, card.id, shaTarget.value!))
      clearWushengMode()
      return
    }

    if (card.kind === 'tao') {
      await act(() => playYuzhoushaCard(state.value!.id, card.id, mySeat.value))
      return
    }

    if (needsOpponentTarget(card) && shaTarget.value != null) {
      await act(() =>
        playYuzhoushaCard(state.value!.id, card.id, {
          targetIndex: shaTarget.value!,
          targetZone: selectedTargetZone.value || undefined,
          targetCardId: selectedTargetCardId.value || undefined,
        }),
      )
      return
    }

    if (needsSelfTarget(card)) {
      await act(() => playYuzhoushaCard(state.value!.id, card.id, mySeat.value))
    }
  }
}

async function submitEndTurn() {
  if (!state.value || !canSubmitEndTurn.value) return
  clearTargeting()
  selectedId.value = ''
  await act(() => endYuzhoushaPlay(state.value!.id))
}

async function submitCancelResponse() {
  if (!state.value || !canSubmitCancel.value) return
  selectedId.value = ''
  await act(() => passYuzhoushaResponse(state.value!.id))
}

async function submitBaguaJudge() {
  if (!state.value || !canSubmitBagua.value) return
  selectedId.value = ''
  await act(() => baguaYuzhoushaJudge(state.value!.id))
}

async function loadGame(gameId: string) {
  loading.value = true
  try {
    const next = await getYuzhoushaState(gameId)
    await runInitialDealAnimation(next)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '加载对局失败')
    await router.replace('/games/yuzhousha/solo/pick')
  } finally {
    loading.value = false
  }
}

async function restart() {
  selectedId.value = ''
  clearTargeting()
  const mode = state.value?.mode
  const pickPath =
    mode && mode !== '1v1'
      ? { path: '/games/yuzhousha/solo/pick', query: { mode } }
      : '/games/yuzhousha/solo/pick'
  await router.push(pickPath)
}

onMounted(() => {
  const gameId = route.params.gameId as string
  if (!gameId) {
    void router.replace('/games/yuzhousha/solo/pick')
    return
  }
  void loadGame(gameId)
})

  const api = {
    act,
    activatableSkillIds,
    activatableSkills,
    activateSkill,
    addTableCard,
    appendDrawnCards,
    applyState,
    attackRangeOf,
    blockFlashSeat,
    canCancelWusheng,
    canInteract,
    canPlayCard,
    canPlaySha,
    canPlayWuxiek,
    canSubmitBagua,
    canSubmitCancel,
    canSubmitEndTurn,
    canSubmitFankui,
    canSubmitGanglieDiscard,
    canSubmitDdzJudgeCancel,
    canSubmitGuicai,
    canSubmitGuidao,
    canSubmitLeiji,
    canSubmitLiuli,
    canSubmitPeekDeck,
    canSubmitPlay,
    canSubmitQixi,
    canSubmitTianxiang,
    canSubmitTuxi,
    canSubmitYijiGive,
    canSubmitYinghunDiscard,
    canTargetOpponentWith,
    canTargetSeat,
    canUsePeekDeckUI,
    cardLabel,
    cardPlaysAsSha,
    cardPlaysAsShan,
    cardPlaysAsTao,
    cardValidForShuangxiong,
    centerHint,
    centerMessage,
    clearFanjianMode,
    clearGuoseMode,
    clearJieyinMode,
    clearQixiMode,
    clearRendeMode,
    clearShuangxiongMode,
    clearSkillSelectModes,
    clearTargeting,
    clearWushengMode,
    clearZhihengMode,
    collectDiscardBatch,
    collectDrawBatch,
    crossSeats,
    discardActorHint,
    discardNeeded,
    displayedDealCounts,
    displayedHand,
    displayedTableCards,
    distanceToSeat,
    drawAreaRef,
    enemySeats,
    enteringDrawCardIds,
    equipSlotOf,
    equipTagLabel,
    equipTagTitle,
    equippedCards,
    fanjianMode,
    fanjianSelectedId,
    fankuiSourceSeat,
    fankuiTargetOptions,
    flashSeatBlocked,
    flashSeatHit,
    ganglieDiscardIds,
    ddzCancelDiscardIds,
    guoseMode,
    guoseSelectedId,
    handAreaRef,
    handLayoutStyle,
    handleTimeout,
    hasMySkill,
    hitFlashSeat,
    hasTeamMode,
    isAnimating,
    isBlackCard,
    isBoltFlying,
    isDealing,
    isDiamondCard,
    isDyingRescue,
    isFanjianSuit,
    isFankui,
    isFinished,
    isGanglieChoice,
    isDdzJudgeCancel,
    isGanglieOffer,
    isGuanYuFollow,
    isGuicai,
    isGuidao,
    isJianxiong,
    isJijiHeal,
    isJijiangRespond,
    isKongchengProtected,
    isLeijiOffer,
    isLiuliOffer,
    isLuanwu,
    isMyDiscard,
    isMyDraw,
    isMyPlay,
    isMyPrepare,
    isMyResponse,
    isMyTurn,
    isPeekDeck,
    isQilinBow,
    isQixiTake,
    isRedCard,
    isResponse,
    isSeatTargetable,
    isSkillActivatable,
    isSkillOnlyResponse,
    isTianxiangOffer,
    isTuxiTake,
    isWuguPick,
    isWuxiekOffer,
    isYijiGive,
    isYijiOffer,
    isYinghunChoice,
    isYinghunDiscard,
    jieyinMode,
    jieyinSelectedIds,
    judgeAreaCards,
    liuliSelectedId,
    loadGame,
    loading,
    myCharacterSkills,
    myHand,
    myPlayer,
    mySeat,
    needsOpponentTarget,
    needsSelfTarget,
    onCharacterSkillClick,
    onPeekDragEnd,
    onPeekDragOver,
    onPeekDragStart,
    onPeekDrop,
    onPeekZoneDrop,
    onTargetOpponent,
    onTargetSeat,
    opponent,
    opponentHandCount,
    opponentHasKongcheng,
    opponentSeat,
    peekDeckBottomIds,
    peekDeckCard,
    peekDeckSkillId,
    peekDeckTopIds,
    peekDrag,
    phase,
    phaseTimerActive,
    pickFankuiTarget,
    pickOpponentCardTarget,
    pickTuxiTarget,
    playAreaRef,
    qilinHorseOptions,
    qixiMode,
    qixiSelectedId,
    qixiSourceSeat,
    qixiTargetOptions,
    removeJudgeCardFromPlayer,
    removeKnownCardFromPlayer,
    rendeMode,
    rendeSelectedIds,
    replayDiscardBatch,
    replayDrawBatch,
    replayEvent,
    responseRequiredKind,
    restart,
    ringDistance,
    route,
    router,
    runInitialDealAnimation,
    runShaFlyBolt,
    seatAt,
    seatHandCount,
    seatPanelClass,
    secondsLeft,
    selectCard,
    selectedCard,
    selectedCardNeedsTargetCard,
    selectedDiscardIds,
    selectedId,
    selectedQilinZone,
    selectedTargetCardId,
    selectedTargetZone,
    setTableCard,
    shaTarget,
    showActionButton,
    showSeatSkillPanels,
    showSeatTimer,
    shuangxiongActive,
    shuangxiongMode,
    shuangxiongRefIsRed,
    shuangxiongSelectedId,
    sleep,
    state,
    submitBaguaJudge,
    submitCancelResponse,
    submitCancelWusheng,
    submitEndTurn,
    submitFanjianSuit,
    submitGanglieDiscard,
    submitGanglieTakeDamage,
    submitPassDraw,
    submitPassPrepare,
    submitPassYijiGive,
    submitPeekDeck,
    submitPlayCard,
    submitSkill,
    submitTuxiSkip,
    submitYinghunDiscard,
    submitYinghunOption,
    syncDisplayFromState,
    syncWeaponSkillTargeting,
    syncWushengFromState,
    tableActionHint,
    tableWrapRef,
    takeableOptionsForPlayer,
    takeableTargetOptions,
    teammateSeat,
    toastError,
    toggleDiscardSelection,
    trickStaysInJudge,
    turnDeadline,
    tuxiSourceSeat,
    tuxiTargetOptions,
    weaponRange,
    wushengMode,
    wushengSkillHint,
    yijiGiveRemaining,
    yijiSelectedIds,
    zhihengMode,
    zhihengSelectedIds,
  }
  return api
}