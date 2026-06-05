import { computed } from 'vue'
import type { YzsCard, YzsSeatSlot } from '../../types/yuzhousha'
import { isIdentityMode } from '../../constants/yzsModes'
import {
  equipSlotOf,
  equippedCards,
  judgeAreaCards,
} from './playerCardHelpers'
import type { YzsTargetingDeps } from './types'

const opponentTargetKinds = new Set(['sha', 'guohe', 'tannang', 'juedou', 'lebu', 'bingliang', 'huogong', 'tiesuo'])
const targetCardKinds = new Set(['guohe', 'tannang'])

function crossSeatsFromMap(seatMap: YzsSeatSlot[] | undefined) {
  if (!seatMap?.length) return []
  return seatMap.map((s) => ({
    seat: s.seat,
    area: s.area,
    placement: (s.placement === 'top' ? 'top' : s.placement === 'left' ? 'left' : 'right') as
      | 'top'
      | 'left'
      | 'right',
    isTeammate: s.is_teammate,
    seatRole: (s.seat_role === 'protect' ||
    s.seat_role === 'mark' ||
    s.seat_role === 'farmer' ||
    s.seat_role === 'landlord' ||
    s.seat_role === 'commander' ||
    s.seat_role === 'forward'
      ? s.seat_role
      : undefined) as 'protect' | 'mark' | 'farmer' | 'landlord' | 'commander' | 'forward' | undefined,
  }))
}

function aliveSeatsExcept(players: { index: number; hp: number }[], mySeat: number) {
  return players.filter((p) => p.index !== mySeat && p.hp > 0).map((p) => p.index)
}

function chainMarkSeat(mySeat: number, playerCount: number) {
  return (mySeat - 1 + playerCount) % playerCount
}

