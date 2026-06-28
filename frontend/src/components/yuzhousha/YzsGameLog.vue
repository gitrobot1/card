<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { GameLogEntry } from '../../types/yuzhousha'

const props = defineProps<{
  entries: GameLogEntry[]
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle'): void
}>()

const listRef = ref<HTMLElement | null>(null)

// 新条目时自动滚动到底部
watch(
  () => props.entries?.length ?? 0,
  async () => {
    await nextTick()
    if (listRef.value) {
      listRef.value.scrollTop = listRef.value.scrollHeight
    }
  },
)

const collapsed = ref(false)
</script>

<template>
  <aside
    class="yzs-game-log"
    :class="{ 'yzs-game-log--hidden': !visible, 'yzs-game-log--collapsed': collapsed }"
  >
    <button
      type="button"
      class="yzs-game-log__toggle"
      @click="collapsed = !collapsed"
      :title="collapsed ? '展开日志' : '折叠日志'"
    >
      {{ collapsed ? '📋' : '📋 战报' }}
    </button>

    <div v-if="!collapsed" class="yzs-game-log__body">
      <div ref="listRef" class="yzs-game-log__list">
        <p v-if="!entries || entries.length === 0" class="yzs-game-log__empty">
          暂无记录
        </p>
        <div
          v-for="(entry, idx) in entries"
          :key="idx"
          class="yzs-game-log__entry"
          :class="{ 'yzs-game-log__entry--turn': entry.msg.startsWith('———') }"
        >
          <span class="yzs-game-log__idx">{{ entries.length - idx }}</span>
          <span class="yzs-game-log__msg">{{ entry.msg }}</span>
        </div>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.yzs-game-log {
  position: fixed;
  right: 0;
  top: 60px;
  bottom: 0;
  width: 240px;
  z-index: 700;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-left: 1px solid #e0e0e0;
  box-shadow: -2px 0 8px rgba(0,0,0,0.06);
  transition: transform 0.25s ease;
}

.yzs-game-log--hidden {
  transform: translateX(100%);
}

.yzs-game-log--collapsed {
  width: auto;
  right: 0;
  bottom: auto;
  top: 60px;
}

.yzs-game-log__toggle {
  flex-shrink: 0;
  padding: 6px 12px;
  border: none;
  border-bottom: 1px solid #eee;
  background: #fafafa;
  color: #333;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s;
}

.yzs-game-log__toggle:hover {
  background: #f0f0f0;
}

.yzs-game-log--collapsed .yzs-game-log__toggle {
  padding: 6px 8px;
  border-radius: 0 0 0 6px;
  border-bottom: none;
  border-left: 1px solid #e0e0e0;
  border-bottom: 1px solid #e0e0e0;
}

.yzs-game-log__body {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.yzs-game-log__list {
  flex: 1;
  overflow-y: auto;
  padding: 6px 8px;
  display: flex;
  flex-direction: column;
}

.yzs-game-log__list::-webkit-scrollbar {
  width: 4px;
}

.yzs-game-log__list::-webkit-scrollbar-thumb {
  background: #ccc;
  border-radius: 2px;
}

.yzs-game-log__empty {
  color: #999;
  font-size: 0.75rem;
  text-align: center;
  padding: 16px 0;
}

.yzs-game-log__entry {
  display: flex;
  gap: 4px;
  padding: 3px 0;
  font-size: 0.75rem;
  line-height: 1.5;
  color: #1a1a1a;
  border-bottom: 1px solid #f0f0f0;
}

.yzs-game-log__entry--turn {
  color: #999;
  font-weight: 600;
  font-size: 0.68rem;
  border-bottom-color: #e8e8e8;
  padding: 5px 0 2px;
}

.yzs-game-log__idx {
  flex-shrink: 0;
  width: 18px;
  text-align: right;
  color: #bbb;
  font-size: 0.62rem;
  font-variant-numeric: tabular-nums;
}

.yzs-game-log__msg {
  flex: 1;
  word-break: break-all;
}
</style>
