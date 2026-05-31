<script setup lang="ts">
import type { UnoCard } from '../../types/uno'
import { unoColorClass } from '../../types/uno'

const props = defineProps<{
  card: UnoCard
  selected?: boolean
  mini?: boolean
  inHand?: boolean
  discardSpot?: boolean
  faceDown?: boolean
}>()
</script>

<template>
  <div
    class="uno-card"
    :class="[
      faceDown ? 'uno-card--back' : unoColorClass(card.color),
      {
        'uno-card--selected': selected,
        'uno-card--mini': mini,
        'uno-card--in-hand': inHand,
        'uno-card--discard-spot': discardSpot,
      },
    ]"
  >
    <template v-if="faceDown">
      <span class="uno-card__logo">UNO</span>
    </template>
    <template v-else>
      <span class="uno-card__center">{{ card.label }}</span>
    </template>
  </div>
</template>
