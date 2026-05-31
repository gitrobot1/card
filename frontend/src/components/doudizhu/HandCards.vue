<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import PlayingCard from './PlayingCard.vue'
import type { Card } from '../../types/doudizhu'

const props = defineProps<{
  cards: Card[]
  selectedIds: string[]
  hintIds: string[]
  interactive?: boolean
  dealing?: boolean
}>()

const emit = defineEmits<{
  'update:selectedIds': [string[]]
}>()

const LONG_PRESS_MS = 180

const rowRef = ref<HTMLElement | null>(null)
const dragStartIndex = ref<number | null>(null)
const pendingIndex = ref<number | null>(null)
const longPressTimer = ref<number | null>(null)
const isPianoMode = ref(false)
const isDragging = ref(false)
const dragMode = ref<'select' | 'deselect' | null>(null)

/** 大牌在左：按点数从大到小 */
const sortedCards = computed(() =>
  [...props.cards].sort((a, b) => {
    if (b.rank !== a.rank) return b.rank - a.rank
    return String(a.suit).localeCompare(String(b.suit))
  }),
)

/** 适度叠放：槽位 step 宽，牌面溢出；不再用负 margin 避免叠太紧 */
const slotStyle = computed(() => {
  const count = sortedCards.value.length
  const cardWidth = 82
  if (count <= 1) {
    return { '--hand-step': `${cardWidth}px`, '--hand-card-width': `${cardWidth}px` }
  }
  const minStep = 34
  const preferStep = 42
  const maxRowWidth = Math.min(window.innerWidth - 48, 980)
  const fitStep = (maxRowWidth - cardWidth) / (count - 1)
  const step = Math.max(minStep, Math.min(preferStep, fitStep))
  return { '--hand-step': `${step}px`, '--hand-card-width': `${cardWidth}px` }
})

const SELECT_SPREAD = 12
const SELECT_LIFT = 14

/** 选中牌从两侧缝隙抬起：左右邻牌让开，不提升 z-index 盖住别家牌 */
const slotLayouts = computed(() => {
  const count = sortedCards.value.length
  const dx = Array(count).fill(0)
  const dy = Array(count).fill(0)
  const selectedSet = new Set(props.selectedIds)

  const indices = sortedCards.value
    .map((card, index) => (selectedSet.has(card.id) ? index : -1))
    .filter((index) => index >= 0)

  if (indices.length === 0) {
    return sortedCards.value.map((_, index) => ({ zIndex: index + 1 }))
  }

  const groups: number[][] = []
  for (const index of indices) {
    const last = groups[groups.length - 1]
    if (last && index === last[last.length - 1] + 1) {
      last.push(index)
    } else {
      groups.push([index])
    }
  }

  for (const group of groups) {
    const left = group[0]
    const right = group[group.length - 1]
    for (let j = 0; j < left; j++) dx[j] -= SELECT_SPREAD
    for (let j = right + 1; j < count; j++) dx[j] += SELECT_SPREAD
    for (const index of group) dy[index] = -SELECT_LIFT
  }

  return sortedCards.value.map((_, index) => ({
    zIndex: index + 1,
    transform: dx[index] || dy[index] ? `translate(${dx[index]}px, ${dy[index]}px)` : undefined,
  }))
})

function clearLongPressTimer() {
  if (longPressTimer.value !== null) {
    window.clearTimeout(longPressTimer.value)
    longPressTimer.value = null
  }
}

function toggleCard(id: string) {
  if (props.selectedIds.includes(id)) {
    emit(
      'update:selectedIds',
      props.selectedIds.filter((item) => item !== id),
    )
  } else {
    emit('update:selectedIds', [...props.selectedIds, id])
  }
}

function applyRange(endIndex: number) {
  if (dragStartIndex.value === null || dragMode.value === null) return
  const start = dragStartIndex.value
  const min = Math.min(start, endIndex)
  const max = Math.max(start, endIndex)
  const rangeIds = sortedCards.value.slice(min, max + 1).map((c) => c.id)

  if (dragMode.value === 'select') {
    emit('update:selectedIds', [...new Set([...props.selectedIds, ...rangeIds])])
  } else {
    const remove = new Set(rangeIds)
    emit('update:selectedIds', props.selectedIds.filter((id) => !remove.has(id)))
  }
}

