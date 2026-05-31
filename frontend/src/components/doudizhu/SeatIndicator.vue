<script setup lang="ts">
import TurnTimer from './TurnTimer.vue'

defineProps<{
  seconds: number
  showTimer: boolean
  showPass: boolean
  /** 相对头像框：left=左侧, right=右侧, top=上方 */
  placement?: 'left' | 'right' | 'top'
}>()
</script>

<template>
  <div
    v-if="showTimer || showPass"
    class="ddz__seat-indicator"
    :class="{
      'ddz__seat-indicator--left': placement === 'left',
      'ddz__seat-indicator--right': placement === 'right',
      'ddz__seat-indicator--top': !placement || placement === 'top',
    }"
  >
    <TurnTimer v-if="showTimer" :seconds="seconds" active />
    <span v-else-if="showPass" class="ddz__pass-tag">不要</span>
  </div>
</template>
