<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  fetchDouDizhuRoom,
  joinDouDizhuRoom,
  leaveDouDizhuRoom,
  readyDouDizhuRoom,
} from '../api/games'
import { loadSession } from '../api/auth'
import type { DouDizhuRoom } from '../types/doudizhu'

const router = useRouter()
const route = useRoute()

const room = ref<DouDizhuRoom | null>(null)
const loading = ref(false)
const error = ref('')
const selfReady = ref(false)

const session = loadSession()
const selfUserId = computed(() => session?.user.id ?? 0)

const slots = computed(() => {
  const players = room.value?.players ?? []
  return Array.from({ length: 3 }, (_, index) => players[index] ?? null)
})

const playerCount = computed(() => room.value?.players.length ?? 0)
const isFull = computed(() => playerCount.value >= 3)

let pollTimer: number | null = null

async function refreshRoom() {
  if (!room.value?.id) return
  try {
    const next = await fetchDouDizhuRoom(room.value.id)
    room.value = next
    const me = next.players.find((p) => p.user_id === selfUserId.value)
    selfReady.value = me?.ready ?? false

    if (next.status === 'playing' && next.game_id) {
      stopPolling()
      router.replace({
        name: 'doudizhu-play',
        params: { gameId: next.game_id },
        query: { room: next.id },
      })
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '同步房间失败'
  }
}

function startPolling() {
  stopPolling()
  pollTimer = window.setInterval(refreshRoom, 1500)
}

function stopPolling() {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

async function enterRoom() {
  loading.value = true
  error.value = ''
  try {
    const inviteRoomId = route.query.room as string | undefined
    room.value = await joinDouDizhuRoom(inviteRoomId)
    const me = room.value.players.find((p) => p.user_id === selfUserId.value)
    selfReady.value = me?.ready ?? false
    if (room.value.status === 'playing' && room.value.game_id) {
      router.replace({
        name: 'doudizhu-play',
        params: { gameId: room.value.game_id },
        query: { room: room.value.id },
      })
      return
    }
    startPolling()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加入房间失败'
  } finally {
    loading.value = false
  }
}

async function toggleReady() {
  if (!room.value || loading.value) return
  loading.value = true
  error.value = ''
  const nextReady = !selfReady.value
  try {
    room.value = await readyDouDizhuRoom(room.value.id, nextReady)
    selfReady.value = nextReady
    if (room.value.status === 'playing' && room.value.game_id) {
      stopPolling()
      router.replace({
        name: 'doudizhu-play',
        params: { gameId: room.value.game_id },
        query: { room: room.value.id },
      })
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '准备失败'
  } finally {
    loading.value = false
  }
}

async function handleLeave() {
  if (!room.value) {
    router.push('/games/doudizhu')
    return
  }
  loading.value = true
  try {
    await leaveDouDizhuRoom(room.value.id)
  } catch {
    // ignore leave errors when navigating away
  } finally {
    loading.value = false
    stopPolling()
    router.push('/games/doudizhu')
  }
}

onMounted(enterRoom)

onUnmounted(stopPolling)
</script>

<template>
  <main class="ddz-room app">
    <section class="hero">
      <div class="hero__top">
        <div>
          <p class="hero__tag">多人联机</p>
          <h1>等待房间</h1>
          <p class="hero__desc">3 人准备后自动开局</p>
        </div>
        <button type="button" class="hero__logout" :disabled="loading" @click="handleLeave">
          ← 离开房间
        </button>
      </div>
    </section>

    <p v-if="error" class="ddz__error">{{ error }}</p>
    <p v-if="loading && !room" class="ddz__loading">正在加入房间...</p>

    <section v-if="room" class="ddz-room__panel">
      <div class="ddz-room__meta">
        <span>房间号</span>
        <strong>{{ room.id.slice(0, 8) }}</strong>
        <span class="ddz-room__count">{{ playerCount }} / 3 人</span>
      </div>

      <div class="ddz-room__slots">
        <div
          v-for="(player, index) in slots"
          :key="index"
          class="ddz-room__slot"
          :class="{ 'ddz-room__slot--filled': !!player, 'ddz-room__slot--ready': player?.ready }"
        >
          <span class="ddz-room__slot-index">座位 {{ index + 1 }}</span>
          <strong v-if="player">{{ player.username }}</strong>
          <span v-else class="ddz-room__slot-empty">等待加入...</span>
          <span v-if="player?.ready" class="ddz__ready-badge">已准备</span>
        </div>
      </div>

      <p v-if="!isFull" class="ddz-room__hint">还差 {{ 3 - playerCount }} 人，可分享房间号邀请好友</p>
      <p v-else-if="!slots.every((p) => p?.ready)" class="ddz-room__hint">已满员，等待全员准备</p>
      <p v-else class="ddz-room__hint">全员已准备，即将开始...</p>

      <button
        type="button"
        class="ddz__btn ddz__btn--primary ddz-room__ready"
        :disabled="loading || selfReady"
        @click="toggleReady"
      >
        {{ selfReady ? '已准备' : '准备' }}
      </button>
    </section>
  </main>
</template>
