import {
  computed,
  inject,
  onMounted,
  onUnmounted,
  provide,
  ref,
  watch,
  type InjectionKey,
} from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePhaseTimer } from '../usePhaseTimer'
import { useYuzhoushaGameSocket } from '../useYuzhoushaSocket'
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
import { skillBlockedInMode } from '../../constants/yzsModes'
import {
  discardYuzhoushaCards,
  endYuzhoushaPlay,
  getYuzhoushaState,
  passYuzhoushaPrepare,
  passYuzhoushaDraw,
  passYuzhoushaResponse,
  passAllWuxiek,
  baguaYuzhoushaJudge,
  playYuzhoushaCard,
  respondYuzhoushaCard,
  respondZhangbaSha,
  useYuzhoushaSkill,
  tickYuzhoushaGame,
} from '../../api/games'
import {
  YZS_CARD_LABELS,
  type GameLogEntry,
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
import { isMyPendingActor } from './pending/helpers'

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
const isOnline = computed(() => {
  const raw = route.query.room
  return typeof raw === 'string' && raw.length > 0
})
const roomId = computed(() => {
  const raw = route.query.room
  return typeof raw === 'string' && raw ? raw : ''
})
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

/** 持久化的游戏日志（右侧面板） */
const gameLog = ref<GameLogEntry[]>([])
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
  if (!players?.length) return 0
  if (state.value?.mode === '3p_chain' && players.length === 3) {
    return (me - 1 + 3) % 3
  }
  // 多队伍模式（2v2, 3p_ddz）：找第一个存活的敌人
  if (players.length > 2) {
    const myTeam = players[me]?.team
    if (myTeam != null) {
      for (const p of players) {
        if (p.index !== me && p.team !== myTeam && p.hp > 0) return p.index
      }
    }
  }
  return 1 - me
})
const myPlayer = computed(() => state.value?.players[mySeat.value])
const opponent = computed(() => state.value?.players[opponentSeat.value])
const seatAt = (seat: number) => state.value?.players[seat]
const isFinished = computed(() => state.value?.phase === 'finished')
const isResponse = computed(() => state.value?.phase === 'response')
const isMyTurn = computed(() => state.value?.current_turn === mySeat.value)
const isMyResponse = computed(() =>
  isMyPendingActor(state.value, mySeat.value),
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
    state.value?.pending?.actor_seat === mySeat.value,
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

const canInteract = computed(() => {
  // AOE/桃园：非目标玩家可以出无懈介入
  const isWuxiekPhase = state.value?.phase === 'response' && state.value?.pending?.response_mode === 'wuxiek_trick'
  const aoeAux = state.value?.phase === 'response' && state.value?.pending?.allow_wuxiek === true && 
    state.value?.pending?.response_mode !== 'wuxiek_trick' && hasWuxiekInHand.value
  const result = !loading.value && !isDealing.value && !isAnimating.value && !isFinished.value &&
    (isMyResponse.value || isMyPlay.value || isMyDiscard.value || isMyPrepare.value || isMyDraw.value || isPeekDeck.value || isJijiHeal.value || aoeAux || (isWuxiekPhase && hasWuxiekInHand.value))
  return result
})

const peekDeckTopIds = ref<string[]>([])
const peekDeckBottomIds = ref<string[]>([])


function opponentHasKongcheng(player = opponent.value) {
  return player?.character?.skill_ids?.includes('kongcheng') ?? false
}

function isKongchengProtected(player = opponent.value) {
  return opponentHasKongcheng(player) && (player?.hand_count ?? 0) === 0
}

const myHand = computed(() => displayedHand.value)
const hasWuxiekInHand = computed(() => myHand.value.some((c) => c.kind === 'wuxiek'))

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
  if (card.kind === 'sha' || card.kind === 'sha_fire' || card.kind === 'sha_thunder') return true
  // 优先使用 ViewAs 统一判断
  if (viewAsSkills.value.length > 0) return cardPlaysAsViaViewAs(card, 'sha')
  // fallback 旧逻辑
  if (hasMySkill('longdan') && card.kind === 'shan') return true
  if (hasMySkill('wusheng') && isRedCard(card)) {
    if (isMyPlay.value && !isMyResponse.value) return wushengMode.value
    return true
  }
  if (hasMySkill('longhun') && card.suit === 'D') return true
  return false
}

function cardPlaysAsShan(card: YzsCard | null | undefined) {
  if (!card) return false
  if (card.kind === 'shan') return true
  if (viewAsSkills.value.length > 0) return cardPlaysAsViaViewAs(card, 'shan')
  if (hasMySkill('longdan') && card.kind === 'sha') return true
  if (hasMySkill('qingguo') && isBlackCard(card)) return true
  if (hasMySkill('longhun') && card.suit === 'C') return true
  return false
}

function cardPlaysAsTao(card: YzsCard | null | undefined) {
  if (!card) return false
  if (card.kind === 'tao') return true
  if (card.kind === 'jiu' && isDyingRescue.value && state.value?.pending?.target_index === mySeat.value) return true
  if (viewAsSkills.value.length > 0) return cardPlaysAsViaViewAs(card, 'tao')
  if (hasMySkill('jiji') && isRedCard(card)) {
    if (isJijiHeal.value || isDyingRescue.value) return state.value?.current_turn !== mySeat.value
  }
  if (hasMySkill('longhun') && card.suit === 'H') return true
  return false
}

