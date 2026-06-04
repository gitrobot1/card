<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { suitColor, suitSymbol } from '../../constants/games'
import { equipMetaForKind, weaponMetaForKind } from '../../constants/yzsWeapons'
import type { YzsCard } from '../../types/yuzhousha'

const props = defineProps<{
  card: YzsCard
  selected?: boolean
  disabled?: boolean
  stacked?: boolean
  /** 武圣等：此牌当前可当【杀】 */
  playsAsSha?: boolean
  /** 龙胆等：此牌当前可当【闪】 */
  playsAsShan?: boolean
}>()

const emit = defineEmits<{ click: [] }>()

const showEffectTip = ref(false)

const weaponMeta = computed(() => weaponMetaForKind(props.card.kind))
const equipMeta = computed(() => equipMetaForKind(props.card.kind))
const displayName = computed(
  () => weaponMeta.value?.name ?? equipMeta.value?.name ?? props.card.name,
)
const isEquipCard = computed(() => !!weaponMeta.value || !!equipMeta.value)
const effectText = computed(() => weaponMeta.value?.effect ?? equipMeta.value?.effect ?? '')
const tipTitle = computed(() => {
  if (weaponMeta.value) {
    return `${displayName.value} · 距离 ${weaponMeta.value.range}\n${effectText.value}`
  }
  if (effectText.value) return `${displayName.value}\n${effectText.value}`
  return displayName.value
})

watch(
  () => props.card.id,
  () => {
    showEffectTip.value = false
  },
)

function toggleEffectTip(event: MouseEvent) {
  event.stopPropagation()
  showEffectTip.value = !showEffectTip.value
}

function onCardClick() {
  if (props.stacked || props.disabled) return
  showEffectTip.value = false
  emit('click')
}
</script>

<template>
  <div
    class="yzs-card"
    :class="[
      `yzs-card--${card.kind}`,
      card.suit ? `yzs-card--${suitColor(card.suit)}` : '',
      {
        'yzs-card--selected': selected,
        'yzs-card--disabled': disabled,
        'yzs-card--stacked': stacked,
        'yzs-card--equip': isEquipCard,
        'yzs-card--as-sha': playsAsSha,
        'yzs-card--as-shan': playsAsShan,
        'yzs-card--clickable': !stacked && !disabled,
      },
    ]"
    :data-card-id="card.id"
    :tabindex="stacked || disabled ? undefined : 0"
    role="button"
    @click="onCardClick"
    @keydown.enter.prevent="onCardClick"
  >
    <button
      v-if="effectText && !stacked"
      type="button"
      class="yzs-card__info"
      :title="tipTitle"
      aria-label="查看牌效果"
      @click="toggleEffectTip"
    >
      ⓘ
    </button>
    <div v-if="showEffectTip" class="yzs-card__popover" role="tooltip" @click.stop>
      <p v-if="weaponMeta" class="yzs-card__popover-range">攻击距离 {{ weaponMeta.range }}</p>
      <p class="yzs-card__popover-text">{{ effectText }}</p>
    </div>
    <span class="yzs-card__corner">
      <span class="yzs-card__label">{{ card.label ?? '?' }}</span>
      <span v-if="card.suit" class="yzs-card__pip">{{ suitSymbol(card.suit) }}</span>
    </span>
    <span v-if="playsAsSha" class="yzs-card__as-sha">杀</span>
    <span v-if="playsAsShan" class="yzs-card__as-shan">闪</span>
    <span class="yzs-card__kind">{{ displayName }}</span>
  </div>
</template>
