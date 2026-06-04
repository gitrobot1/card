<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PlayingCard from '../../components/doudizhu/PlayingCard.vue'
import SeatIndicator from '../../components/doudizhu/SeatIndicator.vue'
import { animateCardsFromCenterBatch } from '../../composables/useDealAnimation'
import { showToast } from '../../composables/useToast'
import {
  fetchZhajinhuaRoom,
  getZhajinhuaState,
  nextZhajinhuaRoom,
  startZhajinhuaGame,
  tickZhajinhuaGame,
  zhajinhuaCompare,
  zhajinhuaFold,
  zhajinhuaFollow,
  zhajinhuaLook,
  zhajinhuaRaise,
} from '../../api/games'
import { HAND_TYPE_LABELS, type ZhajinhuaRoom, type ZhajinhuaEvent, type ZhajinhuaState } from '../../types/zhajinhua'
import type { Card } from '../../types/doudizhu'

const CARDS_PER_PLAYER = 3
const DEAL_SEAT_MS = 160
const SEAT_SPEECH_MS = 4500
const COMPARE_BEAM_SHOOT_MS = 360
const COMPARE_RESOLVE_MS = 320
const SEAT_ACTION_MS = 2400

const DEFAULT_HAND_MULTIPLIERS: Record<string, number> = {
  '235': 12,
  leopard: 10,
  straight_flush: 6,
  flush: 4,
  straight: 3,
  pair: 2,
  high_card: 1,
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function useSimpleTimer(getDeadline: () => number | undefined, active: () => boolean) {
  const secondsLeft = ref(35)
  let timerId: ReturnType<typeof setInterval> | null = null
  function tick() {
    const d = getDeadline()
    if (!d || !active()) {
      secondsLeft.value = 35
      return
    }
    secondsLeft.value = Math.max(0, Math.ceil(d - Date.now() / 1000))
  }
  watch(
    () => [getDeadline(), active()],
    () => {
      if (timerId) clearInterval(timerId)
      tick()
      timerId = setInterval(tick, 500)
    },
    { immediate: true },
  )
  onUnmounted(() => {
    if (timerId) clearInterval(timerId)
  })
  return { secondsLeft }
}

const router = useRouter()
const route = useRoute()

const state = ref<ZhajinhuaState | null>(null)
const loading = ref(false)
const isDealing = ref(false)
const isAnimating = ref(false)
const raiseAmount = ref(0)
const compareTarget = ref<number | null>(null)
const showMultiplierHelp = ref(false)
const replaySeatIndex = ref<number | null>(null)
const replaySeatSeconds = ref(35)
const seatActionLabels = ref<Record<number, string>>({})
const seatSpeechLabels = ref<Record<number, string>>({})
const compareBeam = ref<{
  x1: number
  y1: number
  x2: number
  y2: number
  length: number
  progress: number
} | null>(null)
const readySeats = ref<Record<number, boolean>>({})
const lastGameId = ref('')
const dealCounts = ref<Record<number, number>>({})
const arenaRef = ref<HTMLElement | null>(null)
const dealOriginRef = ref<HTMLElement | null>(null)

const isOnline = computed(() => route.name === 'zhajinhua-play')
const roomId = computed(() => {
  const value = route.query.room
  return typeof value === 'string' ? value : ''
})
const mySeat = computed(() => state.value?.human_player ?? 0)
const isGameFinished = computed(() => state.value?.phase === 'finished')
const isHumanReady = computed(() => readySeats.value[mySeat.value] ?? false)
const showSettleReady = computed(
  () => isGameFinished.value && !loading.value && !isDealing.value && !isAnimating.value && !isHumanReady.value,
)

const compareBeamHead = computed(() => {
  const beam = compareBeam.value
  if (!beam) return { x: 0, y: 0 }
  return {
    x: beam.x1 + (beam.x2 - beam.x1) * beam.progress,
    y: beam.y1 + (beam.y2 - beam.y1) * beam.progress,
  }
})

const compareBeamLineStyle = computed(() => {
  const beam = compareBeam.value
  if (!beam) return {}
  const drawn = Math.max(beam.length * beam.progress, 0.001)
  return {
    strokeDasharray: `${drawn} ${beam.length}`,
  }
})

const isMyTurn = computed(
  () =>
    state.value?.phase === 'betting' &&
    state.value.current_turn === mySeat.value &&
    !state.value.players[mySeat.value]?.folded,
)
const canAct = computed(
  () => isMyTurn.value && !loading.value && !isAnimating.value && !isDealing.value,
)
const canLook = computed(
  () =>
    state.value?.phase === 'betting' &&
    !state.value.players[mySeat.value]?.folded &&
    !state.value.players[mySeat.value]?.looked &&
    !(state.value.my_hand?.length) &&
    !loading.value &&
    !isAnimating.value &&
    !isDealing.value,
)
const hasLooked = computed(() => {
  const player = state.value?.players[mySeat.value]
  return Boolean(player?.looked || state.value?.my_hand?.length)
})
const compareTargets = computed(() =>
  (state.value?.players ?? []).filter((p) => p.index !== mySeat.value && !p.folded),
)
const canPickCompareTarget = computed(() => compareTargets.value.some((p) => p.looked))

function toastError(message: string) {
  showToast(message, 'error')
}

const { secondsLeft } = useSimpleTimer(
  () => state.value?.turn_deadline_unix,
  () =>
    state.value?.phase === 'betting' &&
    !isAnimating.value &&
    !isDealing.value &&
    !state.value.players[state.value.current_turn]?.folded,
)

let pollTimer: number | null = null
let finishedRoomPollTimer: number | null = null
const seatActionTimers = new Map<number, number>()
const seatSpeechTimers = new Map<number, number>()

function resetReadyState() {
  const next: Record<number, boolean> = {}
  for (const p of state.value?.players ?? []) {
    next[p.index] = false
  }
  readySeats.value = next
}

function syncReadyFromRoom(room: ZhajinhuaRoom) {
  const next: Record<number, boolean> = {}
  room.players.forEach((player, seat) => {
    next[seat] = player.ready
  })
  readySeats.value = next
}

const multiplierRows = computed(() => {
  const m = state.value?.hand_multipliers ?? {}
  return Object.entries(m)
    .map(([key, val]) => ({
      key,
      label: HAND_TYPE_LABELS[key] ?? key,
      value: val,
    }))
    .sort((a, b) => b.value - a.value)
})

function clearAllSeatSpeeches() {
  for (const timer of seatSpeechTimers.values()) {
    window.clearTimeout(timer)
  }
  seatSpeechTimers.clear()
  seatSpeechLabels.value = {}
}

function clearSeatSpeech(index: number) {
  const timer = seatSpeechTimers.get(index)
  if (timer !== undefined) {
    window.clearTimeout(timer)
    seatSpeechTimers.delete(index)
  }
  if (!seatSpeechLabels.value[index]) return
  const next = { ...seatSpeechLabels.value }
  delete next[index]
  seatSpeechLabels.value = next
}

function setSeatSpeech(index: number, text: string) {
  clearSeatSpeech(index)
  seatSpeechLabels.value = { ...seatSpeechLabels.value, [index]: text }
  seatSpeechTimers.set(
    index,
    window.setTimeout(() => clearSeatSpeech(index), SEAT_SPEECH_MS),
  )
}

function seatSpeech(index: number) {
  return seatSpeechLabels.value[index] ?? ''
}

function clearAllSeatActions() {
  clearAllSeatSpeeches()
  compareBeam.value = null
  for (const timer of seatActionTimers.values()) {
    window.clearTimeout(timer)
  }
  seatActionTimers.clear()
  seatActionLabels.value = {}
}

function clearSeatAction(index: number) {
  const timer = seatActionTimers.get(index)
  if (timer !== undefined) {
    window.clearTimeout(timer)
    seatActionTimers.delete(index)
  }
  if (!seatActionLabels.value[index]) return
  const next = { ...seatActionLabels.value }
  delete next[index]
  seatActionLabels.value = next
}

function setSeatAction(index: number, label: string) {
  clearSeatAction(index)
  seatActionLabels.value = { ...seatActionLabels.value, [index]: label }
  seatActionTimers.set(
    index,
    window.setTimeout(() => clearSeatAction(index), SEAT_ACTION_MS),
  )
}

function seatActionShort(event: ZhajinhuaEvent): string {
  switch (event.type) {
    case 'look':
      return '看牌'
    case 'check':
      return '过牌'
    case 'follow':
      return event.amount ? `跟注 +${event.amount}` : '跟注'
    case 'raise': {
      if (event.message) {
        const delta = event.message.match(/（\+(\d+)）/)
        if (delta) return `加注 +${delta[1]}`
      }
      return event.amount ? `加注 +${event.amount}` : '加注'
    }
    case 'fold':
      return '弃牌'
    case 'compare':
      return event.target_name ? `比 ${event.target_name}` : '比牌'
    case 'game_over':
      return event.hand_label ? `${event.hand_label} 胜` : '获胜'
    default:
      return ''
  }
}

function showSeatTimer(index: number) {
  if (isDealing.value) return false
  if (replaySeatIndex.value === index) return true
  return (
    state.value?.phase === 'betting' &&
    state.value.current_turn === index &&
    !state.value.players[index]?.folded &&
    !isAnimating.value
  )
}

function showSeatAction(index: number) {
  return (
    !!seatActionLabels.value[index] &&
    !showSeatTimer(index) &&
    replaySeatIndex.value !== index
  )
}

function seatActionLabel(index: number) {
  return seatActionLabels.value[index] ?? ''
}

function seatTimerSeconds(index: number) {
  if (replaySeatIndex.value === index) return replaySeatSeconds.value
  return secondsLeft.value
}

function seatIndicatorPlacement(index: number): 'left' | 'right' | 'top' {
  const total = state.value?.players.length ?? 1
  const rel = (index - mySeat.value + total) % total
  if (rel === 0) return 'top'
  const angle = Math.PI / 2 + (2 * Math.PI * rel) / total
  return Math.cos(angle) < 0 ? 'right' : 'left'
}

function botCountFromRoute() {
  return Math.min(7, Math.max(1, Number(route.query.bots) || 2))
}

function initPlaceholderTable(bots = botCountFromRoute()) {
  const players = [
    {
      index: 0,
      name: '我',
      is_ai: false,
      looked: false,
      folded: false,
      chips: 2000,
      bet_round: 0,
      total_bet: 0,
      card_count: 0,
    },
    ...Array.from({ length: bots }, (_, i) => ({
      index: i + 1,
      name: `电脑${i + 1}`,
      is_ai: true,
      looked: false,
      folded: false,
      chips: 2000,
      bet_round: 0,
      total_bet: 0,
      card_count: 0,
    })),
  ]
  state.value = {
    id: 'pending',
    phase: 'betting',
    players,
    human_player: 0,
    dealer_index: 0,
    current_turn: 1 % players.length,
    pot: 0,
    current_bet: 10,
    base_ante: 10,
    min_raise: 10,
    compare_cost: 10,
    message: '准备发牌…',
    hand_multipliers: DEFAULT_HAND_MULTIPLIERS,
    turn_deadline_unix: 0,
  }
}

function isFreshTable(next: ZhajinhuaState) {
  if (next.phase !== 'betting') return false
  return next.players.every((p) => !p.looked && !p.folded && p.total_bet === next.base_ante)
}

function prepareDealState(next: ZhajinhuaState): ZhajinhuaState {
  return {
    ...next,
    events: [],
    my_hand: undefined,
    message: '正在发牌…',
    pot: 0,
    current_bet: next.base_ante,
    players: next.players.map((p) => ({
      ...p,
      looked: false,
      folded: false,
      hand: undefined,
      hand_label: undefined,
      multiplier: undefined,
      bet_round: 0,
      chips: p.chips + p.bet_round,
      card_count: 0,
    })),
  }
}

function seatStyle(index: number) {
  const total = state.value?.players.length ?? 1
  const rel = (index - mySeat.value + total) % total
  const angle = Math.PI / 2 + (2 * Math.PI * rel) / total
  const rx = 42
  const ry = rel === 0 ? 28 : 34
  const yNudge = rel === 0 ? -5 : 0
  return {
    left: `${50 + rx * Math.cos(angle)}%`,
    top: `${50 + ry * Math.sin(angle) + yNudge}%`,
  }
}

function isActive(index: number) {
  if (isDealing.value) return false
  return showSeatTimer(index) || replaySeatIndex.value === index
}

function seatBackCount(index: number) {
  if (isDealing.value) return dealCounts.value[index] ?? 0
  const player = state.value?.players.find((p) => p.index === index)
  if (!player || player.folded) return 0
  if (showSeatHand(player)) return 0
  return CARDS_PER_PLAYER
}

function seatStatusLabel(player: {
  index: number
  folded: boolean
  looked: boolean
}) {
  if (player.folded) return '已弃'
  if (player.looked || (player.index === mySeat.value && state.value?.my_hand?.length)) return '已看'
  return '闷牌'
}

function seatStatusClass(player: {
  index: number
  folded: boolean
  looked: boolean
}) {
  if (player.folded) return 'zjh__tag--fold'
  if (player.looked || (player.index === mySeat.value && state.value?.my_hand?.length)) return 'zjh__tag--look'
  return 'zjh__tag--blind'
}

function seatHand(player: { index: number; hand?: Card[] }) {
  if (player.index === mySeat.value && state.value?.my_hand?.length) {
    return state.value.my_hand
  }
  return player.hand
}

function showSeatHand(player: { index: number; hand?: Card[]; folded: boolean; looked: boolean }) {
  if (isDealing.value || player.folded) return false
  const hand = seatHand(player)
  return !!hand?.length
}

function eventLabel(event: ZhajinhuaEvent): string {
  if (event.message) return event.message
  switch (event.type) {
    case 'look':
      return `${event.player_name} 看牌`
    case 'check':
      return `${event.player_name} 过牌`
    case 'follow':
      return `${event.player_name} 跟注 +${event.amount ?? 0}`
    case 'raise':
      return `${event.player_name} 加注 +${event.amount ?? 0}`
    case 'fold':
      return `${event.player_name} 弃牌`
    case 'compare':
      return `${event.player_name} 与 ${event.target_name ?? '?'} 比牌`
    case 'game_over':
      return `${event.player_name} 获胜`
    default:
      return event.player_name
  }
}

function eventDelay(event: ZhajinhuaEvent): number {
  switch (event.type) {
    case 'look':
      return 550
    case 'check':
      return 450
    case 'fold':
      return 500
    case 'follow':
      return 650
    case 'raise':
      return 750
    case 'compare':
      return 950
    case 'game_over':
      return 1100
    default:
      return 600
  }
}

async function showThinking(seatIndex: number, ms: number) {
  replaySeatSeconds.value = Math.max(5, Math.min(18, 6 + Math.floor(ms / 100)))
  if (seatIndex === mySeat.value) {
    await sleep(Math.min(ms, 280))
    return
  }
  replaySeatIndex.value = seatIndex
  await sleep(ms)
  replaySeatIndex.value = null
}

function patchStateForEvent(base: ZhajinhuaState, event: ZhajinhuaEvent): ZhajinhuaState {
  const players = base.players.map((p) => ({ ...p }))
  const p = players[event.player_index]
  if (!p) return base

  const next: ZhajinhuaState = {
    ...base,
    players,
    message: eventLabel(event),
    current_turn: event.player_index,
  }

  switch (event.type) {
    case 'look':
      p.looked = true
      break
    case 'fold':
      p.folded = true
      break
    case 'follow':
      if (event.amount && event.amount > 0) {
        p.chips = Math.max(0, p.chips - event.amount)
        p.bet_round += event.amount
        p.total_bet += event.amount
        next.pot += event.amount
      }
      break
    case 'raise':
      if (event.amount && event.amount > 0) {
        p.chips = Math.max(0, p.chips - event.amount)
        p.bet_round += event.amount
        p.total_bet += event.amount
        next.pot += event.amount
      }
      if (event.message) {
        const m = event.message.match(/至\s*(\d+)/)
        if (m) next.current_bet = Number(m[1])
      }
      break
    case 'compare':
      if (event.amount && event.amount > 0) {
        p.chips = Math.max(0, p.chips - event.amount)
        p.bet_round += event.amount
        p.total_bet += event.amount
        next.pot += event.amount
      }
      if (event.message) {
        const match = event.message.match(/，(.+) 出局/)
        if (match) {
          const loser = players.find((x) => x.name === match[1])
          if (loser) loser.folded = true
        }
      }
      break
    case 'game_over':
      next.phase = 'finished'
      next.winner_index = event.player_index
      next.win_hand_label = event.hand_label
      next.win_multiplier = event.multiplier
      if (event.hand_label) {
        p.hand_label = event.hand_label
        p.multiplier = event.multiplier
      }
      break
    default:
      break
  }

  return next
}

function patchCompareCharges(base: ZhajinhuaState, event: ZhajinhuaEvent): ZhajinhuaState {
  const players = base.players.map((p) => ({ ...p }))
  const p = players[event.player_index]
  if (!p) return base
  const next: ZhajinhuaState = {
    ...base,
    players,
    message: eventLabel(event),
    current_turn: event.player_index,
  }
  if (event.amount && event.amount > 0) {
    p.chips = Math.max(0, p.chips - event.amount)
    p.bet_round += event.amount
    p.total_bet += event.amount
    next.pot += event.amount
  }
  return next
}

function applyCompareFold(loserIndex: number) {
  if (!state.value) return
  state.value = {
    ...state.value,
    players: state.value.players.map((p) =>
      p.index === loserIndex ? { ...p, folded: true } : p,
    ),
  }
}

function parseCompareLoserIndex(event: ZhajinhuaEvent): number | null {
  const match = event.message?.match(/，(.+) 出局/)
  if (!match || !state.value) return null
  return state.value.players.find((p) => p.name === match[1])?.index ?? null
}

async function measureCompareBeam(fromIndex: number, toIndex: number) {
  await nextTick()
  const arena = arenaRef.value
  if (!arena) return null
  const pickHandAnchor = (seatIndex: number) =>
    arena.querySelector(`[data-seat="${seatIndex}"] .zjh__hand-slot`) ??
    arena.querySelector(`[data-seat="${seatIndex}"] .zjh__seat-card`)
  const fromEl = pickHandAnchor(fromIndex)
  const toEl = pickHandAnchor(toIndex)
  if (!fromEl || !toEl) return null
  const arenaRect = arena.getBoundingClientRect()
  const fromRect = fromEl.getBoundingClientRect()
  const toRect = toEl.getBoundingClientRect()
  return {
    x1: fromRect.left + fromRect.width / 2 - arenaRect.left,
    y1: fromRect.top + fromRect.height / 2 - arenaRect.top,
    x2: toRect.left + toRect.width / 2 - arenaRect.left,
    y2: toRect.top + toRect.height / 2 - arenaRect.top,
  }
}

async function shootCompareBeam(fromIndex: number, toIndex: number) {
  const coords = await measureCompareBeam(fromIndex, toIndex)
  if (!coords) {
    await sleep(COMPARE_BEAM_SHOOT_MS)
    return
  }
  const dx = coords.x2 - coords.x1
  const dy = coords.y2 - coords.y1
  const length = Math.sqrt(dx * dx + dy * dy)
  compareBeam.value = { ...coords, length, progress: 0 }
  await nextTick()

  const start = performance.now()
  await new Promise<void>((resolve) => {
    function frame(now: number) {
      const progress = Math.min(1, (now - start) / COMPARE_BEAM_SHOOT_MS)
      if (compareBeam.value) {
        compareBeam.value = { ...compareBeam.value, progress }
      }
      if (progress < 1) {
        requestAnimationFrame(frame)
      } else {
        resolve()
      }
    }
    requestAnimationFrame(frame)
  })

  compareBeam.value = null
}

async function replayCompareEvent(event: ZhajinhuaEvent) {
  const initiator = event.player_index
  const target = event.target_index ?? -1
  const humanInvolved = initiator === mySeat.value || target === mySeat.value
  const humanInitiated = initiator === mySeat.value
  const loserIndex = parseCompareLoserIndex(event)

  if (!humanInvolved) {
    if (state.value) state.value = patchStateForEvent(state.value, event)
    await showThinking(initiator, eventDelay(event))
    const short = seatActionShort(event)
    if (short) setSeatAction(initiator, short)
    return
  }

  if (state.value) {
    state.value = patchCompareCharges(state.value, event)
  }

  setSeatSpeech(initiator, '我要验牌！！！')
  replaySeatIndex.value = initiator

  if (target >= 0) {
    await shootCompareBeam(initiator, target)
  } else {
    await sleep(COMPARE_BEAM_SHOOT_MS)
  }

  replaySeatIndex.value = null
  await sleep(COMPARE_RESOLVE_MS)

  if (target >= 0) {
    setSeatAction(initiator, '比牌')
    setSeatAction(target, '比牌')
    await sleep(300)
  }

  if (loserIndex !== null) {
    applyCompareFold(loserIndex)
    await sleep(280)

    const humanLost = loserIndex === mySeat.value
    const winnerIndex = loserIndex === initiator ? target : initiator
    if (humanLost) {
      setSeatSpeech(mySeat.value, '我去擦皮鞋。。。')
      if (winnerIndex >= 0) setSeatSpeech(winnerIndex, '路边站着去！！')
    } else if (humanInitiated || target === mySeat.value) {
      setSeatSpeech(mySeat.value, '给我擦皮鞋！！！')
      setSeatSpeech(loserIndex, '牌没有问题')
    }
  }

  await sleep(400)
}

async function replayEvent(event: ZhajinhuaEvent, pendingHand?: Card[]) {
  if (event.type === 'compare') {
    await replayCompareEvent(event)
    return
  }

  if (state.value && event.type !== 'game_over') {
    state.value = patchStateForEvent(state.value, event)
    if (event.type === 'look' && event.player_index === mySeat.value && pendingHand?.length) {
      state.value = { ...state.value, my_hand: pendingHand }
      state.value.players[mySeat.value].looked = true
    }
  }

  await showThinking(event.player_index, eventDelay(event))

  const short = seatActionShort(event)
  if (short) setSeatAction(event.player_index, short)
}

async function runZjhDealAnimation() {
  if (!state.value) return
  isDealing.value = true
  dealCounts.value = Object.fromEntries(state.value.players.map((p) => [p.index, 0]))

  const total = state.value.players.length
  const start = (state.value.dealer_index + 1) % total

  for (let i = 0; i < total; i++) {
    const player = state.value.players[(start + i) % total]
    if (!player) continue
    dealCounts.value[player.index] = CARDS_PER_PLAYER
    await nextTick()
    const backs = Array.from(
      arenaRef.value?.querySelectorAll(`[data-seat="${player.index}"] .zjh__back`) ?? [],
    ) as HTMLElement[]
    if (backs.length && dealOriginRef.value) {
      await animateCardsFromCenterBatch(backs, dealOriginRef.value)
    }
    await sleep(DEAL_SEAT_MS)
  }

  isDealing.value = false
}

async function applyState(next: ZhajinhuaState, options: { skipDeal?: boolean } = {}) {
  const isNewGame = lastGameId.value !== next.id
  const shouldDeal = !options.skipDeal && isNewGame && isFreshTable(next)

  if (shouldDeal) {
    lastGameId.value = next.id
    state.value = prepareDealState(next)
    await runZjhDealAnimation()
    state.value = {
      ...next,
      events: [],
      my_hand: undefined,
      message: next.message || '底注已下，准备开始',
      players: next.players.map((p) => ({
        ...p,
        looked: false,
        folded: false,
        hand: undefined,
        hand_label: undefined,
        multiplier: undefined,
        card_count: CARDS_PER_PLAYER,
      })),
    }

    const synced = await tickZhajinhuaGame(next.id)
    await applyState(synced, { skipDeal: true })
    return
  }

  if (isNewGame) {
    lastGameId.value = next.id
  }

  const events = next.events ?? []

  if (events.length === 0) {
    state.value = next
    raiseAmount.value = Math.max(next.current_bet + next.min_raise, raiseAmount.value || 0)
    return
  }

  isAnimating.value = true
  try {
    for (const event of events) {
      await replayEvent(event, next.my_hand)
      if (next.phase === 'finished' && event.type === 'game_over') break
    }
    state.value = { ...next, events: [] }
    raiseAmount.value = Math.max(next.current_bet + next.min_raise, raiseAmount.value || 0)
  } finally {
    isAnimating.value = false
    replaySeatIndex.value = null
  }
}

async function beginSolo() {
  const bots = botCountFromRoute()
  loading.value = true
  isAnimating.value = false
  isDealing.value = false
  lastGameId.value = ''
  clearAllSeatActions()
  resetReadyState()
  initPlaceholderTable(bots)
  try {
    const next = await startZhajinhuaGame(bots)
    await applyState(next)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '开局失败')
  } finally {
    loading.value = false
  }
}

