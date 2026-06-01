<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import UnoCard from '../components/uno/UnoCard.vue'
import UnoHand from '../components/uno/UnoHand.vue'
import SeatIndicator from '../components/doudizhu/SeatIndicator.vue'
import DiceTablePair from '../components/dice/DiceTablePair.vue'
import {
  drawUnoCard,
  fetchUnoRoom,
  getUnoState,
  nextUnoRoom,
  playUnoCard,
  rollUnoFirst,
  startUnoGame,
  tickUnoGame,
  voteEndUno,
} from '../api/games'
import { loadSession } from '../api/auth'
import { animateUnoDealToSeat, animateUnoDrawEvent, animateUnoPlayEvent, animateUnoRevealTopCard } from '../composables/useUnoPlayAnimation'
import { useDoubleDiceRoll } from '../composables/useDiceRoll'
import { useUnoTurnTimer } from '../composables/useUnoTurnTimer'
import { showToast } from '../composables/useToast'
import type { UnoRoom } from '../types/uno'
import {
  UNO_COLOR_LABELS,
  UNO_PLAY_COLORS,
  canPlayUnoCard,
  unoColorClass,
  type UnoCard as UnoCardType,
  type UnoColor,
  type UnoEvent,
  type UnoState,
} from '../types/uno'

interface UnoCenterPlay {
  player_index: number
  player_name: string
  card: UnoCardType
}

interface SeatPlayBadge {
  color: UnoColor
  label: string
  uno?: boolean
}

const SEAT_PLAY_BADGE_MS = 20_000
const SEAT_DICE_BADGE_MS = 5000
const TABLE_DICE_SIZE = 72

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

const router = useRouter()
const route = useRoute()
const session = loadSession()

const isOnline = computed(() => route.name === 'uno-play')
const isSoloMode = computed(() => !isOnline.value)
const roomId = computed(() => {
  const raw = route.query.room
  return typeof raw === 'string' && raw ? raw : ''
})
const selfUserId = computed(() => session?.user.id ?? 0)
const roomHostUserId = ref<number | null>(null)
const isHost = computed(() => roomHostUserId.value === selfUserId.value)

const state = ref<UnoState | null>(null)
const loading = ref(false)
const isDealing = ref(false)
const isAnimating = ref(false)
const selectedId = ref<string | null>(null)
const showColorPicker = ref(false)
const pendingWildCard = ref<UnoCardType | null>(null)
const displayedHand = ref<UnoCardType[]>([])
const displayedTopCard = ref<UnoCardType | null>(null)
const centerPlay = ref<UnoCenterPlay | null>(null)
const displayedDealCounts = ref<Record<number, number>>({})
const lastGameId = ref('')

const discardAreaRef = ref<HTMLElement | null>(null)
const drawAreaRef = ref<HTMLElement | null>(null)
const tableRef = ref<HTMLElement | null>(null)

const replaySeatIndex = ref<number | null>(null)
const replaySeatSeconds = ref(12)
const replaySeatActionLabel = ref<string | undefined>(undefined)
const readySeats = ref<Record<number, boolean>>({})
const seatPlayBadges = ref<Record<number, SeatPlayBadge>>({})
const seatBadgeTimerIds = new Map<number, number>()
const seatDiceBadges = ref<Record<number, string>>({})
const isRollingForFirst = ref(false)
const tableDiceVisible = ref(false)
let seatDiceBadgeClearTimer: number | null = null

const { diceA, diceB, rollPair } = useDoubleDiceRoll(true)

const dice1Value = diceA.value
const dice2Value = diceB.value
const dice1Rolling = diceA.rolling
const dice2Rolling = diceB.rolling
const dice1Rotation = diceA.rotation
const dice2Rotation = diceB.rotation

const BOTTOM_ZONE_HEIGHT = 210
const PROMPT_HEIGHT = 88

const promptFixedStyle = ref<Record<string, string>>({
  visibility: 'hidden',
})

let tableResizeObserver: ResizeObserver | null = null
let timeoutTriggered = false

const secondsLeft = ref(20)
const activeSeatSeconds = ref(20)

/** 单机：人类回合计时仅前端展示；联机：服务端 turn_deadline_unix + tick */
const SOLO_TURN_SECONDS = 20

let soloTimerId: ReturnType<typeof setInterval> | null = null
let soloDeadlineAt = 0
let onlinePollTimer: number | null = null
let finishedRoomPollTimer: number | null = null

const mySeat = computed(() => state.value?.human_player ?? 0)
const myHand = computed(() =>
  isDealing.value || (isAnimating.value && displayedHand.value.length > 0)
    ? displayedHand.value
    : (state.value?.my_hand ?? []),
)
const topCard = computed(() => {
  if (isDealing.value) return null
  if (state.value?.opening_turn) return null
  return displayedTopCard.value ?? state.value?.top_card ?? null
})
const isMyTurn = computed(
  () =>
    state.value?.phase === 'playing' &&
    state.value.current_turn === mySeat.value &&
    !playerByIndex(mySeat.value)?.eliminated,
)

const canHoverHand = computed(
  () =>
    !!state.value &&
    state.value.phase === 'playing' &&
    !isDealing.value &&
    !isFinished.value &&
    !isPlayerEliminated(mySeat.value) &&
    myHand.value.length > 0,
)

const canInteractHand = computed(
  () => isMyTurn.value && canHoverHand.value && !isAnimating.value,
)
const isFinished = computed(() => state.value?.phase === 'finished')
const isRollPhase = computed(() => state.value?.phase === 'roll_for_first')

const playableIds = computed(() => {
  if (!state.value || !isMyTurn.value) return []
  return myHand.value
    .filter((c) => canPlayUnoCard(c, state.value!, myHand.value))
    .map((c) => c.id)
})

const canDraw = computed(() => {
  if (!isMyTurn.value || isFinished.value || isDealing.value || !state.value) return false
  const pending = state.value.pending_draw_penalty ?? 0
  if (pending > 0) return true
  if (state.value.must_play_after_stack && playableIds.value.length > 0) return false
  return true
})

const drawButtonLabel = computed(() => {
  const pending = state.value?.pending_draw_penalty ?? 0
  if (pending > 0) return `摸 ${pending} 张`
  return '摸牌'
})

const hasTurnActions = computed(() => {
  const pending = state.value?.pending_draw_penalty ?? 0
  if (pending > 0) return playableIds.value.length > 0 || canDraw.value
  return playableIds.value.length > 0 || canDraw.value
})

const canAct = computed(
  () =>
    isMyTurn.value &&
    !loading.value &&
    !isAnimating.value &&
    !isDealing.value &&
    !isFinished.value &&
    hasTurnActions.value,
)

async function handleOnlineTurnTimeout() {
  if (!state.value || loading.value || isAnimating.value || isDealing.value) return
  try {
    loading.value = true
    await applyState(await tickUnoGame(state.value.id))
  } catch {
    // ignore transient timeout errors
  } finally {
    loading.value = false
  }
}

const { secondsLeft: onlineSecondsLeft } = useUnoTurnTimer(state, isMyTurn, handleOnlineTurnTimeout)

