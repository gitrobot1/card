<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import YzsActionBar from '../../components/yuzhousha/YzsActionBar.vue'
import YzsHandZone from '../../components/yuzhousha/YzsHandZone.vue'
import { resolveYzsLayout, yzsLayoutSubtitle } from '../../components/yuzhousha/layouts/index'
import { provideYzsGame, useYzsGame } from '../../composables/yuzhousha/useYzsGame'

const router = useRouter()
const game = useYzsGame()
provideYzsGame(game)

const {
  state,
  loading,
  isDealing,
  isAnimating,
  isFinished,
  restart,
} = game

const layoutComponent = computed(() => resolveYzsLayout(state.value?.layout_key))
const tableSubtitle = computed(() =>
  yzsLayoutSubtitle(state.value?.layout_key, state.value?.mode),
)
</script>

<template>
  <main class="ddz app">
    <header class="ddz__header">
      <button type="button" class="ddz__back" @click="router.push('/games/yuzhousha')">← 返回</button>
      <div>
        <h1>宇宙杀</h1>
        <p class="ddz__subtitle">{{ tableSubtitle }}</p>
      </div>
      <button
        v-if="isFinished"
        type="button"
        class="ddz__restart"
        :disabled="loading || isDealing || isAnimating"
        @click="restart"
      >
        再来一局
      </button>
      <div v-else class="zjh__header-spacer" aria-hidden="true" />
    </header>

    <p v-if="loading && !state" class="ddz__loading">正在开局…</p>

    <section
      v-if="state"
      :ref="(el) => { game.tableWrapRef.value = el as HTMLElement | null }"
      class="ddz__table yzs__table-wrap"
    >
      <div v-if="!isDealing" class="yzs__deck-count">剩余 {{ state.draw_count }}</div>
      <div :ref="(el) => { game.drawAreaRef.value = el as HTMLElement | null }" class="yzs__draw-anchor" aria-hidden="true" />

      <YzsActionBar />
      <component :is="layoutComponent" />
      <YzsHandZone />
    </section>
  </main>
</template>
