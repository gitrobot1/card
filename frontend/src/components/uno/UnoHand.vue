<script setup lang="ts">
import { computed, ref } from 'vue'
import UnoCard from './UnoCard.vue'
import type { UnoCard as UnoCardType } from '../../types/uno'

const props = defineProps<{
  cards: UnoCardType[]
  selectedId: string | null
  /** 是否可点选出牌 */
  interactive?: boolean
  /** 非己方回合也允许 hover 抬牌 */
  hoverable?: boolean
}>()

const emit = defineEmits<{
  select: [string | null]
}>()

const hoverIndex = ref<number | null>(null)

const canHover = computed(() => props.hoverable === true)

const sortedCards = computed(() =>
  [...props.cards].sort((a, b) => {
    const order = (c: UnoCardType) => {
      if (c.color === 'wild') return 100
      const colors = ['red', 'yellow', 'green', 'blue']
      return colors.indexOf(c.color) * 20 + (Number.parseInt(c.value, 10) || 50)
    }
    return order(a) - order(b)
  }),
)

const slotStyle = computed(() => {
  const count = sortedCards.value.length
  const cardWidth = 82
  if (count <= 1) {
    return { '--hand-step': `${cardWidth}px`, '--hand-card-width': `${cardWidth}px` }
  }
  const minStep = 40
  const preferStep = 50
  const maxRowWidth = Math.min(window.innerWidth - 48, 980)
  const fitStep = (maxRowWidth - cardWidth) / (count - 1)
  const step = Math.max(minStep, Math.min(preferStep, fitStep))
  return { '--hand-step': `${step}px`, '--hand-card-width': `${cardWidth}px` }
})

const SPREAD = 14
const LIFT = 16

const spreadIndex = computed(() => {
  if (props.selectedId) {
    const idx = sortedCards.value.findIndex((c) => c.id === props.selectedId)
    if (idx >= 0) return idx
  }
  return hoverIndex.value
})

const slotLayouts = computed(() => {
  const count = sortedCards.value.length
  const dx = Array(count).fill(0)
  const dy = Array(count).fill(0)
  const index = spreadIndex.value

  if (index === null || index < 0) {
    return sortedCards.value.map((_, i) => ({ zIndex: i + 1 }))
  }

  for (let j = 0; j < index; j++) dx[j] -= SPREAD
  for (let j = index + 1; j < count; j++) dx[j] += SPREAD
  dy[index] = -LIFT

  return sortedCards.value.map((_, i) => ({
    zIndex: i === index ? count + 2 : i + 1,
    transform: dx[i] || dy[i] ? `translate(${dx[i]}px, ${dy[i]}px)` : undefined,
  }))
})

function onClick(card: UnoCardType) {
  if (!props.interactive) return
  emit('select', props.selectedId === card.id ? null : card.id)
}

function onEnter(index: number) {
  if (!canHover.value) return
  hoverIndex.value = index
}

function onLeave() {
  hoverIndex.value = null
}
</script>

<template>
  <div
    class="hand-cards"
    :class="{
      'hand-cards--view-only': !interactive,
      'hand-cards--hoverable': canHover,
    }"
  >
    <div class="hand-cards__row" :style="slotStyle">
      <button
        v-for="(card, index) in sortedCards"
        :key="card.id"
        type="button"
        class="hand-cards__slot"
        :class="{ 'hand-cards__slot--selected': selectedId === card.id }"
        :style="slotLayouts[index]"
        :data-card-id="card.id"
        @click="onClick(card)"
        @mouseenter="onEnter(index)"
        @mouseleave="onLeave"
      >
        <UnoCard
          :card="card"
          in-hand
          :selected="selectedId === card.id"
        />
      </button>
    </div>
  </div>
</template>
