<script setup lang="ts">
import TurnTimer from './TurnTimer.vue'
import UnoSeatPlayBadge from '../uno/UnoSeatPlayBadge.vue'
import UnoSeatDiceBadge from '../uno/UnoSeatDiceBadge.vue'
import type { UnoColor } from '../../types/uno'

defineProps<{
  seconds: number
  showTimer: boolean
  showPass?: boolean
  actionLabel?: string
  playBadge?: { color: UnoColor; label: string; uno?: boolean } | null
  diceBadge?: string | null
  /** 相对头像框：left=左侧, right=右侧, top=上方 */
  placement?: 'left' | 'right' | 'top'
}>()
</script>

<template>
  <div
    v-if="diceBadge || playBadge || showTimer || showPass || actionLabel"
    class="ddz__seat-indicator"
    :class="{
      'ddz__seat-indicator--left': placement === 'left',
      'ddz__seat-indicator--right': placement === 'right',
      'ddz__seat-indicator--top': placement === 'top',
      'ddz__seat-indicator--bottom': placement === 'bottom',
    }"
  >
    <UnoSeatDiceBadge v-if="diceBadge" :label="diceBadge" />
    <UnoSeatPlayBadge
      v-else-if="playBadge"
      :color="playBadge.color"
      :label="playBadge.label"
      :uno="playBadge.uno"
    />
    <TurnTimer v-else-if="showTimer" :seconds="seconds" active />
    <span v-else-if="actionLabel" class="ddz__pass-tag zjh__seat-action">{{ actionLabel }}</span>
    <span v-else-if="showPass" class="ddz__pass-tag">不要</span>
  </div>
</template>
