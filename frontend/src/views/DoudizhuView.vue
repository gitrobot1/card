<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PlayingCard from '../components/doudizhu/PlayingCard.vue'
import HandCards from '../components/doudizhu/HandCards.vue'
import StackedCards from '../components/doudizhu/StackedCards.vue'
import SeatIndicator from '../components/doudizhu/SeatIndicator.vue'
import {
  callLandlord,
  fetchDouDizhuHint,
  fetchDouDizhuRoom,
  getDouDizhuState,
  nextDouDizhuRoom,
  passTurn,
  playCards,
  startDouDizhuGame,
  tickDouDizhuGame,
} from '../api/games'
import { animateCardsFromCenter, animateOpponentDeal } from '../composables/useDealAnimation'
import {
  animatePlayEvent,
  removeCardsFromHand,
  showCallBubble,
} from '../composables/usePlayAnimation'
import { useTurnTimer } from '../composables/useTurnTimer'
import type { Card, DouDizhuState, GameEvent, PlayRecord } from '../types/doudizhu'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const error = ref('')
const hintMessage = ref('')
const state = ref<DouDizhuState | null>(null)
const selectedIds = ref<string[]>([])
const hintIds = ref<string[]>([])

const displayedHand = ref<Card[]>([])
const isDealing = ref(false)
const isAnimating = ref(false)
const opponentCounts = ref({ left: 0, right: 0 })
const centerPlay = ref<PlayRecord | null>(null)
const actionToast = ref('')
const readySeats = ref<Record<number, boolean>>({ 0: false, 1: false, 2: false })
const passLabels = ref<Record<number, boolean>>({})
const replaySeatIndex = ref<number | null>(null)
const replaySeatSeconds = ref(35)

const PASS_LABEL_MS = 10000
const passLabelTimers = new Map<number, number>()

const dealOriginRef = ref<HTMLElement | null>(null)
const topDeckRef = ref<HTMLElement | null>(null)
const playAreaRef = ref<HTMLElement | null>(null)
const handRowRef = ref<InstanceType<typeof HandCards> | null>(null)
const tableRef = ref<HTMLElement | null>(null)
const lastGameId = ref('')

const BOTTOM_ZONE_HEIGHT = 210
const PROMPT_HEIGHT = 88

const promptFixedStyle = ref<Record<string, string>>({
  visibility: 'hidden',
})

let tableResizeObserver: ResizeObserver | null = null
let onlinePollTimer: number | null = null
let finishedRoomPollTimer: number | null = null

const isOnline = computed(() => route.name === 'doudizhu-play')
const roomId = computed(() => {
  const value = route.query.room
  return typeof value === 'string' ? value : ''
})
const mySeatIndex = computed(() => state.value?.human_player ?? 0)
const leftSeatIndex = computed(() => (mySeatIndex.value + 2) % 3)
const rightSeatIndex = computed(() => (mySeatIndex.value + 1) % 3)
const modeSubtitle = computed(() =>
  isOnline.value ? '多人联机 · 3人对战' : '单机对战电脑',
)

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
    visibility: 'visible',
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

const activeTurnIndex = computed(() => {
  if (!state.value || state.value.phase === 'finished') return -1
  if (state.value.phase === 'calling') return state.value.calling_index
  return state.value.current_turn
})

const isMyTurn = computed(() => {
  if (!state.value || isDealing.value) return false
  if (state.value.phase === 'finished') return false
  if (state.value.phase === 'calling') {
    return state.value.calling_index === state.value.human_player
  }
  return state.value.phase === 'playing' && state.value.current_turn === state.value.human_player
})

const canAct = computed(() => isMyTurn.value && !isAnimating.value)

const canPass = computed(() => {
  if (!state.value?.last_play) return false
  return state.value.last_play.player_index !== state.value.human_player
})

const isGameFinished = computed(() => state.value?.phase === 'finished')

const isHumanWinner = computed(
  () =>
    isGameFinished.value &&
    state.value?.winner_index != null &&
    state.value.winner_index === state.value.human_player,
)

const settleTitle = computed(() => {
  if (!isGameFinished.value) return ''
  return isHumanWinner.value ? '你赢了！' : '你输了'
})