const turnSecondsLeft = computed(() =>
  isOnline.value ? onlineSecondsLeft.value : secondsLeft.value,
)

const showPlayButtons = computed(() => canAct.value && !showColorPicker.value)
const showColorPickerPrompt = computed(() => showColorPicker.value && isMyTurn.value)

const isHumanWinner = computed(
  () => isFinished.value && state.value?.winner_index === mySeat.value,
)

const humanFinishRank = computed(() => {
  const p = state.value?.players.find((pl) => pl.index === mySeat.value)
  return p?.finish_rank ?? 0
})

const placementLines = computed(() => {
  if (!state.value?.placements?.length) return []
  return state.value.placements.map((seat, i) => {
    const p = state.value!.players.find((pl) => pl.index === seat)
    const name = seat === mySeat.value ? '我' : (p?.name ?? '玩家')
    return `${i + 1}. ${name}`
  })
})

const settleTitle = computed(() => {
  if (!isFinished.value) return ''
  if (humanFinishRank.value === 1) return '你赢了！'
  if (humanFinishRank.value > 0) return `你获得第 ${humanFinishRank.value} 名`
  return isHumanWinner.value ? '你赢了！' : '本局结束'
})

const settleSubtitle = computed(() => {
  if (!state.value || !isFinished.value) return ''
  if (placementLines.value.length > 0) return placementLines.value.join(' · ')
  const winner = state.value.players.find((p) => p.index === state.value?.winner_index)
  return `${winner?.name ?? '玩家'} 获胜`
})

const activePlayerCount = computed(
  () => state.value?.players.filter((p) => !p.eliminated).length ?? 0,
)

const endVoteCount = computed(() => state.value?.end_votes?.length ?? 0)

const hasVotedEnd = computed(
  () => state.value?.end_votes?.includes(mySeat.value) ?? false,
)

const showEndVoteButton = computed(
  () =>
    !!state.value?.can_vote_to_end &&
    !playerByIndex(mySeat.value)?.eliminated &&
    !hasVotedEnd.value &&
    !isFinished.value &&
    !loading.value &&
    !isAnimating.value &&
    !isDealing.value,
)

const isHumanReady = computed(() => readySeats.value[mySeat.value] ?? false)

const allReady = computed(() => {
  if (!isFinished.value || !state.value) return false
  return state.value.players.every((p) => readySeats.value[p.index])
})

const showSettleReady = computed(
  () =>
    isFinished.value &&
    !loading.value &&
    !isAnimating.value &&
    !isDealing.value &&
    !isHumanReady.value,
)

const promptVisible = computed(
  () =>
    showPlayButtons.value ||
    showColorPickerPrompt.value ||
    showSettleReady.value ||
    showEndVoteButton.value,
)

const promptMyTurn = computed(
  () => showPlayButtons.value || showColorPickerPrompt.value || showSettleReady.value,
)

const promptBannerActive = computed(
  () =>
    showPlayButtons.value ||
    showColorPickerPrompt.value ||
    showSettleReady.value ||
    showEndVoteButton.value,
)

const promptStatusText = computed(() => {
  if (!state.value) return '\u00a0'
  if (isFinished.value) {
    if (allReady.value) return '即将开始…'
    if (isHumanReady.value) return '已准备 · 等待其他玩家'
    return settleSubtitle.value
      ? `${settleTitle.value} · ${settleSubtitle.value}`
      : settleTitle.value
  }
  if (showEndVoteButton.value && !canAct.value) {
    return `剩余 ${activePlayerCount.value} 人 · ${endVoteCount.value}/${activePlayerCount.value} 同意结束`
  }
  if (hasVotedEnd.value && state.value.can_vote_to_end) {
    return `已同意结束 · ${endVoteCount.value}/${activePlayerCount.value}`
  }
  if (showColorPicker.value) return '选择要打出的颜色'
  if (canAct.value) {
    if (
      (state.value.pending_draw_penalty ?? 0) > 0 ||
      state.value.must_play_after_stack
    ) {
      return state.value.message
    }
    return `轮到你 · 剩余 ${turnSecondsLeft.value} 秒`
  }
  return '\u00a0'
})

const centerPlayName = computed(() => {
  if (!centerPlay.value) return ''
  return centerPlay.value.player_index === mySeat.value ? '我' : centerPlay.value.player_name
})

const centerTurnHint = computed(() => {
  if (!state.value || isDealing.value || isRollPhase.value || isFinished.value) return ''
  if (state.value.opening_turn && !centerPlay.value) return ''
  if (!centerPlay.value) return ''
  return centerPlayName.value
})

const opponentIndices = computed(() => {
  const total = state.value?.players.length ?? 0
  const my = mySeat.value
  const others: number[] = []
  for (let i = 1; i < total; i++) {
    others.push((my + i) % total)
  }
  return others
})

function opponentRingPos(seatIndex: number) {
  const opponents = opponentIndices.value
  return opponents.indexOf(seatIndex)
}

/** 左右下家贴底，次邻坐左/右腰，其余均匀排在顶部 */
function opponentSeatStyle(seatIndex: number) {
  const opponents = opponentIndices.value
  const n = opponents.length
  const pos = opponentRingPos(seatIndex)
  if (pos < 0) return {}

  if (n === 1) {
    return { left: '50%', top: '8%', transform: 'translate(-50%, 0)' }
  }

  if (pos === 0) {
    return { left: '2%', bottom: '4%', top: 'auto', transform: 'none' }
  }
  if (pos === n - 1) {
    return { right: '2%', left: 'auto', bottom: '4%', top: 'auto', transform: 'none' }
  }

  if (n >= 4 && pos === 1) {
    return { left: '2%', top: '44%', transform: 'translate(0, -50%)' }
  }
  if (n >= 4 && pos === n - 2) {
    return { right: '2%', left: 'auto', top: '44%', transform: 'translate(0, -50%)' }
  }

  const topStart = n >= 4 ? 2 : 1
  const topEnd = n >= 4 ? n - 2 : n - 1
  const topCount = topEnd - topStart
  const topIndex = pos - topStart

  if (topCount <= 0 || topIndex < 0 || topIndex >= topCount) {
    return { left: '50%', top: '8%', transform: 'translate(-50%, 0)' }
  }

  const t = topCount === 1 ? 0.5 : (topIndex + 1) / (topCount + 1)
  const spread = topCount >= 3 ? 44 : 36
  const angle = Math.PI - t * Math.PI

  return {
    left: `${50 + spread * Math.cos(angle)}%`,
    top: '7%',
    transform: 'translate(-50%, 0)',
  }
}

function seatIndicatorPlacement(seatIndex: number): 'left' | 'right' | 'top' {
  const n = opponentIndices.value.length
  const pos = opponentRingPos(seatIndex)
  if (pos < 0) return 'top'
  if (n === 1) return 'top'
  if (pos === 0 || (n >= 4 && pos === 1)) return 'right'
  if (pos === n - 1 || (n >= 4 && pos === n - 2)) return 'left'
  return 'top'
}

function handCountForSeat(index: number) {
  if (isDealing.value) return displayedDealCounts.value[index] ?? 0
  if (index === mySeat.value) return myHand.value.length
  return playerByIndex(index)?.hand_count ?? 0
}

