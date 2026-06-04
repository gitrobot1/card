<script setup lang="ts">
import YzsCardView from './YzsCardView.vue'
import SeatIndicator from '../doudizhu/SeatIndicator.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

const g = useYzsGameInject()
const {
  blockFlashSeat,
  canInteract,
  cardPlaysAsShan,
  clearFanjianMode,
  clearGuoseMode,
  clearJieyinMode,
  clearQixiMode,
  clearRendeMode,
  clearShuangxiongMode,
  clearZhihengMode,
  enteringDrawCardIds,
  equipTagLabel,
  equipTagTitle,
  equippedCards,
  fanjianMode,
  fanjianSelectedId,
  ganglieDiscardIds,
  ddzCancelDiscardIds,
  guoseMode,
  guoseSelectedId,
  handLayoutStyle,
  hitFlashSeat,
  isDealing,
  isFinished,
  isMyDiscard,
  isMyTurn,
  isRedCard,
  isResponse,
  isSkillActivatable,
  jieyinMode,
  jieyinSelectedIds,
  judgeAreaCards,
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
  zhihengSelectedIds
} = g
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
                  :disabled="!canInteract"
                  :plays-as-sha="wushengMode && isRedCard(card) && card.kind !== 'sha'"
                  :plays-as-shan="cardPlaysAsShan(card) && card.kind !== 'shan'"
                  @click="selectCard(card.id)"
                />
              </div>
            </div>
          </div>

          <div class="ddz__hand-side yzs__hand-side">
            <div class="ddz__seat-stack ddz__seat-stack--self">
              <div
                class="ddz__player ddz__player--self ddz__seat-anchor"
                :class="{
                  'ddz__player--active': (isMyTurn || isMyDiscard) && !isFinished && !isResponse,
                  'yzs__seat--hit': hitFlashSeat === mySeat,
                  'yzs__seat--block': blockFlashSeat === mySeat,
                }"
                :data-seat="mySeat"
              >
                <span class="ddz__badge ddz__badge--role">{{ myPlayer?.character.name?.slice(0, 1) }}</span>
                <span>我</span>
                <span class="yzs__hero-name">{{ myPlayer?.character.name }}</span>
                <span class="yzs__hp">♥ {{ myPlayer?.hp }}/{{ myPlayer?.max_hp }}</span>
                <span class="ddz__count">{{ myHand.length }} 张</span>
                <span v-if="myPlayer?.drunk" class="yzs__equip-tag yzs__equip-tag--buff">酒</span>
              </div>
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
                <button
                  v-if="qixiMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--confirm"
                  :disabled="qixiSelectedId === '' || !canInteract"
                  @click="submitSkill('qixi')"
                >
                  发动奇袭
                </button>
                <button
                  v-if="qixiMode"
                  type="button"
                  class="yzs__skill-btn yzs__skill-btn--cancel"
                  @click="clearQixiMode"
                >
                  取消
                </button>
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
                <p v-if="wushengSkillHint" class="yzs__skill-hint">{{ wushengSkillHint }}</p>
              </div>
              <div v-if="equippedCards(myPlayer).length" class="yzs__equip-row">
                <span
                  v-for="equip in equippedCards(myPlayer)"
                  :key="equip.id"
                  class="yzs__equip-tag"
                  :title="equipTagTitle(equip)"
                >
                  {{ equipTagLabel(equip) }}
                </span>
              </div>
              <div v-if="judgeAreaCards(myPlayer).length" class="yzs__equip-row">
                <span
                  v-for="judge in judgeAreaCards(myPlayer)"
                  :key="judge.id"
                  class="yzs__equip-tag yzs__equip-tag--judge"
                >
                  {{ judge.name }}
                </span>
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
</template>
