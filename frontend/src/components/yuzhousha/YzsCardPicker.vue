<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { YzsCard, YzsPlayer } from '../../types/yuzhousha'
import { equippedCards, equipSlotOf, judgeAreaCards } from '../../composables/yuzhousha/playerCardHelpers'

export interface PickerOption {
  zone: string
  card: YzsCard
}

const props = defineProps<{
  title?: string
  player: YzsPlayer | null
  /** 是否多选，默认 false（单选） */
  multi?: boolean
  /** 多选上限（仅 multi=true 时生效），不设则无上限 */
  maxSelect?: number
  /** 是否显示确认/取消按钮 */
  showActions?: boolean
  confirmText?: string
  cancelText?: string
  visible: boolean
  /** 是否显示手牌内容，false 则显示牌背面 */
  showHand?: boolean
}>()

const emit = defineEmits<{
  confirm: [selections: PickerOption[]]
  cancel: []
}>()

// ---- 本地选中状态 ----
const selectedIds = ref<Set<string>>(new Set())

watch(() => props.visible, (v) => {
  if (!v) selectedIds.value = new Set()
})

function isSelected(cardId: string) {
  return selectedIds.value.has(cardId)
}

function toggleOption(option: PickerOption) {
  if (option.zone === 'hand' && !props.showHand) {
    // 牌背面不可见手牌内容时，仍然可选中（用于过河拆桥选牌）
  }
  if (props.multi) {
    const next = new Set(selectedIds.value)
    if (next.has(option.card.id)) {
      next.delete(option.card.id)
    } else {
      if (props.maxSelect && next.size >= props.maxSelect) return
      next.add(option.card.id)
    }
    selectedIds.value = next
  } else {
    // 单选：点击即选中（替换之前选中）
    selectedIds.value = new Set([option.card.id])
  }
}

function onConfirm() {
  const result: PickerOption[] = []
  for (const id of selectedIds.value) {
    // 从所有选项中找
    const opt = allOptions.value.find(o => o.card.id === id)
    if (opt) result.push(opt)
  }
  emit('confirm', result)
}

// ---- 数据 ----
const handCards = computed<PickerOption[]>(() => {
  if (!props.player) return []
  if (props.showHand) {
    return (props.player.hand ?? []).map(c => ({ zone: 'hand', card: c }))
  }
  const count = props.player.hand_count ?? 0
  return Array.from({ length: count }, (_, i) => ({
    zone: 'hand',
    // 背面牌 id 留空，后端 TakeOne 取手牌时总是取第一张，cardID 参数被忽略
    card: { id: `_hand_${i}`, kind: 'hand_back', suit: '', name: '', label: '' } as YzsCard,
  }))
})

const equipCards = computed<PickerOption[]>(() => {
  if (!props.player) return []
  return equippedCards(props.player).map(equip => ({
    zone: equipSlotOf(equip),
    card: equip,
  }))
})

const judgeCards = computed<PickerOption[]>(() => {
  if (!props.player) return []
  return judgeAreaCards(props.player).map(c => ({ zone: 'judge', card: c }))
})

const allOptions = computed(() => [...handCards.value, ...equipCards.value, ...judgeCards.value])
</script>

