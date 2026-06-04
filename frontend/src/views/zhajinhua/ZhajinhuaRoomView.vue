<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  fetchZhajinhuaRoom,
  joinZhajinhuaRoom,
  leaveZhajinhuaRoom,
  readyZhajinhuaRoom,
  startZhajinhuaRoom,
} from '../../api/games'
import { loadSession } from '../../api/auth'
import { showToast } from '../../composables/useToast'
import type { ZhajinhuaRoom } from '../../types/zhajinhua'

const router = useRouter()
const route = useRoute()

const room = ref<ZhajinhuaRoom | null>(null)
const loading = ref(false)
const selfReady = ref(false)

const session = loadSession()
const selfUserId = computed(() => session?.user.id ?? 0)
const maxSeats = 8

const slots = computed(() => {
  const players = room.value?.players ?? []
  return Array.from({ length: maxSeats }, (_, index) => players[index] ?? null)
})

const playerCount = computed(() => room.value?.players.length ?? 0)
const isHost = computed(() => room.value?.host_user_id === selfUserId.value)
/** 房主无需准备；其他玩家全部准备即可开局 */
const othersAllReady = computed(() => {
  const players = room.value?.players ?? []
  const hostId = room.value?.host_user_id
  if (playerCount.value < 2) return false
  return players.every((p) => p.user_id === hostId || p.ready)
})
const canStart = computed(
  () => isHost.value && othersAllReady.value && room.value?.status === 'waiting',
)
const invitePath = computed(() => {
  if (!room.value) return ''
  return `/games/zhajinhua/online?room=${room.value.id}`
})

let pollTimer: number | null = null

function toastError(message: string) {
  showToast(message, 'error')
}

function goToGame(next: ZhajinhuaRoom) {
  if (next.status !== 'playing' || !next.game_id) return
  stopPolling()
  router.replace({
    name: 'zhajinhua-play',
    params: { gameId: next.game_id },
    query: { room: next.id },
  })
}

async function refreshRoom() {
  if (!room.value?.id) return
  try {
    const next = await fetchZhajinhuaRoom(room.value.id)
    room.value = next
    const me = next.players.find((p) => p.user_id === selfUserId.value)
    selfReady.value = me?.ready ?? false
    goToGame(next)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '同步房间失败')
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
  try {
    const inviteRoomId = route.query.room as string | undefined
    room.value = await joinZhajinhuaRoom(inviteRoomId)
    const me = room.value.players.find((p) => p.user_id === selfUserId.value)
    selfReady.value = me?.ready ?? false
    if (room.value.status === 'playing' && room.value.game_id) {
      goToGame(room.value)
      return
    }
    startPolling()
  } catch (err) {
    toastError(err instanceof Error ? err.message : '加入房间失败')
  } finally {
    loading.value = false
  }
}

async function toggleReady() {
  if (!room.value || loading.value) return
  loading.value = true
  const nextReady = !selfReady.value
  try {
    room.value = await readyZhajinhuaRoom(room.value.id, nextReady)
    selfReady.value = nextReady
    goToGame(room.value)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '准备失败')
  } finally {
    loading.value = false
  }
}

async function handleStart() {
  if (!room.value || loading.value || !canStart.value) return
  loading.value = true
  try {
    room.value = await startZhajinhuaRoom(room.value.id)
    goToGame(room.value)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '开始失败')
  } finally {
    loading.value = false
  }
}

async function handleLeave() {
  if (!room.value) {
    router.push('/games/zhajinhua')
    return
  }
  loading.value = true
  try {
    await leaveZhajinhuaRoom(room.value.id)
  } catch {
    // ignore leave errors when navigating away
  } finally {
    loading.value = false
    stopPolling()
    router.push('/games/zhajinhua')
  }
}

async function copyInvite() {
  if (!invitePath.value) return
  const url = `${window.location.origin}${invitePath.value}`
  try {
    await navigator.clipboard.writeText(url)
    showToast('邀请链接已复制', 'success')
  } catch {
    showToast(`邀请链接：${url}`, 'info', 5000)
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
          <p class="hero__tag">扎金花 · 联机</p>
          <h1>等待房间</h1>
          <p class="hero__desc">2-8 人，其他玩家准备后房主开始</p>
        </div>
        <button type="button" class="hero__logout" :disabled="loading" @click="handleLeave">
          ← 离开房间
        </button>
      </div>
    </section>

    <p v-if="loading && !room" class="ddz__loading">正在加入房间...</p>

    <section v-if="room" class="ddz-room__panel">
      <div class="ddz-room__meta">
        <span>房间号</span>
        <strong>{{ room.id.slice(0, 8) }}</strong>
        <span class="ddz-room__count">{{ playerCount }} / {{ maxSeats }} 人</span>
      </div>

      <div class="ddz-room__slots ddz-room__slots--8">
        <div
          v-for="(player, index) in slots"
          :key="index"
          class="ddz-room__slot"
          :class="{
            'ddz-room__slot--filled': !!player,
            'ddz-room__slot--ready': !!player?.ready && player.user_id !== room.host_user_id,
          }"
        >
          <span class="ddz-room__slot-index">座位 {{ index + 1 }}</span>
          <strong v-if="player">
            {{ player.username }}
            <span v-if="player.user_id === room.host_user_id" class="zjh-room__host">房主</span>
          </strong>
          <span v-else class="ddz-room__slot-empty">空位</span>
          <span v-if="player?.ready && player.user_id !== room.host_user_id" class="ddz__ready-badge">已准备</span>
          <span v-else-if="player && player.user_id !== room.host_user_id" class="zjh-room__waiting">未准备</span>
        </div>
      </div>

      <p v-if="playerCount < 2" class="ddz-room__hint">
        还差 {{ 2 - playerCount }} 人，可分享房间邀请好友
      </p>
      <p v-else-if="isHost && !othersAllReady" class="ddz-room__hint">
        当前 {{ playerCount }} 人，等待其他玩家准备
      </p>
      <p v-else-if="isHost" class="ddz-room__hint">其他玩家已准备，点击开始游戏</p>
      <p v-else-if="!othersAllReady" class="ddz-room__hint">当前 {{ playerCount }} 人，请点击准备</p>
      <p v-else class="ddz-room__hint">已准备，等待房主开始</p>

      <div class="zjh-room__actions">
        <button type="button" class="ddz__btn" :disabled="loading" @click="copyInvite">
          复制邀请链接
        </button>
        <button
          v-if="!isHost"
          type="button"
          class="ddz__btn ddz__btn--primary ddz-room__ready"
          :disabled="loading || playerCount < 2"
          @click="toggleReady"
        >
          {{ selfReady ? '取消准备' : '准备' }}
        </button>
        <button
          v-if="isHost"
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="loading || !canStart"
          @click="handleStart"
        >
          开始游戏
        </button>
      </div>
    </section>
  </main>
</template>