function cardPlaysAsWuxiek(card: YzsCard | null | undefined) {
  if (!card) return false
  if (card.kind === 'wuxiek') return true
  if (viewAsSkills.value.length > 0) return cardPlaysAsViaViewAs(card, 'wuxiek')
  if (hasMySkill('longhun') && card.suit === 'S') return true
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
/** 五谷丰登已选牌追踪：cardId -> 选牌者名字 */
const wuguPickedCards = ref<Record<string, string>>({})
/** 五谷丰登框是否可见（含延迟消失） */
const isWuguBoardVisible = ref(false)

/** 五谷丰登是否正在流程中（不含延迟） */
const isWuguActive = computed(
  () => isResponse.value && (state.value?.pending?.response_mode === 'wugu_pick' || state.value?.pending?.response_mode === 'wuxiek_trick') && (state.value?.pending?.revealed_cards?.length ?? 0) > 0,
)

let wuguHideTimer: ReturnType<typeof setTimeout> | null = null

/** 五谷丰登初始亮牌列表（pending 清空后仍然保留用于展示） */
const wuguRevealedAllCache = ref<YzsCard[]>([])

// 监听 state 变化追踪五谷选牌
watch(
  () => [state.value?.pending?.response_mode, state.value?.events?.length ?? 0] as const,
  ([newMode, _len], [oldMode]) => {
    const s = state.value
    if (!s) return

    // 新五谷开始（wuxiek_trick 或 wugu_pick）：清空记录，延迟显示框（等飞线动画）
    if ((newMode === 'wuxiek_trick' || newMode === 'wugu_pick') && 
        oldMode !== 'wuxiek_trick' && oldMode !== 'wugu_pick') {
      if (wuguHideTimer) { clearTimeout(wuguHideTimer); wuguHideTimer = null }
      wuguPickedCards.value = {}
      // 缓存初始亮牌列表
      if (s.pending?.wugu_revealed_all?.length) {
        wuguRevealedAllCache.value = [...s.pending.wugu_revealed_all]
      }
      // 延迟 450ms 等飞线动画完成后才亮出选牌框
      setTimeout(() => {
        isWuguBoardVisible.value = true
      }, 450)
    }

    // 追踪 wugu_pick 事件（必须在结束检测之前，确保最后选牌也被记录）
    const events = s.events ?? []
    for (const e of events) {
      if (e.type === 'wugu_pick' && e.card?.id && e.player_index != null) {
        const pickerName = s.players.find((p) => p.index === e.player_index)?.name ?? `玩家${e.player_index}`
        wuguPickedCards.value = { ...wuguPickedCards.value, [e.card.id]: pickerName }
      }
    }

    // 五谷结束：延迟1秒后隐藏框，确保最后选牌结果可见
    if ((oldMode === 'wugu_pick' || oldMode === 'wuxiek_trick') && newMode !== 'wugu_pick' && newMode !== 'wuxiek_trick') {
      if (wuguHideTimer) clearTimeout(wuguHideTimer)
      wuguHideTimer = setTimeout(() => {
        wuguPickedCards.value = {}
        wuguRevealedAllCache.value = []
        isWuguBoardVisible.value = false
      }, 1000)
    }
  },
)

const selectedQilinZone = ref('')
const rendeMode = ref(false)
const rendeSelectedIds = ref<string[]>([])
const zhihengMode = ref(false)
const zhihengSelectedIds = ref<string[]>([])
const jieyinMode = ref(false)
const jieyinSelectedIds = ref<string[]>([])
const fanjianMode = ref(false)
// [已废弃] 变牌旧模式，已被 activeViewAs 统一接管
// const zhangbaMode = ref(false)
// const zhangbaSelectedIds = ref<string[]>([])
// ===== ViewAs 统一变牌系统 =====
const viewAsSkills = computed(() => state.value?.view_as_skills ?? [])

/** 当前激活的变牌技能（统一状态，替换 wushengMode/qixiMode/guoseMode/shuangxiongMode） */
const activeViewAs = ref<{
  skillId: string
  asKind: string
  selectCount: number
  selectedIds: string[]
} | null>(null)

const viewAsSkillHint = computed(() => {
  const av = activeViewAs.value
  if (!av) return ''
  const vas = viewAsSkills.value.find(v => v.skill_id === av.skillId)
  if (!vas) return ''
  return `【${vas.skill_name}】已发动：${vas.prompt}。已选 ${av.selectedIds.length}/${av.selectCount} 张`
})

function isCardSelectableForViewAs(card: YzsCard): boolean {
  const av = activeViewAs.value
  if (!av) return false
  const vas = viewAsSkills.value.find(v => v.skill_id === av.skillId)
  if (!vas) return false
  // 用声明式过滤条件判断
  if (vas.filter_kinds && vas.filter_kinds.length > 0) {
    if (!vas.filter_kinds.includes(card.kind)) return false
  }
  if (vas.filter_suits && vas.filter_suits.length > 0) {
    if (!vas.filter_suits.includes(card.suit)) return false
  }
  if (vas.filter_suit_color === 'red' && !isRedCard(card)) return false
  if (vas.filter_suit_color === 'black' && !isBlackCard(card)) return false
  return true
}

/** 统一变牌判断：用后端 view_as_skills 的声明式过滤条件判断，不硬编码任何技能名 */
function cardPlaysAsViaViewAs(card: YzsCard | null | undefined, asKind: string): boolean {
  if (!card) return false
  if (card.kind === asKind) return true
  if (asKind === 'sha' && (card.kind === 'sha_fire' || card.kind === 'sha_thunder')) return true
  for (const vas of viewAsSkills.value) {
    if (vas.as_kind !== asKind) continue
    if (!vas.passive && !vas.is_active) continue
    if (vas.filter_kinds && vas.filter_kinds.length > 0) {
      if (!vas.filter_kinds.includes(card.kind)) continue
    }
    if (vas.filter_suits && vas.filter_suits.length > 0) {
      if (!vas.filter_suits.includes(card.suit)) continue
    }
    if (vas.filter_suit_color === 'red' && !isRedCard(card)) continue
    if (vas.filter_suit_color === 'black' && !isBlackCard(card)) continue
    return true
  }
  return false
}
// ===== ViewAs 结束 =====

// 借刀杀人：双目标选择模式（被借刀者 + 出杀目标）
const jiedaoMode = ref(false)
const jiedaoWeaponHolder = ref<number | null>(null)  // 被借刀者
const jiedaoShaTarget = ref<number | null>(null)      // 出杀目标
// 方天画戟：多目标杀（最后一张手牌出杀，可选1-3个目标）
const fangtianMode = ref(false)
const fangtianTargets = ref<number[]>([])
// 铁索连环：多目标选择模式（1-2 个目标，或重铸）
const tiesuoMode = ref(false)
const tiesuoTargets = ref<number[]>([])
const fanjianSelectedId = ref('')
// [已废弃] 变牌旧模式，已被 activeViewAs 统一接管
// const qixiMode = ref(false)
// const qixiSelectedId = ref('')
// const guoseMode = ref(false)
// const shuangxiongMode = ref(false)
// const shuangxiongSelectedId = ref('')
// const guoseSelectedId = ref('')
const guoseTarget = ref(-1)
const liuliSelectedId = ref('')
// [已废弃] 变牌旧模式，已被 activeViewAs 统一接管
// const wushengMode = ref(false)
const ganglieDiscardIds = ref<string[]>([])
const ddzCancelDiscardIds = ref<string[]>([])
const yijiSelectedIds = ref<string[]>([])

const activatableSkills = computed(() => state.value?.activatable_skills ?? [])

const myCharacterSkills = computed(() => {
  const skills = [...(myPlayer.value?.character?.skills ?? [])]
  // 丈八蛇矛：装备时始终显示，手牌不够时按钮不可用
  if (myPlayer.value?.weapon?.kind === 'weapon_10') {
    skills.push({ id: 'zhangba', name: '丈八蛇矛', description: '将两张手牌当杀使用' })
  }
  return skills
})

const activatableSkillIds = computed(
  () => new Set(activatableSkills.value.map((s) => s.id)),
)

const wushengSkillHint = computed(() => {
  if (!wushengMode.value) return ''
  return '【武圣】已发动：可选红色牌当【杀】。点上方「取消武圣」恢复正常出牌'
})

const zhangbaSkillHint = computed(() => {
  if (!zhangbaMode.value) return ''
  return '【丈八蛇矛】已发动：选2张手牌当【杀】。点上方「取消丈八」恢复正常出牌'
})

const canCancelWusheng = computed(
  () => wushengMode.value && (isMyPlay.value || isMyResponse.value) && canInteract.value,
)

const canCancelZhangba = computed(
  () => zhangbaMode.value && isMyPlay.value && canInteract.value,
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
const isChixiong = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'weapon_8',
)
const canSubmitChixiong = computed(() => isChixiong.value && selectedId.value !== '')

async function submitChixiongDiscard() {
  if (!state.value || !selectedId.value) return
  await act(() => respondYuzhoushaCard(state.value!.id, selectedId.value))
}

const isGuanshifu = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'weapon_9',
)
const guanshifuDiscardIds = ref<string[]>([])
const canSubmitGuanshifu = computed(() => isGuanshifu.value && guanshifuDiscardIds.value.length === 2)

