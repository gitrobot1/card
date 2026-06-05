<script setup lang="ts">
import { computed } from 'vue'
import SeatIndicator from '../doudizhu/SeatIndicator.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'
import { isIdentityMode } from '../../constants/yzsModes'

const props = withDefaults(
  defineProps<{
    seat: number
    placement: 'left' | 'right' | 'top'
    isTeammate?: boolean
    seatRole?: 'protect' | 'mark' | 'landlord' | 'farmer' | 'commander' | 'forward'
    stackClass?: string
  }>(),
  { isTeammate: false, stackClass: '' },
)

const {
  state,
  seatAt,
  seatPanelClass,
  isSeatTargetable,
  onTargetSeat,
  seatHandCount,
  equippedCards,
  equipTagTitle,
  equipTagLabel,
  judgeAreaCards,
  showSeatSkillPanels,
  isQilinBow,
  qilinHorseOptions,
  selectedQilinZone,
  isFankui,
  fankuiSourceSeat,
  fankuiTargetOptions,
  pickFankuiTarget,
  isTuxiTake,
  tuxiSourceSeat,
  tuxiTargetOptions,
  pickTuxiTarget,
  isQixiTake,
  qixiSourceSeat,
  qixiTargetOptions,
  isMyPlay,
  selectedCardNeedsTargetCard,
  takeableTargetOptions,
  selectedTargetZone,
  selectedTargetCardId,
  pickOpponentCardTarget,
  showSeatTimer,
  secondsLeft,
} = useYzsGameInject()

const player = computed(() => seatAt(props.seat))
const isIdentityModeActive = computed(() => isIdentityMode(state.value?.mode))

function identityLabel(identity?: string) {
  switch (identity) {
    case 'lord':
      return '主'
    case 'loyalist':
      return '忠'
    case 'spy':
      return '内'
    case 'rebel':
      return '反'
    default:
      return ''
  }
}

const roleBadge = computed(() => {
  const p = player.value
  if (p?.identity) {
    if (p.identity === 'lord' || p.identity_revealed) {
      return identityLabel(p.identity) || '?'
    }
    return '?'
  }
  if (props.seatRole === 'protect') return '保'
  if (props.seatRole === 'mark') return '杀'
  if (props.seatRole === 'landlord') return '主'
  if (props.seatRole === 'farmer') return '农'
  if (props.seatRole === 'commander') return '帅'
  if (props.seatRole === 'forward') return '锋'
  if (props.isTeammate) return '友'
  return player.value?.character.name?.slice(0, 1) ?? '敌'
})
const identityBadgeClass = computed(() => {
  const p = player.value
  if (!p?.identity) return {}
  if (!p.identity_revealed && p.identity !== 'lord') {
    return { 'ddz__badge--identity-hidden': true }
  }
  return {
    'ddz__badge--lord': p.identity === 'lord',
    'ddz__badge--loyalist': p.identity === 'loyalist',
    'ddz__badge--spy': p.identity === 'spy',
    'ddz__badge--rebel': p.identity === 'rebel',
  }
})
const isProtectSeat = computed(
  () => !isIdentityModeActive.value && (props.isTeammate || props.seatRole === 'protect'),
)
</script>