const settleSubtitle = computed(() => {
  if (!state.value || !isGameFinished.value) return ''
  const role = state.value.winner_role === 'landlord' ? '地主' : '农民'
  const winner = state.value.players.find((p) => p.index === state.value?.winner_index)
  return `${winner?.name ?? '玩家'} 作为${role}获胜`
})

const canSelectCards = computed(
  () => !!state.value && state.value.phase === 'playing' && !isDealing.value,
)

const canHint = computed(
  () => !isOnline.value && state.value?.phase === 'playing' && isMyTurn.value,
)

const { secondsLeft } = useTurnTimer(state, canAct, async () => {
  if (!state.value || isAnimating.value || isGameFinished.value) return
  try {
    const next = await tickDouDizhuGame(state.value.id)
    await applyGameState(next)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '超时处理失败'
  }
})

const promptShowActions = computed(
  () =>
    !!state.value &&
    !isGameFinished.value &&
    !isDealing.value &&
    canAct.value &&
    (state.value.phase === 'calling' || state.value.phase === 'playing'),
)

const showCallButtons = computed(
  () => promptShowActions.value && state.value?.phase === 'calling',
)

const showPlayButtons = computed(
  () => promptShowActions.value && state.value?.phase === 'playing',
)

const isHumanReady = computed(() => readySeats.value[mySeatIndex.value] ?? false)

const allReady = computed(
  () =>
    isGameFinished.value &&
    readySeats.value[0] &&
    readySeats.value[1] &&
    readySeats.value[2],
)

const showSettleReady = computed(
  () => isGameFinished.value && !loading.value && !isHumanReady.value,
)

const promptBannerActive = computed(() => promptShowActions.value || showSettleReady.value)

const promptMyTurn = computed(() => showPlayButtons.value || showCallButtons.value)

const bottomCards = computed(() => state.value?.bottom_cards ?? [])

const showBottomCards = computed(
  () =>
    bottomCards.value.length > 0 &&
    !!state.value &&
    state.value.phase !== 'calling' &&
    !isDealing.value,
)

const centerPlayRole = computed(() => {
  if (!centerPlay.value) return null
  return seatRole(centerPlay.value.player_index)
})

const centerPlayName = computed(() => {
  if (!centerPlay.value) return ''
  return centerPlay.value.player_index === state.value?.human_player
    ? '我'
    : centerPlay.value.player_name
})

const myHand = computed(() => state.value?.my_hand ?? [])

const revealedHands = computed(() => state.value?.revealed_hands ?? [])

function copyHand(hand: Card[] | null | undefined) {
  return hand ? [...hand] : []
}

const promptStatusText = computed(() => {
  if (!state.value) return '\u00a0'
  if (isGameFinished.value) {
    if (allReady.value) return '即将开始...'
    if (isHumanReady.value) return '已准备 · 等待其他玩家'
    return settleSubtitle.value ? `${settleTitle.value} · ${settleSubtitle.value}` : settleTitle.value
  }
  if (canAct.value && !isDealing.value) {
    return `轮到你 · 剩余 ${secondsLeft.value} 秒`
  }
  if (isDealing.value) return '发牌中...'
  if (isAnimating.value && actionToast.value) return actionToast.value
  if (isAnimating.value) return '出牌回放中...'
  if (actionToast.value) return actionToast.value
  if (hintMessage.value) return hintMessage.value
  if (state.value.message) return state.value.message
  const name = seatByIndex(activeTurnIndex.value)?.name || '玩家'
  return `${name} 思考中 · ${secondsLeft.value}s`
})

