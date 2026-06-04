import { nextTick, type ComputedRef, type Ref } from 'vue'
import {
  animateYzsDrawBatch,
  animateYzsDiscardBatch,
  animateYzsShaFlyBolt,
} from './useYzsPlayAnimation'
import type { YuzhoushaState, YzsCard, YzsEvent, YzsPlayer } from '../../types/yuzhousha'
import {
  equipSlotOf,
  judgeAreaCards,
  removeKnownCardFromPlayer,
  trickStaysInJudge,
} from './playerCardHelpers'
import {
  replayRegisteredEvent,
  shouldPrefetchEventMessage,
  type EventReplayContext,
} from './eventReplayerRegistry'

export const YZS_INITIAL_HAND = 4
export const YZS_TURN_DRAW = 2

export interface YzsAnimationsDeps {
  state: Ref<YuzhoushaState | null>
  mySeat: ComputedRef<number>
  opponentSeat: ComputedRef<number>
  myPlayer: ComputedRef<YzsPlayer | undefined>
  drawAreaRef: Ref<HTMLElement | null>
  handAreaRef: Ref<HTMLElement | null>
  playAreaRef: Ref<HTMLElement | null>
  tableWrapRef: Ref<HTMLElement | null>
  isBoltFlying: Ref<boolean>
  hitFlashSeat: Ref<number | null>
  blockFlashSeat: Ref<number | null>
  isDealing: Ref<boolean>
  isAnimating: Ref<boolean>
  displayedHand: Ref<YzsCard[]>
  displayedTableCards: Ref<YzsCard[]>
  displayedDealCounts: Ref<Record<number, number>>
  enteringDrawCardIds: Ref<string[]>
  centerMessage: Ref<string>
  tableActionHint: Ref<string>
  selectedId: Ref<string>
  selectedDiscardIds: Ref<string[]>
  syncWeaponSkillTargeting: (next: YuzhoushaState) => void
  syncWushengFromState: () => void
  clearTargeting: () => void
}

