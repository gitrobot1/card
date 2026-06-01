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
          <h1>斗地主 - 选择模式</h1>
          <p class="hero__desc">单机练习或三人联机对战</p>
        </div>
        <button type="button" class="hero__logout" @click="router.push('/')">← 返回大厅</button>
      </div>
    </section>

    <section class="game-grid game-grid--modes">
      <div class="game-card game-card--setup">
        <span class="game-card__tag">单机</span>
        <h2>对战电脑</h2>
        <p class="zjh-mode__hint">你 + 电脑甲 + 电脑乙，固定 3 人一桌</p>
        <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goSolo">
          开始单机
        </button>
      </div>

      <div class="game-card game-card--setup">
        <span class="game-card__tag game-card__tag--online">联机</span>
        <h2>多人联机</h2>
        <p class="zjh-mode__hint">3 真人同桌，人满即开</p>
        <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goOnline">
          进入联机房间
        </button>
      </div>
    </section>
  </main>
</template>