function handCountLabel(index: number) {
  const p = playerByIndex(index)
  if (p?.eliminated) {
    return p.finish_rank ? `第${p.finish_rank}名` : '已出完'
  }
  return `${handCountForSeat(index)} 张`
}

function isPlayerEliminated(index: number) {
  return playerByIndex(index)?.eliminated ?? false
}

function isUnoAlert(index: number) {
  if (isDealing.value || isRollPhase.value || isFinished.value) return false
  if (state.value?.phase !== 'playing') return false
  if (isPlayerEliminated(index)) return false
  return handCountForSeat(index) === 1
}

function seatPlayerClass(index: number) {
  return {
    'ddz__player--active': isSeatHighlighted(index),
    'uno__seat--alert': isUnoAlert(index),
    'uno__seat--out': isPlayerEliminated(index),
  }
}

function selfPlayerClass() {
  return {
    'ddz__player--active': isMyTurn.value && !isFinished.value,
    'uno__seat--alert': isUnoAlert(mySeat.value),
    'uno__seat--out': isPlayerEliminated(mySeat.value),
  }
}

function playerByIndex(index: number) {
  return state.value?.players.find((p) => p.index === index)
}

function handCountClass(index: number) {
  return {
    'uno__hand-count': true,
    'uno__hand-count--alert': isUnoAlert(index),
  }
}

function isSeatActive(index: number) {
  return state.value?.current_turn === index && !isFinished.value
}

function isSeatHighlighted(index: number) {
  if (replaySeatIndex.value === index) return true
  return isSeatActive(index)
}

/** 当前回合玩家只显示计时/操作提示，不显示出牌徽章 */
function isSeatOnTurn(index: number) {
  return state.value?.phase === 'playing' && state.value.current_turn === index
}

/** 徽章区是否应显示出牌徽章（与计时/摸牌/跳过互斥） */
function seatIndicatorShowsPlayBadge(index: number) {
  if (!seatPlayBadges.value[index]) return false
  if (replaySeatIndex.value === index && !replaySeatActionLabel.value) return false
  if (isSeatOnTurn(index) && !isAnimating.value) return false
  return true
}

function showSeatTimer(index: number) {
  if (isRollPhase.value) return false
  if (isPlayerEliminated(index)) return false
  if (seatDiceBadges.value[index]) return false
  if (seatIndicatorShowsPlayBadge(index)) return false
  if (isDealing.value) return false
  if (replaySeatIndex.value === index) return !replaySeatActionLabel.value
  if (isSeatOnTurn(index) && !isFinished.value && !isAnimating.value) return true
  return false
}

function showSeatActionLabel(index: number) {
  if (isRollPhase.value) return undefined
  if (seatDiceBadges.value[index]) return undefined
  if (seatIndicatorShowsPlayBadge(index)) return undefined
  if (replaySeatIndex.value === index) return replaySeatActionLabel.value
  return undefined
}

function seatPlayBadge(index: number) {
  if (!seatIndicatorShowsPlayBadge(index)) return null
  return seatPlayBadges.value[index] ?? null
}

function syncActiveTurnIndicator(
  turn: number | undefined,
  phase: UnoState['phase'] | undefined = state.value?.phase,
) {
  if (turn == null || turn < 0 || phase !== 'playing') return
  clearSeatPlayBadge(turn)
}

function seatDiceBadge(index: number) {
  return seatDiceBadges.value[index] ?? null
}

function setSeatDiceBadge(seat: number, label: string) {
  seatDiceBadges.value = { ...seatDiceBadges.value, [seat]: label }
}

function clearSeatDiceBadge(seat: number) {
  if (seatDiceBadges.value[seat]) {
    const next = { ...seatDiceBadges.value }
    delete next[seat]
    seatDiceBadges.value = next
  }
}

function clearAllSeatDiceBadges() {
  seatDiceBadges.value = {}
}

function cancelDiceBadgeClear() {
  if (seatDiceBadgeClearTimer != null) {
    window.clearTimeout(seatDiceBadgeClearTimer)
    seatDiceBadgeClearTimer = null
  }
}

function scheduleDiceBadgeClear() {
  cancelDiceBadgeClear()
  seatDiceBadgeClearTimer = window.setTimeout(() => {
    clearAllSeatDiceBadges()
    seatDiceBadgeClearTimer = null
  }, SEAT_DICE_BADGE_MS)
}

/** 清骰子徽章与桌面骰子（正式开局前调用） */
function clearDiceDisplay() {
  cancelDiceBadgeClear()
  clearAllSeatDiceBadges()
  tableDiceVisible.value = false
}

function tryStartTurnTimer() {
  if (
    isSoloMode.value &&
    !isAnimating.value &&
    !isDealing.value &&
    state.value?.phase === 'playing' &&
    state.value.current_turn != null &&
    state.value.current_turn >= 0
  ) {
    resetActiveTurnTimer()
  }
}

function clearSeatBadgeTimer(seat: number) {
  const id = seatBadgeTimerIds.get(seat)
  if (id != null) {
    window.clearTimeout(id)
    seatBadgeTimerIds.delete(seat)
  }
}

function clearSeatPlayBadge(seat: number) {
  clearSeatBadgeTimer(seat)
  if (seatPlayBadges.value[seat]) {
    const next = { ...seatPlayBadges.value }
    delete next[seat]
    seatPlayBadges.value = next
  }
}

function clearAllSeatPlayBadges() {
  for (const seat of Object.keys(seatPlayBadges.value).map(Number)) {
    clearSeatBadgeTimer(seat)
  }
  seatPlayBadges.value = {}
}

function badgeFromPlayEvent(event: UnoEvent, handCountAfter: number): SeatPlayBadge | null {
  if (!event.card) return null
  const card = event.card
  let color: UnoColor = 'red'
  if (event.color && event.color !== 'wild') {
    color = event.color
  } else if (card.color !== 'wild') {
    color = card.color
  }
  return {
    color,
    label: card.label,
    uno: handCountAfter === 1,
  }
}

function setSeatPlayBadge(seat: number, badge: SeatPlayBadge) {
  clearSeatPlayBadge(seat)
  seatPlayBadges.value = { ...seatPlayBadges.value, [seat]: badge }
  seatBadgeTimerIds.set(
    seat,
    window.setTimeout(() => clearSeatPlayBadge(seat), SEAT_PLAY_BADGE_MS),
  )
}

const POST_REPLAY_EVENT_TYPES = new Set(['play', 'draw', 'pass', 'player_out', 'vote_end', 'game_over'])

function postDealReplayEvents(events: UnoEvent[]) {
  return events.filter((e) => POST_REPLAY_EVENT_TYPES.has(e.type))
}

