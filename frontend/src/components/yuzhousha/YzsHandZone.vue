<script setup lang="ts">
import { computed, ref } from 'vue'
import YzsCardView from './YzsCardView.vue'
import YzsCardPicker from './YzsCardPicker.vue'
import SeatIndicator from '../doudizhu/SeatIndicator.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'
import { useYuzhoushaSkill } from '../../api/games'
import { equippedCards, equipSlotOf, judgeAreaCards } from '../../composables/yuzhousha/playerCardHelpers'
import { suitColor, suitSymbol } from '../../constants/games'

const g = useYzsGameInject()
const {
  blockFlashSeat,
  canInteract,
  cardPlaysAsShan,
  clearFanjianMode,
  clearGuoseMode,
  clearJieyinMode,
  clearRendeMode,
  clearShuangxiongMode,
  clearZhihengMode,
  enteringDrawCardIds,
  equipTagLabel,
  equipTagTitle,
  fanjianMode,
  fanjianSelectedId,
  ganglieDiscardIds,
  ddzCancelDiscardIds,
  guoseMode,
  guoseSelectedId,
  handLayoutStyle,
  hitFlashSeat,
  isBlackCard,
  isDealing,
  isDiamondCard,
  isFinished,
  isMyDiscard,
  isMyTurn,
  isRedCard,
  isResponse,
  isSkillActivatable,
  cardValidForShuangxiong,
  jieyinMode,
  jieyinSelectedIds,
  liuliSelectedId,
  myCharacterSkills,
  myHand,
  myPlayer,
  mySeat,
  onCharacterSkillClick,
  qixiMode,
  qixiSelectedId,
  rendeMode,
  rendeSelectedIds,
  secondsLeft,
  selectCard,
  selectedDiscardIds,
  selectedId,
  shaTarget,
  showSeatTimer,
  shuangxiongMode,
  shuangxiongSelectedId,
  submitSkill,
  wushengMode,
  wushengSkillHint,
  yijiSelectedIds,
  zhihengMode,
  zhihengSelectedIds,
  state,
  isFankui,
  isTuxiTake,
  isQixiTake,
  fankuiTargetOptions,
  tuxiTargetOptions,
  qixiTargetOptions,
  weaponRange,
  tiesuoMode,
  tiesuoTargets,
  handleSeatTarget,
} = g

// 铁索连环：自己也可被选为目标
const isSelfTiesuoTargetable = computed(
  () => tiesuoMode.value && !!myPlayer.value && (myPlayer.value.hp ?? 0) > 0,
)
const isSelfTiesuoSelected = computed(
  () => tiesuoMode.value && tiesuoTargets.value.includes(mySeat.value),
)
const isSelfChained = computed(
  () => !!myPlayer.value?.skill_counters?.chained,
)
function onSelfSeatClick() {
  console.log('[tiesuo] onSelfSeatClick: tiesuoMode=', tiesuoMode.value, 'isSelfTiesuoTargetable=', isSelfTiesuoTargetable.value, 'mySeat=', mySeat.value)
  if (isSelfTiesuoTargetable.value) {
    handleSeatTarget(mySeat.value)
  }
}

// 取牌选择器状态（破军已移至 YuzhoushaView 弹窗）
const pickerState = computed<{
  title: string
  skillId: string
  options: { zone: string; cardId: string }[]
  zoneFilter?: string[]
  showHand: boolean
} | null>(() => {
  if (isFankui.value && fankuiTargetOptions.value.length) {
    return { title: '【反馈】：选择来源的一张牌', skillId: 'fankui', options: fankuiTargetOptions.value, showHand: false }
  }
  if (isTuxiTake.value && tuxiTargetOptions.value.length) {
    return { title: '【突袭】：选择获得对手的一张牌', skillId: 'tuxi', options: tuxiTargetOptions.value, showHand: false }
  }
  if (isQixiTake.value && qixiTargetOptions.value.length) {
    return { title: '【奇袭】：选择一张黑色牌', skillId: 'qixi', options: qixiTargetOptions.value, showHand: true }
  }
  return null
})

const pickerPlayer = computed(() => {
  if (!pickerState.value) return null
  const seat = state.value?.pending?.subject_seat ?? state.value?.pending?.target_index ?? -1
  return state.value?.players?.[seat] ?? null
})

const pickerSelectedIds = ref<string[]>([])

async function onPickerToggle(option: { zone: string; card: { id: string } }) {
  if (!pickerState.value) return
  const gameId = state.value?.id
  if (!gameId) return
  await useYuzhoushaSkill(gameId, pickerState.value.skillId, {
    targetZone: option.zone,
    targetCardId: option.card.id,
  })
}