export function useYzsAnimations(deps: YzsAnimationsDeps) {
  const {
    state,
    mySeat,
    opponentSeat,
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
  } = deps

  async function flashSeatHit(seat: number) {
    hitFlashSeat.value = seat
    await sleep(260)
    if (hitFlashSeat.value === seat) {
      hitFlashSeat.value = null
    }
  }

  async function flashSeatBlocked(seat: number) {
    blockFlashSeat.value = seat
    await sleep(260)
    if (blockFlashSeat.value === seat) {
      blockFlashSeat.value = null
    }
  }

  async function runShaFlyBolt(source: number, target: number) {
    if (isBoltFlying.value) return
    isBoltFlying.value = true
    try {
      await animateYzsShaFlyBolt(
        source,
        target,
        mySeat.value,
        handAreaRef.value,
        'dash-flow',
        tableWrapRef.value,
      )
    } finally {
      isBoltFlying.value = false
    }
  }

  function syncDisplayFromState(next: YuzhoushaState) {
  displayedHand.value = [...(next.my_hand ?? [])]
  centerMessage.value = next.message
  syncWeaponSkillTargeting(next)
  syncWushengFromState()
}

function appendDrawnCards(cards: YzsCard[]) {
  if (cards.length === 0) return
  const existingIds = new Set(displayedHand.value.map((c) => c.id))
  const incoming = cards.filter((card) => !existingIds.has(card.id))
  if (incoming.length === 0) return

  displayedHand.value = [...displayedHand.value, ...incoming]
  enteringDrawCardIds.value = incoming.map((card) => card.id)
  window.setTimeout(() => {
    const active = new Set(enteringDrawCardIds.value)
    for (const card of incoming) active.delete(card.id)
    enteringDrawCardIds.value = [...active]
  }, 420)
}

function setTableCard(card: YzsCard) {
  displayedTableCards.value = [card]
}

function addTableCard(card: YzsCard) {
  if (displayedTableCards.value.some((c) => c.id === card.id)) return
  displayedTableCards.value = [...displayedTableCards.value, card]
}

function collectDiscardBatch(events: YzsEvent[], start: number): YzsEvent[] {
  const first = events[start]
  if (first?.type !== 'discard' || !first.card) return []
  const seat = first.player_index
  const batch: YzsEvent[] = []
  for (let i = start; i < events.length; i++) {
    const ev = events[i]
    if (ev.type !== 'discard' || !ev.card) break
    if (seat != null && ev.player_index !== seat) break
    batch.push(ev)
  }
  return batch
}

function collectDrawBatch(events: YzsEvent[], start: number): YzsEvent[] {
  const first = events[start]
  if (first?.type !== 'draw' || first.player_index == null) return []
  const seat = first.player_index
  const batch: YzsEvent[] = []
  for (let i = start; i < events.length; i++) {
    const ev = events[i]
    if (ev.type !== 'draw' || ev.player_index !== seat) break
    batch.push(ev)
  }
  return batch
}

async function replayDrawBatch(batch: YzsEvent[]) {
  if (batch.length === 0) return
  const seat = batch[0].player_index!
  const lastMsg = batch[batch.length - 1]?.message
  if (lastMsg) centerMessage.value = lastMsg

  const cards = batch.flatMap((ev) => (ev.card ? [ev.card] : []))

  await animateYzsDrawBatch(
    drawAreaRef.value,
    seat,
    batch.length,
    mySeat.value,
    handAreaRef.value,
    () => {
      if (!state.value) return
      if (seat === mySeat.value) {
        appendDrawnCards(cards)
      }
      state.value = {
        ...state.value,
        draw_count: Math.max(0, state.value.draw_count - batch.length),
        players: state.value.players.map((p) =>
          p.index === seat ? { ...p, hand_count: p.hand_count + batch.length } : p,
        ),
      }
    },
  )
}

function discardActorHint(seat: number | undefined | null): string {
  if (seat == null) return '弃牌'
  const player = state.value?.players.find((p) => p.index === seat)
  return `${player?.name ?? '玩家'} 弃牌`
}

async function replayDiscardBatch(batch: YzsEvent[]) {
  const cards = batch.flatMap((ev) => (ev.card ? [ev.card] : []))
  if (cards.length === 0) return

  const seat = batch[0]?.player_index
  const hint = discardActorHint(seat)
  tableActionHint.value = hint
  centerMessage.value = hint
  displayedTableCards.value = []

  await animateYzsDiscardBatch(
    batch,
    playAreaRef.value,
    mySeat.value,
    handAreaRef.value,
    () => {
      for (const card of cards) {
        addTableCard(card)
      }
      if (seat === mySeat.value) {
        const ids = new Set(cards.map((c) => c.id))
        displayedHand.value = displayedHand.value.filter((c) => !ids.has(c.id))
      }
      if (state.value && seat != null) {
        state.value = {
          ...state.value,
          message: hint,
          discard_count: state.value.discard_count + cards.length,
          players: state.value.players.map((p) =>
            p.index === seat
              ? { ...p, hand_count: Math.max(0, p.hand_count - cards.length) }
              : p,
          ),
        }
      }
    },
  )

  await sleep(120)
  tableActionHint.value = ''
}

async function replayEvent(event: YzsEvent) {
  if (shouldPrefetchEventMessage(event)) {
    centerMessage.value = event.message!
  }

  const ctx: EventReplayContext = {
    event,
    state,
    mySeat: mySeat.value,
    centerMessage,
    tableActionHint,
    displayedHand,
    displayedTableCards,
    drawAreaRef,
    playAreaRef,
    handAreaRef,
    appendDrawnCards,
    setTableCard,
    sleep,
    flashSeatHit,
    flashSeatBlocked,
    runShaFlyBolt,
    nextTick,
  }
  await replayRegisteredEvent(ctx)
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function applyState(next: YuzhoushaState) {
  const events = next.events ?? []
  clearTargeting()
  selectedId.value = ''
  selectedDiscardIds.value = []
  tableActionHint.value = ''

  const skipReplay =
    next.pending?.response_mode === 'peek_deck' || next.turn_step === 'prepare'

  if (events.length === 0 || skipReplay) {
    state.value = { ...next, events: [] }
    syncDisplayFromState(next)
    return
  }

  isAnimating.value = true
  try {
    let i = 0
    while (i < events.length) {
      if (events[i]?.type === 'draw') {
        const batch = collectDrawBatch(events, i)
        await replayDrawBatch(batch)
        i += batch.length
        continue
      }
      if (events[i]?.type === 'discard') {
        const batch = collectDiscardBatch(events, i)
        await replayDiscardBatch(batch)
        i += batch.length
        continue
      }
      await replayEvent(events[i])
      i++
    }
    state.value = { ...next, events: [] }
    syncDisplayFromState(next)
  } finally {
    isAnimating.value = false
    await nextTick()
  }
}

async function runInitialDealAnimation(next: YuzhoushaState) {
  isDealing.value = true
  isAnimating.value = true
  displayedHand.value = []
  displayedTableCards.value = []
  displayedDealCounts.value = Object.fromEntries(next.players.map((p) => [p.index, 0]))
  centerMessage.value = '发牌中…'

  state.value = {
    ...next,
    my_hand: [],
    players: next.players.map((p) => ({ ...p, hand_count: 0 })),
    events: [],
    message: '发牌中…',
  }
  await nextTick()

  const origin = drawAreaRef.value
  const seats = [mySeat.value, opponentSeat.value]

  for (const seat of seats) {
    const dealCount = seat === mySeat.value ? YZS_INITIAL_HAND : YZS_INITIAL_HAND
    await animateYzsDrawBatch(
      origin,
      seat,
      dealCount,
      mySeat.value,
      handAreaRef.value,
      () => {
        displayedDealCounts.value = {
          ...displayedDealCounts.value,
          [seat]: dealCount,
        }
        if (seat === mySeat.value && next.my_hand) {
          displayedHand.value = next.my_hand.slice(0, dealCount)
        }
        if (state.value) {
          state.value = {
            ...state.value,
            players: state.value.players.map((p) =>
              p.index === seat ? { ...p, hand_count: dealCount } : p,
            ),
          }
        }
      },
    )
    await sleep(50)
  }

  const openingDraw = Math.min(
    YZS_TURN_DRAW,
    Math.max(0, (next.my_hand?.length ?? 0) - YZS_INITIAL_HAND),
  )
  if (openingDraw > 0 && next.current_turn === mySeat.value) {
    await animateYzsDrawBatch(
      origin,
      mySeat.value,
      openingDraw,
      mySeat.value,
      handAreaRef.value,
      () => {
        if (next.my_hand) appendDrawnCards(next.my_hand.slice(YZS_INITIAL_HAND))
        if (state.value) {
          state.value = {
            ...state.value,
            players: state.value.players.map((p) =>
              p.index === mySeat.value
                ? { ...p, hand_count: next.my_hand?.length ?? p.hand_count }
                : p,
            ),
          }
        }
      },
    )
  }

  await sleep(80)
  state.value = { ...next, events: [] }
  syncDisplayFromState(next)
  isDealing.value = false
  isAnimating.value = false
}

  return {
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
    equipSlotOf,
    judgeAreaCards,
    trickStaysInJudge,
    removeKnownCardFromPlayer,
    discardActorHint,
  }
}
