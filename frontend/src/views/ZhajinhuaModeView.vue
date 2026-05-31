<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const botCount = ref(2)

function goSolo() {
  router.push({ path: '/games/zhajinhua/solo', query: { bots: String(botCount.value) } })
}

function goOnline() {
  router.push('/games/zhajinhua/online')
}
</script>

<template>
  <main class="app">
    <section class="hero">
      <div class="hero__top">
        <div>
          <p class="hero__tag">扎金花</p>
          <h1>选择模式</h1>
          <p class="hero__desc">2-8 人比牌，牌型越高结算倍率越大</p>
        </div>
        <button type="button" class="hero__logout" @click="router.push('/')">← 返回大厅</button>
      </div>
    </section>

    <section class="game-grid game-grid--modes">
      <div class="game-card game-card--setup">
        <span class="game-card__tag">单机</span>
        <h2>对战电脑</h2>
        <p class="zjh-mode__hint">你 + 电脑，共 2-8 人一桌</p>
        <label class="zjh-mode__bots">
          <span>电脑数量</span>
          <input v-model.number="botCount" type="range" min="1" max="7" step="1" />
          <strong>{{ botCount }} 个电脑（共 {{ botCount + 1 }} 人）</strong>
        </label>
        <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goSolo">
          开始单机
        </button>
      </div>

      <button type="button" class="game-card" @click="goOnline">
        <span class="game-card__tag game-card__tag--online">联机</span>
        <h2>多人联机</h2>
        <p>2-8 真人同桌，全员准备后由房主开始</p>
      </button>
    </section>

    <section class="zjh-multiplier-legend">
      <h3>牌型倍率</h3>
      <ul>
        <li><span class="zjh-mul zjh-mul--max">×12</span> 235（专克豹子）</li>
        <li><span class="zjh-mul">×10</span> 豹子</li>
        <li><span class="zjh-mul">×6</span> 顺金</li>
        <li><span class="zjh-mul">×4</span> 金花</li>
        <li><span class="zjh-mul">×3</span> 顺子</li>
        <li><span class="zjh-mul">×2</span> 对子</li>
        <li><span class="zjh-mul zjh-mul--min">×1</span> 单牌</li>
      </ul>
    </section>
  </main>
</template>