/** 从右到左命中整张牌区域 */
function cardIndexFromEvent(event: MouseEvent) {
  const row = rowRef.value?.querySelector<HTMLElement>('.hand-cards__row')
  if (!row) return -1
  const slots = row.querySelectorAll<HTMLElement>('.hand-cards__slot')
  for (let i = slots.length - 1; i >= 0; i--) {
    const cardEl = slots[i].querySelector<HTMLElement>('.playing-card')
    const rect = (cardEl ?? slots[i]).getBoundingClientRect()
    if (
      event.clientX >= rect.left &&
      event.clientX <= rect.right &&
      event.clientY >= rect.top &&
      event.clientY <= rect.bottom
    ) {
      return i
    }
  }
  return -1
}

function beginPianoMode(index: number) {
  isPianoMode.value = true
  isDragging.value = true
  dragStartIndex.value = index
  const id = sortedCards.value[index].id
  dragMode.value = props.selectedIds.includes(id) ? 'deselect' : 'select'
  applyRange(index)
}

function onRowMouseDown(event: MouseEvent) {
  if (!props.interactive || props.dealing) return
  const index = cardIndexFromEvent(event)
  if (index < 0) return
  event.preventDefault()
  pendingIndex.value = index
  clearLongPressTimer()
  longPressTimer.value = window.setTimeout(() => {
    longPressTimer.value = null
    beginPianoMode(index)
  }, LONG_PRESS_MS)
}

function onWindowMouseMove(event: MouseEvent) {
  if (!props.interactive || props.dealing) return

  // 按住并滑到其它牌：立刻进入钢琴连选，不必等长按计时
  if (longPressTimer.value !== null && pendingIndex.value !== null) {
    const index = cardIndexFromEvent(event)
    if (index >= 0 && index !== pendingIndex.value) {
      clearLongPressTimer()
      beginPianoMode(pendingIndex.value)
      applyRange(index)
    }
    return
  }

  if (!isPianoMode.value) return
  const index = cardIndexFromEvent(event)
  if (index >= 0) applyRange(index)
}

function onMouseUp() {
  clearLongPressTimer()
  if (!isPianoMode.value && pendingIndex.value !== null && props.interactive) {
    toggleCard(sortedCards.value[pendingIndex.value].id)
  }
  isPianoMode.value = false
  isDragging.value = false
  dragStartIndex.value = null
  dragMode.value = null
  pendingIndex.value = null
}

onMounted(() => {
  window.addEventListener('mouseup', onMouseUp)
  window.addEventListener('mousemove', onWindowMouseMove)
})

onUnmounted(() => {
  window.removeEventListener('mouseup', onMouseUp)
  window.removeEventListener('mousemove', onWindowMouseMove)
  clearLongPressTimer()
})

defineExpose({ rowRef, sortedCards })
</script>

<template>
  <div
    ref="rowRef"
    class="hand-cards"
    :class="{
      'hand-cards--dragging': isDragging,
      'hand-cards--piano-select': isPianoMode && dragMode === 'select',
      'hand-cards--piano-deselect': isPianoMode && dragMode === 'deselect',
      'hand-cards--dealing': dealing,
    }"
  >
    <div class="hand-cards__row" :style="slotStyle" @mousedown="onRowMouseDown">
      <div
        v-for="(card, index) in sortedCards"
        :key="card.id"
        class="hand-cards__slot"
        :class="{ 'hand-cards__slot--selected': selectedIds.includes(card.id) }"
        :style="slotLayouts[index]"
        :data-card-id="card.id"
      >
        <PlayingCard
          :card="card"
          stacked
          :selected="selectedIds.includes(card.id)"
          :hint="hintIds.includes(card.id)"
          :interactive="false"
          :dealing="dealing"
        />
      </div>
    </div>
  </div>
</template>