<template>
  <div class="ddz__seat-stack" :class="stackClass">
    <button
      type="button"
      class="ddz__player ddz__player--compact ddz__seat-anchor yzs__opponent-seat"
      :class="seatPanelClass(seat, isTeammate, seatRole)"
      :data-seat="seat"
      :disabled="isProtectSeat || !isSeatTargetable(seat)"
      @click="onTargetSeat(seat)"
    >
      <span
        class="ddz__badge ddz__badge--role"
        :class="[
          identityBadgeClass,
          {
            'ddz__badge--ally': isProtectSeat,
            'ddz__badge--mark': !isIdentityModeActive && (seatRole === 'mark' || seatRole === 'farmer'),
          },
        ]"
      >
        {{ roleBadge }}
      </span>
      <span>{{ player?.name }}</span>
      <span class="yzs__hero-name">{{ player?.character.name }}</span>
      <span class="yzs__hp">♥ {{ player?.hp }}/{{ player?.max_hp }}</span>
      <span class="ddz__count">{{ seatHandCount(seat) }} 张</span>
      <span v-if="player?.drunk" class="yzs__equip-tag yzs__equip-tag--buff">酒</span>
    </button>
    <div v-if="equippedCards(player).length" class="yzs__equip-row">
      <span
        v-for="equip in equippedCards(player)"
        :key="equip.id"
        class="yzs__equip-tag"
        :title="equipTagTitle(equip)"
      >
        {{ equipTagLabel(equip) }}
      </span>
    </div>
    <div v-if="judgeAreaCards(player).length" class="yzs__equip-row">
      <span
        v-for="judge in judgeAreaCards(player)"
        :key="judge.id"
        class="yzs__equip-tag yzs__equip-tag--judge"
      >
        {{ judge.name }}
      </span>
    </div>
    <template v-if="showSeatSkillPanels(seat)">
      <div v-if="isQilinBow && qilinHorseOptions.length" class="yzs__target-card-row">
        <button
          v-for="option in qilinHorseOptions"
          :key="option.zone"
          type="button"
          class="yzs__target-card-btn"
          :class="{ 'yzs__target-card-btn--active': selectedQilinZone === option.zone }"
          @click="selectedQilinZone = option.zone"
        >
          弃 {{ option.label }}
        </button>
      </div>
      <div
        v-if="isFankui && fankuiSourceSeat === seat && fankuiTargetOptions.length"
        class="yzs__target-card-row"
      >
        <button
          v-for="option in fankuiTargetOptions"
          :key="`fk-${option.zone}:${option.cardId}`"
          type="button"
          class="yzs__target-card-btn"
          :class="{
            'yzs__target-card-btn--active':
              selectedTargetZone === option.zone && selectedTargetCardId === option.cardId,
          }"
          @click="pickFankuiTarget(option.zone, option.cardId)"
        >
          {{ option.label }}
        </button>
      </div>
      <div
        v-if="isTuxiTake && tuxiSourceSeat === seat && tuxiTargetOptions.length"
        class="yzs__target-card-row"
      >
        <button
          v-for="option in tuxiTargetOptions"
          :key="`tx-${option.zone}:${option.cardId}`"
          type="button"
          class="yzs__target-card-btn"
          :class="{
            'yzs__target-card-btn--active':
              selectedTargetZone === option.zone && selectedTargetCardId === option.cardId,
          }"
          @click="pickTuxiTarget(option.zone, option.cardId)"
        >
          {{ option.label }}
        </button>
      </div>
      <div
        v-if="isQixiTake && qixiSourceSeat === seat && qixiTargetOptions.length"
        class="yzs__target-card-row"
      >
        <button
          v-for="option in qixiTargetOptions"
          :key="`qx-${option.zone}:${option.cardId}`"
          type="button"
          class="yzs__target-card-btn"
          :class="{
            'yzs__target-card-btn--active':
              selectedTargetZone === option.zone && selectedTargetCardId === option.cardId,
          }"
          @click="pickTuxiTarget(option.zone, option.cardId)"
        >
          {{ option.label }}
        </button>
      </div>
      <div
        v-if="isMyPlay && selectedCardNeedsTargetCard() && takeableTargetOptions().length"
        class="yzs__target-card-row"
      >
        <button
          v-for="option in takeableTargetOptions()"
          :key="`${option.zone}:${option.cardId}`"
          type="button"
          class="yzs__target-card-btn"
          :class="{
            'yzs__target-card-btn--active':
              selectedTargetZone === option.zone && selectedTargetCardId === option.cardId,
          }"
          @click="pickOpponentCardTarget(option.zone, option.cardId)"
        >
          {{ option.label }}
        </button>
      </div>
    </template>
    <SeatIndicator
      :placement="placement"
      :show-timer="showSeatTimer(seat)"
      :seconds="secondsLeft"
    />
  </div>
</template>
