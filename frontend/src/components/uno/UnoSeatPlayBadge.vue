<script setup lang="ts">
import type { UnoColor } from '../../types/uno'

defineProps<{
  color: UnoColor
  label: string
  uno?: boolean
}>()
</script>

<template>
  <div
    class="uno__seat-play-badge"
    :class="`uno-card--${color}`"
  >
    <span class="uno__seat-play-badge__label">{{ label }}</span>
    <span v-if="uno" class="uno__seat-play-badge__uno">UNO</span>
  </div>
</template>
