<script setup lang="ts">
import { computed } from 'vue'
import YzsCardView from './YzsCardView.vue'
import type { YzsCard } from '../../types/yuzhousha'

const props = withDefaults(
  defineProps<{
    cards: YzsCard[]
    maxWidth?: number
  }>(),
  {
    maxWidth: undefined,
  },
)

const rowStyle = computed(() => {
  const count = props.cards.length
  const cardWidth = 64
  if (count <= 1) {
    return { '--stack-step': `${cardWidth}px`, '--stack-card-width': `${cardWidth}px` }
  }
  const minStep = 36
  const preferStep = 44
  const maxRowWidth = props.maxWidth ?? Math.min(window.innerWidth - 120, 520)
  const fitStep = (maxRowWidth - cardWidth) / (count - 1)
  const step = Math.max(minStep, Math.min(preferStep, fitStep))
  return { '--stack-step': `${step}px`, '--stack-card-width': `${cardWidth}px` }
})
</script>

<template>
  <div class="stacked-cards yzs-stacked-cards">
    <div class="stacked-cards__row yzs-stacked-cards__row" :style="rowStyle">
      <div
        v-for="(card, index) in cards"
        :key="card.id"
        class="stacked-cards__slot"
        :style="{ zIndex: index + 1 }"
      >
        <YzsCardView :card="card" stacked disabled />
      </div>
    </div>
  </div>
</template>
