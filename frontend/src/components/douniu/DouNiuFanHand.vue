<script setup lang="ts">
import { computed } from 'vue'
import PlayingCard from '../doudizhu/PlayingCard.vue'
import { cardsFromLayout, isSpecialHandType } from '../../utils/douniuHand'
import type { Card } from '../../types/doudizhu'
import type { DouNiuHandLayout } from '../../types/douniu'

const props = withDefaults(
  defineProps<{
    cards: Card[]
    layout?: DouNiuHandLayout | null
    handLabel?: string
    handMultiplier?: number
    handType?: string
    /** 高亮右侧「牛」牌（自己看牌时） */
    highlightNiu?: boolean
    /** 结算亮牌：更大徽章 */
    reveal?: boolean
  }>(),
  {
    layout: null,
    handLabel: '',
    handMultiplier: 1,
    handType: '',
    highlightNiu: false,
    reveal: false,
  },
)

const groups = computed(() => cardsFromLayout(props.cards, props.layout))
const hasSplit = computed(() => groups.value.niu.length > 0)
const showNiuHighlight = computed(() => {
  if (!hasSplit.value || isSpecialHandType(props.handType)) return false
  return props.highlightNiu || props.reveal
})
const badgeClass = computed(() => {
  if (isSpecialHandType(props.handType)) return 'dn__hand-type-badge dn__hand-type-badge--special'
  if (props.handType === 'niu_niu') return 'dn__hand-type-badge dn__hand-type-badge--niu'
  if (props.handType === 'none') return 'dn__hand-type-badge dn__hand-type-badge--none'
  return 'dn__hand-type-badge'
})
</script>

<template>
  <div
    class="dn__fan-wrap"
    :class="{ 'dn__fan-wrap--reveal': reveal, 'dn__fan-wrap--split': hasSplit }"
  >
    <div class="dn__fan-hand" :class="{ 'dn__fan-hand--split': hasSplit }">
      <div class="dn__fan-group">
        <PlayingCard
          v-for="c in groups.head"
          :key="c.id"
          :card="c"
          stacked
          mini
        />
      </div>
      <div v-if="hasSplit" class="dn__fan-divider" aria-hidden="true" />
      <div v-if="hasSplit" class="dn__fan-group dn__fan-group--niu">
        <PlayingCard
          v-for="c in groups.niu"
          :key="c.id"
          :card="c"
          stacked
          mini
          :niu-highlight="showNiuHighlight"
        />
      </div>
    </div>
    <div :class="[badgeClass, { 'dn__hand-type-badge--empty': !handLabel }]">
      <template v-if="handLabel">
        {{ handLabel }}<span v-if="handMultiplier && handMultiplier > 1"> ×{{ handMultiplier }}</span>
      </template>
    </div>
  </div>
</template>