/** 发牌动画前还原已被后端 AI 执行过的步数，便于再回放 */
function buildPreReplayState(finalState: UnoState, events: UnoEvent[]): UnoState {
  if (events.length === 0) return finalState

  let players = finalState.players.map((p) => ({ ...p }))
  let myHand = [...(finalState.my_hand ?? [])]
  let drawCount = finalState.draw_count

  for (let i = events.length - 1; i >= 0; i--) {
    const e = events[i]
    if (e.type === 'play' && e.card) {
      players = players.map((p) =>
        p.index === e.player_index ? { ...p, hand_count: p.hand_count + 1 } : p,
      )
      if (e.player_index === finalState.human_player) {
        myHand = [...myHand, e.card]
      }
    } else if (e.type === 'draw') {
      const amount = e.amount ?? (e.card ? 1 : 0)
      players = players.map((p) =>
        p.index === e.player_index
          ? { ...p, hand_count: Math.max(0, p.hand_count - amount) }
          : p,
      )
      if (e.player_index === finalState.human_player && e.card) {
        myHand = myHand.filter((c) => c.id !== e.card!.id)
      }
      drawCount += amount
    }
  }

  const firstActor = events.find(
    (e) => e.type === 'play' || e.type === 'draw' || e.type === 'pass',
  )

  return {
    ...finalState,
    players,
    my_hand: myHand,
    draw_count: drawCount,
    opening_turn: true,
    current_turn: firstActor?.player_index ?? finalState.current_turn,
  }
}

async function replayPostDealEvents(finalState: UnoState, events: UnoEvent[]) {
  clearDiceDisplay()
  if (events.length === 0) {
    tryStartTurnTimer()
    return
  }
  isAnimating.value = true
  try {
    for (const event of events) {
      await replayEvent(event)
      if (finalState.phase === 'finished' && event.type === 'game_over') break
    }
    state.value = { ...finalState, events: [] }
    syncDisplayFromState(finalState)
    syncActiveTurnIndicator(finalState.current_turn, finalState.phase)
  } finally {
    isAnimating.value = false
  }
}

function seatTimerSeconds(index: number) {
  if (replaySeatIndex.value === index) return replaySeatSeconds.value
  if (isOnline.value && isSeatOnTurn(index)) return onlineSecondsLeft.value
  return activeSeatSeconds.value
}

async function animateRollRound(rollEvents: UnoEvent[]) {
  if (rollEvents.length === 0) return
  const focus =
    rollEvents.find((e) => e.player_index === mySeat.value) ?? rollEvents[0]
  tableDiceVisible.value = true
  await rollPair({
    d1: focus.dice1,
    d2: focus.dice2,
  })
  for (const event of rollEvents) {
    setSeatDiceBadge(
      event.player_index,
      `${event.dice1}+${event.dice2}=${event.amount}`,
    )
  }
  tableDiceVisible.value = false
}

async function runRollForFirst() {
  if (!state.value) return
  isRollingForFirst.value = true
  cancelDiceBadgeClear()
  clearAllSeatDiceBadges()
  let finalPlayingState: UnoState | null = null
  let postDealEvents: UnoEvent[] = []
  try {
    while (state.value?.phase === 'roll_for_first') {
      const next = await rollUnoFirst(state.value.id)
      const events = next.events ?? []
      const rollEvents = events.filter((e) => e.type === 'roll_dice')
      const tieEvent = events.find((e) => e.type === 'roll_tie')
      const firstEvent = events.find((e) => e.type === 'first_player')

      await animateRollRound(rollEvents)

      if (tieEvent) {
        showToast(tieEvent.message ?? '平局，重掷')
        for (const seat of tieEvent.tied_seats ?? []) {
          clearSeatDiceBadge(seat)
        }
        await sleep(320)
      }
      if (firstEvent) {
        showToast(firstEvent.message ?? '先手已确定')
      }

      if (next.phase === 'playing') {
        finalPlayingState = { ...next, events: [] }
        postDealEvents = postDealReplayEvents(events)
      }
      state.value = { ...next, events: [] }
    }
    if (finalPlayingState) {
      isRollingForFirst.value = false
      await nextTick()
      scheduleDiceBadgeClear()
      const dealState = buildPreReplayState(finalPlayingState, postDealEvents)
      await runUnoDealAnimation(dealState)
      await replayPostDealEvents(finalPlayingState, postDealEvents)
    }
  } catch (err) {
    toastError(err instanceof Error ? err.message : '掷骰失败')
  } finally {
    isRollingForFirst.value = false
    tableDiceVisible.value = false
  }
}

function updatePromptFixedPosition() {
  const table = tableRef.value
  if (!table) {
    promptFixedStyle.value = { visibility: 'hidden' }
    return
  }

  const rect = table.getBoundingClientRect()
  const width = Math.min(560, Math.max(0, rect.width - 48))
  const left = rect.left + (rect.width - width) / 2
  const top = rect.bottom - BOTTOM_ZONE_HEIGHT - PROMPT_HEIGHT

  promptFixedStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    width: `${width}px`,
    height: `${PROMPT_HEIGHT}px`,
    visibility: promptVisible.value ? 'visible' : 'hidden',
  }
}

function bindPromptPositionObserver() {
  tableResizeObserver?.disconnect()
  tableResizeObserver = null

  const table = tableRef.value
  if (!table) {
    updatePromptFixedPosition()
    return
  }

  updatePromptFixedPosition()
  tableResizeObserver = new ResizeObserver(updatePromptFixedPosition)
  tableResizeObserver.observe(table)
}

function clearActiveTurnTimer() {
  if (soloTimerId) {
    clearInterval(soloTimerId)
    soloTimerId = null
  }
}

function resetActiveTurnTimer() {
  clearActiveTurnTimer()
  const turn = state.value?.current_turn
  if (
    !isSoloMode.value ||
    turn == null ||
    turn < 0 ||
    isAnimating.value ||
    isDealing.value ||
    isFinished.value ||
    state.value?.phase !== 'playing'
  ) {
    return
  }

  // 单机：后端 deadline 在 API 返回前已开始走表，发牌/回放过场期间玩家无法操作，
  // 以「界面可行动」时刻在客户端重新计满 20 秒，避免首回合只剩 ~10 秒。
  soloDeadlineAt = Date.now() + SOLO_TURN_SECONDS * 1000

  const tick = () => {
    const left = Math.max(0, Math.ceil((soloDeadlineAt - Date.now()) / 1000))
    activeSeatSeconds.value = left
    if (turn === mySeat.value) {
      secondsLeft.value = left
      if (left <= 0 && !timeoutTriggered && canAct.value) {
        timeoutTriggered = true
        void handleSoloTurnTimeout()
      }
    }
  }

  tick()
  soloTimerId = setInterval(tick, 200)
}

function pickTimeoutWildColor(): UnoColor {
  const current = state.value?.current_color
  if (current && current !== 'wild') return current
  for (const card of myHand.value) {
    if (card.color !== 'wild') return card.color
  }
  return 'red'
}

async function handleSoloTurnTimeout() {
  if (!state.value || loading.value || isAnimating.value) return
  try {
    const pending = state.value.pending_draw_penalty ?? 0
    if (pending > 0) {
      await act(() => drawUnoCard(state.value!.id))
      return
    }
    const playable = playableIds.value
    if (playable.length > 0) {
      const card = myHand.value.find((c) => playable.includes(c.id))
      if (card) {
        if (card.color === 'wild') {
          await act(() => playUnoCard(state.value!.id, card.id, pickTimeoutWildColor()))
        } else {
          await act(() => playUnoCard(state.value!.id, card.id))
        }
        return
      }
    }
    if (canDraw.value) {
      await act(() => drawUnoCard(state.value!.id))
    }
  } finally {
    timeoutTriggered = false
    if (
      isSoloMode.value &&
      isMyTurn.value &&
      canAct.value &&
      !loading.value &&
      !isAnimating.value &&
      !isDealing.value
    ) {
      resetActiveTurnTimer()
    }
  }
}

