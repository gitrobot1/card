<script setup lang="ts">
import { computed } from 'vue'
import type { YuzhoushaState, YzsGameOverStats } from '../../types/yuzhousha'
import { getSettlementDisplay, type SettlementDisplay } from '../../composables/yuzhousha/settlementDisplay'

const props = defineProps<{
  state: YuzhoushaState | null
}>()

const emit = defineEmits<{
  restart: []
}>()

const display = computed<SettlementDisplay>(() => getSettlementDisplay(props.state))
const isWin = computed(() => {
  if (!props.state || !display.value.isFinished) return false
  // 判断人类玩家是否获胜
  const humanSeat = props.state.human_player ?? 0
  return display.value.winnerIndex === humanSeat
})

const resultClass = computed(() => ({
  'yzs-game-over--win': isWin.value,
  'yzs-game-over--lose': !isWin.value && display.value.isFinished,
}))

// 游戏结束统计（如果后端返回了的话）
const gameOverStats = computed<YzsGameOverStats | null>(() => {
  return props.state?.game_over_stats ?? null
})
</script>

<template>
  <Transition name="yzs-game-over-fade">
    <div v-if="display.isFinished" class="yzs-game-over" :class="resultClass">
      <div class="yzs-game-over__backdrop" />
      <div class="yzs-game-over__card">
        <div class="yzs-game-over__icon">
          <span v-if="isWin">🏆</span>
          <span v-else>💀</span>
        </div>

        <h2 class="yzs-game-over__title">
          <span v-if="isWin">胜利！</span>
          <span v-else>败北</span>
        </h2>

        <p class="yzs-game-over__message">{{ display.centerHint }}</p>

        <!-- 玩家表现统计（如果后端返回了统计数据） -->
        <div v-if="gameOverStats?.player_stats?.length" class="yzs-game-over__stats">
          <div class="yzs-game-over__stat">
            <span class="yzs-game-over__stat-label">模式</span>
            <span class="yzs-game-over__stat-value">{{ state?.mode ?? '未知' }}</span>
          </div>
          <!-- 未来可扩展：展示MVP、伤害统计等 -->
          <div v-if="gameOverStats.reason" class="yzs-game-over__stat">
            <span class="yzs-game-over__stat-label">结束原因</span>
            <span class="yzs-game-over__stat-value">{{ gameOverStats.reason }}</span>
          </div>
        </div>

        <div class="yzs-game-over__actions">
          <button
            type="button"
            class="yzs-game-over__btn yzs-game-over__btn--primary"
            @click="emit('restart')"
          >
            再来一局
          </button>
          <!-- 预留：返回大厅、查看详情等按钮 -->
          <button
            type="button"
            class="yzs-game-over__btn yzs-game-over__btn--secondary"
            @click="$router.push('/games/yuzhousha')"
          >
            返回大厅
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.yzs-game-over {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: grid;
  place-items: center;
}

.yzs-game-over__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.yzs-game-over__card {
  position: relative;
  width: min(420px, 90vw);
  padding: 40px 32px;
  border-radius: 20px;
  background: var(--color-surface);
  box-shadow: var(--shadow-lg);
  text-align: center;
  animation: yzs-game-over-enter 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes yzs-game-over-enter {
  from {
    opacity: 0;
    transform: translateY(32px) scale(0.96);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.yzs-game-over--win .yzs-game-over__card {
  border: 2px solid var(--color-success);
  box-shadow: var(--shadow-lg), 0 0 40px rgba(21, 128, 61, 0.15);
}

.yzs-game-over--lose .yzs-game-over__card {
  border: 2px solid var(--color-error);
  box-shadow: var(--shadow-lg), 0 0 40px rgba(220, 38, 38, 0.1);
}

.yzs-game-over__icon {
  font-size: 64px;
  line-height: 1;
  margin-bottom: 16px;
}

.yzs-game-over__title {
  margin: 0 0 12px;
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
}

.yzs-game-over--win .yzs-game-over__title {
  color: var(--color-success);
}

.yzs-game-over--lose .yzs-game-over__title {
  color: var(--color-error);
}

.yzs-game-over__message {
  margin: 0 0 24px;
  font-size: 1.1rem;
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.yzs-game-over__stats {
  display: flex;
  justify-content: center;
  gap: 24px;
  margin-bottom: 32px;
  padding: 16px;
  border-radius: 12px;
  background: var(--color-surface-soft);
}

.yzs-game-over__stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.yzs-game-over__stat-label {
  font-size: 12px;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.yzs-game-over__stat-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-text);
}

.yzs-game-over__actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.yzs-game-over__btn {
  padding: 12px 28px;
  border: none;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.yzs-game-over__btn--primary {
  background: var(--color-accent);
  color: white;
}

.yzs-game-over__btn--primary:hover {
  background: #1d4ed8;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

.yzs-game-over__btn--secondary {
  background: var(--color-surface-muted);
  color: var(--color-text-secondary);
}

.yzs-game-over__btn--secondary:hover {
  background: var(--color-border);
}

/* Transition */
.yzs-game-over-fade-enter-active,
.yzs-game-over-fade-leave-active {
  transition: opacity 0.3s;
}

.yzs-game-over-fade-enter-from,
.yzs-game-over-fade-leave-to {
  opacity: 0;
}
</style>