function toggleGuanshifuCard(cardId: string) {
  const idx = guanshifuDiscardIds.value.indexOf(cardId)
  if (idx >= 0) {
    guanshifuDiscardIds.value.splice(idx, 1)
  } else if (guanshifuDiscardIds.value.length < 2) {
    guanshifuDiscardIds.value = [...guanshifuDiscardIds.value, cardId]
  }
}

async function submitGuanshifuDiscard() {
  if (!state.value || guanshifuDiscardIds.value.length !== 2) return
  await act(() => discardYuzhoushaCards(state.value!.id, [...guanshifuDiscardIds.value]))
  guanshifuDiscardIds.value = []
}

async function submitGuanshifuSkip() {
  if (!state.value) return
  guanshifuDiscardIds.value = []
  await act(() => passYuzhoushaResponse(state.value!.id))
}

async function submitChixiongSkip() {
  if (!state.value) return
  selectedId.value = ''
  await act(() => passYuzhoushaResponse(state.value!.id))
}
const isTuxiTake = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_tuxi',
)
const isQixiTake = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_qixi',
)
const isPojun = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_pojun',
)
const isYinghunChoice = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_yinghun',
)
const isYinghunDiscard = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'skill_yinghun_discard',
)
const isGuicai = computed(
  () => isResponse.value && (state.value?.pending?.response_mode === 'skill_guicai' || state.value?.pending?.response_mode === 'skill_guicai_guidao'),
)
const isGuidao = computed(
  () => isResponse.value && (state.value?.pending?.response_mode === 'skill_guidao' || state.value?.pending?.response_mode === 'skill_guicai_guidao'),
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
const isGuoHeTake = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'guohe',
)
const isTanNangTake = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'tannang',
)
/** TakeWindow 是否激活（不含延迟） */
const isTakeWindowRaw = computed(
  () => isGuoHeTake.value || isTanNangTake.value,
)
/** TakeWindow 延迟显示（等飞线+无懈动画） */
const isTakeWindowVisible = ref(false)
let takeWindowDelayTimer: ReturnType<typeof setTimeout> | null = null
watch(isTakeWindowRaw, (val) => {
  if (takeWindowDelayTimer) { clearTimeout(takeWindowDelayTimer); takeWindowDelayTimer = null }
  if (val) {
    // 等待动画完成后再弹出，给用户看清飞线的反应时间
    takeWindowDelayTimer = setTimeout(() => { isTakeWindowVisible.value = true }, 800)
  } else {
    isTakeWindowVisible.value = false
  }
})
const isTakeWindow = computed(() => isTakeWindowVisible.value)
const takeWindowTargetOptions = computed(() => {
  if (!isTakeWindow.value) return []
  const takenSeat = state.value?.pending?.subject_seat ?? -1
  if (takenSeat < 0) return []
  return takeableOptionsForPlayer(takenSeat)
})
const isSkillOnlyResponse = computed(() => pendingIsSkillOnly(state.value))

const qilinHorseOptions = computed(() => {
  if (!isQilinBow.value) return []
  const target = state.value?.pending?.effect_target ?? opponentSeat.value
  const player = state.value?.players[target]
  const options: { zone: string; label: string }[] = []
  if (player?.plus_horse) options.push({ zone: 'plus_horse', label: player.plus_horse.name })
  if (player?.minus_horse) options.push({ zone: 'minus_horse', label: player.minus_horse.name })
  return options
})

