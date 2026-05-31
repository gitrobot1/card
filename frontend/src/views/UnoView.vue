<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import UnoCard from '../components/uno/UnoCard.vue'
import UnoHand from '../components/uno/UnoHand.vue'
import SeatIndicator from '../components/doudizhu/SeatIndicator.vue'
import {
  drawUnoCard,
  playUnoCard,
  startUnoGame,
} from '../api/games'
import { animateUnoDealToSeat, animateUnoDrawEvent, animateUnoPlayEvent, animateUnoRevealTopCard } from '../composables/useUnoPlayAnimation'
import { showToast } from '../composables/useToast'
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

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

const router = useRouter()
const route = useRoute()

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

const BOTTOM_ZONE_HEIGHT = 210
const PROMPT_HEIGHT = 88

const promptFixedStyle = ref<Record<string, string>>({
  visibility: 'hidden',
})

let tableResizeObserver: ResizeObserver | null = null
let timeoutTriggered = false

const secondsLeft = ref(20)

/** 单机模式：人类回合计时仅前端展示；联机时再对接服务端 deadline + tick */
const SOLO_TURN_SECONDS = 20
const isSoloMode = true

let soloTimerId: ReturnType<typeof setInterval> | null = null
let soloDeadlineAt = 0

const mySeat = computed(() => state.value?.human_player ?? 0)
const myHand = computed(() =>
  isDealing.value || (isAnimating.value && displayedHand.value.length > 0)
    ? displayedHand.value
    : (state.value?.my_hand ?? []),
)
const topCard = computed(() => {
  if (isDealing.value) return null
  return displayedTopCard.value ?? state.value?.top_card ?? null
})
const isMyTurn = computed(
  () => state.value?.phase === 'playing' && state.value.current_turn === mySeat.value,
)
const isFinished = computed(() => state.value?.phase === 'finished')

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

const showPlayButtons = computed(() => canAct.value && !showColorPicker.value)
const showColorPickerPrompt = computed(() => showColorPicker.value && isMyTurn.value)

const isHumanWinner = computed(
  () => isFinished.value && state.value?.winner_index === mySeat.value,
)

const settleTitle = computed(() => {
  if (!isFinished.value) return ''
  return isHumanWinner.value ? '你赢了！' : '你输了'
})

const settleSubtitle = computed(() => {
  if (!state.value || !isFinished.value) return ''
  const winner = state.value.players.find((p) => p.index === state.value?.winner_index)
  return `${winner?.name ?? '玩家'} 获胜`
})

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
  () => showPlayButtons.value || showColorPickerPrompt.value || showSettleReady.value,
)

const promptMyTurn = computed(
  () => showPlayButtons.value || showColorPickerPrompt.value || showSettleReady.value,
)

const promptBannerActive = computed(
  () => showPlayButtons.value || showColorPickerPrompt.value || showSettleReady.value,
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
  if (showColorPicker.value) return '选择要打出的颜色'
  if (canAct.value) {
    if (
      (state.value.pending_draw_penalty ?? 0) > 0 ||
      state.value.must_play_after_stack
    ) {
      return state.value.message
    }
    return `轮到你 · 剩余 ${secondsLeft.value} 秒`
  }
  return '\u00a0'
})

