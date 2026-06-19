<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { fetchYuzhoushaModes } from '../../api/games'
import { showToast } from '../../composables/useToast'
import type { YzsModeMeta } from '../../types/yuzhousha'

type TabKey = 'bot' | 'online' | 'codex'

const router = useRouter()
const modes = ref<YzsModeMeta[]>([])
const loading = ref(true)
const activeTab = ref<TabKey>('bot')

const tabs: { key: TabKey; label: string; icon: string }[] = [
  { key: 'bot', label: '人机', icon: '🤖' },
  { key: 'online', label: '联机', icon: '🎮' },
  { key: 'codex', label: '图鉴', icon: '📖' },
]

/** 人机模式：human_seats 只有玩家一人 */
const botModes = computed(() =>
  modes.value.filter(m => m.human_seats?.length === 1)
)

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
          <h1>宇宙杀</h1>
          <p class="hero__desc">选择游戏模式开始对局</p>
        </div>
        <button type="button" class="hero__logout" @click="router.push('/')">← 返回大厅</button>
      </div>
    </section>

    <!-- Tab 栏 -->
    <div class="yzs-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="yzs-tabs__btn"
        :class="{ 'yzs-tabs__btn--active': activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        <span class="yzs-tabs__icon">{{ tab.icon }}</span>
        {{ tab.label }}
      </button>
    </div>

    <p v-if="loading" class="ddz__loading">加载模式中…</p>

    <section v-else class="yzs-tab-panel">
      <!-- 人机 -->
      <div v-if="activeTab === 'bot'" class="game-grid game-grid--modes">
        <div v-for="mode in botModes" :key="mode.id" class="game-card game-card--setup">
          <span class="game-card__tag">{{ mode.tag || mode.id }}</span>
          <h2>{{ mode.name }}</h2>
          <p v-if="mode.hint" class="zjh-mode__hint">{{ mode.hint }}</p>
          <ul v-if="mode.rules?.length" class="yzs-mode__rules">
            <li v-for="(rule, i) in mode.rules" :key="i">{{ rule }}</li>
          </ul>
          <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goPick(mode)">
            开始对局
          </button>
        </div>
        <p v-if="botModes.length === 0" class="yzs-tab__empty">暂无可用人机模式</p>
      </div>

      <!-- 联机 -->
      <div v-if="activeTab === 'online'" class="game-grid game-grid--modes">
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
        <div class="game-card game-card--setup game-card--online">
          <span class="game-card__tag">联机 3v3</span>
          <h2>3v3 竞技 6 人</h2>
          <p class="zjh-mode__hint">暖色 vs 冷色 · 击败敌方主帅获胜 · 按加入顺序分配座位</p>
          <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goOnline('3v3')">
            进入 3v3 房间
          </button>
        </div>
        <div class="game-card game-card--setup game-card--online">
          <span class="game-card__tag">联机身份</span>
          <h2>5 人身份局</h2>
          <p class="zjh-mode__hint">主公+忠臣+内奸+2 反贼 · 开局随机身份 · 主公公开 · 需满 5 人</p>
          <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goOnline('identity_5')">
            进入 5 人房间
          </button>
        </div>
        <div class="game-card game-card--setup game-card--online">
          <span class="game-card__tag">联机身份</span>
          <h2>8 人身份局</h2>
          <p class="zjh-mode__hint">标准八人身份 · 开局随机身份 · 主公公开 · 需满 8 人</p>
          <button type="button" class="ddz__btn ddz__btn--primary zjh-mode__start" @click="goOnline('identity_8')">
            进入 8 人房间
          </button>
        </div>
      </div>

      <!-- 图鉴 -->
      <div v-if="activeTab === 'codex'" class="yzs-tab-panel__codex">
        <div class="game-card game-card--setup game-card--codex">
          <span class="game-card__tag">图鉴</span>
          <h2>武将 & 牌库</h2>
          <p class="zjh-mode__hint">查看全部武将技能、基础牌、武器与锦囊效果说明</p>
          <button type="button" class="ddz__btn ddz__btn--hint zjh-mode__start" @click="router.push('/games/yuzhousha/codex')">
            打开图鉴
          </button>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.yzs-tabs {
  display: flex;
  gap: 12px;
  padding: 0 24px;
  margin: 24px 0 20px;
  justify-content: center;
}

.yzs-tabs__btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-width: 120px;
  padding: 10px 20px;
  border: 2px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-surface);
  color: var(--color-text-secondary);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.yzs-tabs__icon {
  font-size: 18px;
}

.yzs-tabs__btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.yzs-tabs__btn--active {
  border-color: #c0392b;
  background: #c0392b;
  color: #fff;
  box-shadow: 0 2px 8px rgba(192, 57, 43, 0.35);
}

.yzs-tab-panel {
  padding: 0 24px;
}

.yzs-tab-panel .game-grid--modes {
  max-width: 900px;
  margin: 0 auto;
}

/* 图鉴卡片与网格卡片保持相同容器宽度 */
.yzs-tab-panel__codex {
  max-width: 900px;
  margin: 0 auto;
  padding: 0;
}

.yzs-tab-panel__codex .game-card {
  width: 100%;
}

.yzs-tab__empty {
  color: var(--color-text-secondary);
  padding: 40px 0;
  text-align: center;
  grid-column: 1 / -1;
}

.yzs-mode__rules {
  margin: 12px 0 16px;
  padding-left: 18px;
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.6;
}
</style>
