<script setup lang="ts">
defineProps<{
  seconds: number
  active?: boolean
}>()
</script>

<template>
  <div class="turn-timer" :class="{ 'turn-timer--active': active, 'turn-timer--warn': active && seconds <= 10 }">
    <svg viewBox="0 0 36 36" class="turn-timer__ring">
      <circle cx="18" cy="18" r="15.5" class="turn-timer__track" />
      <circle
        cx="18"
        cy="18"
        r="15.5"
        class="turn-timer__progress"
        :style="{ strokeDashoffset: `${97.4 * (1 - seconds / 35)}` }"
      />
    </svg>
    <span class="turn-timer__text">{{ seconds }}s</span>
  </div>
</template>