const centerPlayName = computed(() => {
  if (!centerPlay.value) return ''
  return centerPlay.value.player_index === mySeat.value ? '我' : centerPlay.value.player_name
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

function isUnoAlert(index: number) {
  return handCountForSeat(index) === 1 && !isFinished.value
}

function seatPlayerClass(index: number) {
  return {
    'ddz__player--active': isSeatHighlighted(index),
    'uno__seat--alert': isUnoAlert(index),
  }
}

function selfPlayerClass() {
  return {
    'ddz__player--active': isMyTurn.value && !isFinished.value,
    'uno__seat--alert': isUnoAlert(mySeat.value),
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

function showSeatTimer(index: number) {
  if (isDealing.value) return false
  if (replaySeatIndex.value === index) return !replaySeatActionLabel.value
  if (index === mySeat.value) {
    return isMyTurn.value && !isFinished.value && !isAnimating.value
  }
  return false
}

function showSeatActionLabel(index: number) {
  if (replaySeatIndex.value === index) return replaySeatActionLabel.value
  return undefined
}

function seatTimerSeconds(index: number) {
  if (replaySeatIndex.value === index) return replaySeatSeconds.value
  return secondsLeft.value
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

function clearSoloTurnTimer() {
  if (soloTimerId) {
    clearInterval(soloTimerId)
    soloTimerId = null
  }
}

function resetSoloTurnTimer() {
  clearSoloTurnTimer()
  if (
    !isSoloMode ||
    !isMyTurn.value ||
    isAnimating.value ||
    isDealing.value ||
    isFinished.value
  ) {
    return
  }
  soloDeadlineAt = Date.now() + SOLO_TURN_SECONDS * 1000
  secondsLeft.value = SOLO_TURN_SECONDS
  soloTimerId = setInterval(() => {
    const left = Math.max(0, Math.ceil((soloDeadlineAt - Date.now()) / 1000))
    secondsLeft.value = left
    if (left <= 0 && !timeoutTriggered && canAct.value) {
      timeoutTriggered = true
      void handleSoloTurnTimeout()
    }
  }, 200)
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
      isSoloMode &&
      isMyTurn.value &&
      canAct.value &&
      !loading.value &&
      !isAnimating.value &&
      !isDealing.value
    ) {
      resetSoloTurnTimer()
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
  displayedTopCard.value = next.top_card
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

  if (origin && discardAreaRef.value) {
    await animateUnoRevealTopCard(
      origin,
      discardAreaRef.value,
      next.top_card.label,
      unoColorClass(next.top_card.color),
    )
  }

  displayedTopCard.value = next.top_card
  state.value = { ...next, events: [] }
  syncDisplayFromState(next)
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
      event.type === 'draw' && isOpponent
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
        message: event.message || state.value!.message,
        players: state.value!.players.map((p) =>
          p.index === event.player_index
            ? { ...p, hand_count: Math.max(0, p.hand_count - 1) }
            : p,
        ),
      }
    })
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

  if (events.length === 0) {
    if (isNewGame) {
      lastGameId.value = next.id
      await runUnoDealAnimation(next)
      return
    }
    state.value = next
    syncDisplayFromState(next)
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

    state.value = { ...next, events: [] }
    syncDisplayFromState(next)
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

async function handleReady() {
  if (!isFinished.value || loading.value || isHumanReady.value) return

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
  resetReadyState()
  lastGameId.value = ''
  try {
    await applyState(await startUnoGame(botCountFromRoute()))
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
      toastError('对局已失效（后端可能已重启），正在重新开局…')
      await beginGame()
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

watch(
  () => [
    isMyTurn.value,
    isAnimating.value,
    isDealing.value,
    state.value?.current_turn,
    state.value?.phase,
    state.value?.must_play_after_stack,
    state.value?.pending_draw_penalty,
  ],
  () => {
    timeoutTriggered = false
    clearSoloTurnTimer()
    if (
      isSoloMode &&
      isMyTurn.value &&
      !isAnimating.value &&
      !isDealing.value &&
      state.value?.phase === 'playing'
    ) {
      resetSoloTurnTimer()
    }
  },
  { immediate: true },
)

watch(
  () => state.value,
  async () => {
    await nextTick()
    bindPromptPositionObserver()
  },
)

watch(promptVisible, async () => {
  await nextTick()
  updatePromptFixedPosition()
})

onMounted(async () => {
  window.addEventListener('resize', updatePromptFixedPosition)
  window.addEventListener('scroll', updatePromptFixedPosition, true)
  await beginGame()
  await nextTick()
  bindPromptPositionObserver()
})

onUnmounted(() => {
  window.removeEventListener('resize', updatePromptFixedPosition)
  window.removeEventListener('scroll', updatePromptFixedPosition, true)
  tableResizeObserver?.disconnect()
  clearSoloTurnTimer()
})
</script>

<template>
  <main class="ddz app">
    <header class="ddz__header">
      <button type="button" class="ddz__back" @click="router.push('/games/uno')">← 返回</button>
      <div>
        <h1>UNO</h1>
        <p class="ddz__subtitle">单机对战电脑 · {{ state?.players.length ?? 0 }} 人</p>
      </div>
      <button
        v-if="!isFinished"
        type="button"
        class="ddz__restart"
        :disabled="loading || isAnimating || isDealing"
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
                <span :class="handCountClass(seatIndex)">{{ handCountForSeat(seatIndex) }} 张</span>
                <span v-if="isFinished && readySeats[seatIndex]" class="ddz__ready-badge">准备</span>
              </div>
              <SeatIndicator
                :placement="seatIndicatorPlacement(seatIndex)"
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
                v-else
                class="ddz__play-by"
                :class="{ 'uno__play-by--hidden': !centerPlay || isDealing }"
              >
                <span>{{ centerPlayName || '\u00a0' }}</span>
                <span class="ddz__play-by-action">出牌</span>
              </p>
            </div>
            <p v-if="isDealing" class="uno__dealing-hint">发牌中…</p>
            <div class="uno__piles" :class="{ 'uno__piles--dealing': isDealing }">
              <div ref="drawAreaRef" class="uno__pile">
                <UnoCard
                  :card="{ id: 'back', color: 'wild', value: 'back', label: '牌堆' }"
                  face-down
                  mini
                />
                <span class="uno__pile-count">剩余 {{ state.draw_count }} 张</span>
              </div>
              <div
                v-if="!isDealing"
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
                <span :class="handCountClass(mySeat)">{{ handCountForSeat(mySeat) }} 张</span>
                <span v-if="isFinished && readySeats[mySeat]" class="ddz__ready-badge">准备</span>
              </div>
              <SeatIndicator
                placement="top"
                :show-timer="showSeatTimer(mySeat)"
                :seconds="seatTimerSeconds(mySeat)"
              />
            </div>
          </div>
          <UnoHand
            :cards="myHand"
            :selected-id="selectedId"
            :interactive="isMyTurn && !loading && !isAnimating && !isDealing"
            :hoverable="!loading && !isDealing"
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