function botCountFromRoute() {
  const raw = route.query.bots
  const n = typeof raw === 'string' ? Number.parseInt(raw, 10) : 1
  return Number.isFinite(n) ? Math.min(7, Math.max(1, n)) : 1
}

function toastError(message: string) {
  showToast(message, 'error')
}

function syncDisplayFromState(next: UnoState) {
  displayedHand.value = [...(next.my_hand ?? [])]
  displayedTopCard.value = next.opening_turn ? null : next.top_card
}

async function runUnoDealAnimation(next: UnoState) {
  isDealing.value = true
  isAnimating.value = true
  displayedHand.value = []
  displayedTopCard.value = null
  centerPlay.value = null
  selectedId.value = null

  const playerCount = next.players.length
  const cardsPerPlayer = next.my_hand?.length ?? 5
  displayedDealCounts.value = Object.fromEntries(next.players.map((p) => [p.index, 0]))

  state.value = {
    ...next,
    my_hand: [],
    players: next.players.map((p) => ({ ...p, hand_count: 0 })),
    events: [],
    message: '发牌中…',
  }
  await nextTick()

  const origin = drawAreaRef.value
  for (let round = 0; round < cardsPerPlayer; round++) {
    for (let seat = 0; seat < playerCount; seat++) {
      const nextCount = (displayedDealCounts.value[seat] ?? 0) + 1
      displayedDealCounts.value = { ...displayedDealCounts.value, [seat]: nextCount }

      if (seat === mySeat.value && next.my_hand) {
        displayedHand.value = next.my_hand.slice(0, nextCount)
      }

      if (origin) {
        await animateUnoDealToSeat(origin, seat)
      }

      if (state.value) {
        state.value = {
          ...state.value,
          players: state.value.players.map((p) =>
            p.index === seat ? { ...p, hand_count: nextCount } : p,
          ),
        }
      }
      await sleep(45)
    }
  }

  await sleep(200)

  if (
    !next.opening_turn &&
    next.top_card?.id &&
    origin &&
    discardAreaRef.value
  ) {
    await animateUnoRevealTopCard(
      origin,
      discardAreaRef.value,
      next.top_card.label,
      unoColorClass(next.top_card.color),
    )
  }

  displayedTopCard.value = next.opening_turn ? null : next.top_card
  state.value = { ...next, events: [] }
  syncDisplayFromState(next)
  syncActiveTurnIndicator(next.current_turn, next.phase)
  isDealing.value = false
  isAnimating.value = false
}

const AI_THINK_MS = 2200
const AI_OPPONENT_DRAW_MS = 320
const AI_THINK_DISPLAY_SEC = 20

async function showSeatThinking(
  seatIndex: number,
  ms: number,
  mode: 'timer' | 'draw' | 'pass',
) {
  if (seatIndex === mySeat.value) return
  clearSeatPlayBadge(seatIndex)
  replaySeatIndex.value = seatIndex
  replaySeatActionLabel.value =
    mode === 'draw' ? '摸牌' : mode === 'pass' ? '跳过' : undefined

  if (replaySeatActionLabel.value) {
    await sleep(ms)
  } else {
    const start = Date.now()
    replaySeatSeconds.value = AI_THINK_DISPLAY_SEC
    while (Date.now() - start < ms) {
      const elapsedSec = Math.floor((Date.now() - start) / 1000)
      replaySeatSeconds.value = Math.max(1, AI_THINK_DISPLAY_SEC - elapsedSec)
      const nextSecondAt = start + (elapsedSec + 1) * 1000
      const remaining = ms - (Date.now() - start)
      await sleep(Math.min(Math.max(0, nextSecondAt - Date.now()), remaining))
    }
  }

  replaySeatIndex.value = null
  replaySeatActionLabel.value = undefined
}

async function replayEvent(event: UnoEvent) {
  if (!state.value) return

  const isOpponent = event.player_index !== mySeat.value

  if (event.type !== 'game_over') {
    const thinkMs =
      isOnline.value && isOpponent
        ? 0
        : event.type === 'draw' && isOpponent
          ? AI_OPPONENT_DRAW_MS
          : event.type === 'pass' || event.type === 'play'
            ? AI_THINK_MS
            : 0
    const mode =
      event.type === 'draw' ? 'draw' : event.type === 'pass' ? 'pass' : 'timer'
    if (thinkMs > 0) {
      await showSeatThinking(event.player_index, thinkMs, mode)
    }
  }

  if (event.type === 'roll_dice') {
    if (event.dice1 != null && event.dice2 != null) {
      setSeatDiceBadge(
        event.player_index,
        `${event.dice1}+${event.dice2}=${event.amount ?? event.dice1 + event.dice2}`,
      )
    }
    await sleep(280)
    return
  }

  if (event.type === 'roll_tie') {
    showToast(event.message ?? '平局，重掷')
    for (const seat of event.tied_seats ?? []) {
      clearSeatDiceBadge(seat)
    }
    await sleep(320)
    return
  }

  if (event.type === 'first_player') {
    showToast(event.message ?? '先手已确定')
    await sleep(280)
    return
  }

  if (event.type === 'play' && event.card) {
    await animateUnoPlayEvent(event, discardAreaRef.value, mySeat.value, () => {
      displayedTopCard.value = event.card!
      centerPlay.value = {
        player_index: event.player_index,
        player_name: event.player_name,
        card: event.card!,
      }
      if (event.player_index === mySeat.value) {
        displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
      }
      state.value = {
        ...state.value!,
        top_card: event.card!,
        current_color: event.color ?? state.value!.current_color,
        opening_turn: false,
        message: event.message || state.value!.message,
        players: state.value!.players.map((p) =>
          p.index === event.player_index
            ? { ...p, hand_count: Math.max(0, p.hand_count - 1) }
            : p,
        ),
      }
      const handAfter =
        state.value.players.find((p) => p.index === event.player_index)?.hand_count ?? 0
      const badge = badgeFromPlayEvent(event, handAfter)
      if (badge) setSeatPlayBadge(event.player_index, badge)
    })
    return
  }

  if (event.type === 'player_out') {
    showToast(event.message ?? '有玩家出完牌')
    if (state.value) {
      state.value = {
        ...state.value,
        message: event.message || state.value.message,
        players: state.value.players.map((p) =>
          p.index === event.player_index
            ? {
                ...p,
                hand_count: 0,
                eliminated: true,
                finish_rank: event.amount ?? p.finish_rank,
              }
            : p,
        ),
      }
    }
    await sleep(900)
    return
  }

  if (event.type === 'vote_end') {
    if (state.value) {
      const votes = [...(state.value.end_votes ?? [])]
      if (!votes.includes(event.player_index)) votes.push(event.player_index)
      state.value = {
        ...state.value,
        end_votes: votes,
        message: event.message || state.value.message,
      }
    }
    await sleep(360)
    return
  }

  if (event.type === 'draw') {
    const amount = event.amount ?? (event.card ? 1 : 0)
    const applyDrawState = () => {
      const current = state.value
      if (!current) return
      if (event.player_index === mySeat.value && event.card && amount === 1) {
        const drawn = event.card
        if (!displayedHand.value.some((c) => c.id === drawn.id)) {
          displayedHand.value = [...displayedHand.value, drawn]
        }
      }
      state.value = {
        ...current,
        message: event.message || current.message,
        draw_count: Math.max(0, current.draw_count - amount),
        players: current.players.map((p) =>
          p.index === event.player_index
            ? { ...p, hand_count: p.hand_count + amount }
            : p,
        ),
      }
    }

    if (event.card || amount > 0) {
      await animateUnoDrawEvent(event, drawAreaRef.value, mySeat.value, applyDrawState)
    } else {
      state.value = { ...state.value!, message: event.message || state.value!.message }
      await sleep(280)
    }
    return
  }

  state.value = { ...state.value, message: event.message || state.value.message }
  await sleep(event.type === 'game_over' ? 700 : 280)
}

