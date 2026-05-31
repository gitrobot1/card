<script setup lang="ts">
import { suitColor, suitSymbol } from '../../constants/games'
import type { Card } from '../../types/doudizhu'

withDefaults(
  defineProps<{
    card: Card
    selected?: boolean
    mini?: boolean
    stacked?: boolean
    interactive?: boolean
    hint?: boolean
    dealing?: boolean
    /** 仅展示点数，不展示花色（结算亮牌） */
    rankOnly?: boolean
  }>(),
  {
    selected: false,
    mini: false,
    stacked: false,
    interactive: false,
    hint: false,
    dealing: false,
    rankOnly: false,
  },
)

const emit = defineEmits<{ click: [] }>()
</script>

<template>
  <component
    :is="stacked ? 'div' : 'button'"
    :type="stacked ? undefined : 'button'"
    class="playing-card"
    :class="[
      `playing-card--${suitColor(card.suit)}`,
      {
        'playing-card--selected': selected,
        'playing-card--mini': mini,
        'playing-card--stacked': stacked,
        'playing-card--interactive': interactive && !mini && !dealing,
        'playing-card--hint': hint,
        'playing-card--dealing': dealing,
        'playing-card--rank-only': rankOnly,
      },
    ]"
    :disabled="stacked ? undefined : dealing"
    @click="!stacked && emit('click')"
  >
    <span class="playing-card__rank">{{ card.label }}</span>
    <span v-if="card.suit !== 'J' && !rankOnly" class="playing-card__suit">{{ suitSymbol(card.suit) }}</span>
  </component>
</template>
