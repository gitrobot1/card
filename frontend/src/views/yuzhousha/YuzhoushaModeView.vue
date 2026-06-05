<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchYuzhoushaModes } from '../../api/games'
import { showToast } from '../../composables/useToast'
import type { YzsModeMeta } from '../../types/yuzhousha'

const router = useRouter()
const modes = ref<YzsModeMeta[]>([])
const loading = ref(true)

onMounted(async () => {
  loading.value = true
  try {
    const res = await fetchYuzhoushaModes()
    modes.value = res.modes
  } catch (err) {
    showToast(err instanceof Error ? err.message : '加载模式失败')
  } finally {
    loading.value = false
  }
})

function goPick(mode: YzsModeMeta) {
  const query = mode.id === '1v1' ? undefined : { mode: mode.id }
  router.push({ path: '/games/yuzhousha/solo/pick', query })
}

function goOnline(mode: string) {
  router.push({ path: '/games/yuzhousha/online', query: { mode } })
}
</script>

<template>
  <main class="app">
    <section class="hero">
      <div class="hero__top">
        <div>
          <h1>宇宙杀 - 选择模式</h1>
          <p class="hero__desc">1v1 / 2v2 / 3v3 / 身份局 · 多种人机规则</p>
        </div>
        <button type="button" class="hero__logout" @click="router.push('/')">← 返回大厅</button>
      </div>
    </section>

    <p v-if="loading" class="ddz__loading">加载模式中…</p>

    <section v-else class="game-grid game-grid--modes">
      <div class="game-card game-card--setup game-card--codex">
        <span class="game-card__tag">图鉴</span>
        <h2>武将 & 牌库</h2>
        <p class="zjh-mode__hint">查看全部武将技能、基础牌、武器与锦囊效果说明</p>
        <button type="button" class="ddz__btn ddz__btn--hint zjh-mode__start" @click="router.push('/games/yuzhousha/codex')">
          打开图鉴
        </button>
      </div>
      <div class="game-card game-card--setup game-card--online">
        <span class="game-card__tag">联机</span>
        <h2>1v1 真人对战</h2>
        <p class="zjh-mode__hint">匹配或邀请好友 · WebSocket 实时同步局面与结算</p>
        <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goOnline('1v1')">
          进入联机房间
        </button>
      </div>
      <div class="game-card game-card--setup game-card--online">
        <span class="game-card__tag">联机 2v2</span>
        <h2>十字阵 4 人</h2>
        <p class="zjh-mode__hint">四人组队 · 你+队友 vs 两侧敌将 · 需满 4 人开局</p>
        <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goOnline('2v2')">
          进入 2v2 房间
        </button>
      </div>
      <div class="game-card game-card--setup game-card--online">
        <span class="game-card__tag">联机 3 人</span>
        <h2>杀上保下</h2>
        <p class="zjh-mode__hint">三人链式 · 杀上家保下家 · 需满 3 人开局</p>
        <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goOnline('3p_chain')">
          进入房间
        </button>
      </div>
      <div v-for="mode in modes" :key="mode.id" class="game-card game-card--setup">
        <span class="game-card__tag">{{ mode.tag || mode.id }}</span>
        <h2>{{ mode.name }}</h2>
        <p v-if="mode.hint" class="zjh-mode__hint">{{ mode.hint }}</p>
        <ul v-if="mode.rules?.length" class="yzs-mode__rules">
          <li v-for="(rule, i) in mode.rules" :key="i">{{ rule }}</li>
        </ul>
        <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goPick(mode)">
          开始 {{ mode.id }}
        </button>
      </div>
    </section>
  </main>
</template>

<style scoped>
.yzs-mode__rules {
  margin: 12px 0 16px;
  padding-left: 18px;
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.6;
}
</style>