async function applyState(next: UnoState) {
  const events = next.events ?? []
  const isNewGame = lastGameId.value !== next.id
  const prevPhase = state.value?.phase

  if (events.length === 0) {
    if (isNewGame) {
      lastGameId.value = next.id
      if (next.phase === 'roll_for_first') {
        state.value = next
        if (isSoloMode.value || isHost.value) {
          await runRollForFirst()
        }
        return
      }
      await runUnoDealAnimation(next)
      return
    }
    if (
      isOnline.value &&
      prevPhase === 'roll_for_first' &&
      next.phase === 'playing'
    ) {
      await runUnoDealAnimation(next)
      return
    }
    state.value = next
    syncDisplayFromState(next)
    syncActiveTurnIndicator(next.current_turn, next.phase)
    selectedId.value = null
    return
  }

  isAnimating.value = true
  try {
    const priorHand = state.value?.my_hand?.length
      ? [...state.value.my_hand]
      : displayedHand.value.length > 0
        ? [...displayedHand.value]
        : [...(next.my_hand ?? [])]
    displayedHand.value = priorHand

    if (!state.value) {
      state.value = { ...next, events: [], my_hand: priorHand }
    }
    await nextTick()

    for (const event of events) {
      await replayEvent(event)
      if (next.phase === 'finished' && event.type === 'game_over') break
    }

    if (next.phase === 'playing' && (prevPhase === 'roll_for_first' || isNewGame)) {
      lastGameId.value = next.id
      isAnimating.value = false
      const postDeal = postDealReplayEvents(events)
      const dealState = buildPreReplayState(next, postDeal)
      await runUnoDealAnimation(dealState)
      await replayPostDealEvents(next, postDeal)
      return
    }

    state.value = { ...next, events: [] }
    syncDisplayFromState(next)
    syncActiveTurnIndicator(next.current_turn, next.phase)
    selectedId.value = null
  } finally {
    isAnimating.value = false
  }
}

function resetReadyState() {
  const next: Record<number, boolean> = {}
  for (const p of state.value?.players ?? []) {
    next[p.index] = false
  }
  readySeats.value = next
}

function syncReadyFromRoom(room: UnoRoom) {
  const next: Record<number, boolean> = {}
  room.players.forEach((player, seat) => {
    next[seat] = player.ready
  })
  readySeats.value = next
}

async function loadRoomMeta() {
  if (!roomId.value) return
  try {
    const room = await fetchUnoRoom(roomId.value)
    roomHostUserId.value = room.host_user_id
    syncReadyFromRoom(room)
  } catch {
    // ignore transient room errors
  }
}

async function loadGame(gameId: string) {
  loading.value = true
  lastGameId.value = ''
  try {
    if (roomId.value) await loadRoomMeta()
    await applyState(await getUnoState(gameId))
  } catch (err) {
    toastError(err instanceof Error ? err.message : '加载对局失败')
  } finally {
    loading.value = false
  }
}

function startOnlinePolling() {
  stopOnlinePolling()
  if (!isOnline.value) return
  onlinePollTimer = window.setInterval(async () => {
    if (
      !state.value ||
      loading.value ||
      isAnimating.value ||
      isDealing.value ||
      isRollingForFirst.value ||
      state.value.phase === 'finished'
    ) {
      return
    }
    try {
      const next = await getUnoState(state.value.id)
      const turnChanged = next.current_turn !== state.value.current_turn
      const phaseChanged = next.phase !== state.value.phase
      const hasEvents = (next.events?.length ?? 0) > 0
      if (hasEvents || turnChanged || phaseChanged) {
        loading.value = true
        try {
          await applyState(next)
        } finally {
          loading.value = false
        }
      }
    } catch {
      // ignore transient polling errors
    }
  }, 1500)
}

function stopOnlinePolling() {
  if (onlinePollTimer !== null) {
    window.clearInterval(onlinePollTimer)
    onlinePollTimer = null
  }
}

async function enterNextOnlineGame(gameId: string) {
  if (!roomId.value) return
  resetReadyState()
  stopOnlinePolling()
  stopFinishedRoomPolling()
  lastGameId.value = ''
  loading.value = true
  try {
    await router.replace({
      name: 'uno-play',
      params: { gameId },
      query: { room: roomId.value },
    })
    await applyState(await getUnoState(gameId))
    startOnlinePolling()
  } catch (err) {
    toastError(err instanceof Error ? err.message : '进入下一局失败')
  } finally {
    loading.value = false
  }
}

function startFinishedRoomPolling() {
  stopFinishedRoomPolling()
  if (!isOnline.value || !roomId.value) return

  const poll = async () => {
    if (!roomId.value || !state.value || state.value.phase !== 'finished') return
    try {
      const room = await fetchUnoRoom(roomId.value)
      syncReadyFromRoom(room)
      if (room.game_id && room.game_id !== state.value.id) {
        await enterNextOnlineGame(room.game_id)
      }
    } catch {
      // ignore transient polling errors
    }
  }

  void poll()
  finishedRoomPollTimer = window.setInterval(poll, 1500)
}

function stopFinishedRoomPolling() {
  if (finishedRoomPollTimer !== null) {
    window.clearInterval(finishedRoomPollTimer)
    finishedRoomPollTimer = null
  }
}

async function waitForNextOnlineGame() {
  if (!roomId.value || !state.value) return
  const currentGameId = state.value.id
  for (let i = 0; i < 40; i++) {
    const room = await fetchUnoRoom(roomId.value)
    syncReadyFromRoom(room)
    if (room.game_id && room.game_id !== currentGameId) {
      await enterNextOnlineGame(room.game_id)
      return
    }
    await sleep(1000)
  }
  toastError('等待其他玩家准备超时')
  readySeats.value = { ...readySeats.value, [mySeat.value]: false }
}