onMounted(async () => {
  window.addEventListener('resize', updatePromptFixedPosition)
  window.addEventListener('scroll', updatePromptFixedPosition, true)
  if (isOnline.value) {
    const gameId = route.params.gameId
    if (typeof gameId === 'string' && gameId) {
      await loadGame(gameId)
    }
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
  clearAllPassLabels()
})

watch(activeTurnIndex, (index) => {
  if (index >= 0) {
    clearSeatPass(index)
  }
})

function clearPassLabelTimer(index: number) {
  const timer = passLabelTimers.get(index)
  if (timer !== undefined) {
    window.clearTimeout(timer)
    passLabelTimers.delete(index)
  }
}

function clearSeatPass(index: number) {
  if (!passLabels.value[index]) return
  const next = { ...passLabels.value }
  delete next[index]
  passLabels.value = next
  clearPassLabelTimer(index)
}

function clearAllPassLabels() {
  for (const index of passLabelTimers.keys()) {
    clearPassLabelTimer(index)
  }
  passLabels.value = {}
}

function markSeatPassed(index: number) {
  clearPassLabelTimer(index)
  passLabels.value = { ...passLabels.value, [index]: true }
  passLabelTimers.set(
    index,
    window.setTimeout(() => clearSeatPass(index), PASS_LABEL_MS),
  )
}

function showSeatTimer(index: number) {
  if (replaySeatIndex.value === index) return true
  return isSeatActive(index)
}

function showSeatPass(index: number) {
  return (
    !!passLabels.value[index] &&
    !isSeatActive(index) &&
    replaySeatIndex.value !== index
  )
}

function seatTimerSeconds(index: number) {
  if (replaySeatIndex.value === index) return replaySeatSeconds.value
  return secondsLeft.value
}

function isSeatHighlighted(index: number) {
  if (replaySeatIndex.value === index) return true
  return isSeatActive(index)
}

function humanSeatIndex() {
  return state.value?.human_player ?? mySeatIndex.value
}

async function showAIThinking(seatIndex: number, ms: number) {
  if (seatIndex === humanSeatIndex()) return
  replaySeatIndex.value = seatIndex
  replaySeatSeconds.value = Math.max(5, Math.min(18, 6 + Math.floor(ms / 100)))
  await sleep(ms)
  replaySeatIndex.value = null
}

function seatRole(index: number): 'landlord' | 'farmer' | null {
  if (!state.value || state.value.phase === 'calling') return null
  const seat = seatByIndex(index)
  if (!seat) return null
  const hasLandlord = state.value.players.some((p) => p.is_landlord)
  if (!hasLandlord && state.value.phase !== 'finished') return null
  return seat.is_landlord ? 'landlord' : 'farmer'
}

function seatRoleLabel(index: number) {
  return seatRole(index) === 'landlord' ? '地主' : '农民'
}

watch(
  () => state.value,
  async () => {
    await nextTick()
    bindPromptPositionObserver()
  },
)

watch([isGameFinished], async () => {
  await nextTick()
  updatePromptFixedPosition()
  if (isGameFinished.value && isOnline.value && roomId.value) {
    startFinishedRoomPolling()
  } else {
    stopFinishedRoomPolling()
  }
})

watch(
  () => state.value?.my_hand,
  async (hand) => {
    if (!state.value || isDealing.value) return

    if (state.value.phase === 'finished') {
      displayedHand.value = hand ? [...hand] : []
      selectedIds.value = []
      return
    }

    if (isAnimating.value) return

    if (!hand?.length) {
      displayedHand.value = []
      selectedIds.value = []
      return
    }

    const handIdSet = new Set(hand.map((c) => c.id))
    selectedIds.value = selectedIds.value.filter((id) => handIdSet.has(id))

    if (state.value.id !== lastGameId.value) {
      return
    }

    const prevIds = new Set(displayedHand.value.map((c) => c.id))
    const newCards = hand.filter((c) => !prevIds.has(c.id))
    if (newCards.length > 0) {
      isDealing.value = true
      for (const card of newCards) {
        displayedHand.value.push(card)
        await nextTick()
        await animateLastCardFromCenter()
        await sleep(80)
      }
      displayedHand.value = [...hand]
      isDealing.value = false
      return
    }

    displayedHand.value = [...hand]
  },
)

function applyFinishedState(next: DouDizhuState) {
  isAnimating.value = false
  loading.value = false
  resetReadyState()
  state.value = {
    ...next,
    phase: 'finished',
    events: [],
  }
  centerPlay.value = next.last_play
  displayedHand.value = copyHand(next.my_hand)
  hintMessage.value = next.message
  selectedIds.value = []
  hintIds.value = []
  actionToast.value = ''
}

function isFinishedState(next: DouDizhuState) {
  return next.phase === 'finished' || next.winner_index != null
}

async function applyGameState(next: DouDizhuState) {
  if (isFinishedState(next)) {
    applyFinishedState(next)
    return
  }

  const events = next.events ?? []
  const pendingHand = copyHand(next.my_hand)
  const isNewGame = lastGameId.value !== next.id

  if (events.length === 0 && isNewGame) {
    lastGameId.value = next.id
    state.value = { ...next, events: [] }
    hintMessage.value = next.message
    centerPlay.value = next.last_play
    await runDealAnimation(pendingHand)
    return
  }

  const prePlayHand = state.value ? copyHand(state.value.my_hand) : [...displayedHand.value]
  const prevPhase = state.value?.phase
  const hasHumanPlay = events.some((event) => event.type === 'play' && event.player_index === 0)
  const deferLandlordHand =
    prevPhase === 'calling' &&
    next.phase === 'playing' &&
    humanIsLandlord(next)

  if (events.length > 0) {
    isAnimating.value = true
  }

  state.value = { ...next, events: [] }
  hintMessage.value = next.message

  if (events.length === 0) {
    displayedHand.value = pendingHand
    centerPlay.value = next.last_play
    return
  }

  displayedHand.value = hasHumanPlay || deferLandlordHand ? prePlayHand : pendingHand
  centerPlay.value = null
  try {
    for (const event of events) {
      if (event.type === 'game_over') continue
      await replayEvent(event)
      if (isFinishedState(next)) break
    }
  } finally {
    isAnimating.value = false
    centerPlay.value = next.last_play
    const handIdSet = new Set(pendingHand.map((c) => c.id))
    selectedIds.value = selectedIds.value.filter((id) => handIdSet.has(id))

    if (deferLandlordHand) {
      const oldIds = new Set(prePlayHand.map((c) => c.id))
      const newCards = pendingHand.filter((c) => !oldIds.has(c.id))
      if (newCards.length > 0) {
        await animateLandlordBottomCards(newCards, pendingHand)
      } else {
        displayedHand.value = pendingHand
      }
    } else {
      displayedHand.value = pendingHand
    }
  }
}

async function replayEvent(event: GameEvent) {
  if (event.type === 'game_over') return

  actionToast.value =
    event.type === 'play'
      ? `${event.player_name} 出牌`
      : event.type === 'pass'
        ? `${event.player_name} 不出`
        : event.type === 'call'
          ? `${event.player_name} ${event.call ? '抢地主' : '不抢'}`
          : ''

  await showAIThinking(
    event.player_index,
    event.type === 'pass' ? 220 : event.type === 'call' ? 180 : 320,
  )

  if (event.type === 'play' && event.cards?.length) {
    if (event.player_index === humanSeatIndex()) {
      displayedHand.value = removeCardsFromHand(displayedHand.value, event.cards.map((c) => c.id))
    }
    const playArea = playAreaRef.value
    if (!playArea) {
      centerPlay.value = {
        player_index: event.player_index,
        player_name: event.player_name,
        cards: event.cards ?? [],
        pattern: 'play',
      }
      return
    }
    await animatePlayEvent(event, playArea, () => {
      centerPlay.value = {
        player_index: event.player_index,
        player_name: event.player_name,
        cards: event.cards ?? [],
        pattern: 'play',
      }
    })
    return
  }

  if (event.type === 'pass') {
    centerPlay.value = null
    markSeatPassed(event.player_index)
    await sleep(180)
    return
  }

  if (event.type === 'call') {
    await showCallBubble(event)
  }
}

function sortHand(cards: Card[]) {
  return [...cards].sort((a, b) => {
    if (b.rank !== a.rank) return b.rank - a.rank
    return String(a.suit).localeCompare(String(b.suit))
  })
}

function humanIsLandlord(game: DouDizhuState) {
  return game.players.some((p) => p.index === game.human_player && p.is_landlord)
}

async function animateLandlordBottomCards(newCards: Card[], fullHand: Card[]) {
  isDealing.value = true
  const origin = topDeckRef.value ?? dealOriginRef.value
  if (!origin) {
    displayedHand.value = fullHand
    isDealing.value = false
    return
  }

  const ordered = sortHand(newCards)
  let current = [...displayedHand.value]

  for (const card of ordered) {
    current = sortHand([...current, card])
    displayedHand.value = current
    await nextTick()

    const row = handRowRef.value?.rowRef
    const originCard = topDeckRef.value?.querySelector<HTMLElement>(
      `[data-bottom-card-id="${card.id}"] .playing-card`,
    )
    const flyOrigin = originCard ?? origin

    if (row) {
      const slot = row.querySelector<HTMLElement>(`.hand-cards__slot[data-card-id="${card.id}"]`)
      const cardEl = slot?.querySelector<HTMLElement>('.playing-card')
      if (cardEl) {
        await animateCardsFromCenter([cardEl], flyOrigin, 0)
        if (originCard) {
          originCard.style.opacity = '0'
        }
      }
    }
    await sleep(100)
  }

  displayedHand.value = fullHand
  isDealing.value = false
}

async function animateLastCardFromCenter() {
  const row = handRowRef.value?.rowRef
  const origin = dealOriginRef.value
  if (!row || !origin) return
  const slots = row.querySelectorAll<HTMLElement>('.hand-cards__slot')
  const lastSlot = slots[slots.length - 1]
  const lastCard = lastSlot?.querySelector<HTMLElement>('.playing-card')
  if (lastCard) {
    await animateCardsFromCenter([lastCard], origin, 0)
  }
}

async function runDealAnimation(hand: Card[]) {
  isDealing.value = true
  displayedHand.value = []
  hintIds.value = []
  selectedIds.value = []
  centerPlay.value = null
  opponentCounts.value = { left: 0, right: 0 }

  const leftTarget =
    state.value?.players.find((p) => p.index === leftSeatIndex.value)?.hand_count ?? 17
  const rightTarget =
    state.value?.players.find((p) => p.index === rightSeatIndex.value)?.hand_count ?? 17

  const opponentPromise = Promise.all([
    animateOpponentDeal(leftTarget, (current) => {
      opponentCounts.value.left = current
    }),
    animateOpponentDeal(rightTarget, (current) => {
      opponentCounts.value.right = current
    }),
  ])

  for (const card of hand) {
    displayedHand.value.push(card)
    await nextTick()
    const row = handRowRef.value?.rowRef
    const origin = dealOriginRef.value
    if (row && origin) {
      const slot = row.querySelector<HTMLElement>(`.hand-cards__slot[data-card-id="${card.id}"]`)
      const cardEl = slot?.querySelector<HTMLElement>('.playing-card')
      if (cardEl) {
        await animateCardsFromCenter([cardEl], origin, 0)
      }
    }
    await sleep(20)
  }

  await opponentPromise
  isDealing.value = false
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function loadGame(gameId: string) {
  loading.value = true
  error.value = ''
  hintMessage.value = ''
  selectedIds.value = []
  hintIds.value = []
  centerPlay.value = null
  resetReadyState()
  clearAllPassLabels()
  lastGameId.value = ''
  try {
    const next = await getDouDizhuState(gameId)
    await applyGameState(next)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载对局失败'
  } finally {
    loading.value = false
  }
}

function startOnlinePolling() {
  stopOnlinePolling()
  if (!isOnline.value) return
  onlinePollTimer = window.setInterval(async () => {
    if (!state.value || loading.value || isAnimating.value || isDealing.value || isGameFinished.value) {
      return
    }
    try {
      const next = await getDouDizhuState(state.value.id)
      const turnChanged =
        next.current_turn !== state.value.current_turn ||
        next.calling_index !== state.value.calling_index
      const hasEvents = (next.events?.length ?? 0) > 0
      if (hasEvents || turnChanged || next.phase !== state.value.phase) {
        await applyGameState(next)
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

function syncReadyFromRoom(room: { players: { ready: boolean }[] }) {
  const next: Record<number, boolean> = { 0: false, 1: false, 2: false }
  room.players.forEach((player, seat) => {
    next[seat] = player.ready
  })
  readySeats.value = next
}

function startFinishedRoomPolling() {
  stopFinishedRoomPolling()
  if (!isOnline.value || !roomId.value) return

  const poll = async () => {
    if (!roomId.value || !state.value || state.value.phase !== 'finished') return
    try {
      const room = await fetchDouDizhuRoom(roomId.value)
      syncReadyFromRoom(room)
      if (room.game_id && room.game_id !== state.value.id) {
        stopFinishedRoomPolling()
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

async function enterNextOnlineGame(gameId: string) {
  if (!roomId.value) return
  resetReadyState()
  stopOnlinePolling()
  await router.replace({
    name: 'doudizhu-play',
    params: { gameId },
    query: { room: roomId.value },
  })
  await loadGame(gameId)
  startOnlinePolling()
}

async function beginGame() {
  loading.value = true
  error.value = ''
  hintMessage.value = ''
  selectedIds.value = []
  hintIds.value = []
  displayedHand.value = []
  centerPlay.value = null
  opponentCounts.value = { left: 0, right: 0 }
  lastGameId.value = ''
  resetReadyState()
  clearAllPassLabels()
  try {
    const next = await startDouDizhuGame()
    await applyGameState(next)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '开局失败'
  } finally {
    loading.value = false
  }
}

function onSelectedIdsUpdate(ids: string[]) {
  hintIds.value = []
  selectedIds.value = ids
}

async function handleHint() {
  if (!state.value || !canHint.value) return
  error.value = ''
  try {
    const hint = await fetchDouDizhuHint(state.value.id)
    hintMessage.value = hint.message
    if (hint.action === 'play') {
      hintIds.value = hint.card_ids
      selectedIds.value = [...hint.card_ids]
    } else {
      hintIds.value = []
      selectedIds.value = []
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取提示失败'
  }
}

async function handleCall(call: boolean) {
  if (!state.value) return
  loading.value = true
  error.value = ''
  try {
    const next = await callLandlord(state.value.id, call)
    await applyGameState(next)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '叫地主失败'
  } finally {
    loading.value = false
  }
}

function resolvePlayCardIds(): string[] {
  if (!state.value) return []
  const handIds = new Set(myHand.value.map((c) => c.id))
  const resolved: string[] = []
  const seen = new Set<string>()
  for (const id of selectedIds.value) {
    if (!handIds.has(id) || seen.has(id)) continue
    seen.add(id)
    resolved.push(id)
  }
  return resolved
}

async function handlePlay() {
  if (!state.value) return
  const cardIds = resolvePlayCardIds()
  if (cardIds.length === 0) {
    error.value = '请选择要出的牌'
    return
  }
  loading.value = true
  error.value = ''
  hintMessage.value = ''
  try {
    const next = await playCards(state.value.id, cardIds)
    selectedIds.value = []
    hintIds.value = []
    await applyGameState(next)
    if (isFinishedState(next)) return
  } catch (err) {
    const message = err instanceof Error ? err.message : '出牌失败'
    error.value = message
    hintMessage.value = message
  } finally {
    loading.value = false
  }
}

async function handlePass() {
  if (!state.value) return
  loading.value = true
  error.value = ''
  try {
    const next = await passTurn(state.value.id)
    selectedIds.value = []
    hintIds.value = []
    await applyGameState(next)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '过牌失败'
  } finally {
    loading.value = false
  }
}

function seatByIndex(index: number) {
  return state.value?.players.find((p) => p.index === index)
}

function revealedHandByIndex(index: number) {
  return revealedHands.value.find((hand) => hand.index === index)
}

function revealedCards(index: number) {
  return revealedHandByIndex(index)?.cards ?? []
}

function resetReadyState() {
  readySeats.value = { 0: false, 1: false, 2: false }
}

async function handleReady() {
  if (!isGameFinished.value || loading.value || isHumanReady.value) return

  if (isOnline.value && roomId.value && state.value) {
    loading.value = true
    error.value = ''
    readySeats.value = { ...readySeats.value, [mySeatIndex.value]: true }
    try {
      const room = await nextDouDizhuRoom(roomId.value, state.value.id, true)
      syncReadyFromRoom(room)
      if (room.game_id && room.game_id !== state.value.id) {
        await enterNextOnlineGame(room.game_id)
        return
      }
      await waitForNextOnlineGame()
    } catch (err) {
      readySeats.value = { ...readySeats.value, [mySeatIndex.value]: false }
      error.value = err instanceof Error ? err.message : '准备失败'
    } finally {
      loading.value = false
    }
    return
  }

  readySeats.value = { ...readySeats.value, 0: true }
  await sleep(300)
  readySeats.value = { 0: true, 1: true, 2: true }
  await sleep(400)
  await beginGame()
}

async function waitForNextOnlineGame() {
  if (!roomId.value || !state.value) return
  const currentGameId = state.value.id
  for (let i = 0; i < 40; i++) {
    const room = await fetchDouDizhuRoom(roomId.value)
    syncReadyFromRoom(room)
    if (room.game_id && room.game_id !== currentGameId) {
      await enterNextOnlineGame(room.game_id)
      return
    }
    await sleep(1000)
  }
  error.value = '等待其他玩家准备超时'
  readySeats.value = { ...readySeats.value, [mySeatIndex.value]: false }
}

function opponentCount(index: number) {
  if (isGameFinished.value) {
    return revealedCards(index).length
  }
  if (isDealing.value) {
    return index === leftSeatIndex.value
      ? opponentCounts.value.left
      : index === rightSeatIndex.value
        ? opponentCounts.value.right
        : 0
  }
  return seatByIndex(index)?.hand_count ?? 0
}

function isSeatActive(index: number) {
  if (!state.value || state.value.phase === 'finished' || isAnimating.value) return false
  return activeTurnIndex.value === index
}
</script>

<template>
  <main class="ddz">
    <header class="ddz__header">
      <button type="button" class="ddz__back" @click="router.push('/games/doudizhu')">← 返回</button>
      <div>
        <h1>斗地主</h1>
        <p class="ddz__subtitle">{{ modeSubtitle }}</p>
      </div>
      <button
        v-if="!isOnline"
        type="button"
        class="ddz__restart"
        :disabled="loading || isDealing || isAnimating || isGameFinished"
        @click="beginGame"
      >
        重新开局
      </button>
    </header>

    <p v-if="error" class="ddz__error">{{ error }}</p>
    <p v-if="loading && !state" class="ddz__loading">准备发牌...</p>

    <section v-if="state" ref="tableRef" class="ddz__table">
      <div v-if="showBottomCards" ref="topDeckRef" class="ddz__top-deck">
        <span class="ddz__label">底牌</span>
        <div class="ddz__bottom-row">
          <div
            v-for="card in bottomCards"
            :key="card.id"
            class="ddz__bottom-card-slot"
            :data-bottom-card-id="card.id"
          >
            <PlayingCard :card="card" stacked />
          </div>
        </div>
      </div>

      <div class="ddz__arena">
        <div class="ddz__seat ddz__seat--left">
          <div class="ddz__seat-column">
            <div class="ddz__seat-stack">
              <div
                class="ddz__player ddz__seat-anchor"
                :class="{ 'ddz__player--active': isSeatHighlighted(leftSeatIndex) }"
                :data-seat="leftSeatIndex"
              >
                <span
                  class="ddz__badge ddz__badge--role"
                  :class="{
                    'ddz__badge--landlord': seatRole(leftSeatIndex) === 'landlord',
                    'ddz__badge--farmer': seatRole(leftSeatIndex) === 'farmer',
                    'ddz__badge--reserved': !seatRole(leftSeatIndex),
                  }"
                >{{ seatRoleLabel(leftSeatIndex) }}</span>
                <span>{{ seatByIndex(leftSeatIndex)?.name }}</span>
                <span v-if="isGameFinished && readySeats[leftSeatIndex]" class="ddz__ready-badge">准备</span>
                <span class="ddz__count">{{ opponentCount(leftSeatIndex) }} 张</span>
              </div>
              <SeatIndicator
                placement="right"
                :seconds="seatTimerSeconds(leftSeatIndex)"
                :show-timer="showSeatTimer(leftSeatIndex)"
                :show-pass="showSeatPass(leftSeatIndex)"
              />
            </div>
            <div v-if="isGameFinished" class="ddz__seat-reveal">
              <StackedCards
                v-if="revealedCards(leftSeatIndex).length"
                reveal
                :cards="revealedCards(leftSeatIndex)"
              />
              <span v-else class="ddz__seat-reveal-empty">已出完</span>
            </div>
          </div>
        </div>

        <div ref="dealOriginRef" class="ddz__center">
          <div class="ddz__center-stage">
            <div v-show="isDealing" class="ddz__deck">
              <div class="ddz__deck-card" />
              <span>发牌中</span>
            </div>

            <div v-show="!isDealing" ref="playAreaRef" class="ddz__play-area">
              <div v-show="centerPlay" class="ddz__last-play">
                <p class="ddz__play-by">
                  <span v-if="centerPlayRole === 'landlord'" class="ddz__badge ddz__badge--landlord">地主</span>
                  <span v-else-if="centerPlayRole === 'farmer'" class="ddz__badge ddz__badge--farmer">农民</span>
                  <span>{{ centerPlayName }}</span>
                  <span class="ddz__play-by-action">出牌</span>
                </p>
                <StackedCards v-if="centerPlay" :cards="centerPlay.cards" :max-width="520" />
              </div>
              <div v-show="!centerPlay" class="ddz__play-area-empty">等待出牌</div>
            </div>
          </div>
        </div>

        <div class="ddz__seat ddz__seat--right">
          <div class="ddz__seat-column">
            <div class="ddz__seat-stack">
              <div
                class="ddz__player ddz__seat-anchor"
                :class="{ 'ddz__player--active': isSeatHighlighted(rightSeatIndex) }"
                :data-seat="rightSeatIndex"
              >
                <span
                  class="ddz__badge ddz__badge--role"
                  :class="{
                    'ddz__badge--landlord': seatRole(rightSeatIndex) === 'landlord',
                    'ddz__badge--farmer': seatRole(rightSeatIndex) === 'farmer',
                    'ddz__badge--reserved': !seatRole(rightSeatIndex),
                  }"
                >{{ seatRoleLabel(rightSeatIndex) }}</span>
                <span>{{ seatByIndex(rightSeatIndex)?.name }}</span>
                <span v-if="isGameFinished && readySeats[rightSeatIndex]" class="ddz__ready-badge">准备</span>
                <span class="ddz__count">{{ opponentCount(rightSeatIndex) }} 张</span>
              </div>
              <SeatIndicator
                placement="left"
                :seconds="seatTimerSeconds(rightSeatIndex)"
                :show-timer="showSeatTimer(rightSeatIndex)"
                :show-pass="showSeatPass(rightSeatIndex)"
              />
            </div>
            <div v-if="isGameFinished" class="ddz__seat-reveal">
              <StackedCards
                v-if="revealedCards(rightSeatIndex).length"
                reveal
                :cards="revealedCards(rightSeatIndex)"
              />
              <span v-else class="ddz__seat-reveal-empty">已出完</span>
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
                :class="{ 'ddz__player--active': isSeatHighlighted(mySeatIndex) }"
                :data-seat="mySeatIndex"
              >
                <span
                  class="ddz__badge ddz__badge--role"
                  :class="{
                    'ddz__badge--landlord': seatRole(mySeatIndex) === 'landlord',
                    'ddz__badge--farmer': seatRole(mySeatIndex) === 'farmer',
                    'ddz__badge--reserved': !seatRole(mySeatIndex),
                  }"
                >{{ seatRoleLabel(mySeatIndex) }}</span>
                <span>我</span>
                <span v-if="isGameFinished && readySeats[mySeatIndex]" class="ddz__ready-badge">准备</span>
                <span class="ddz__count">{{ displayedHand.length || myHand.length }} 张</span>
              </div>
              <SeatIndicator
                placement="top"
                :seconds="seatTimerSeconds(mySeatIndex)"
                :show-timer="showSeatTimer(mySeatIndex)"
                :show-pass="showSeatPass(mySeatIndex)"
              />
            </div>
          </div>
          <HandCards
            ref="handRowRef"
            :cards="displayedHand"
            :selected-ids="selectedIds"
            :hint-ids="hintIds"
            :interactive="canSelectCards"
            :dealing="isDealing"
            @update:selected-ids="onSelectedIdsUpdate"
          />
        </div>
      </div>
    </section>

    <Teleport to="body">
      <div v-if="state" class="ddz__prompt-slot" :style="promptFixedStyle">
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
                :class="{ 'ddz__prompt-actions--visible': showCallButtons }"
              >
                <button type="button" class="ddz__btn ddz__btn--primary" :disabled="loading" @click="handleCall(true)">
                  抢地主
                </button>
                <button type="button" class="ddz__btn" :disabled="loading" @click="handleCall(false)">不抢</button>
              </div>

              <div
                class="ddz__prompt-actions"
                :class="{ 'ddz__prompt-actions--visible': showPlayButtons }"
              >
                <button
                  v-if="!isOnline"
                  type="button"
                  class="ddz__btn ddz__btn--hint"
                  :disabled="loading"
                  @click="handleHint"
                >
                  提示
                </button>
                <button
                  type="button"
                  class="ddz__btn ddz__btn--primary"
                  :class="{ 'ddz__btn--play-emphasis': showPlayButtons }"
                  :disabled="loading || resolvePlayCardIds().length === 0"
                  @click="handlePlay"
                >
                  出牌
                </button>
                <button type="button" class="ddz__btn" :disabled="loading || !canPass" @click="handlePass">不出</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </main>
</template>
