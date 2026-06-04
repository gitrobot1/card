<script setup lang="ts">
import YzsCardView from './YzsCardView.vue'
import YzsStackedCards from './YzsStackedCards.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

defineProps<{
  centerClass?: string
}>()

const ctx = useYzsGameInject()
const {
  tableActionHint,
  isFinished,
  isDealing,
  centerHint,
  isMyDiscard,
  isMyPlay,
  isMyResponse,
  isMyPrepare,
  isPeekDeck,
  isGuicai,
  isGuidao,
  state,
  onPeekDragOver,
  onPeekZoneDrop,
  peekDeckTopIds,
  onPeekDragStart,
  onPeekDrop,
  onPeekDragEnd,
  peekDeckCard,
  peekDeckBottomIds,
  isWuguPick,
  selectedId,
  canInteract,
  selectCard,
  displayedTableCards,
} = ctx
</script>

<template>
  <div class="yzs__middle" :class="centerClass">
    <div class="ddz__center uno__center-ring yzs__center-ring">
      <div class="ddz__center-stage uno__center-stage yzs__center-stage">
        <div class="yzs__center-hint-slot">
          <p v-if="tableActionHint" class="yzs__table-action">{{ tableActionHint }}</p>
          <p
            v-else-if="isFinished && !isDealing"
            class="yzs__center-hint yzs__center-hint--result"
          >
            {{ centerHint }}
          </p>
          <p
            v-else-if="centerHint && centerHint !== '\u00a0'"
            class="yzs__center-hint"
            :class="{
              'yzs__center-hint--active':
                isMyDiscard || isMyPlay || isMyResponse || isMyPrepare || isPeekDeck,
            }"
          >
            {{ centerHint }}
          </p>
        </div>

        <div :ref="(el) => { ctx.playAreaRef.value = el as HTMLElement | null }" class="yzs__play-area">
          <div
            v-if="(isGuicai || isGuidao) && state?.pending?.judge_card"
            class="yzs__guicai-judge"
          >
            <p class="yzs__guicai-judge-label">当前判定</p>
            <YzsCardView :card="state.pending.judge_card" stacked />
          </div>
          <div
            v-if="isPeekDeck && state?.pending?.revealed_cards?.length"
            class="yzs__peek-deck"
          >
            <p class="yzs__peek-deck-title">观星 · 拖拽排列</p>
            <div class="yzs__peek-deck-col">
              <p class="yzs__peek-deck-label">
                牌堆顶 <span class="yzs__peek-deck-hint">最左先判定</span>
              </p>
              <div
                class="yzs__peek-deck-cards yzs__peek-deck-cards--ordered"
                @dragover="onPeekDragOver"
                @drop.prevent="onPeekZoneDrop('top')"
              >
                <div
                  v-for="(cardId, index) in peekDeckTopIds"
                  :key="'top-' + cardId"
                  class="yzs__peek-deck-slot"
                  draggable="true"
                  @dragstart="onPeekDragStart($event, 'top', index)"
                  @dragover="onPeekDragOver"
                  @drop.prevent.stop="onPeekDrop('top', index)"
                  @dragend="onPeekDragEnd"
                >
                  <YzsCardView v-if="peekDeckCard(cardId)" :card="peekDeckCard(cardId)!" stacked />
                </div>
              </div>
            </div>
            <div class="yzs__peek-deck-col">
              <p class="yzs__peek-deck-label">
                牌堆底 <span class="yzs__peek-deck-hint">拖入此区</span>
              </p>
              <div
                class="yzs__peek-deck-cards yzs__peek-deck-cards--ordered yzs__peek-deck-cards--bottom"
                @dragover="onPeekDragOver"
                @drop.prevent="onPeekZoneDrop('bottom')"
              >
                <div
                  v-for="(cardId, index) in peekDeckBottomIds"
                  :key="'bottom-' + cardId"
                  class="yzs__peek-deck-slot"
                  draggable="true"
                  @dragstart="onPeekDragStart($event, 'bottom', index)"
                  @dragover="onPeekDragOver"
                  @drop.prevent.stop="onPeekDrop('bottom', index)"
                  @dragend="onPeekDragEnd"
                >
                  <YzsCardView v-if="peekDeckCard(cardId)" :card="peekDeckCard(cardId)!" stacked />
                </div>
              </div>
            </div>
          </div>
          <div
            v-else-if="isWuguPick && state?.pending?.revealed_cards?.length"
            class="yzs__wugu-pick"
          >
            <button
              v-for="card in state.pending.revealed_cards"
              :key="card.id"
              type="button"
              class="yzs__wugu-pick-btn"
              :class="{ 'yzs__wugu-pick-btn--selected': selectedId === card.id }"
              :disabled="!canInteract"
              @click="selectCard(card.id)"
            >
              <YzsCardView :card="card" :selected="selectedId === card.id" />
            </button>
          </div>
          <div v-if="displayedTableCards.length" class="yzs__last-play">
            <YzsStackedCards :cards="displayedTableCards" :max-width="520" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