async function handleReady() {
  if (!isFinished.value || loading.value || isHumanReady.value) return

  if (isOnline.value && roomId.value && state.value) {
    loading.value = true
    readySeats.value = { ...readySeats.value, [mySeat.value]: true }
    try {
      const room = await nextUnoRoom(roomId.value, true)
      syncReadyFromRoom(room)
      if (room.game_id && room.game_id !== state.value.id) {
        await enterNextOnlineGame(room.game_id)
        return
      }
      await waitForNextOnlineGame()
    } catch (err) {
      readySeats.value = { ...readySeats.value, [mySeat.value]: false }
      toastError(err instanceof Error ? err.message : '准备失败')
    } finally {
      loading.value = false
    }
    return
  }

  readySeats.value = { ...readySeats.value, [mySeat.value]: true }
  await sleep(300)

  const next = { ...readySeats.value }
  for (const p of state.value?.players ?? []) {
    if (p.is_ai) next[p.index] = true
  }
  readySeats.value = next
  await sleep(400)
  await beginGame()
}

function isUnoGameLost(err: unknown) {
  if (!(err instanceof Error)) return false
  const msg = err.message
  return (
    msg.includes('game not found') ||
    msg.includes('对局不存在') ||
    msg.includes('已过期')
  )
}

async function beginGame() {
  loading.value = true
  centerPlay.value = null
  showColorPicker.value = false
  pendingWildCard.value = null
  displayedDealCounts.value = {}
  selectedId.value = null
  clearAllSeatPlayBadges()
  cancelDiceBadgeClear()
  clearAllSeatDiceBadges()
  resetReadyState()
  lastGameId.value = ''
  try {
    const next = await startUnoGame(botCountFromRoute())
    loading.value = false
    await applyState(next)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '开局失败')
  } finally {
    loading.value = false
  }
}

async function act(fn: () => Promise<UnoState>) {
  if (!state.value || loading.value || isAnimating.value || isDealing.value) return
  loading.value = true
  try {
    const next = await fn()
    loading.value = false
    await applyState(next)
  } catch (err) {
    if (isUnoGameLost(err)) {
      if (isOnline.value) {
        toastError('对局已失效，请返回房间重新开局')
      } else {
        toastError('对局已失效（后端可能已重启），正在重新开局…')
        await beginGame()
      }
      return
    }
    toastError(err instanceof Error ? err.message : '操作失败')
  } finally {
    loading.value = false
  }
}

function onSelectCard(id: string | null) {
  if (!isMyTurn.value || loading.value || isAnimating.value || isDealing.value) return
  selectedId.value = id
}

watch(isMyTurn, (turn) => {
  if (!turn) selectedId.value = null
})

async function handlePlay(card: UnoCardType) {
  if (!state.value) return
  if (card.color === 'wild' || card.value === 'wild' || card.value === 'wild4') {
    pendingWildCard.value = card
    showColorPicker.value = true
    return
  }
  await act(() => playUnoCard(state.value!.id, card.id))
}

async function pickColor(color: UnoColor) {
  showColorPicker.value = false
  const card = pendingWildCard.value
  pendingWildCard.value = null
  if (!card || !state.value) return
  await act(() => playUnoCard(state.value!.id, card.id, color))
}

function cancelColorPick() {
  showColorPicker.value = false
  pendingWildCard.value = null
}

async function handlePlaySelected() {
  const card = myHand.value.find((c) => c.id === selectedId.value)
  if (!card || !playableIds.value.includes(card.id)) return
  await handlePlay(card)
}

async function handleDraw() {
  if (!state.value || !canDraw.value) return
  await act(() => drawUnoCard(state.value!.id))
}

async function handleVoteEnd() {
  if (!state.value || !showEndVoteButton.value) return
  await act(() => voteEndUno(state.value!.id))
}

watch(
  () => [
    isAnimating.value,
    isDealing.value,
    state.value?.current_turn,
    state.value?.phase,
    state.value?.must_play_after_stack,
    state.value?.pending_draw_penalty,
  ],
  () => {
    timeoutTriggered = false
    clearActiveTurnTimer()
    if (
      isSoloMode.value &&
      !isAnimating.value &&
      !isDealing.value &&
      state.value?.phase === 'playing' &&
      state.value.current_turn != null &&
      state.value.current_turn >= 0
    ) {
      resetActiveTurnTimer()
    }
  },
  { immediate: true },
)

watch(
  () => state.value?.phase,
  (phase) => {
    if (phase === 'finished') {
      resetReadyState()
      if (isOnline.value && roomId.value) {
        stopOnlinePolling()
        startFinishedRoomPolling()
      }
    } else {
      stopFinishedRoomPolling()
    }
  },
)

watch(
  () => state.value,
  async () => {
    await nextTick()
    bindPromptPositionObserver()
  },
)

watch(
  () => state.value?.current_turn,
  (turn) => {
    syncActiveTurnIndicator(turn)
  },
)

watch(
  () => [isMyTurn.value, isAnimating.value] as const,
  ([myTurn, animating]) => {
    if (myTurn && !animating) {
      clearSeatPlayBadge(mySeat.value)
    }
  },
)

watch(promptVisible, async () => {
  await nextTick()
  updatePromptFixedPosition()
})

onMounted(async () => {
  window.addEventListener('resize', updatePromptFixedPosition)
  window.addEventListener('scroll', updatePromptFixedPosition, true)
  if (isOnline.value && route.params.gameId) {
    await loadGame(String(route.params.gameId))
    startOnlinePolling()
  } else {
    await beginGame()
  }
  await nextTick()
  bindPromptPositionObserver()
})

onUnmounted(() => {
  window.removeEventListener('resize', updatePromptFixedPosition)
  window.removeEventListener('scroll', updatePromptFixedPosition, true)
  tableResizeObserver?.disconnect()
  stopOnlinePolling()
  stopFinishedRoomPolling()
  clearActiveTurnTimer()
  cancelDiceBadgeClear()
  clearAllSeatPlayBadges()
  clearAllSeatDiceBadges()
})
</script>

