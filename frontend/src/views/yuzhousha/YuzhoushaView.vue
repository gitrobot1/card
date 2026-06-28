<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import YzsActionBar from '../../components/yuzhousha/YzsActionBar.vue'
import YzsHandZone from '../../components/yuzhousha/YzsHandZone.vue'
import YzsGameOver from '../../components/yuzhousha/YzsGameOver.vue'
import YzsGameLog from '../../components/yuzhousha/YzsGameLog.vue'
import YzsCardPicker from '../../components/yuzhousha/YzsCardPicker.vue'
import type { PickerOption } from '../../components/yuzhousha/YzsCardPicker.vue'
import { useYuzhoushaSkill } from '../../api/games'
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
  isTakeWindow,
  isGuoHeTake,
  isTanNangTake,
  isPojun,
  mySeat,
  seatAt,
  selectedTargetZone,
  selectedTargetCardId,
  submitSkill,
  act,
  gameLog,
} = game

const showGameLog = ref(true)

const layoutComponent = computed(() => resolveYzsLayout(state.value?.layout_key))
const tableSubtitle = computed(() =>
  yzsLayoutSubtitle(state.value?.layout_key, state.value?.mode),
)

// ---- 对手牌框（始终可用，点击按钮弹出查看装备/判定区） ----
const opponentSeat = computed(() => {
  const players = state.value?.players
  if (!players) return -1
  for (let i = 0; i < players.length; i++) {
    if (i !== mySeat.value && players[i].hp > 0) return i
  }
  return -1
})
const opponentPlayer = computed(() => {
  if (opponentSeat.value < 0) return null
  return seatAt(opponentSeat.value)
})

// 手动打开的查看弹窗（只读模式）
const peekVisible = ref(false)

function openPeek() {
  peekVisible.value = true
}

function closePeek() {
  peekVisible.value = false
}

// ---- TakeWindow 拆牌弹窗 ----
const takenSeat = computed(() => state.value?.pending?.subject_seat ?? -1)
const takeWindowPlayer = computed(() => {
  if (takenSeat.value < 0) return null
  return seatAt(takenSeat.value)
})
const takeWindowTitle = computed(() => {
  if (isGuoHeTake.value) return '【过河拆桥】选择要拆掉的牌'
  if (isTanNangTake.value) return '【顺手牵羊】选择要获得的牌'
  return '选择牌'
})
const showTakeWindow = computed(() => isTakeWindow.value && takenSeat.value >= 0)

function onTakeCardConfirm(selections: PickerOption[]) {
  if (selections.length === 0) return
  selectedTargetZone.value = selections[0].zone
  // 手牌背面占位 ID（如 _hand_0）不传给后端，让后端取第一张
  selectedTargetCardId.value = selections[0].card.id.startsWith('_hand_') ? '' : selections[0].card.id
  submitSkill('')
}

function onTakeCardCancel() {
  selectedTargetZone.value = ''
  selectedTargetCardId.value = ''
  submitSkill('')
}

// ---- 破军选牌弹窗 ----
const pojunSubjectSeat = computed(() => state.value?.pending?.subject_seat ?? -1)
const pojunPickerPlayer = computed(() => {
  if (pojunSubjectSeat.value < 0) return null
  return seatAt(pojunSubjectSeat.value)
})
const pojunRemaining = computed(() => {
  const p = state.value?.pending
  return Math.max(0, (p?.pojun_max ?? 0) - (p?.pojun_placed ?? 0))
})
const pojunPickerTitle = computed(() => {
  return `【破军】将至多 ${pojunRemaining.value} 张牌置于「营」`
})
const showPojunPicker = computed(() => isPojun.value && pojunSubjectSeat.value >= 0 && pojunRemaining.value > 0)

function onPojunConfirm(selections: PickerOption[]) {
  if (selections.length === 0) {
    submitSkill('pojun')
    return
  }
  // 一次性提交所有选中的牌
  const cardIds = selections.map(s => s.card.id)
  act(async () => useYuzhoushaSkill(state.value!.id, 'pojun', { cardIds }))
}

function onPojunCancel() {
  submitSkill('pojun')
}
</script>

<template>
  <main class="ddz app">
    <header class="ddz__header">
      <button type="button" class="ddz__back" @click="router.push('/games/yuzhousha')">← 返回</button>
      <div>
        <h1>宇宙杀</h1>
        <p class="ddz__subtitle">{{ tableSubtitle }}</p>
      </div>
      <div v-if="!isFinished" class="zjh__header-spacer" aria-hidden="true" />
    </header>

    <p v-if="loading && !state" class="ddz__loading">正在开局…</p>

    <section
      v-if="state"
      :ref="(el) => { game.tableWrapRef.value = el as HTMLElement | null }"
      class="ddz__table yzs__table-wrap"
      :class="{ 'yzs__table-wrap--finished': isFinished }"
    >
      <div v-if="!isDealing" class="yzs__deck-count">剩余 {{ state.draw_count }}</div>
      <div :ref="(el) => { game.drawAreaRef.value = el as HTMLElement | null }" class="yzs__draw-anchor" aria-hidden="true" />

      <YzsActionBar />
      <component :is="layoutComponent" />
      <YzsHandZone />
    </section>

    <YzsGameOver :state="state" @restart="restart" />

    <!-- 右侧游戏日志 -->
    <YzsGameLog
      :entries="gameLog ?? []"
      :visible="showGameLog"
      @toggle="showGameLog = !showGameLog"
    />

    <!-- 始终显示的"查看对手牌"按钮 -->
    <button
      v-if="state && !isFinished && opponentPlayer"
      type="button"
      class="yzs-peek-btn"
      @click="openPeek"
    >
      👁 查看 {{ opponentPlayer.character.name }}
    </button>

    <!-- 手动打开的对手牌查看弹窗（只读，不显示手牌） -->
    <YzsCardPicker
      :player="opponentPlayer"
      :visible="peekVisible"
      :show-hand="false"
      :show-actions="false"
      @cancel="closePeek"
    />

    <!-- TakeWindow 拆牌弹窗（手牌显示背面但可选中） -->
    <YzsCardPicker
      :title="takeWindowTitle"
      :player="takeWindowPlayer"
      :visible="showTakeWindow"
      :show-hand="false"
      :show-actions="true"
      confirm-text="确定"
      cancel-text="取消"
      @confirm="onTakeCardConfirm"
      @cancel="onTakeCardCancel"
    />

    <!-- 破军选牌弹窗（多选，手牌可见） -->
    <YzsCardPicker
      :title="pojunPickerTitle"
      :player="pojunPickerPlayer"
      :visible="showPojunPicker"
      :show-hand="true"
      :multi="true"
      :max-select="pojunRemaining"
      :show-actions="true"
      confirm-text="置于「营」"
      cancel-text="跳过"
      @confirm="onPojunConfirm"
      @cancel="onPojunCancel"
    />
  </main>
</template>

<style scoped>
.yzs-peek-btn {
  position: fixed;
  right: 16px;
  bottom: 160px;
  z-index: 800;
  padding: 8px 16px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-surface);
  color: var(--color-text);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: var(--shadow-sm);
}

.yzs-peek-btn:hover {
  background: var(--color-accent);
  color: white;
  border-color: var(--color-accent);
  transform: translateY(-1px);
}
</style>
