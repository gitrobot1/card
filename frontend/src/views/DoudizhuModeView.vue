<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goSolo() {
  router.push('/games/doudizhu/solo')
}

function goOnline() {
  router.push('/games/doudizhu/online')
}
</script>

<template>
  <main class="app">
    <section class="hero">
      <div class="hero__top">
        <div>
          <p class="hero__tag">斗地主</p>
          <h1>选择模式</h1>
          <p class="hero__desc">单机练习或三人联机对战</p>
        </div>
        <button type="button" class="hero__logout" @click="router.push('/')">← 返回大厅</button>
      </div>
    </section>

    <section class="game-grid game-grid--modes">
      <button type="button" class="game-card" @click="goSolo">
        <span class="game-card__tag">单机</span>
        <h2>对战电脑</h2>
        <p>你 + 电脑甲 + 电脑乙，立即开局</p>
      </button>

      <button type="button" class="game-card" @click="goOnline">
        <span class="game-card__tag game-card__tag--online">联机</span>
        <h2>多人联机</h2>
        <p>3 真人同桌，人满即开</p>
      </button>
    </section>
  </main>
</template>