export function useYzsTargeting(deps: YzsTargetingDeps) {
  const {
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
    isPojunDiscard,
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
    pojunVictimSeat,
  } = deps

  const hasTeamMode = computed(
    () => state.value?.players.some((p) => p.team != null) ?? false,
  )

  const enemySeats = computed(() => {
    const players = state.value?.players
    if (!players?.length) return [opponentSeat.value]
    if (isIdentityMode(state.value?.mode)) {
      return aliveSeatsExcept(players, mySeat.value)
    }
    if (state.value?.mode === '3p_chain' && players.length === 3) {
      return [chainMarkSeat(mySeat.value, 3)]
    }
    const me = players[mySeat.value]
    if (me?.team == null) return [opponentSeat.value]
    return players.filter((p) => p.team !== me.team).map((p) => p.index)
  })

  const teammateSeat = computed(() => {
    const fromMap = state.value?.seat_map?.find((s) => s.is_teammate)
    if (fromMap) return fromMap.seat
    const me = myPlayer.value
    if (me?.team == null) return -1
    return (
      state.value?.players.find((p) => p.team === me.team && p.index !== mySeat.value)?.index ?? -1
    )
  })

  const crossSeats = computed(() => crossSeatsFromMap(state.value?.seat_map))

  function ringDistance(from: number, to: number) {
    const n = state.value?.players?.length ?? 2
    if (n <= 2) return 1
    let diff = Math.abs(from - to)
    if (n - diff < diff) diff = n - diff
    return Math.max(1, diff)
  }

  function distanceToSeat(seat: number) {
    const me = myPlayer.value
    const target = seatAt(seat)
    if (!me || !target) return 99
    let dist = ringDistance(mySeat.value, seat)
    if (me.minus_horse) dist -= 1
    if (target.plus_horse) dist += 1
    return Math.max(1, dist)
  }

  function takeableOptionsForPlayer(seat: number) {
    const player = state.value?.players[seat]
    if (!player) return []
    const options: { zone: string; cardId: string; label: string }[] = []
    const handCount = seat === mySeat.value ? myHand.value.length : (player.hand_count ?? 0)
    if (handCount > 0) {
      options.push({ zone: 'hand', cardId: '', label: `手牌 ${handCount} 张` })
    }
    for (const equip of equippedCards(player)) {
      options.push({ zone: equipSlotOf(equip), cardId: equip.id, label: equipTagLabel(equip) })
    }
    for (const judge of judgeAreaCards(player)) {
      options.push({ zone: 'judge', cardId: judge.id, label: judge.name })
    }
    return options
  }

  function takeableTargetOptions() {
    const seat = shaTarget.value ?? opponentSeat.value
    return takeableOptionsForPlayer(seat)
  }

  const fankuiTargetOptions = computed(() =>
    isFankui.value ? takeableOptionsForPlayer(fankuiSourceSeat.value) : [],
  )
  const tuxiTargetOptions = computed(() =>
    isTuxiTake.value ? takeableOptionsForPlayer(tuxiSourceSeat.value) : [],
  )
  const qixiTargetOptions = computed(() => {
    if (!isQixiTake.value) return []
    return takeableOptionsForPlayer(qixiSourceSeat.value).filter((o) => o.zone === 'hand')
  })
  const pojunTargetOptions = computed(() =>
    isPojun.value ? takeableOptionsForPlayer(pojunVictimSeat.value) : [],
  )

  function selectedCardNeedsTargetCard(card = selectedCard.value) {
    return !!card && targetCardKinds.has(card.kind)
  }

  function canTargetSeat(seat: number, card: YzsCard | null | undefined) {
    if (!card) return false
    const target = seatAt(seat)
    if (!target || !enemySeats.value.includes(seat)) return false
    if (cardPlaysAsSha(card)) {
      if (isKongchengProtected(target)) return false
      return canPlaySha.value && distanceToSeat(seat) <= attackRangeOf()
    }
    if (!needsOpponentTarget(card)) return false
    if (card.kind === 'juedou' && isKongchengProtected(target)) return false
    if ((card.kind === 'guohe' || card.kind === 'tannang') && takeableOptionsForPlayer(seat).length === 0) {
      return false
    }
    if (card.kind === 'bingliang' && distanceToSeat(seat) > 1) return false
    return true
  }

  function canTargetOpponentWith(card: YzsCard | null | undefined) {
    if (!card) return false
    return enemySeats.value.some((s) => canTargetSeat(s, card))
  }

  function isSeatTargetable(seat: number) {
    return isMyPlay.value && canTargetSeat(seat, selectedCard.value)
  }

  function seatPanelClass(
    seat: number,
    isTeammate: boolean,
    seatRole?: 'protect' | 'mark' | 'farmer' | 'landlord' | 'commander' | 'forward',
  ) {
    const isProtect = isTeammate || seatRole === 'protect'
    return {
      'ddz__player--active': state.value?.current_turn === seat && !isFinished.value && !isResponse.value,
      'yzs__opponent-seat--targetable': !isProtect && isSeatTargetable(seat),
      'yzs__opponent-seat--targeted': shaTarget.value === seat,
      'yzs__seat--hit': hitFlashSeat.value === seat,
      'yzs__seat--block': blockFlashSeat.value === seat,
      'yzs__seat--teammate': isProtect,
      'yzs__seat--mark': seatRole === 'mark' || seatRole === 'farmer',
      'yzs__seat--commander': seatRole === 'commander',
    }
  }

  function onTargetSeat(seat: number) {
    if (!isMyPlay.value || !canTargetSeat(seat, selectedCard.value)) return
    shaTarget.value = seat
    if (selectedCardNeedsTargetCard() && selectedTargetZone.value === '') {
      const first = takeableOptionsForPlayer(seat)[0]
      if (first) {
        selectedTargetZone.value = first.zone
        selectedTargetCardId.value = first.cardId
      }
    }
  }

  function onTargetOpponent() {
    onTargetSeat(opponentSeat.value)
  }

  function pickFankuiTarget(zone: string, cardId = '') {
    if (!isFankui.value) return
    selectedTargetZone.value = zone
    selectedTargetCardId.value = cardId
  }

  function pickTuxiTarget(zone: string, cardId = '') {
    if (!isTuxiTake.value) return
    selectedTargetZone.value = zone
    selectedTargetCardId.value = cardId
  }

  function pickPojunTarget(zone: string, cardId = '') {
    if (!isPojun.value && !isPojunDiscard.value) return
    if (isPojunDiscard.value) {
      selectedTargetCardId.value = cardId
      return
    }
    selectedTargetZone.value = zone
    selectedTargetCardId.value = cardId
  }

  function pickOpponentCardTarget(zone: string, cardId = '') {
    if (!isMyPlay.value || !selectedCardNeedsTargetCard() || !canTargetOpponentWith(selectedCard.value)) return
    shaTarget.value = enemySeats.value[0] ?? opponentSeat.value
    selectedTargetZone.value = zone
    selectedTargetCardId.value = cardId
  }

  function syncWeaponSkillTargeting(next: import('../../types/yuzhousha').YuzhoushaState) {
    selectedQilinZone.value = ''
    if (next.phase === 'response' && next.pending?.response_mode === 'guanyu_follow') {
      shaTarget.value = next.pending.effect_target ?? opponentSeat.value
    }
  }

  return {
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
    opponentTargetKinds,
    targetCardKinds,
  }
}

export type YzsTargetingApi = ReturnType<typeof useYzsTargeting>
