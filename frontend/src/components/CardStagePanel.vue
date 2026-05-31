<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { CardStage } from '../game/CardStage'
import type { AppConfig } from '../config/loadConfig'

defineProps<{
  appName: string
  config: AppConfig
}>()

const stageHost = ref<HTMLElement | null>(null)
let stage: CardStage | null = null

onMounted(async () => {
  if (!stageHost.value) {
    return
  }
  stage = new CardStage()
  await stage.mount(stageHost.value)
})

onBeforeUnmount(() => {
  stage?.destroy()
  stage = null
})
</script>

<template>
  <section class="stage-panel">
    <header class="stage-panel__header">
      <div>
        <p class="stage-panel__eyebrow">PixiJS + GSAP</p>
        <h2>{{ appName }}</h2>
      </div>
      <code>{{ config.apiBaseUrl }}</code>
    </header>
    <div ref="stageHost" class="stage-panel__canvas" />
  </section>
</template>