async function enterNextOnlineGame(gameId: string) {
  if (!roomId.value) return
  stopFinishedRoomPolling()
  resetReadyState()
  lastGameId.value = ''
  loading.value = true
  try {
    await router.replace({
      name: 'zhajinhua-play',
      params: { gameId },
      query: { room: roomId.value },
    })
    await applyState(await getZhajinhuaState(gameId))
    startPolling()
  } catch (err) {
    toastError(err instanceof Error ? err.message : '进入下一局失败')
  } finally {
    loading.value = false
  }
}

async function waitForNextOnlineGame() {
  if (!roomId.value || !state.value) return
  const currentGameId = state.value.id
  for (let i = 0; i < 40; i++) {
    const room = await fetchZhajinhuaRoom(roomId.value)
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
  if (!isGameFinished.value || loading.value || isHumanReady.value) return

  if (isOnline.value && roomId.value && state.value) {
    loading.value = true
    readySeats.value = { ...readySeats.value, [mySeat.value]: true }
    try {
      const room = await nextZhajinhuaRoom(roomId.value, true)
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

  const players = state.value?.players ?? []
  readySeats.value = { ...readySeats.value, [mySeat.value]: true }
  await sleep(300)
  const allReady: Record<number, boolean> = { ...readySeats.value }
  for (const p of players) {
    allReady[p.index] = true
  }
  readySeats.value = allReady
  await sleep(400)
  await beginSolo()
}

async function loadGame(gameId: string) {
  loading.value = true
  lastGameId.value = ''
  try {
    await applyState(await getZhajinhuaState(gameId))
  } catch (err) {
    toastError(err instanceof Error ? err.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function act(fn: () => Promise<ZhajinhuaState>) {
  if (!state.value || loading.value || isAnimating.value || isDealing.value) return
  loading.value = true
  try {
    await applyState(await fn())
  } catch (err) {
    toastError(err instanceof Error ? err.message : '操作失败')
  } finally {
    loading.value = false
  }
}

async function handleCompare() {
  if (!state.value || loading.value || isAnimating.value || isDealing.value) return
  if (!hasLooked.value) {
    toastError('请先点击「看牌」后再比牌')
    return
  }
  if (compareTarget.value === null) {
    toastError('请选择比牌对象')
    return
  }
  const target = state.value.players[compareTarget.value]
  if (!target || target.folded) {
    toastError('比牌对象无效')
    return
  }
  if (!target.looked) {
    toastError('对方尚未看牌，无法比牌')
    return
  }
  const me = state.value.players[mySeat.value]
  const cost = state.value.compare_cost ?? 10
  if (me && me.chips < cost) {
    toastError(`比牌需要 ${cost} 币，当前筹码不足`)
    return
  }
  await act(() => zhajinhuaCompare(state.value!.id, compareTarget.value!))
}

function startPolling() {
  stopPolling()
  if (!isOnline.value) return
  pollTimer = window.setInterval(async () => {
    if (
      !state.value ||
      loading.value ||
      isAnimating.value ||
      isDealing.value ||
      state.value.phase === 'finished'
    ) {
      return
    }
    try {
      const next = await getZhajinhuaState(state.value.id)
      if (next.current_turn !== state.value.current_turn || (next.events?.length ?? 0) > 0) {
        loading.value = true
        try {
          await applyState(next, { skipDeal: true })
        } finally {
          loading.value = false
        }
      }
    } catch {
      /* ignore */
    }
  }, 1500)
}

function stopPolling() {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

function startFinishedRoomPolling() {
  stopFinishedRoomPolling()
  if (!isOnline.value || !roomId.value) return

  const poll = async () => {
    if (!roomId.value || !state.value || state.value.phase !== 'finished') return
    try {
      const room = await fetchZhajinhuaRoom(roomId.value)
      syncReadyFromRoom(room)
      if (room.game_id && room.game_id !== state.value.id) {
        await enterNextOnlineGame(room.game_id)
      }
    } catch {
      /* ignore */
    }
  }

  poll()
  finishedRoomPollTimer = window.setInterval(poll, 1500)
}

function stopFinishedRoomPolling() {
  if (finishedRoomPollTimer !== null) {
    window.clearInterval(finishedRoomPollTimer)
    finishedRoomPollTimer = null
  }
}

onMounted(async () => {
  if (isOnline.value && route.params.gameId) {
    await loadGame(String(route.params.gameId))
    startPolling()
  } else {
    await beginSolo()
  }
})

onUnmounted(() => {
  stopPolling()
  stopFinishedRoomPolling()
})

watch(
  () => state.value?.phase,
  (phase) => {
    if (phase === 'finished') {
      resetReadyState()
      if (isOnline.value && roomId.value) {
        stopPolling()
        startFinishedRoomPolling()
      }
    } else {
      stopFinishedRoomPolling()
    }
  },
)
</script>

<template>
  <main class="zjh">
    <header class="zjh__header">
      <button type="button" class="ddz__back" @click="router.push('/games/zhajinhua')">← 返回</button>
      <div>
        <h1>扎金花</h1>
        <p class="ddz__subtitle">{{ isOnline ? '多人联机' : '单机对战电脑' }}</p>
      </div>
      <button
        v-if="!isOnline && !isGameFinished"
        type="button"
        class="ddz__restart"
        :disabled="loading || isAnimating || isDealing"
        @click="beginSolo"
      >
        重新开局
      </button>
      <div v-else class="zjh__header-spacer" aria-hidden="true" />
    </header>

    <section v-if="state" class="zjh__table">
      <div ref="arenaRef" class="zjh__arena">
        <button
          type="button"
          class="zjh__help-btn"
          title="牌型倍率"
          @click="showMultiplierHelp = !showMultiplierHelp"
        >
          ?
        </button>
        <div v-if="showMultiplierHelp" class="zjh__help-popover">
          <h3>牌型倍率</h3>
          <ul>
            <li v-for="row in multiplierRows" :key="row.key">
              <span>{{ row.label }}</span>
              <strong>×{{ row.value }}</strong>
            </li>
          </ul>
        </div>

        <div ref="dealOriginRef" class="zjh__deck-origin">
          <div v-show="isDealing" class="ddz__deck">
            <span class="ddz__deck-card" />
            <span>发牌中</span>
          </div>
        </div>

        <div class="zjh__pot">
          <span class="zjh__pot-label">底池</span>
          <strong>{{ state.pot }}</strong>
          <span class="zjh__pot-sub">当前注 {{ state.current_bet }}</span>
        </div>

        <svg v-if="compareBeam" class="zjh__compare-beam" aria-hidden="true">
          <line
            :x1="compareBeam.x1"
            :y1="compareBeam.y1"
            :x2="compareBeam.x2"
            :y2="compareBeam.y2"
            class="zjh__compare-beam-line"
            :style="compareBeamLineStyle"
          />
          <circle
            :cx="compareBeamHead.x"
            :cy="compareBeamHead.y"
            r="5"
            class="zjh__compare-beam-head"
          />
        </svg>

        <div
          v-for="player in state.players"
          :key="player.index"
          class="zjh__seat"
          :data-seat="player.index"
          :class="{
            'zjh__seat--active': isActive(player.index),
            'zjh__seat--folded': player.folded,
            'zjh__seat--self': player.index === mySeat,
            'zjh__seat--thinking': replaySeatIndex === player.index,
          }"
          :style="seatStyle(player.index)"
        >
          <div class="zjh__seat-stack">
            <div v-if="seatSpeech(player.index)" class="zjh__seat-speech">{{ seatSpeech(player.index) }}</div>
            <div class="zjh__seat-card">
              <strong class="zjh__seat-name">{{ player.index === mySeat ? '我' : player.name }}</strong>
              <span v-if="isGameFinished && readySeats[player.index]" class="ddz__ready-badge">准备</span>
              <span v-else class="zjh__tag" :class="seatStatusClass(player)">{{ seatStatusLabel(player) }}</span>
              <span class="zjh__chips">{{ player.chips }} 币</span>
              <span class="zjh__bet-line">
                <span v-if="player.bet_round > 0" class="zjh__bet">注 {{ player.bet_round }}</span>
                <span v-else class="zjh__bet zjh__bet--placeholder">&nbsp;</span>
              </span>
              <span class="zjh__hand-type-slot">
                <span v-if="player.hand_label" class="zjh__hand-type">{{ player.hand_label }} ×{{ player.multiplier }}</span>
              </span>
            </div>
            <SeatIndicator
              :placement="seatIndicatorPlacement(player.index)"
              :seconds="seatTimerSeconds(player.index)"
              :show-timer="showSeatTimer(player.index)"
              :action-label="showSeatAction(player.index) ? seatActionLabel(player.index) : undefined"
            />
          </div>
          <div class="zjh__hand-slot">
            <div v-if="showSeatHand(player)" class="zjh__mini-hand">
              <PlayingCard
                v-for="c in seatHand(player)"
                :key="c.id"
                :card="c"
                stacked
                mini
              />
            </div>
            <div v-else-if="!player.folded" class="zjh__backs">
              <span
                v-for="n in CARDS_PER_PLAYER"
                :key="`${player.index}-${n}`"
                class="zjh__back"
                :class="{ 'zjh__back--empty': n > seatBackCount(player.index) }"
              />
            </div>
          </div>
        </div>

        <div class="zjh__prompt">
          <div v-if="state.phase === 'finished'" class="zjh__settle">
            <div class="zjh__result">
              胜者：{{ state.players[state.winner_index ?? 0]?.name }}
              · {{ state.win_hand_label }} ×{{ state.win_multiplier }}
            </div>
            <p v-if="isHumanReady && !isOnline" class="zjh__settle-hint">电脑已准备，即将发牌…</p>
            <p v-else-if="isHumanReady && isOnline" class="zjh__settle-hint">已准备 · 等待其他玩家</p>
            <button
              v-if="showSettleReady"
              type="button"
              class="ddz__btn ddz__btn--primary"
              :disabled="loading || isDealing || isAnimating"
              @click="handleReady"
            >
              准备
            </button>
          </div>
          <div v-else-if="!isDealing && (canLook || canAct)" class="zjh__actions">
            <button
              v-if="canLook"
              type="button"
              class="ddz__btn ddz__btn--hint"
              :disabled="loading || isAnimating"
              @click="act(() => zhajinhuaLook(state!.id))"
            >
              看牌
            </button>
            <template v-if="canAct">
              <button type="button" class="ddz__btn" :disabled="loading" @click="act(() => zhajinhuaFollow(state!.id))">
                跟注
              </button>
              <div class="zjh__raise">
                <input v-model.number="raiseAmount" type="number" :min="state.current_bet + state.min_raise" />
                <button
                  type="button"
                  class="ddz__btn ddz__btn--primary"
                  :disabled="loading"
                  @click="act(() => zhajinhuaRaise(state!.id, raiseAmount))"
                >
                  加注
                </button>
              </div>
              <div class="zjh__compare">
                <select v-model.number="compareTarget" :disabled="!canPickCompareTarget">
                  <option :value="null" disabled>
                    {{ canPickCompareTarget ? '比牌对象' : '暂无可比牌玩家' }}
                  </option>
                  <option
                    v-for="p in compareTargets.filter((x) => x.looked)"
                    :key="p.index"
                    :value="p.index"
                  >
                    {{ p.name }}
                  </option>
                </select>
                <button
                  type="button"
                  class="ddz__btn ddz__btn--hint"
                  :disabled="loading || compareTarget === null"
                  @click="handleCompare"
                >
                  比牌
                </button>
              </div>
              <button type="button" class="ddz__btn" :disabled="loading" @click="act(() => zhajinhuaFold(state!.id))">
                弃牌
              </button>
            </template>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>
