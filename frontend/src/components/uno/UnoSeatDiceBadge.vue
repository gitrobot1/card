<script setup lang="ts">
defineProps<{
  label: string
}>()
</script>

<template>
  <div class="uno__seat-dice-badge">
    <span class="uno__seat-dice-badge__label">{{ label }}</span>
  </div>
</template>