<template>
  <main class="ddz app">
    <header class="ddz__header">
      <button type="button" class="ddz__back" @click="router.push('/games/uno')">← 返回</button>
      <div>
        <h1>UNO</h1>
        <p class="ddz__subtitle">
          {{ isOnline ? '多人联机' : '单机对战电脑' }} · {{ state?.players.length ?? 0 }} 人
        </p>
      </div>
      <button
        v-if="!isOnline && !isFinished"
        type="button"
        class="ddz__restart"
        :disabled="loading || isAnimating || isDealing || isRollingForFirst"
        @click="beginGame"
      >
        重新开局
      </button>
    </header>

    <p v-if="loading && !state" class="ddz__loading">准备发牌...</p>

    <section v-if="state" ref="tableRef" class="ddz__table">
      <div
        class="ddz__arena uno__arena"
        :class="{ 'uno__arena--many': opponentIndices.length >= 5 }"
      >
        <div class="uno__ring">
          <div
            v-for="seatIndex in opponentIndices"
            :key="seatIndex"
            class="uno__ring-seat"
            :style="opponentSeatStyle(seatIndex)"
          >
            <div class="ddz__seat-stack">
              <div
                class="ddz__player ddz__player--compact ddz__seat-anchor"
                :class="seatPlayerClass(seatIndex)"
                :data-seat="seatIndex"
              >
                <span>{{ playerByIndex(seatIndex)?.name }}</span>
                <span :class="handCountClass(seatIndex)">
                  <template v-if="isUnoAlert(seatIndex)">
                    <span class="uno__hand-count-num">1</span>
                    张
                  </template>
                  <template v-else>{{ handCountLabel(seatIndex) }}</template>
                </span>
                <span v-if="isFinished && readySeats[seatIndex]" class="ddz__ready-badge">准备</span>
              </div>
              <SeatIndicator
                :placement="seatIndicatorPlacement(seatIndex)"
                :dice-badge="seatDiceBadge(seatIndex)"
                :play-badge="seatPlayBadge(seatIndex)"
                :show-timer="showSeatTimer(seatIndex)"
                :seconds="seatTimerSeconds(seatIndex)"
                :action-label="showSeatActionLabel(seatIndex)"
              />
            </div>
          </div>
        </div>

        <div class="ddz__center uno__center-ring">
          <div class="ddz__center-stage uno__center-stage">
            <div class="uno__play-by-slot">
              <p
                v-if="isFinished && !isDealing"
                class="uno__center-result"
                :class="{ 'uno__center-result--win': isHumanWinner, 'uno__center-result--lose': !isHumanWinner }"
              >
                {{ settleTitle }} · {{ settleSubtitle }}
              </p>
              <p
                v-else-if="isRollPhase && !isDealing && !tableDiceVisible"
                class="uno__dealing-hint"
              >
                掷骰定先手…
              </p>
              <p
                v-else-if="!isRollPhase"
                class="ddz__play-by"
                :class="{ 'uno__play-by--hidden': !centerTurnHint || isDealing }"
              >
                <span>{{ centerTurnHint || '\u00a0' }}</span>
                <span class="ddz__play-by-action">出牌</span>
              </p>
            </div>
            <p v-if="isDealing" class="uno__dealing-hint">发牌中…</p>
            <div v-if="isRollPhase || tableDiceVisible" class="uno__roll-stage">
              <DiceTablePair
                :visible="tableDiceVisible"
                :dice1-value="dice1Value"
                :dice2-value="dice2Value"
                :dice1-rolling="dice1Rolling"
                :dice2-rolling="dice2Rolling"
                :dice1-rotation="dice1Rotation"
                :dice2-rotation="dice2Rotation"
                :size="TABLE_DICE_SIZE"
              />
            </div>
            <div
              v-else
              class="uno__piles"
              :class="{ 'uno__piles--dealing': isDealing }"
            >
              <div ref="drawAreaRef" class="uno__pile">
                <UnoCard
                  :card="{ id: 'back', color: 'wild', value: 'back', label: '牌堆' }"
                  face-down
                  mini
                />
                <span class="uno__pile-count">剩余 {{ state.draw_count }} 张</span>
              </div>
              <div
                v-if="!isDealing && !state.opening_turn"
                class="uno__color-badge"
                :class="`uno-card--${state.current_color}`"
              >
                当前 · {{ UNO_COLOR_LABELS[state.current_color] ?? state.current_color }}
              </div>
              <div ref="discardAreaRef" class="uno__pile uno__pile--discard">
                <UnoCard v-if="topCard" :card="topCard" discard-spot />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="ddz__bottom-zone">
        <div class="ddz__hand">
          <div class="ddz__hand-side">
            <div class="ddz__seat-stack ddz__seat-stack--self">
              <div
                class="ddz__player ddz__player--self ddz__seat-anchor"
                :class="selfPlayerClass()"
                :data-seat="mySeat"
              >
                <span>我</span>
                <span :class="handCountClass(mySeat)">
                  <template v-if="isUnoAlert(mySeat)">
                    <span class="uno__hand-count-num">1</span>
                    张
                  </template>
                  <template v-else>{{ handCountLabel(mySeat) }}</template>
                </span>
                <span v-if="isFinished && readySeats[mySeat]" class="ddz__ready-badge">准备</span>
              </div>
              <SeatIndicator
                placement="top"
                :dice-badge="seatDiceBadge(mySeat)"
                :play-badge="seatPlayBadge(mySeat)"
                :show-timer="showSeatTimer(mySeat)"
                :seconds="seatTimerSeconds(mySeat)"
              />
            </div>
          </div>
          <UnoHand
            :cards="myHand"
            :selected-id="selectedId"
            :interactive="canInteractHand"
            :hoverable="canHoverHand"
            @select="onSelectCard"
          />
        </div>
      </div>
    </section>

    <Teleport to="body">
      <div v-if="state && promptVisible" class="ddz__prompt-slot" :style="promptFixedStyle">
        <div
          class="ddz__prompt-inner"
          :class="{ 'ddz__prompt-inner--my-turn': promptMyTurn }"
        >
          <div class="ddz__prompt-content">
            <p
              class="ddz__prompt-banner"
              :class="{
                'ddz__prompt-banner--active': promptBannerActive,
                'ddz__prompt-banner--my-turn': promptMyTurn,
              }"
            >
              {{ promptStatusText }}
            </p>

            <div class="ddz__prompt-body">
              <div
                class="ddz__prompt-actions"
                :class="{ 'ddz__prompt-actions--visible': showEndVoteButton }"
              >
                <button
                  type="button"
                  class="ddz__btn ddz__btn--primary"
                  :disabled="loading || isDealing || isAnimating"
                  @click="handleVoteEnd"
                >
                  结束对局
                </button>
              </div>

              <div
                class="ddz__prompt-actions"
                :class="{ 'ddz__prompt-actions--visible': showSettleReady }"
              >
                <button
                  type="button"
                  class="ddz__btn ddz__btn--primary"
                  :disabled="loading || isDealing || isAnimating"
                  @click="handleReady"
                >
                  准备
                </button>
              </div>

              <div
                class="ddz__prompt-actions"
                :class="{ 'ddz__prompt-actions--visible': showColorPickerPrompt }"
              >
                <div class="uno__color-prompt">
                  <button
                    v-for="color in UNO_PLAY_COLORS"
                    :key="color"
                    type="button"
                    class="uno__color-prompt-btn"
                    :class="`uno-card--${color}`"
                    @click="pickColor(color)"
                  >
                    {{ UNO_COLOR_LABELS[color] }}
                  </button>
                  <button type="button" class="ddz__btn" @click="cancelColorPick">取消</button>
                </div>
              </div>

              <div
                class="ddz__prompt-actions"
                :class="{ 'ddz__prompt-actions--visible': showPlayButtons }"
              >
                <button
                  type="button"
                  class="ddz__btn ddz__btn--primary ddz__btn--play-emphasis"
                  :disabled="loading || !selectedId || !playableIds.includes(selectedId ?? '')"
                  @click="handlePlaySelected"
                >
                  出牌
                </button>
                <button
                  type="button"
                  class="ddz__btn"
                  :disabled="loading || !canDraw"
                  @click="handleDraw"
                >
                  {{ drawButtonLabel }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </main>
</template>
