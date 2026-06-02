<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DouNiuFanHand from '../components/douniu/DouNiuFanHand.vue'
import SeatIndicator from '../components/doudizhu/SeatIndicator.vue'
import { animateCardsFromCenterBatch } from '../composables/useDealAnimation'
import { useDouNiuGameSocket, useDouNiuRoomSocket } from '../composables/useDouNiuSocket'
import { usePhaseTimer } from '../composables/usePhaseTimer'
import { showToast } from '../composables/useToast'
import {
  betDouNiu,
  fetchDouNiuRoom,
  getDouNiuState,
  grabDouNiuBanker,
  nextDouNiuRoom,
  startDouNiuGame,
  tickDouNiuGame,
} from '../api/games'
import { DOUNIU_HAND_LABELS, type DouNiuEvent, type DouNiuHandLayout, type DouNiuPlayer, type DouNiuRoom, type DouNiuState } from '../types/douniu'
import type { Card } from '../types/doudizhu'

const CARDS_PER_PLAYER = 5
const DEAL_SEAT_MS = 160
const PHASE_EVENT_MS = 320
const GRAB_UNSET = -1

const DEFAULT_HAND_MULTIPLIERS: Record<string, number> = {
  five_small: 6,
  bomb: 5,
  five_flower: 4,
  niu_niu: 3,
  niu_9: 2,
  niu_8: 2,
  niu_7: 2,
  niu_6: 1,
  niu_5: 1,
  niu_4: 1,
  niu_3: 1,
  niu_2: 1,
  niu_1: 1,
  none: 1,
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

const router = useRouter()
const route = useRoute()

const state = ref<DouNiuState | null>(null)
const loading = ref(false)
const isDealing = ref(false)
const isAnimating = ref(false)
const readySeats = ref<Record<number, boolean>>({})
const lastGameId = ref('')
const dealCounts = ref<Record<number, number>>({})
const showMultiplierHelp = ref(false)
const arenaRef = ref<HTMLElement | null>(null)
const dealOriginRef = ref<HTMLElement | null>(null)

const isOnline = computed(() => route.name === 'douniu-play')
const roomId = computed(() => {
  const raw = route.query.room
  return typeof raw === 'string' && raw ? raw : ''
})

const mySeat = computed(() => state.value?.human_player ?? 0)
const isFinished = computed(() => state.value?.phase === 'finished')
const isGrabPhase = computed(() => state.value?.phase === 'grab_banker')
const isBetPhase = computed(() => state.value?.phase === 'betting')
const isHumanReady = computed(() => readySeats.value[mySeat.value] ?? false)

const showSettleReady = computed(
  () => isFinished.value && !loading.value && !isDealing.value && !isAnimating.value && !isHumanReady.value,
)

const myPlayer = computed(() => state.value?.players[mySeat.value])
const canGrab = computed(
  () =>
    isGrabPhase.value &&
    !loading.value &&
    !isDealing.value &&
    !isAnimating.value &&
    myPlayer.value &&
    !myPlayer.value.grab_done,
)
const canBet = computed(
  () =>
    isBetPhase.value &&
    !loading.value &&
    !isDealing.value &&
    !isAnimating.value &&
    state.value &&
    mySeat.value !== state.value.banker_index &&
    myPlayer.value &&
    !myPlayer.value.bet_done,
)

const needsAction = computed(() => canGrab.value || canBet.value)
const isMyTurn = computed(() => Boolean(needsAction.value))

async function handleTimeout() {
  if (!state.value || !needsAction.value) return
  try {
    loading.value = true
    if (canGrab.value) {
      await applyState(await grabDouNiuBanker(state.value.id, 0), { skipDeal: true })
    } else if (canBet.value) {
      await applyState(await betDouNiu(state.value.id, 1), { skipDeal: true })
    }
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
}

const turnDeadline = computed(() => state.value?.turn_deadline_unix)
const phase = computed(() => state.value?.phase)
const { secondsLeft } = usePhaseTimer(turnDeadline, phase, isMyTurn, handleTimeout)

const multiplierRows = computed(() => {
  const m = state.value?.hand_multipliers ?? DEFAULT_HAND_MULTIPLIERS
  return Object.entries(m)
    .map(([key, val]) => ({
      key,
      label: DOUNIU_HAND_LABELS[key] ?? key,
      value: val,
    }))
    .sort((a, b) => b.value - a.value)
})

function grabActionLabel(mult: number) {
  return mult === 0 ? '不抢' : `抢庄 ×${mult}`
}

const promptText = computed(() => {
  if (!state.value) return ''
  if (state.value.id === 'pending') return '准备发牌…'
  if (isDealing.value) return '发牌中…'
  if (isFinished.value) return ''
  if (canGrab.value) return `看牌后选择是否抢庄 · 剩余 ${secondsLeft.value} 秒`
  if (canBet.value) return `选择下注倍数 · 剩余 ${secondsLeft.value} 秒`
  if (isGrabPhase.value) return state.value.message || '等待其他玩家选择'
  if (isBetPhase.value) return state.value.message || '等待其他玩家下注'
  return state.value.message
})

let pollTimer: number | null = null
let finishedRoomPollTimer: number | null = null
const wsGameConnected = ref(false)
const wsRoomConnected = ref(false)

const onlineGameId = computed(() => {
  if (!isOnline.value) return ''
  const fromRoute = route.params.gameId
  if (typeof fromRoute === 'string' && fromRoute) return fromRoute
  return state.value?.id ?? ''
})

const gameSocketEnabled = computed(
  () => isOnline.value && Boolean(onlineGameId.value) && !isFinished.value,
)
const roomSocketEnabled = computed(() => isOnline.value && Boolean(roomId.value) && isFinished.value)

async function applyRemoteGameState(next: DouNiuState) {
  if (!state.value || loading.value || isDealing.value || isAnimating.value) return
  if (state.value.id !== next.id && next.id !== 'pending') {
    await applyState(next)
    return
  }
  loading.value = true
  try {
    await applyState(next, { skipDeal: true })
  } finally {
    loading.value = false
  }
}

async function applyRemoteRoom(next: DouNiuRoom) {
  syncReadyFromRoom(next)
  if (next.game_id && state.value && next.game_id !== state.value.id) {
    await enterNextOnlineGame(next.game_id)
  }
}

const { reconnect: reconnectGameSocket } = useDouNiuGameSocket({
  gameId: onlineGameId,
  enabled: gameSocketEnabled,
  currentState: state,
  onStatus: (status) => {
    wsGameConnected.value = status === 'open'
  },
  onState: applyRemoteGameState,
})

useDouNiuRoomSocket({
  roomId,
  enabled: roomSocketEnabled,
  currentRoom: ref(null),
  onStatus: (status) => {
    wsRoomConnected.value = status === 'open'
  },
  onRoom: applyRemoteRoom,
})

function toastError(message: string) {
  showToast(message, 'error')
}

function botCountFromRoute() {
  const raw = route.query.bots
  const n = typeof raw === 'string' ? Number.parseInt(raw, 10) : 1
  return Number.isFinite(n) ? Math.min(7, Math.max(1, n)) : 1
}

function initPlaceholderTable(bots = botCountFromRoute(), prevPlayers?: DouNiuPlayer[]) {
  const chipsAt = (index: number, fallback = 2000) => prevPlayers?.[index]?.chips ?? fallback
  const players: DouNiuPlayer[] = [
    {
      index: 0,
      name: '我',
      is_ai: false,
      chips: chipsAt(0),
      grab_mult: GRAB_UNSET,
      bet_mult: 0,
      grab_done: false,
      bet_done: false,
      card_count: 0,
    },
    ...Array.from({ length: bots }, (_, i) => ({
      index: i + 1,
      name: `电脑${i + 1}`,
      is_ai: true,
      chips: chipsAt(i + 1),
      grab_mult: GRAB_UNSET,
      bet_mult: 0,
      grab_done: false,
      bet_done: false,
      card_count: 0,
    })),
  ]
  state.value = {
    id: 'pending',
    phase: 'grab_banker',
    players,
    human_player: 0,
    banker_index: -1,
    base_ante: 10,
    message: '准备发牌…',
    hand_multipliers: DEFAULT_HAND_MULTIPLIERS,
  }
}

function seatStyle(index: number) {
  const total = state.value?.players.length ?? 1
  const rel = (index - mySeat.value + total) % total
  const angle = Math.PI / 2 + (2 * Math.PI * rel) / total
  const rx = 39
  const ry = rel === 0 ? 26 : 31
  const yNudge = rel === 0 ? -2 : 0
  return {
    left: `${50 + rx * Math.cos(angle)}%`,
    top: `${50 + ry * Math.sin(angle) + yNudge}%`,
  }
}

function seatIndicatorPlacement(index: number): 'left' | 'right' | 'top' {
  const total = state.value?.players.length ?? 1
  const rel = (index - mySeat.value + total) % total
  if (rel === 0) return 'top'
  const angle = Math.PI / 2 + (2 * Math.PI * rel) / total
  return Math.cos(angle) < 0 ? 'right' : 'left'
}

function seatLabel(index: number) {
  if (index === mySeat.value) return '我'
  return state.value?.players[index]?.name ?? `玩家${index + 1}`
}

function seatBackCount(index: number) {
  if (isDealing.value) return dealCounts.value[index] ?? 0
  if (showSeatHand(index)) return 0
  return CARDS_PER_PLAYER
}

function seatHandCards(index: number): Card[] {
  if (index === mySeat.value && state.value?.my_hand?.length) {
    return state.value.my_hand
  }
  const p = state.value?.players[index]
  return p?.hand ?? []
}

function showSeatHand(index: number) {
  if (isDealing.value) return false
  return seatHandCards(index).length > 0
}

function seatStatusLabel(player: DouNiuPlayer) {
  if (showSeatHand(player.index)) return '已看牌'
  return '待发牌'
}

function seatStatusClass(player: DouNiuPlayer) {
  if (showSeatHand(player.index)) return 'zjh__tag--look'
  return 'zjh__tag--blind'
}

function isRoundDraw(index: number) {
  return isFinished.value && (state.value?.players[index]?.round_delta ?? 0) === 0
}

function isBanker(index: number) {
  return state.value != null && state.value.banker_index >= 0 && state.value.banker_index === index && !isGrabPhase.value
}

function isRoundWinner(index: number) {
  const delta = state.value?.players[index]?.round_delta ?? 0
  return isFinished.value && delta > 0
}

function isRoundLoser(index: number) {
  const delta = state.value?.players[index]?.round_delta ?? 0
  return isFinished.value && delta < 0
}

function seatHandLayout(index: number): DouNiuHandLayout | null {
  if (index === mySeat.value && state.value?.my_hand_layout) {
    return state.value.my_hand_layout
  }
  return state.value?.players[index]?.hand_layout ?? null
}

function seatHandMeta(index: number) {
  if (index === mySeat.value && state.value?.my_hand?.length) {
    return {
      label: state.value.my_hand_label ?? '',
      type: state.value.my_hand_type ?? '',
      multiplier: state.value.my_hand_multiplier ?? 1,
    }
  }
  const p = state.value?.players[index]
  return {
    label: p?.hand_label ?? '',
    type: p?.hand_type ?? '',
    multiplier: p?.hand_multiplier ?? 1,
  }
}

function showHandLabel(index: number) {
  const meta = seatHandMeta(index)
  if (!meta.label) return false
  if (isFinished.value) return true
  return index === mySeat.value && showSeatHand(index)
}

function patchGrabMult(event: DouNiuEvent, current: number) {
  return typeof event.grab_mult === 'number' ? event.grab_mult : current
}

function patchBetMult(event: DouNiuEvent, current: number) {
  return typeof event.bet_mult === 'number' ? event.bet_mult : current
}

function seatBadgeLabel(index: number): string | undefined {
  const player = state.value?.players[index]
  if (!player || isDealing.value) return undefined
  if (isBetPhase.value) {
    if (player.bet_done && index !== state.value?.banker_index) {
      return `下注 ×${player.bet_mult}`
    }
    if (player.grab_done && player.grab_mult >= 0) {
      return grabActionLabel(player.grab_mult)
    }
    return undefined
  }
  if (isGrabPhase.value && player.grab_done && player.grab_mult >= 0) {
    return grabActionLabel(player.grab_mult)
  }
  return undefined
}

function showSeatBadge(index: number): boolean {
  if (showSeatTimer(index)) return false
  return Boolean(seatBadgeLabel(index))
}

function showSeatTimer(index: number): boolean {
  if (isDealing.value || isAnimating.value) return false
  if (index !== mySeat.value) return false
  return Boolean(needsAction.value)
}

function prepareDealState(next: DouNiuState): DouNiuState {
  return {
    ...next,
    events: [],
    my_hand: undefined,
    message: '正在发牌…',
    banker_index: -1,
    players: next.players.map((p) => ({
      ...p,
      grab_mult: GRAB_UNSET,
      bet_mult: 0,
      grab_done: false,
      bet_done: false,
      hand_type: undefined,
      hand_label: undefined,
      hand_multiplier: undefined,
      round_delta: undefined,
      hand: undefined,
      card_count: 0,
    })),
  }
}

async function runDouNiuDealAnimation() {
  if (!state.value) return
  isDealing.value = true
  dealCounts.value = Object.fromEntries(state.value.players.map((p) => [p.index, 0]))

  const total = state.value.players.length
  for (let i = 0; i < total; i++) {
    const player = state.value.players[i]
    if (!player) continue
    dealCounts.value = { ...dealCounts.value, [player.index]: CARDS_PER_PLAYER }
    await nextTick()
    const backs = Array.from(
      arenaRef.value?.querySelectorAll(
        `[data-seat="${player.index}"] .dn__back:not(.dn__back--empty)`,
      ) ?? [],
    ) as HTMLElement[]
    if (backs.length && dealOriginRef.value) {
      await animateCardsFromCenterBatch(backs, dealOriginRef.value, 0.18)
    }
    await sleep(DEAL_SEAT_MS)
  }

  isDealing.value = false
}

function isFreshTable(next: DouNiuState) {
  if (next.phase !== 'grab_banker') return false
  return next.players.every((p) => !p.grab_done)
}

function patchStateForEvent(current: DouNiuState, event: DouNiuEvent): DouNiuState {
  if (event.type === 'grab_banker') {
    return {
      ...current,
      players: current.players.map((p) =>
        p.index === event.player_index
          ? { ...p, grab_mult: patchGrabMult(event, p.grab_mult), grab_done: true }
          : p,
      ),
    }
  }
  if (event.type === 'place_bet') {
    return {
      ...current,
      players: current.players.map((p) =>
        p.index === event.player_index
          ? { ...p, bet_mult: patchBetMult(event, p.bet_mult), bet_done: true }
          : p,
      ),
    }
  }
  if (event.type === 'banker_set') {
    return {
      ...current,
      banker_index: event.player_index,
      message: event.message || current.message,
    }
  }
  return current
}

function shouldToastEvent(event: DouNiuEvent) {
  return event.type === 'game_over' && Boolean(event.message)
}

async function replayPhaseEvent(event: DouNiuEvent) {
  if (state.value && (event.type === 'grab_banker' || event.type === 'place_bet' || event.type === 'banker_set')) {
    state.value = patchStateForEvent(state.value, event)
  }
  if (shouldToastEvent(event)) {
    showToast(event.message!, 'info', 2200)
  }
  await sleep(event.type === 'game_over' ? 400 : PHASE_EVENT_MS)
}

async function applyState(next: DouNiuState, options: { skipDeal?: boolean } = {}) {
  const isNewGame = lastGameId.value !== next.id
  const shouldDeal = !options.skipDeal && isNewGame && isFreshTable(next)

  if (shouldDeal) {
    lastGameId.value = next.id
    state.value = prepareDealState(next)
    await runDouNiuDealAnimation()
    state.value = {
      ...next,
      events: [],
      message: next.message || '看牌后选择是否抢庄',
      players: next.players.map((p) => ({
        ...p,
        hand: undefined,
        hand_type: undefined,
        hand_label: undefined,
        hand_multiplier: undefined,
        card_count: CARDS_PER_PLAYER,
      })),
    }

    const synced = await tickDouNiuGame(next.id)
    await applyState(synced, { skipDeal: true })
    return
  }

  if (isNewGame) {
    lastGameId.value = next.id
  }

  const events = next.events ?? []
  if (events.length === 0) {
    state.value = { ...next, events: [] }
    return
  }

  isAnimating.value = true
  try {
    for (const event of events) {
      await replayPhaseEvent(event)
      if (next.phase === 'finished' && event.type === 'game_over') break
    }
    state.value = { ...next, events: [] }
  } finally {
    isAnimating.value = false
  }
}

async function act(fn: () => Promise<DouNiuState>) {
  if (!state.value || loading.value || isDealing.value) return
  loading.value = true
  try {
    await applyState(await fn(), { skipDeal: true })
  } catch (err) {
    toastError(err instanceof Error ? err.message : '操作失败')
  } finally {
    loading.value = false
  }
}

async function beginSolo(carryChips = false) {
  const bots = botCountFromRoute()
  const previousGameId = carryChips && state.value?.id ? state.value.id : undefined
  const prevPlayers = carryChips ? state.value?.players : undefined
  loading.value = true
  isAnimating.value = false
  isDealing.value = false
  lastGameId.value = ''
  dealCounts.value = {}
  readySeats.value = {}
  initPlaceholderTable(bots, prevPlayers)
  try {
    const next = await startDouNiuGame(bots, previousGameId)
    await applyState(next)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '开局失败')
  } finally {
    loading.value = false
  }
}

async function loadGame(gameId: string) {
  loading.value = true
  try {
    const next = await getDouNiuState(gameId)
    initPlaceholderTable(Math.max(1, next.players.length - 1))
    lastGameId.value = ''
    await applyState(next)
  } catch (err) {
    lastGameId.value = ''
    toastError(err instanceof Error ? err.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function startPolling() {
  stopPolling()
  if (!isOnline.value || wsGameConnected.value) return
  pollTimer = window.setInterval(async () => {
    if (!state.value || loading.value || isDealing.value || isAnimating.value || isFinished.value) return
    if (wsGameConnected.value) return
    try {
      const next = await getDouNiuState(state.value.id)
      if (next.phase !== state.value.phase || (next.events?.length ?? 0) > 0) {
        await applyRemoteGameState(next)
      }
    } catch {
      // ignore
    }
  }, 5000)
}

function stopPolling() {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

function syncReadyFromRoom(room: DouNiuRoom) {
  const next: Record<number, boolean> = {}
  room.players.forEach((p, seat) => {
    next[seat] = p.ready
  })
  readySeats.value = next
}

async function enterNextOnlineGame(gameId: string) {
  if (!roomId.value) return
  stopPolling()
  stopFinishedRoomPolling()
  lastGameId.value = ''
  loading.value = true
  initPlaceholderTable(Math.max(1, (state.value?.players.length ?? 2) - 1), state.value?.players)
  try {
    await router.replace({
      name: 'douniu-play',
      params: { gameId },
      query: { room: roomId.value },
    })
    await applyState(await getDouNiuState(gameId))
    reconnectGameSocket()
    startPolling()
  } catch (err) {
    toastError(err instanceof Error ? err.message : '进入下一局失败')
  } finally {
    loading.value = false
  }
}

function startFinishedRoomPolling() {
  stopFinishedRoomPolling()
  if (!isOnline.value || !roomId.value || wsRoomConnected.value) return
  const poll = async () => {
    if (!roomId.value || !state.value || !isFinished.value || wsRoomConnected.value) return
    try {
      await applyRemoteRoom(await fetchDouNiuRoom(roomId.value))
    } catch {
      // ignore
    }
  }
  void poll()
  finishedRoomPollTimer = window.setInterval(poll, 5000)
}

function stopFinishedRoomPolling() {
  if (finishedRoomPollTimer !== null) {
    window.clearInterval(finishedRoomPollTimer)
    finishedRoomPollTimer = null
  }
}

async function handleReady() {
  if (!isFinished.value || loading.value || isHumanReady.value) return
  if (isOnline.value && roomId.value && state.value) {
    loading.value = true
    readySeats.value = { ...readySeats.value, [mySeat.value]: true }
    try {
      const room = await nextDouNiuRoom(roomId.value, true)
      syncReadyFromRoom(room)
      if (room.game_id && room.game_id !== state.value.id) {
        await enterNextOnlineGame(room.game_id)
      }
    } catch (err) {
      readySeats.value = { ...readySeats.value, [mySeat.value]: false }
      toastError(err instanceof Error ? err.message : '准备失败')
    } finally {
      loading.value = false
    }
    return
  }
  readySeats.value = { ...readySeats.value, [mySeat.value]: true }
  await beginSolo(true)
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

watch(wsGameConnected, (open) => {
  if (open) stopPolling()
  else if (isOnline.value && !isFinished.value) startPolling()
})

watch(wsRoomConnected, (open) => {
  if (open) stopFinishedRoomPolling()
  else if (isOnline.value && isFinished.value) startFinishedRoomPolling()
})

watch(
  () => state.value?.phase,
  (p) => {
    if (p === 'finished') {
      resetReadyOnFinish()
      if (isOnline.value && roomId.value) {
        stopPolling()
        startFinishedRoomPolling()
      }
    } else {
      stopFinishedRoomPolling()
    }
  },
)

function resetReadyOnFinish() {
  const next: Record<number, boolean> = {}
  for (const p of state.value?.players ?? []) {
    next[p.index] = false
  }
  readySeats.value = next
}
</script>

<template>
  <main class="zjh dn">
    <header class="zjh__header">
      <button type="button" class="ddz__back" @click="router.push('/games/douniu')">← 返回</button>
      <div>
        <h1>斗牛</h1>
        <p class="ddz__subtitle">{{ isOnline ? '多人联机' : '单机对战电脑' }} · 看牌抢庄</p>
      </div>
      <button
        v-if="!isOnline && !isFinished"
        type="button"
        class="ddz__restart"
        :disabled="loading || isDealing || isAnimating"
        @click="() => beginSolo()"
      >
        重新开局
      </button>
      <div v-else class="zjh__header-spacer" aria-hidden="true" />
    </header>

    <section v-if="state" class="zjh__table">
      <div ref="arenaRef" class="zjh__arena dn__arena">
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
          <div v-show="isDealing || state.id === 'pending'" class="ddz__deck">
            <span class="ddz__deck-card" />
            <span>发牌中</span>
          </div>
        </div>

        <div class="dn__ante">
          <span class="dn__ante-label">底注</span>
          <strong>{{ state.base_ante }}</strong>
          <span v-if="state.banker_index >= 0 && !isGrabPhase" class="dn__ante-sub">
            庄家 · {{ seatLabel(state.banker_index) }}
          </span>
        </div>

        <div
          v-for="player in state.players"
          :key="player.index"
          class="zjh__seat"
          :class="{
            'zjh__seat--self': player.index === mySeat,
            'dn__seat--banker': isBanker(player.index),
            'dn__seat--winner': isRoundWinner(player.index),
            'dn__seat--loser': isRoundLoser(player.index),
          }"
          :data-seat="player.index"
          :style="seatStyle(player.index)"
        >
          <div class="zjh__seat-stack">
            <div class="zjh__seat-card dn__seat-card">
              <div class="dn__seat-title">
                <strong class="zjh__seat-name">{{ seatLabel(player.index) }}</strong>
                <div class="dn__role-badges">
                  <span v-if="isBanker(player.index)" class="dn__role-badge dn__role-badge--banker">庄</span>
                  <span v-if="isRoundWinner(player.index)" class="dn__role-badge dn__role-badge--win">赢</span>
                  <span v-else-if="isRoundLoser(player.index)" class="dn__role-badge dn__role-badge--lose">输</span>
                  <span v-else-if="isRoundDraw(player.index)" class="dn__role-badge dn__role-badge--draw">平</span>
                </div>
              </div>
              <div class="dn__status-row">
                <span v-if="isFinished && readySeats[player.index]" class="ddz__ready-badge">准备</span>
                <span v-else-if="!isFinished" class="zjh__tag" :class="seatStatusClass(player)">{{ seatStatusLabel(player) }}</span>
              </div>
              <span class="zjh__chips">{{ player.chips }} 币</span>
              <div class="dn__delta-row">
                <span v-if="isFinished && player.round_delta" class="dn__delta" :class="{ 'dn__delta--win': player.round_delta > 0 }">
                  {{ player.round_delta > 0 ? '+' : '' }}{{ player.round_delta }}
                </span>
              </div>
            </div>
            <SeatIndicator
              :placement="seatIndicatorPlacement(player.index)"
              :show-timer="showSeatTimer(player.index)"
              :seconds="secondsLeft"
              :action-label="showSeatBadge(player.index) ? seatBadgeLabel(player.index) : undefined"
            />
          </div>
          <div class="dn__hand-slot">
            <DouNiuFanHand
              v-if="showSeatHand(player.index)"
              :cards="seatHandCards(player.index)"
              :layout="seatHandLayout(player.index)"
              :hand-label="showHandLabel(player.index) ? seatHandMeta(player.index).label : ''"
              :hand-type="seatHandMeta(player.index).type"
              :hand-multiplier="showHandLabel(player.index) ? seatHandMeta(player.index).multiplier : 1"
              :highlight-niu="player.index === mySeat && showSeatHand(player.index) && !isFinished"
              :reveal="isFinished"
            />
            <div v-else class="dn__backs">
              <span
                v-for="n in CARDS_PER_PLAYER"
                :key="`${player.index}-${n}`"
                class="dn__back"
                :class="{ 'dn__back--empty': n > seatBackCount(player.index) }"
              />
            </div>
          </div>
        </div>

        <div class="zjh__prompt dn__prompt">
          <p v-if="promptText" class="dn__prompt-text">{{ promptText }}</p>

          <div v-if="isFinished" class="zjh__settle dn__settle">
            <p v-if="isHumanReady && !isOnline" class="zjh__settle-hint">即将重新发牌…</p>
            <p v-else-if="isHumanReady && isOnline" class="zjh__settle-hint">已准备 · 等待其他玩家</p>
            <button
              v-if="showSettleReady"
              type="button"
              class="ddz__btn ddz__btn--primary"
              :disabled="loading || isDealing || isAnimating"
              @click="handleReady"
            >
              {{ isOnline ? '准备下一局' : '再来一局' }}
            </button>
          </div>

          <div v-else-if="canGrab" class="dn__mult-row">
            <button
              v-for="m in state.grab_options ?? [0, 1, 2, 3, 4]"
              :key="m"
              type="button"
              class="ddz__btn"
              :class="{ 'ddz__btn--primary': m > 0 }"
              :disabled="loading || isDealing"
              @click="act(() => grabDouNiuBanker(state!.id, m))"
            >
              {{ grabActionLabel(m) }}
            </button>
          </div>

          <div v-else-if="canBet" class="dn__mult-row">
            <button
              v-for="m in state.bet_options ?? [1, 2, 3, 5]"
              :key="m"
              type="button"
              class="ddz__btn ddz__btn--primary"
              :disabled="loading || isDealing"
              @click="act(() => betDouNiu(state!.id, m))"
            >
              下注 ×{{ m }}
            </button>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>