async function onPickerCancel() {
  if (!pickerState.value) return
  const gameId = state.value?.id
  if (!gameId) return
  await useYuzhoushaSkill(gameId, pickerState.value.skillId, {})
  pickerSelectedIds.value = []
}
</script>

<template>
      <div class="ddz__bottom-zone yzs__bottom-zone">
        <div class="ddz__hand yzs__hand-row">
          <div
            :ref="(el) => { g.handAreaRef.value = el as HTMLElement | null }"
            class="hand-cards yzs__hand-cards"
            :class="{ 'hand-cards--dealing': isDealing }"
            :style="handLayoutStyle"
          >
            <div class="hand-cards__row">
              <div
                v-for="card in myHand"
                :key="card.id"
                class="hand-cards__slot"
                :class="{ 'yzs__hand-slot--draw-enter': enteringDrawCardIds.includes(card.id) }"
              >
                <YzsCardView
                  :card="card"
                  :selected="isMyDiscard ? selectedDiscardIds.includes(card.id) : selectedId === card.id || rendeSelectedIds.includes(card.id) || zhihengSelectedIds.includes(card.id) || jieyinSelectedIds.includes(card.id) || fanjianSelectedId === card.id || qixiSelectedId === card.id || guoseSelectedId === card.id || liuliSelectedId === card.id || ganglieDiscardIds.includes(card.id) || ddzCancelDiscardIds.includes(card.id) || yijiSelectedIds.includes(card.id)"
                  :disabled="!canInteract || (wushengMode && !isRedCard(card)) || (qixiMode && !isBlackCard(card)) || (guoseMode && !isDiamondCard(card)) || (shuangxiongMode && !cardValidForShuangxiong(card))"
                  :plays-as-sha="wushengMode && isRedCard(card) && card.kind !== 'sha'"
                  :plays-as-shan="cardPlaysAsShan(card) && card.kind !== 'shan'"
                  @click="selectCard(card.id)"
                />
              </div>
            </div>
          </div>

          <div class="ddz__hand-side yzs__hand-side">
            <!-- 自己竖长头像卡片（左右布局：左侧血量竖排，右侧装备+技能） -->
            <div
              class="yzs__hero-card yzs__hero-card--self"
              :class="{
                'yzs__hero-card--active': (isMyTurn || isMyDiscard) && !isFinished && !isResponse,
                'yzs__seat--hit': hitFlashSeat === mySeat,
                'yzs__seat--block': blockFlashSeat === mySeat,
                'yzs__hero-card--tiesuo-targetable': isSelfTiesuoTargetable,
                'yzs__hero-card--tiesuo-selected': isSelfTiesuoSelected,
                'yzs__hero-card--chained': isSelfChained,
                'yzs__hero-card--dead': (myPlayer?.hp ?? 0) <= 0,
              }"
              :data-seat="mySeat"
              @click="onSelfSeatClick"
            >
              <!-- 左侧：血量竖排 + 手牌数（右下角，手牌数在最底部） -->
              <div class="yzs__hero-left">
                <div class="yzs__hero-hp-col">
                  <span v-for="i in (myPlayer?.max_hp ?? 0)" :key="i" class="yzs__hp-dot" :class="{ 'yzs__hp-dot--lost': i > (myPlayer?.hp ?? 0) }" />
                </div>
                <span v-if="myPlayer" class="yzs__hero-hand-tag">{{ myHand.length }}</span>
              </div>

              <!-- 右侧：武将名 + 装备区 + 判定区 + 标记 -->
              <div class="yzs__hero-right">
                <!-- 武将名（居中） -->
                <div class="yzs__hero-name-btn yzs__hero-name-btn--self">
                  <span>{{ myPlayer?.character.name }}</span>
                </div>

                <!-- 装备区（4行固定占位，变牌模式下可点击选择） -->
                <div class="yzs__hero-equips">
                  <div
                    class="yzs__equip-line"
                    :class="{
                      'yzs__equip-line--filled': !!myPlayer?.weapon,
                      'yzs__equip-line--clickable': !!myPlayer?.weapon && canInteract && (
                        (wushengMode && isRedCard(myPlayer?.weapon)) ||
                        (qixiMode && isBlackCard(myPlayer?.weapon)) ||
                        (guoseMode && isDiamondCard(myPlayer?.weapon)) ||
                        shuangxiongMode
                      )
                    }"
                    :title="myPlayer?.weapon ? equipTagTitle(myPlayer.weapon) : '武器'"
                    @click="myPlayer?.weapon && selectCard(myPlayer.weapon.id)"
                  >
                    <template v-if="myPlayer?.weapon">
                      <span class="yzs__equip-suit" :class="`yzs__equip-suit--${suitColor(myPlayer.weapon.suit)}`">{{ suitSymbol(myPlayer.weapon.suit) }}</span>
                      <span class="yzs__equip-name">{{ equipTagLabel(myPlayer.weapon) }}</span>
                      <span class="yzs__equip-range">{{ weaponRange(myPlayer.weapon.kind) }}</span>
                    </template>
                    <template v-else>
                      <span class="yzs__equip-placeholder">武器</span>
                    </template>
                  </div>
                  <div
                    class="yzs__equip-line"
                    :class="{
                      'yzs__equip-line--filled': !!myPlayer?.armor,
                      'yzs__equip-line--clickable': !!myPlayer?.armor && canInteract && (
                        (wushengMode && isRedCard(myPlayer?.armor)) ||
                        (qixiMode && isBlackCard(myPlayer?.armor)) ||
                        (guoseMode && isDiamondCard(myPlayer?.armor)) ||
                        shuangxiongMode
                      )
                    }"
                    :title="myPlayer?.armor ? equipTagTitle(myPlayer.armor) : '防具'"
                    @click="myPlayer?.armor && selectCard(myPlayer.armor.id)"
                  >
                    <template v-if="myPlayer?.armor">
                      <span class="yzs__equip-suit" :class="`yzs__equip-suit--${suitColor(myPlayer.armor.suit)}`">{{ suitSymbol(myPlayer.armor.suit) }}</span>
                      <span class="yzs__equip-name">{{ equipTagLabel(myPlayer.armor) }}</span>
                    </template>
                    <template v-else>
                      <span class="yzs__equip-placeholder">防具</span>
                    </template>
                  </div>
                  <div
                    class="yzs__equip-line"
                    :class="{
                      'yzs__equip-line--filled': !!myPlayer?.plus_horse,
                      'yzs__equip-line--clickable': !!myPlayer?.plus_horse && canInteract && (
                        (wushengMode && isRedCard(myPlayer?.plus_horse)) ||
                        (qixiMode && isBlackCard(myPlayer?.plus_horse)) ||
                        (guoseMode && isDiamondCard(myPlayer?.plus_horse)) ||
                        shuangxiongMode
                      )
                    }"
                    :title="myPlayer?.plus_horse ? equipTagTitle(myPlayer.plus_horse) : '+1马'"
                    @click="myPlayer?.plus_horse && selectCard(myPlayer.plus_horse.id)"
                  >
                    <template v-if="myPlayer?.plus_horse">
                      <span class="yzs__equip-suit" :class="`yzs__equip-suit--${suitColor(myPlayer.plus_horse.suit)}`">{{ suitSymbol(myPlayer.plus_horse.suit) }}</span>
                      <span class="yzs__equip-name">+1马</span>
                    </template>
                    <template v-else>
                      <span class="yzs__equip-placeholder">+1马</span>
                    </template>
                  </div>
                  <div
                    class="yzs__equip-line"
                    :class="{
                      'yzs__equip-line--filled': !!myPlayer?.minus_horse,
                      'yzs__equip-line--clickable': !!myPlayer?.minus_horse && canInteract && (
                        (wushengMode && isRedCard(myPlayer?.minus_horse)) ||
                        (qixiMode && isBlackCard(myPlayer?.minus_horse)) ||
                        (guoseMode && isDiamondCard(myPlayer?.minus_horse)) ||
                        shuangxiongMode
                      )
                    }"
                    :title="myPlayer?.minus_horse ? equipTagTitle(myPlayer.minus_horse) : '-1马'"
                    @click="myPlayer?.minus_horse && selectCard(myPlayer.minus_horse.id)"
                  >
                    <template v-if="myPlayer?.minus_horse">
                      <span class="yzs__equip-suit" :class="`yzs__equip-suit--${suitColor(myPlayer.minus_horse.suit)}`">{{ suitSymbol(myPlayer.minus_horse.suit) }}</span>
                      <span class="yzs__equip-name">-1马</span>
                    </template>
                    <template v-else>
                      <span class="yzs__equip-placeholder">-1马</span>
                    </template>
                  </div>
                </div>

                <!-- 判定区 -->
                <div v-if="judgeAreaCards(myPlayer).length" class="yzs__hero-judge">
                  <div v-for="judge in judgeAreaCards(myPlayer)" :key="judge.id" class="yzs__equip-line yzs__equip-line--judge">
                    <span class="yzs__equip-name">{{ judge.name }}</span>
                  </div>
                </div>

                <!-- 标记 -->
                <div v-if="myPlayer?.drunk || myPlayer?.skill_counters?.pojun_gain_pending" class="yzs__hero-marks">
                  <span v-if="myPlayer?.drunk" class="yzs__equip-tag yzs__equip-tag--buff">酒</span>
                  <span v-if="myPlayer?.skill_counters?.pojun_gain_pending" class="yzs__equip-tag yzs__equip-tag--mark">营</span>
                </div>
              </div>
            </div>

            <div class="ddz__seat-stack ddz__seat-stack--self">
              <div v-if="myCharacterSkills.length" class="yzs__skill-bar">
                <button
                  v-for="skill in myCharacterSkills"
                  :key="skill.id"
                  type="button"
                  class="yzs__skill-btn"
                  :class="{
                    'yzs__skill-btn--active':
                      (skill.id === 'wusheng' && wushengMode) ||
                      (skill.id === 'rende' && rendeMode) ||
                      (skill.id === 'zhiheng' && zhihengMode) ||
                      (skill.id === 'jieyin' && jieyinMode) ||
                      (skill.id === 'fanjian' && fanjianMode) ||
                      (skill.id === 'qixi' && qixiMode) ||
                      (skill.id === 'guose' && guoseMode) ||
                      (skill.id === 'shuangxiong' && shuangxiongMode),
                    'yzs__skill-btn--passive': skill.kind === 'passive',
                    'yzs__skill-btn--awakening': skill.kind === 'awakening',
                    'yzs__skill-btn--locked': skill.id === 'paoxiao' || skill.id === 'longdan',
                    'yzs__skill-btn--lord-inactive': skill.inactive_in_1v1,
                  }"
                  :title="skill.desc"
                  :disabled="!canInteract || !isSkillActivatable(skill)"
                  @click="onCharacterSkillClick(skill)"
                >
                  {{ skill.name }}
                </button>
                <button
                  v-if="rendeMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--confirm"
                  :disabled="rendeSelectedIds.length === 0 || shaTarget == null || !canInteract"
                  @click="submitSkill('rende')"
                >
                  发动仁德
                </button>
                <button
                  v-if="rendeMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--cancel"
                  @click="clearRendeMode"
                >
                  取消
                </button>
                <button
                  v-if="zhihengMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--confirm"
                  :disabled="zhihengSelectedIds.length === 0 || !canInteract"
                  @click="submitSkill('zhiheng')"
                >
                  发动制衡
                </button>
                <button
                  v-if="zhihengMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--cancel"
                  @click="clearZhihengMode"
                >
                  取消
                </button>
                <button
                  v-if="jieyinMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--confirm"
                  :disabled="jieyinSelectedIds.length !== 2 || shaTarget == null || !canInteract"
                  @click="submitSkill('jieyin')"
                >
                  发动结姻
                </button>
                <button
                  v-if="jieyinMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--cancel"
                  @click="clearJieyinMode"
                >
                  取消
                </button>
                <button
                  v-if="fanjianMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--confirm"
                  :disabled="fanjianSelectedId === '' || !canInteract"
                  @click="submitSkill('fanjian')"
                >
                  发动反间
                </button>
                <button
                  v-if="fanjianMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--cancel"
                  @click="clearFanjianMode"
                >
                  取消
                </button>
                <!-- 奇袭不再需要单独的取消按钮，再点技能按钮即可取消 -->
                <button
                  v-if="guoseMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--confirm"
                  :disabled="guoseSelectedId === '' || !canInteract"
                  @click="submitSkill('guose')"
                >
                  发动国色
                </button>
                <button
                  v-if="guoseMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--cancel"
                  @click="clearGuoseMode"
                >
                  取消
                </button>
                <button
                  v-if="shuangxiongMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--confirm"
                  :disabled="shuangxiongSelectedId === '' || !canInteract"
                  @click="submitSkill('shuangxiong')"
                >
                  当决斗
                </button>
                <button
                  v-if="shuangxiongMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--cancel"
                  @click="clearShuangxiongMode"
                >
                  取消
                </button>

              </div>
              <SeatIndicator
                placement="top"
                :show-timer="showSeatTimer(mySeat)"
                :seconds="secondsLeft"
              />
            </div>
          </div>
        </div>
      </div>

  <YzsCardPicker
    v-if="pickerState && pickerPlayer"
    :title="pickerState.title"
    :player="pickerPlayer"
    :show-hand="pickerState.showHand"
    :zone-filter="pickerState.zoneFilter"
    :multi="false"
    :max-select="1"
    :selected-ids="pickerSelectedIds"
    :show-actions="pickerState.options.length > 0"
    :visible="!!pickerState"
    @toggle="onPickerToggle"
    @cancel="onPickerCancel"
  />
</template>