const selectedCard = computed(() => {
  const c = myHand.value.find((c) => c.id === selectedId.value)
  if (c) return c
  // ViewAs 变牌模式下装备区也可被选中
  if (activeViewAs.value) {
    const av = activeViewAs.value
    const vas = viewAsSkills.value.find(v => v.skill_id === av.skillId)
    if (vas && (vas.position === 'he' || vas.position === 'e')) {
      return equippedCards(myPlayer.value).find((c) => c.id === selectedId.value) ?? null
    }
  }
  // [已废弃] 旧兼容，已被 activeViewAs 替代
  // if (wushengMode.value || qixiMode.value || guoseMode.value || shuangxiongMode.value) {
  //   return equippedCards(myPlayer.value).find((c) => c.id === selectedId.value) ?? null
  // }
  return state.value?.pending?.revealed_cards?.find((c) => c.id === selectedId.value) ?? null
})
const selfTargetKinds = new Set([
  'tao',
  'taoyuan',
  'wuzhong',
  'wugu',
  'shandian',
  'nanman',
  'wanjian',
  'jiu',
  'tiesuo',
  'weapon_1',
  'weapon_2',
  'weapon_3',
  'weapon_4',
  'weapon_5',
  'weapon_6',
  'weapon_7',
  'weapon_8',
  'weapon_9',
  'armor',
  'armor_vine',
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
/** 仅群体锦囊的无懈窗口（南蛮/万箭/桃园/五谷），才显示"本轮都不出" */
const isAoeWuxiekOffer = computed(
  () => isResponse.value && state.value?.pending?.response_mode === 'wuxiek_trick',
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
    myPlayer.value.armor.kind === 'armor' &&
    !state.value?.pending?.bagua_used &&
    !state.value?.pending?.ignore_armor &&
    !loading.value &&
    !isAnimating.value,
)

function cardLabel(kind: string | undefined) {
  if (!kind) return '牌'
  return YZS_CARD_LABELS[kind] ?? kind
}

const opponentTargetKinds = new Set(['sha', 'guohe', 'tannang', 'juedou', 'lebu', 'bingliang', 'huogong', 'tiesuo', 'jiedao'])

function needsOpponentTarget(card: YzsCard | null | undefined) {
  if (!card) return false
  if (cardPlaysAsSha(card)) return true
  // ViewAs 变牌模式：检查是否需要对手目标
  if (activeViewAs.value) {
    const av = activeViewAs.value
    const vas = viewAsSkills.value.find(v => v.skill_id === av.skillId)
    if (vas) {
      // 杀/过河拆桥/顺手牵羊/决斗/乐不思蜀/兵粮寸断 需要对对手
      if (['sha', 'guohe', 'tannang', 'juedou', 'lebu', 'bingliang'].includes(vas.as_kind)) return true
    }
  }
  // [已废弃] 旧兼容
  // if (qixiMode.value && isBlackCard(card)) return true
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
  isPojun,
  selectedCard,
  canPlaySha,
  cardPlaysAsSha,
  needsOpponentTarget,
  equipTagLabel,
  isKongchengProtected,
  attackRangeOf,
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
  pojunTargetOptions,
  selectedCardNeedsTargetCard,
  canTargetSeat,
  canTargetOpponentWith,
  isSeatTargetable,
  seatPanelClass,
  onTargetSeat,
  onTargetOpponent,
  pickFankuiTarget,
  pickTuxiTarget,
  pickPojunTarget,
  pickOpponentCardTarget,
  syncWeaponSkillTargeting,
} = targeting

// 包装 onTargetSeat：铁索连环模式下走多目标选择
const _origOnTargetSeat = onTargetSeat
function handleSeatTarget(seat: number) {
  if (tiesuoMode.value && isMyPlay.value) {
    toggleTiesuoTarget(seat)
    return
  }
  if (jiedaoMode.value && isMyPlay.value) {
    toggleJiedaoTarget(seat)
    return
  }
  if (fangtianMode.value && isMyPlay.value) {
    toggleFangtianTarget(seat)
    return
  }
  _origOnTargetSeat(seat)
}

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
    selectedDiscardIds,
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
    pojunTargetOptions,
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
    // [已废弃] 丈八旧模式，已被 activeViewAs 接管
    // if (zhangbaMode.value) { return zhangbaSelectedIds.value.length === 2 && shaTarget.value != null }
    // 借刀杀人：必须选完两个目标
    if (jiedaoMode.value) {
      return jiedaoWeaponHolder.value !== null && jiedaoShaTarget.value !== null
    }
    // 方天画戟：选1-3个目标
    if (fangtianMode.value) {
      return fangtianTargets.value.length >= 1 && fangtianTargets.value.length <= 3
    }
    // ViewAs 变牌模式：检查选牌是否满足要求
    if (activeViewAs.value) {
      const av = activeViewAs.value
      if (av.selectCount > 1) {
        // 多牌模式（丈八）：需要选满
        return av.selectedIds.length === av.selectCount && shaTarget.value != null
      }
      // 单牌模式（武圣/奇袭/国色/双雄）：需要选牌 + 目标
      const card = selectedCard.value
      if (!card) return false
      // 用声明式过滤条件判断
      const vas = viewAsSkills.value.find(v => v.skill_id === av.skillId)
      if (!vas) return false
      if (vas.filter_kinds && vas.filter_kinds.length > 0) {
        if (!vas.filter_kinds.includes(card.kind)) return false
      }
      if (vas.filter_suits && vas.filter_suits.length > 0) {
        if (!vas.filter_suits.includes(card.suit)) return false
      }
      if (vas.filter_suit_color === 'red' && !isRedCard(card)) return false
      if (vas.filter_suit_color === 'black' && !isBlackCard(card)) return false
      // 需要对对手目标的牌型
      if (['sha', 'guohe', 'tannang', 'juedou', 'lebu', 'bingliang'].includes(vas.as_kind)) {
        return shaTarget.value != null
      }
      return true
    }
    // [已废弃] 旧变牌分支，已被 activeViewAs 接管
    // if (qixiMode.value) { ... }
    // if (guoseMode.value) { ... }
    // if (shuangxiongMode.value) { ... }
    if (tiesuoMode.value) {
      return tiesuoTargets.value.length >= 1 && tiesuoTargets.value.length <= 2
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
      // 过河拆桥/顺手牵羊：具体选牌在 TakeWindow 弹窗中处理，出牌阶段只需选目标
      if (card.kind === 'guohe' || card.kind === 'tannang') return true
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
    (isMyResponse.value || isMyPlay.value || isMyDiscard.value || isMyPrepare.value || isMyDraw.value || isPeekDeck.value || isJijiHeal.value ||
     // AOE/桃园阶段：非目标玩家也可以出无懈可击
     (state.value?.phase === 'response' && state.value?.pending?.allow_wuxiek === true &&
      state.value?.pending?.response_mode !== 'wuxiek_trick') ||
     // 无懈窗口阶段：任何人可出无懈
     (state.value?.phase === 'response' && state.value?.pending?.response_mode === 'wuxiek_trick' && hasWuxiekInHand.value)),
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
  // 五谷丰登选牌阶段：非选牌者不能取消
  if (isWuguPick.value && state.value?.pending?.actor_seat !== mySeat.value) return false
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
const canSubmitPojun = computed(() => {
  if (!isPojun.value) return false
  const ctx = makePendingContext()
  if (!ctx) return false
  return pendingCanSubmitSkill(ctx, 'pojun') ?? false
})
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
  tiesuoMode,
  tiesuoTargets,
  jiedaoMode,
  jiedaoWeaponHolder,
  jiedaoShaTarget,
  fangtianMode,
  fangtianTargets,
  gameLog,
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
  isAoeWuxiekOffer,
    isWuguPick,
    isWuguActive,
    isWuguBoardVisible,
    wuguPickedCards,
    wuguRevealedAllCache,
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
  const p = state.value?.pending
  return (
    (isQilinBow.value && (p?.effect_target ?? opponentSeat.value) === seat) ||
    (p?.subject_seat != null && p.subject_seat === seat) ||
    takeableHere
  )
}

function showSeatTimer(seat: number) {
  if (isDealing.value || isFinished.value) return false
  if (isResponse.value) {
    return state.value?.pending?.actor_seat === seat
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

function clearOtherModes() {
  activeViewAs.value = null
  // [已废弃] 旧变量清理，已被 activeViewAs 替代
  // wushengMode.value = false; qixiMode.value = false; guoseMode.value = false
  // shuangxiongMode.value = false; zhangbaMode.value = false; zhangbaSelectedIds.value = []
  rendeMode.value = false
  zhihengMode.value = false
  jieyinMode.value = false
  fanjianMode.value = false
}

// [已废弃] 已被 activeViewAs 替代
// function clearWushengMode() { wushengMode.value = false }

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

function clearFangtianMode() {
  fangtianMode.value = false
  fangtianTargets.value = []
}

function clearJiedaoMode() {
  jiedaoMode.value = false
  jiedaoWeaponHolder.value = null
  jiedaoShaTarget.value = null
}

// [已废弃] 已被 activeViewAs 替代
// function clearZhangbaMode() { zhangbaMode.value = false; zhangbaSelectedIds.value = [] }

function clearTiesuoMode() {
  tiesuoMode.value = false
  tiesuoTargets.value = []
}

// 铁索连环重铸：弃置此牌摸一张
async function submitTiesuoRecast() {
  if (!tiesuoMode.value) return
  const card = selectedCard.value
  if (!card) return
  await act(() =>
    playYuzhoushaCard(state.value!.id, card.id, {
      targetIndex: mySeat.value,
      targetZone: 'recast',
    }),
  )
  clearTiesuoMode()
}

// 铁索连环：点击座位切换目标选择
function toggleTiesuoTarget(seat: number) {
  if (!tiesuoMode.value) return
  console.log('[tiesuo] toggleTiesuoTarget seat=', seat, 'before=', [...tiesuoTargets.value])
  const idx = tiesuoTargets.value.indexOf(seat)
  if (idx >= 0) {
    tiesuoTargets.value = tiesuoTargets.value.filter((s) => s !== seat)
  } else if (tiesuoTargets.value.length < 2) {
    tiesuoTargets.value = [...tiesuoTargets.value, seat]
  }
  console.log('[tiesuo] toggleTiesuoTarget seat=', seat, 'after=', [...tiesuoTargets.value])
}

// 借刀杀人双目标选择：
// 第一步：选有武器的角色（不能是自己）
// 第二步：选该角色攻击范围内的任意角色（包括使用者自己）
function toggleJiedaoTarget(seat: number) {
  if (!jiedaoMode.value) return
  const player = state.value?.players[seat]
  if (!player || (player.hp ?? 0) <= 0) return

  // 第一步：选被借刀者（必须有武器）
  if (jiedaoWeaponHolder.value === null) {
    if (seat === mySeat.value) return // 不能选自己
    if (!player.weapon) return        // 必须有武器
    jiedaoWeaponHolder.value = seat
    jiedaoShaTarget.value = null      // 清除之前选的出杀目标
    return
  }

  // 第二步：选出杀目标（在被借刀者攻击范围内，不能是被借刀者自己）
  if (jiedaoShaTarget.value === null) {
    if (seat === jiedaoWeaponHolder.value) return // 不能选被借刀者自己
    // 检查是否在攻击范围内（简化：距离 ≤ 武器范围）
    const holder = state.value?.players[jiedaoWeaponHolder.value]
    const weaponRange = holder?.weapon ? getWeaponRange(holder.weapon.kind) : 1
    const dist = getDistance(jiedaoWeaponHolder.value, seat)
    if (dist > weaponRange) return // 不在攻击范围内
    jiedaoShaTarget.value = seat
    return
  }

  // 已选完两个目标，点击已选的可取消
  if (seat === jiedaoWeaponHolder.value) {
    jiedaoWeaponHolder.value = null
    jiedaoShaTarget.value = null
  } else if (seat === jiedaoShaTarget.value) {
    jiedaoShaTarget.value = null
  }
}

// 借刀杀人可选目标判定
function isJiedaoWeaponHolderTarget(seat: number): boolean {
  if (!jiedaoMode.value) return false
  if (seat === mySeat.value) return false
  const player = state.value?.players[seat]
  if (!player || (player.hp ?? 0) <= 0) return false
  return !!player.weapon // 必须有武器
}

function isJiedaoShaTargetable(seat: number): boolean {
  if (!jiedaoMode.value || jiedaoWeaponHolder.value === null) return false
  if (seat === jiedaoWeaponHolder.value) return false
  const player = state.value?.players[seat]
  if (!player || (player.hp ?? 0) <= 0) return false
  // 检查是否在被借刀者的攻击范围内
  const holder = state.value?.players[jiedaoWeaponHolder.value]
  const weaponRange = holder?.weapon ? getWeaponRange(holder.weapon.kind) : 1
  const dist = getDistance(jiedaoWeaponHolder.value, seat)
  return dist <= weaponRange
}

// 辅助函数
function getWeaponRange(kind: string): number {
  const ranges: Record<string, number> = {
    weapon_1: 1, weapon_2: 2, weapon_3: 3, weapon_4: 4, weapon_5: 5,
    weapon_6: 2, weapon_7: 4, weapon_8: 2, weapon_9: 3, weapon_10: 3,
  }
  return ranges[kind] ?? 1
}

// 方天画戟：装备 weapon_4 且手牌只剩最后一张时触发
function canFangtianTrigger(): boolean {
  const weapon = myPlayer.value?.weapon
  if (!weapon || weapon.kind !== 'weapon_4') return false
  return myHand.value.length === 1 // 只剩最后一张手牌
}

function toggleFangtianTarget(seat: number) {
  if (!fangtianMode.value) return
  if (seat === mySeat.value) return // 不能选自己
  const player = state.value?.players[seat]
  if (!player || (player.hp ?? 0) <= 0) return
  // 检查是否在攻击范围内
  const weaponRange = getWeaponRange('weapon_4')
  const dist = getDistance(mySeat.value, seat)
  if (dist > weaponRange) return

  const idx = fangtianTargets.value.indexOf(seat)
  if (idx >= 0) {
    fangtianTargets.value = fangtianTargets.value.filter((s) => s !== seat)
  } else if (fangtianTargets.value.length < 3) {
    fangtianTargets.value = [...fangtianTargets.value, seat]
  }
}

// 方天画戟可选目标判定
function isFangtianTargetable(seat: number): boolean {
  if (!fangtianMode.value) return false
  if (seat === mySeat.value) return false
  const player = state.value?.players[seat]
  if (!player || (player.hp ?? 0) <= 0) return false
  const dist = getDistance(mySeat.value, seat)
  return dist <= getWeaponRange('weapon_4')
}

function clearAllSkillModes() {
  clearRendeMode()
  clearZhihengMode()
  clearJieyinMode()
  clearFanjianMode()
  clearQixiMode()
  clearGuoseMode()
  clearShuangxiongMode()
  clearTiesuoMode()
  clearZhangbaMode()
  clearJiedaoMode()
  clearFangtianMode()
}

function getDistance(a: number, b: number): number {
  const n = state.value?.players.length ?? 0
  if (n <= 1) return 0
  const forward = (b - a + n) % n
  const backward = (a - b + n) % n
  return Math.min(forward, backward)
}

function clearFanjianMode() {
  fanjianMode.value = false
  fanjianSelectedId.value = ''
}

// [已废弃] 已被 activeViewAs 替代
// function clearQixiMode() { qixiMode.value = false; qixiSelectedId.value = '' }
// function clearGuoseMode() { guoseMode.value = false; guoseSelectedId.value = ''; guoseTarget.value = -1 }
// function clearShuangxiongMode() { shuangxiongMode.value = false; shuangxiongSelectedId.value = '' }

function clearSkillSelectModes() {
  clearRendeMode()
  clearZhihengMode()
  clearJieyinMode()
  clearFanjianMode()
  clearQixiMode()
  clearGuoseMode()
  clearShuangxiongMode()
  clearTiesuoMode()
  clearZhangbaMode()
  clearJiedaoMode()
  clearFangtianMode()
}

function clearTargeting() {
  shaTarget.value = null
  selectedTargetZone.value = ''
  selectedTargetCardId.value = ''
  selectedQilinZone.value = ''
  hitFlashSeat.value = null
  blockFlashSeat.value = null
  // 不清除变牌模式（武圣/奇袭），用户取消选牌后仍可继续用技能
  // clearSkillSelectModes()
  // clearWushengMode()
}


const wushengModeInitialized = ref(false)
const qixiModeInitialized = ref(false)

function syncWushengFromState() {
  // 只在首次加载时从后端同步，后续由前端自己管理状态
  if (!wushengModeInitialized.value) {
    wushengMode.value = (myPlayer.value?.skill_counters?.wusheng_active ?? 0) > 0
    wushengModeInitialized.value = true
  }
}

function syncQixiFromState() {
  if (!qixiModeInitialized.value) {
    qixiMode.value = (myPlayer.value?.skill_counters?.qixi_active ?? 0) > 0
    qixiModeInitialized.value = true
  }
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
  syncQixiFromState,
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

const wsGameConnected = ref(false)
let pollTimer: number | null = null

async function applyRemoteGameState(next: YuzhoushaState) {
  if (!state.value || loading.value || isDealing.value || isAnimating.value) return
  if (state.value.id !== next.id) {
    await applyState(next)
    return
  }

  // 在 applyState 前提取 events 用于日志
  const newEvents = next.events ?? []
  const currentTurn = next.current_turn
  const round = next.round ?? 0

  loading.value = true
  try {
    await applyState(next)
    // 追加到持久化日志（直接使用后端 Message，不做前端转换）
    for (const e of newEvents) {
      if (e.message) {
        gameLog.value = [...gameLog.value, { round, turn: currentTurn, msg: e.message }].slice(-200)
      }
    }
  } finally {
    loading.value = false
  }
}

useYuzhoushaGameSocket({
  gameId: computed(() => {
    const fromRoute = route.params.gameId
    if (typeof fromRoute === 'string' && fromRoute) return fromRoute
    return state.value?.id ?? ''
  }),
  enabled: computed(
    () => isOnline.value && Boolean(state.value?.id) && state.value?.phase !== 'finished',
  ),
  currentState: state,
  onStatus: (status) => {
    wsGameConnected.value = status === 'open'
  },
  onState: applyRemoteGameState,
})

function stopPollFallback() {
  if (pollTimer != null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

function startPollFallback() {
  if (pollTimer != null) return
  pollTimer = window.setInterval(async () => {
    if (!isOnline.value || wsGameConnected.value || !state.value?.id || loading.value || isAnimating.value) {
      return
    }
    try {
      const next = await getYuzhoushaState(state.value.id)
      await applyRemoteGameState(next)
    } catch {
      // ignore poll errors
    }
  }, 2000)
}

watch(wsGameConnected, (open) => {
  if (isOnline.value && !open) startPollFallback()
  else stopPollFallback()
})

onUnmounted(() => stopPollFallback())

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
  // 从手牌或装备区查找卡牌（武圣/奇袭/国色等变牌模式下可用装备区牌）
  let card = myHand.value.find((c) => c.id === id)
  if (!card) {
    const equips = equippedCards(myPlayer.value)
    card = equips.find((c: YzsCard) => c.id === id)
  }
  // 五谷丰登/观星等：从 revealed_cards 中查找
  if (!card) {
    card = state.value?.pending?.revealed_cards?.find((c) => c.id === id)
  }
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

  // ===== ViewAs 统一变牌选牌（替换 zhangbaMode/qixiMode/guoseMode/shuangxiongMode） =====
  if (activeViewAs.value && (isMyPlay.value || isMyResponse.value)) {
    const av = activeViewAs.value
    const cardObj = myHand.value.find(c => c.id === id)
    if (cardObj && isCardSelectableForViewAs(cardObj)) {
      const idx = av.selectedIds.indexOf(id)
      if (idx >= 0) {
        av.selectedIds.splice(idx, 1)
      } else if (av.selectedIds.length < av.selectCount) {
        av.selectedIds = [...av.selectedIds, id]
      }
    }
    return
  }

  // 保留旧兼容（丈八蛇矛，G6后续步骤清理）
  if (zhangbaMode.value && (isMyPlay.value || isMyResponse.value)) {
    if (zhangbaSelectedIds.value.includes(id)) {
      zhangbaSelectedIds.value = zhangbaSelectedIds.value.filter((x) => x !== id)
    } else if (zhangbaSelectedIds.value.length < 2) {
      zhangbaSelectedIds.value = [...zhangbaSelectedIds.value, id]
    }
    return
  }

  // [已废弃] 旧变牌选牌分支，已被 activeViewAs 接管
  // if (qixiMode.value && isMyPlay.value) { ... }
  // if (guoseMode.value && isMyPlay.value) { ... }
  // if (shuangxiongMode.value && isMyPlay.value) { ... }
  // ===== ViewAs 变牌选牌结束 =====

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
    if (card.kind === 'tiesuo') clearTiesuoMode()
    if (card.kind === 'jiedao') clearJiedaoMode()
    return
  }

  selectedId.value = id
  // 方天画戟：最后一张手牌是杀时，进入多目标选择模式
  if (isMyPlay.value && cardPlaysAsSha(card) && canFangtianTrigger()) {
    fangtianMode.value = true
    fangtianTargets.value = []
    clearAllSkillModes()
    return
  }
  // 借刀杀人：进入双目标选择模式（先选被借刀者，再选出杀目标）
  if (isMyPlay.value && card.kind === 'jiedao') {
    jiedaoMode.value = true
    jiedaoWeaponHolder.value = null
    jiedaoShaTarget.value = null
    clearRendeMode()
    clearZhihengMode()
    clearJieyinMode()
    clearFanjianMode()
    clearQixiMode()
    clearGuoseMode()
    clearShuangxiongMode()
    clearTiesuoMode()
    clearZhangbaMode()
    clearWushengMode()
    return
  }
  // 铁索连环：进入多目标选择模式
  if (isMyPlay.value && card.kind === 'tiesuo') {
    tiesuoMode.value = true
    tiesuoTargets.value = []
    clearRendeMode()
    clearZhihengMode()
    clearJieyinMode()
    clearFanjianMode()
    clearWushengMode()
  }
  if (!(isMyPlay.value && canTargetOpponentWith(card))) {
    clearTargeting()
  }
}


async function act(fn: () => Promise<YuzhoushaState>, opts?: { allowAnimating?: boolean }) {
  if (!state.value || loading.value || isDealing.value) return
  if (!opts?.allowAnimating && isAnimating.value) return
  loading.value = true
  try {
    const next = await fn()
    // 写入日志（单机模式不经过 applyRemoteGameState）
    const newEvents = next.events ?? []
    for (const e of newEvents) {
      if (e.message) {
        gameLog.value = [...gameLog.value, { round: next.round ?? 0, turn: next.current_turn, msg: e.message }].slice(-200)
      }
    }
    await applyState(next)
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
  wushengMode.value = false
}

async function submitCancelZhangba() {
  if (!canCancelZhangba.value) return
  clearZhangbaMode()
  selectedId.value = ''
  clearTargeting()
}

async function submitSkill(skillId: string) {
  if (!state.value || loading.value) return

  // 过河拆桥/顺手牵羊 TakeWindow：走 pending handler
  if (skillId === '' && isTakeWindow.value) {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, ''))) return
    return
  }

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
    // 响应阶段（TakeWindow 选牌）：通过 pending handler 处理
    if (isQixiTake.value) {
      const ctx = makePendingContext()
      if (ctx && (await pendingSubmitSkill(ctx, 'qixi'))) return
    }
    return
  }
  // 空 skillId：过河拆桥/顺手牵羊的 TakeWindow 选牌，走 pending handler
  if (skillId === '' && isMyResponse.value) {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, ''))) return
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
  if (skillId === 'pojun') {
    const ctx = makePendingContext()
    if (ctx && (await pendingSubmitSkill(ctx, 'pojun'))) return
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

async function submitYinghunOption(option: 'opp_draw_x_discard_1' | 'opp_draw_1_discard_x') {
  const ctx = makePendingContext()
  if (!ctx) return
  const action = option === 'opp_draw_1_discard_x' ? 'yinghun_opp_draw_1_discard_x' : 'yinghun_opp_draw_x_discard_1'
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
  if (skillBlockedInMode(skill, state.value?.mode)) return false
  if (skill.id === 'longdan' || skill.id === 'paoxiao' || skill.id === 'kongcheng') return false
  // 武圣/奇袭 可随时 toggle（激活后可取消）
  if (skill.id === 'wusheng' && isMyPlay.value) return true
  if (skill.id === 'qixi' && isMyPlay.value) return true
  // 丈八蛇矛：装备时显示为可激活技能（出牌阶段 + 响应阶段均可）
  if (skill.id === 'zhangba' && (isMyPlay.value || isMyResponse.value)) {
    return myPlayer.value?.weapon?.kind === 'weapon_10' && myHand.value.length >= 2
  }
  const ctx = makePendingContext()
  if (ctx && isMyResponse.value) {
    const handled = pendingCanSubmitSkill(ctx, skill.id)
    if (handled !== undefined) return handled
    if (skill.id === 'yiji' && isYijiOffer.value) return true
  }
  return activatableSkillIds.value.has(skill.id)
}

async function onCharacterSkillClick(skill: YzsSkillMeta) {
  if (skillBlockedInMode(skill, state.value?.mode)) return
  if (skill.id === 'rende') {
    await activateSkill('rende')
    return
  }
  if (skill.id === 'zhiheng') {
    await activateSkill('zhiheng')
    return
  }
  if (skill.id === 'jieyin') {
    await activateSkill('jieyin')
    return
  }
  if (skill.id === 'fanjian') {
    await activateSkill('fanjian')
    return
  }
  if (skill.id === 'wusheng') {
    await activateSkill('wusheng')
    return
  }
  if (skill.id === 'qixi') {
    await activateSkill('qixi')
    return
  }
  if (skill.id === 'guose') {
    await activateSkill('guose')
    return
  }
  if (skill.id === 'shuangxiong') {
    if (isMyDraw.value) {
      void submitSkill('shuangxiong')
      return
    }
    await activateSkill('shuangxiong')
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
  if (skill.id === 'pojun' && isPojun.value) {
    void submitSkill('pojun')
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

async function activateSkill(skillId: string) {
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
    // ViewAs 统一激活：奇袭
    if (activeViewAs.value?.skillId === 'qixi') {
      activeViewAs.value = null
    } else {
      activeViewAs.value = { skillId: 'qixi', asKind: 'guohe', selectCount: 1, selectedIds: [] }
      await act(() => useYuzhoushaSkill(state.value!.id, 'qixi'))
    }
    qixiMode.value = !qixiMode.value // 保留旧兼容
    selectedId.value = ''
    return
  }
  if (skillId === 'guose') {
    activeViewAs.value = { skillId: 'guose', asKind: 'lebu', selectCount: 1, selectedIds: [] }
    guoseMode.value = true // 保留旧兼容
    guoseSelectedId.value = ''
    selectedId.value = ''
    return
  }
  if (skillId === 'wusheng' && isMyPlay.value) {
    if (activeViewAs.value?.skillId === 'wusheng') {
      activeViewAs.value = null
    } else {
      activeViewAs.value = { skillId: 'wusheng', asKind: 'sha', selectCount: 1, selectedIds: [] }
      await act(() => useYuzhoushaSkill(state.value!.id, 'wusheng'))
    }
    wushengMode.value = !wushengMode.value // 保留旧兼容
    selectedId.value = ''
    return
  }
  // ===== ViewAs 统一激活 =====
  const vas = viewAsSkills.value.find(v => v.skill_id === skillId)
  if (vas && (isMyPlay.value || isMyResponse.value)) {
    if (activeViewAs.value?.skillId === skillId) {
      activeViewAs.value = null
    } else {
      activeViewAs.value = { skillId, asKind: vas.as_kind, selectCount: vas.select_card, selectedIds: [] }
      // 需要后端 toggle 的技能（武圣/奇袭等）
      if (skillId === 'wusheng' || skillId === 'qixi') {
        await act(() => useYuzhoushaSkill(state.value!.id, skillId))
      }
    }
    selectedId.value = ''
    return
  }
  // 保留旧兼容（丈八蛇矛/shuangxiong，后续步骤清理）
  if (skillId === 'zhangba' && isMyPlay.value) {
    zhangbaMode.value = !zhangbaMode.value
    zhangbaSelectedIds.value = []
    selectedId.value = ''
    clearOtherModes()
    return
  }
  if (skillId === 'shuangxiong' && isMyPlay.value) {
    activeViewAs.value = { skillId: 'shuangxiong', asKind: 'juedou', selectCount: 1, selectedIds: [] }
    shuangxiongMode.value = true // 保留旧兼容
    shuangxiongSelectedId.value = ''
    selectedId.value = ''
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

  // AOE/桃园/五谷阶段：非目标玩家出无懈可击介入
  if (!isMyResponse.value && state.value?.phase === 'response' && (state.value?.pending?.allow_wuxiek === true || state.value?.pending?.response_mode === 'wuxiek_trick')) {
    const card = selectedCard.value
    if (card && card.kind === 'wuxiek') {
      await act(() => respondYuzhoushaCard(state.value!.id, card.id))
      return
    }
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
    // ViewAs 单牌变牌提交（奇袭/国色/双雄等）
    if (activeViewAs.value && activeViewAs.value.selectCount === 1 && selectedCard.value) {
      if (shaTarget.value != null) {
        await act(() =>
          playYuzhoushaCard(state.value!.id, selectedCard.value!.id, {
            targetIndex: shaTarget.value!,
            targetZone: selectedTargetZone.value || undefined,
            targetCardId: selectedTargetCardId.value || undefined,
          }),
        )
      } else {
        await act(() => playYuzhoushaCard(state.value!.id, selectedCard.value!.id, mySeat.value))
      }
      activeViewAs.value = null
      return
    }
    // [已废弃] 旧变牌提交分支，已被 activeViewAs 接管
    // if (qixiMode.value && selectedCard.value && shaTarget.value != null) { ... }
    // if (guoseMode.value && guoseSelectedId.value !== '') { ... }
    // if (shuangxiongMode.value && shuangxiongSelectedId.value !== '') { ... }
    // 铁索连环：提交1-2个目标，或重铸
    if (tiesuoMode.value) {
      const card = selectedCard.value
      if (!card) return
      if (tiesuoTargets.value.length === 0) return
      const target1 = tiesuoTargets.value[0]
      const target2 = tiesuoTargets.value[1] ?? -1
      console.log('[tiesuo] submit: targets=', [...tiesuoTargets.value], 'target1=', target1, 'target2=', target2, 'secondTargetIndex=', target2 >= 0 ? target2 : undefined)
      await act(() =>
        playYuzhoushaCard(state.value!.id, card.id, {
          targetIndex: target1,
          secondTargetIndex: target2 >= 0 ? target2 : undefined,
        }),
      )
      clearTiesuoMode()
      return
    }
    // 方天画戟：提交多目标杀
    if (fangtianMode.value && fangtianTargets.value.length >= 1) {
      const card = selectedCard.value
      if (!card) return
      const primary = fangtianTargets.value[0]
      const extra = fangtianTargets.value.slice(1)
      await act(() =>
        playYuzhoushaCard(state.value!.id, card.id, {
          targetIndex: primary,
          fangtianExtraTargets: extra,
        }),
      )
      clearFangtianMode()
      return
    }
    // 借刀杀人：提交双目标
    if (jiedaoMode.value && jiedaoWeaponHolder.value !== null && jiedaoShaTarget.value !== null) {
      const card = selectedCard.value
      if (!card) return
      await act(() =>
        playYuzhoushaCard(state.value!.id, card.id, {
          targetIndex: jiedaoWeaponHolder.value!,
          secondTargetIndex: jiedaoShaTarget.value!,
        }),
      )
      clearJiedaoMode()
      return
    }
    // ===== ViewAs 统一变牌提交（替换 zhangbaMode/wushengMode/qixiMode 硬编码） =====
    if (activeViewAs.value) {
      const av = activeViewAs.value
      if (av.selectedIds.length === av.selectCount) {
        if (av.selectCount === 2) {
          // 多牌合一（丈八蛇矛等）
          if (isMyResponse.value) {
            await act(() => respondZhangbaSha(state.value!.id, [av.selectedIds[0], av.selectedIds[1]]))
          } else if (shaTarget.value != null) {
            await act(() =>
              playYuzhoushaCard(state.value!.id, av.selectedIds[0], {
                targetIndex: shaTarget.value!,
                zhangbaSecondCardId: av.selectedIds[1],
              }),
            )
          }
        } else {
          // 单牌变牌（武圣/奇袭/国色等）
          if (isMyResponse.value) {
            await act(() => respondYuzhoushaCard(state.value!.id, av.selectedIds[0]))
          } else if (shaTarget.value != null) {
            await act(() => playYuzhoushaCard(state.value!.id, av.selectedIds[0], shaTarget.value!))
          } else if (isMyPlay.value) {
            await act(() => playYuzhoushaCard(state.value!.id, av.selectedIds[0], mySeat.value))
          }
        }
        activeViewAs.value = null
        return
      }
    }

    // [已废弃] 旧丈八提交，已被 activeViewAs 接管
    // if (zhangbaMode.value && zhangbaSelectedIds.value.length === 2) { ... }

    const card = selectedCard.value
    if (!card) return

    // [已废弃] 旧武圣提交，已被 activeViewAs 接管
    // if (cardPlaysAsSha(card) && shaTarget.value != null) { ... clearWushengMode() ... }
    // ===== ViewAs 变牌提交结束 =====

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
  if (!state.value) return
  if (loading.value || isAnimating.value) return
  selectedId.value = ''
  await act(() => passYuzhoushaResponse(state.value!.id))
}

async function submitPassAllWuxiek() {
  if (!state.value) return
  if (loading.value) return
  selectedId.value = ''
  await act(() => passAllWuxiek(state.value!.id), { allowAnimating: true })
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
    // 写入初始日志
    const newEvents = next.events ?? []
    for (const e of newEvents) {
      if (e.message) {
        gameLog.value = [...gameLog.value, { round: next.round ?? 0, turn: next.current_turn, msg: e.message }].slice(-200)
      }
    }
    await runInitialDealAnimation(next)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '加载对局失败')
    if (isOnline.value && roomId.value) {
      const mode = state.value?.mode ?? '1v1'
      await router.replace({ path: '/games/yuzhousha/online', query: { room: roomId.value, mode } })
    } else {
      await router.replace('/games/yuzhousha/solo/pick')
    }
  } finally {
    loading.value = false
  }
}

async function restart() {
  selectedId.value = ''
  clearTargeting()
  if (isOnline.value && roomId.value) {
    const mode = state.value?.mode ?? '1v1'
    await router.push({ path: '/games/yuzhousha/online', query: { room: roomId.value, mode } })
    return
  }
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
    canCancelZhangba,
    submitCancelZhangba,
    canInteract,
    canPlayCard,
    canPlaySha,
    canPlayWuxiek,
    canSubmitBagua,
    canSubmitCancel,
    canSubmitChixiong,
    canSubmitGuanshifu,
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
    canSubmitPojun,
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
    isChixiong,
    isGuanshifu,
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
    isPojun,
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
  isWuguActive,
  isWuguBoardVisible,
  wuguPickedCards,
  wuguRevealedAllCache,
  isWuxiekOffer,
  isAoeWuxiekOffer,
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
    handleSeatTarget,
    tiesuoMode,
    tiesuoTargets,
    submitTiesuoRecast,
    clearTiesuoMode,
    jiedaoMode,
    jiedaoWeaponHolder,
    jiedaoShaTarget,
    isJiedaoWeaponHolderTarget,
    isJiedaoShaTargetable,
    fangtianMode,
    fangtianTargets,
    isFangtianTargetable,
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
    pickPojunTarget,
    pickTuxiTarget,
    playAreaRef,
    pojunTargetOptions,
    qilinHorseOptions,
    qixiMode,
    qixiSelectedId,
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
    submitPassAllWuxiek,
    submitCancelWusheng,
    submitChixiongDiscard,
    submitChixiongSkip,
    guanshifuDiscardIds,
    toggleGuanshifuCard,
    submitGuanshifuDiscard,
    submitGuanshifuSkip,
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
    syncQixiFromState,
    tableActionHint,
    tableWrapRef,
    isGuoHeTake,
    isTanNangTake,
    isTakeWindow,
    takeWindowTargetOptions,
    takeableOptionsForPlayer,
    takeableTargetOptions,
    teammateSeat,
    toastError,
    toggleDiscardSelection,
    trickStaysInJudge,
    turnDeadline,
    tuxiTargetOptions,
    weaponRange,
    viewAsSkills,
    activeViewAs,
    viewAsSkillHint,
    clearOtherModes,
    wushengMode,
    wushengSkillHint,
    zhangbaMode,
    zhangbaSkillHint,
    zhangbaSelectedIds,
    yijiGiveRemaining,
    yijiSelectedIds,
    zhihengMode,
    zhihengSelectedIds,
  }
  return api
}