<script setup lang="ts">
import { useToastState } from '../composables/useToast'

const { toasts } = useToastState()
</script>

<template>
  <Teleport to="body">
    <div class="app-toast-stack" aria-live="polite">
      <TransitionGroup name="app-toast">
        <div
          v-for="item in toasts"
          :key="item.id"
          class="app-toast"
          :class="`app-toast--${item.type}`"
        >
          {{ item.message }}
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
