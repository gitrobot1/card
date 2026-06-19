<script setup lang="ts">
import { computed } from 'vue'
import SeatIndicator from '../doudizhu/SeatIndicator.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'
import { isIdentityMode } from '../../constants/yzsModes'
import { suitColor, suitSymbol } from '../../constants/games'

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
  equippedCards,
  equipTagTitle,
  equipTagLabel,
  judgeAreaCards,
  showSeatSkillPanels,
  isQilinBow,
  qilinHorseOptions,
  selectedQilinZone,
  isFankui,
  fankuiTargetOptions,
  pickFankuiTarget,
  isTuxiTake,
  tuxiTargetOptions,
  pickTuxiTarget,
  isQixiTake,
  qixiTargetOptions,
  isMyPlay,
  selectedCardNeedsTargetCard,
  takeableTargetOptions,
  selectedTargetZone,
  selectedTargetCardId,
  pickOpponentCardTarget,
  showSeatTimer,
  secondsLeft,
  weaponRange,
} = useYzsGameInject()

const player = computed(() => seatAt(props.seat))
const isIdentityModeActive = computed(() => isIdentityMode(state.value?.mode))
const takenSeat = computed(() => state.value?.pending?.subject_seat ?? -1)

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
    <!-- 竖长角色卡片（左右布局：左侧血量竖排，右侧装备+技能） -->
    <div
      class="yzs__hero-card"
      :class="[
        `yzs__hero-card--${placement}`,
        seatPanelClass(seat, isTeammate, seatRole),
        { 'yzs__hero-card--targetable': isSeatTargetable(seat) },
      ]"
      :data-seat="seat"
      @click="isSeatTargetable(seat) && onTargetSeat(seat)"
    >

      <!-- 左侧：血量竖排 + 手牌数（右下角，手牌数在最底部） -->
      <div class="yzs__hero-left">
        <div class="yzs__hero-hp-col">
          <span v-for="i in (player?.max_hp ?? 0)" :key="i" class="yzs__hp-dot" :class="{ 'yzs__hp-dot--lost': i > (player?.hp ?? 0) }" />
        </div>
        <span v-if="player" class="yzs__hero-hand-tag">{{ player.hand_count ?? 0 }}</span>
      </div>

      <!-- 右侧：武将名 + 装备区 + 判定区 + 标记 -->
      <div class="yzs__hero-right">
        <!-- 武将名（居中） -->
        <div class="yzs__hero-name-btn yzs__hero-name-btn--self">
          <span>{{ player?.character.name }}</span>
        </div>

        <!-- 装备区（4行固定占位） -->
        <div class="yzs__hero-equips">
          <div class="yzs__equip-line" :class="{ 'yzs__equip-line--filled': !!player?.weapon }" :title="player?.weapon ? equipTagTitle(player.weapon) : '武器'">
            <template v-if="player?.weapon">
              <span class="yzs__equip-suit" :class="`yzs__equip-suit--${suitColor(player.weapon.suit)}`">{{ suitSymbol(player.weapon.suit) }}</span>
              <span class="yzs__equip-name">{{ equipTagLabel(player.weapon) }}</span>
              <span class="yzs__equip-range">{{ weaponRange(player.weapon.kind) }}</span>
            </template>
            <template v-else>
              <span class="yzs__equip-placeholder">武器</span>
            </template>
          </div>
          <div class="yzs__equip-line" :class="{ 'yzs__equip-line--filled': !!player?.armor }" :title="player?.armor ? equipTagTitle(player.armor) : '防具'">
            <template v-if="player?.armor">
              <span class="yzs__equip-suit" :class="`yzs__equip-suit--${suitColor(player.armor.suit)}`">{{ suitSymbol(player.armor.suit) }}</span>
              <span class="yzs__equip-name">{{ equipTagLabel(player.armor) }}</span>
            </template>
            <template v-else>
              <span class="yzs__equip-placeholder">防具</span>
            </template>
          </div>
          <div class="yzs__equip-line" :class="{ 'yzs__equip-line--filled': !!player?.plus_horse }" :title="player?.plus_horse ? equipTagTitle(player.plus_horse) : '+1马'">
            <template v-if="player?.plus_horse">
              <span class="yzs__equip-suit" :class="`yzs__equip-suit--${suitColor(player.plus_horse.suit)}`">{{ suitSymbol(player.plus_horse.suit) }}</span>
              <span class="yzs__equip-name">+1马</span>
            </template>
            <template v-else>
              <span class="yzs__equip-placeholder">+1马</span>
            </template>
          </div>
          <div class="yzs__equip-line" :class="{ 'yzs__equip-line--filled': !!player?.minus_horse }" :title="player?.minus_horse ? equipTagTitle(player.minus_horse) : '-1马'">
            <template v-if="player?.minus_horse">
              <span class="yzs__equip-suit" :class="`yzs__equip-suit--${suitColor(player.minus_horse.suit)}`">{{ suitSymbol(player.minus_horse.suit) }}</span>
              <span class="yzs__equip-name">-1马</span>
            </template>
            <template v-else>
              <span class="yzs__equip-placeholder">-1马</span>
            </template>
          </div>
        </div>

        <!-- 判定区 -->
        <div v-if="judgeAreaCards(player).length" class="yzs__hero-judge">
          <div v-for="judge in judgeAreaCards(player)" :key="judge.id" class="yzs__equip-line yzs__equip-line--judge">
            <span class="yzs__equip-name">{{ judge.name }}</span>
          </div>
        </div>

        <!-- 标记（酒/营等） -->
        <div v-if="player?.drunk || player?.skill_counters?.pojun_gain_pending" class="yzs__hero-marks">
          <span v-if="player?.drunk" class="yzs__equip-tag yzs__equip-tag--buff">酒</span>
          <span v-if="player?.skill_counters?.pojun_gain_pending" class="yzs__equip-tag yzs__equip-tag--mark">营</span>
        </div>
      </div>
    </div>

    <!-- 可选目标/技能面板 -->
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
        v-if="isFankui && takenSeat === seat && fankuiTargetOptions.length"
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
        v-if="isTuxiTake && takenSeat === seat && tuxiTargetOptions.length"
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
        v-if="isQixiTake && takenSeat === seat && qixiTargetOptions.length"
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
