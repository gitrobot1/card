<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchGameCatalog } from '../api/games'
import { GAME_ROUTES } from '../constants/games'
import type { AppConfig } from '../config/loadConfig'
import type { GameMeta } from '../types/doudizhu'
import type { User } from '../types/auth'

defineProps<{
  appName: string
  config: AppConfig
  user: User
}>()

const emit = defineEmits<{ logout: [] }>()
const router = useRouter()
const games = ref<GameMeta[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const result = await fetchGameCatalog()
    games.value = result.games
  } catch {
    games.value = [
      { type: 'doudizhu', name: '斗地主', description: '三人扑克，抢地主对战', enabled: true },
      { type: 'zhajinhua', name: '扎金花', description: '比牌博弈，敬请期待', enabled: false },
      { type: 'douniu', name: '斗牛', description: '凑十比点数，敬请期待', enabled: false },
      { type: 'sanguosha', name: '三国杀', description: '身份策略卡牌，敬请期待', enabled: false },
      { type: 'uno', name: 'UNO', description: '经典变色牌，敬请期待', enabled: false },
    ]
  } finally {
    loading.value = false
  }
})

function enterGame(game: GameMeta) {
  const route = GAME_ROUTES[game.type]
  if (route && game.enabled) {
    router.push(route)
  }
}
</script>

<template>
  <main class="app">
    <section class="hero">
      <div class="hero__top">
        <div>
          <p class="hero__tag">游戏大厅</p>
          <h1>{{ appName }}</h1>
          <p class="hero__desc">
            欢迎，<strong>{{ user.nickname }}</strong>
            <span class="hero__id">#{{ user.id }}</span>
          </p>
        </div>
        <button class="hero__logout" type="button" @click="emit('logout')">切换账号</button>
      </div>
    </section>

    <section class="game-grid">
      <p v-if="loading" class="game-grid__loading">加载游戏列表...</p>
      <button
        v-for="game in games"
        :key="game.type"
        type="button"
        class="game-card"
        :class="{ 'game-card--disabled': !game.enabled }"
        :disabled="!game.enabled"
        @click="enterGame(game)"
      >
        <span class="game-card__tag">{{ game.enabled ? '可玩' : '开发中' }}</span>
        <h2>{{ game.name }}</h2>
        <p>{{ game.description }}</p>
      </button>
    </section>
  </main>
</template>
