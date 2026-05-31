<script setup lang="ts">
import { computed } from 'vue'
import PlayingCard from './PlayingCard.vue'
import type { Card } from '../../types/doudizhu'

const props = withDefaults(
  defineProps<{
    cards: Card[]
    /** 可用宽度上限，默认随屏幕 */
    maxWidth?: number
    mini?: boolean
    /** 结算亮牌：间距更大，保证每张牌点数可见 */
    reveal?: boolean
  }>(),
  {
    maxWidth: undefined,
    mini: false,
    reveal: false,
  },
)

const rowStyle = computed(() => {
  const count = props.cards.length
  const cardWidth = props.reveal ? 34 : props.mini ? 40 : 82
  if (count <= 1) {
    return { '--stack-step': `${cardWidth}px`, '--stack-card-width': `${cardWidth}px` }
  }
  const minStep = props.reveal ? 18 : props.mini ? 10 : 28
  const preferStep = props.reveal ? 22 : props.mini ? 14 : 38
  const maxRowWidth =
    props.maxWidth ??
    (props.reveal ? 520 : props.mini ? 140 : Math.min(window.innerWidth - 120, 560))
  const fitStep = (maxRowWidth - cardWidth) / (count - 1)
  const step = Math.max(minStep, Math.min(preferStep, fitStep))
  return { '--stack-step': `${step}px`, '--stack-card-width': `${cardWidth}px` }
})
</script>

<template>
  <div
    class="stacked-cards"
    :class="{
      'stacked-cards--mini': mini && !reveal,
      'stacked-cards--reveal': reveal,
    }"
  >
    <div class="stacked-cards__row" :style="rowStyle">
      <div
        v-for="(card, index) in cards"
        :key="card.id"
        class="stacked-cards__slot"
        :style="{ zIndex: index + 1 }"
      >
        <PlayingCard :card="card" stacked :mini="mini || reveal" :rank-only="reveal" />
      </div>
    </div>
  </div>
</template>