<template>
  <Transition name="yzs-picker-fade">
    <div v-if="visible" class="yzs-picker" @click.self="emit('cancel')">
      <div class="yzs-picker__backdrop" />
      <div class="yzs-picker__card">
        <div class="yzs-picker__header">
          <div class="yzs-picker__title">{{ title || '' }}</div>
          <button type="button" class="yzs-picker__close" @click="emit('cancel')">✕</button>
        </div>

        <!-- 手牌区 -->
        <div class="yzs-picker__section">
          <div class="yzs-picker__section-title">手牌 · {{ handCards.length }}</div>
          <div class="yzs-picker__section-cards">
            <template v-if="handCards.length">
              <div
                v-for="opt in handCards"
                :key="opt.card.id"
                class="yzs-card-back"
                :class="{ 'yzs-card-back--selected': isSelected(opt.card.id) }"
                @click="toggleOption(opt)"
              />
            </template>
            <div v-else class="yzs-picker__empty">—</div>
          </div>
        </div>

        <!-- 装备区 -->
        <div class="yzs-picker__section">
          <div class="yzs-picker__section-title">装备</div>
          <div class="yzs-picker__section-cards">
            <template v-if="equipCards.length">
              <div
                v-for="opt in equipCards"
                :key="opt.card.id"
                class="yzs-card-real"
                :class="{ 'yzs-card-real--selected': isSelected(opt.card.id) }"
                @click="toggleOption(opt)"
              >
                <span class="yzs-card-suit">{{ opt.card.suit === 'H' ? '♥' : opt.card.suit === 'D' ? '♦' : opt.card.suit === 'S' ? '♠' : opt.card.suit === 'C' ? '♣' : '' }}</span>
                <span class="yzs-card-label">{{ opt.card.name || opt.card.label }}</span>
              </div>
            </template>
            <div v-else class="yzs-picker__empty">—</div>
          </div>
        </div>

        <!-- 判定区 -->
        <div class="yzs-picker__section">
          <div class="yzs-picker__section-title">判定区</div>
          <div class="yzs-picker__section-cards">
            <template v-if="judgeCards.length">
              <div
                v-for="opt in judgeCards"
                :key="opt.card.id"
                class="yzs-card-real yzs-card-real--judge"
                :class="{ 'yzs-card-real--selected': isSelected(opt.card.id) }"
                @click="toggleOption(opt)"
              >
                <span class="yzs-card-label">{{ opt.card.name }}</span>
              </div>
            </template>
            <div v-else class="yzs-picker__empty">—</div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div v-if="showActions" class="yzs-picker__actions">
          <button
            type="button"
            class="yzs-picker__btn yzs-picker__btn--primary"
            :disabled="selectedIds.size === 0"
            @click="onConfirm"
          >
            {{ confirmText || '确定' }}
          </button>
          <button
            type="button"
            class="yzs-picker__btn yzs-picker__btn--secondary"
            @click="emit('cancel')"
          >
            {{ cancelText || '取消' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.yzs-picker {
  position: fixed;
  inset: 0;
  z-index: 900;
  display: grid;
  place-items: center;
}

.yzs-picker__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.yzs-picker__card {
  position: relative;
  width: min(480px, 90vw);
  max-height: 82vh;
  padding: 20px 24px 16px;
  border-radius: 16px;
  background: var(--color-surface);
  box-shadow: var(--shadow-lg);
  overflow-y: auto;
  animation: yzs-picker-enter 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes yzs-picker-enter {
  from { opacity: 0; transform: translateY(24px) scale(0.96); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.yzs-picker__header {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-bottom: 12px;
  min-height: 32px;
}

.yzs-picker__title {
  flex: 1;
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.yzs-picker__close {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: var(--color-surface-muted);
  color: var(--color-text-secondary);
  font-size: 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  flex-shrink: 0;
}

.yzs-picker__close:hover {
  background: #ef4444;
  color: white;
}

/* 区域 */
.yzs-picker__section {
  margin-bottom: 14px;
}

.yzs-picker__section-title {
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--color-text-secondary);
  margin-bottom: 8px;
  padding-bottom: 4px;
  border-bottom: 1.5px solid var(--color-border);
}

.yzs-picker__section-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  min-height: 54px;
  padding: 6px;
  border-radius: 8px;
  background: var(--color-surface-muted);
  align-items: flex-start;
}

.yzs-picker__empty {
  width: 100%;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 0.85rem;
  padding: 8px 0;
  user-select: none;
}

/* 牌背面 */
.yzs-card-back {
  width: 64px;
  height: 90px;
  border-radius: 6px;
  background: linear-gradient(135deg, #1a3a6b, #0d2240);
  border: 2px solid #2a5aaa;
  cursor: pointer;
  transition: all 0.15s;
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
}

.yzs-card-back::after {
  content: '';
  position: absolute;
  inset: 3px;
  border-radius: 3px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 2px,
    rgba(255, 255, 255, 0.03) 2px,
    rgba(255, 255, 255, 0.03) 4px
  );
}

.yzs-card-back:hover {
  border-color: #4a8aee;
  transform: translateY(-3px);
  box-shadow: 0 4px 12px rgba(42, 90, 170, 0.3);
}

.yzs-card-back--selected {
  border-color: #f59e0b;
  background: linear-gradient(135deg, #5a3a0b, #3a2000);
  box-shadow: 0 0 14px rgba(245, 158, 11, 0.5);
  transform: translateY(-4px);
}

/* 装备/判定牌 */
.yzs-card-real {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 90px;
  padding: 4px;
  border-radius: 6px;
  background: #fefce8;
  border: 2px solid #d4c8a0;
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;
  gap: 2px;
}

.yzs-card-real:hover {
  border-color: #2563eb;
  transform: translateY(-3px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
}

.yzs-card-real--selected {
  border-color: #f59e0b;
  background: #fef3c7;
  box-shadow: 0 0 14px rgba(245, 158, 11, 0.5);
  transform: translateY(-4px);
}

.yzs-card-real--judge {
  border-color: #a78bfa;
  background: #f5f3ff;
}

.yzs-card-real--judge:hover {
  border-color: #7c3aed;
}

.yzs-card-suit {
  font-size: 1rem;
  line-height: 1;
}

.yzs-card-label {
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--color-text);
  text-align: center;
  line-height: 1.15;
  word-break: break-all;
}

/* 按钮 */
.yzs-picker__actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--color-border);
}

.yzs-picker__btn {
  padding: 10px 28px;
  border: none;
  border-radius: 10px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.yzs-picker__btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.yzs-picker__btn--primary {
  background: var(--color-accent);
  color: white;
}

.yzs-picker__btn--primary:hover:not(:disabled) {
  background: #1d4ed8;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

.yzs-picker__btn--secondary {
  background: var(--color-surface-muted);
  color: var(--color-text-secondary);
}

.yzs-picker__btn--secondary:hover {
  background: var(--color-border);
}

/* Transition */
.yzs-picker-fade-enter-active,
.yzs-picker-fade-leave-active {
  transition: opacity 0.25s;
}
.yzs-picker-fade-enter-from,
.yzs-picker-fade-leave-to {
  opacity: 0;
}
</style>
