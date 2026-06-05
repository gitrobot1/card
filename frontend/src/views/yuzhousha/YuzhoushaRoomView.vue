<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  fetchYuzhoushaHeroes,
  fetchYuzhoushaRoom,
  joinYuzhoushaRoom,
  leaveYuzhoushaRoom,
  readyYuzhoushaRoom,
  setYuzhoushaRoomHero,
  startYuzhoushaRoom,
} from '../../api/games'
import { loadSession } from '../../api/auth'
import { useYuzhoushaRoomSocket } from '../../composables/useYuzhoushaSocket'
import { heroAccentColor } from '../../composables/yuzhousha/resolveYzsHeroDisplay'
import { skillBlockedInMode } from '../../constants/yzsModes'
import {
  normalizeOnlineMode,
  onlineModeMeta,
  YZS_ONLINE_3V3_SEAT_ROLES,
} from '../../constants/yzsOnlineModes'
import { showToast } from '../../composables/useToast'
import type { YuzhoushaRoom, YzsCharacter } from '../../types/yuzhousha'
import { YZS_KINGDOM_LABELS } from '../../types/yuzhousha'

const router = useRouter()
const route = useRoute()

const room = ref<YuzhoushaRoom | null>(null)
const heroes = ref<YzsCharacter[]>([])
const selectedHeroId = ref('')
const loading = ref(false)
const heroesLoading = ref(false)
const selfReady = ref(false)
const wsRoomConnected = ref(false)

const session = loadSession()
const selfUserId = computed(() => session?.user.id ?? 0)
const roomEnabled = computed(() => Boolean(room.value?.id) && room.value?.status === 'waiting')

const roomMode = computed(() => normalizeOnlineMode(room.value?.mode ?? (route.query.mode as string)))
const modeMeta = computed(() => onlineModeMeta(roomMode.value))
const requiredPlayers = computed(() => modeMeta.value.playerCount)

const isHost = computed(() => room.value?.host_user_id === selfUserId.value)
const playerCount = computed(() => room.value?.players.length ?? 0)
const othersAllReady = computed(() => {
  const players = room.value?.players ?? []
  const hostId = room.value?.host_user_id
  if (playerCount.value < requiredPlayers.value) return false
  return players.every((p) => p.user_id === hostId || p.ready)
})
const canStart = computed(
  () => isHost.value && othersAllReady.value && room.value?.status === 'waiting',
)
const invitePath = computed(() => {
  if (!room.value) return ''
  return `/games/yuzhousha/online?room=${room.value.id}&mode=${roomMode.value}`
})

function seatLabel(index: number) {
  if (roomMode.value !== '3v3') return `${index + 1} 号位`
  return YZS_ONLINE_3V3_SEAT_ROLES[index] ?? `${index + 1} 号位`
}

let pollTimer: number | null = null

function toastError(message: string) {
  showToast(message, 'error')
}

function syncFromRoom(next: YuzhoushaRoom) {
  room.value = next
  const self = next.players.find((p) => p.user_id === selfUserId.value)
  selfReady.value = self?.ready ?? false
  if (self?.character_id) selectedHeroId.value = self.character_id
  goToGame(next)
}

async function applyRemoteRoom(next: YuzhoushaRoom) {
  syncFromRoom(next)
}

useYuzhoushaRoomSocket({
  roomId: computed(() => room.value?.id),
  enabled: roomEnabled,
  currentRoom: room,
  onStatus: (status) => {
    wsRoomConnected.value = status === 'open'
  },
  onRoom: applyRemoteRoom,
})

function goToGame(next: YuzhoushaRoom) {
  if (next.status !== 'playing' || !next.game_id) return
  stopPolling()
  router.replace({
    name: 'yuzhousha-play',
    params: { gameId: next.game_id },
    query: { room: next.id },
  })
}

async function loadHeroes() {
  heroesLoading.value = true
  try {
    const res = await fetchYuzhoushaHeroes({ mode: roomMode.value, page: 1, page_size: 48 })
    heroes.value = res.heroes
  } catch (err) {
    toastError(err instanceof Error ? err.message : '加载武将失败')
  } finally {
    heroesLoading.value = false
  }
}

async function refreshRoom() {
  if (!room.value?.id) return
  try {
    syncFromRoom(await fetchYuzhoushaRoom(room.value.id))
  } catch (err) {
    toastError(err instanceof Error ? err.message : '同步房间失败')
  }
}

function startPolling() {
  stopPolling()
  if (wsRoomConnected.value) return
  pollTimer = window.setInterval(refreshRoom, 5000)
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
    const mode = normalizeOnlineMode(route.query.mode as string | undefined)
    room.value = await joinYuzhoushaRoom(
      inviteRoomId ? { room_id: inviteRoomId } : { mode },
    )
    syncFromRoom(room.value)
    if (room.value.status === 'playing' && room.value.game_id) {
      goToGame(room.value)
      return
    }
    await loadHeroes()
    startPolling()
  } catch (err) {
    toastError(err instanceof Error ? err.message : '加入房间失败')
  } finally {
    loading.value = false
  }
}

async function pickHero(heroId: string) {
  if (!room.value || loading.value) return
  selectedHeroId.value = heroId
  loading.value = true
  try {
    syncFromRoom(await setYuzhoushaRoomHero(room.value.id, heroId))
  } catch (err) {
    toastError(err instanceof Error ? err.message : '选将失败')
  } finally {
    loading.value = false
  }
}

async function toggleReady() {
  if (!room.value || loading.value || !selectedHeroId.value) {
    if (!selectedHeroId.value) toastError('请先选择武将')
    return
  }
  loading.value = true
  const nextReady = !selfReady.value
  try {
    syncFromRoom(await readyYuzhoushaRoom(room.value.id, nextReady))
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
    syncFromRoom(await startYuzhoushaRoom(room.value.id))
    goToGame(room.value)
  } catch (err) {
    toastError(err instanceof Error ? err.message : '开始失败')
  } finally {
    loading.value = false
  }
}

async function handleLeave() {
  if (!room.value) {
    router.push('/games/yuzhousha')
    return
  }
  loading.value = true
  try {
    await leaveYuzhoushaRoom(room.value.id)
  } catch {
    // ignore
  } finally {
    loading.value = false
    stopPolling()
    router.push('/games/yuzhousha')
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

function kingdomLabel(k?: string) {
  if (!k) return ''
  return YZS_KINGDOM_LABELS[k] ?? k
}

onMounted(async () => {
  await enterRoom()
})
onUnmounted(stopPolling)

watch(wsRoomConnected, (open) => {
  if (open) stopPolling()
  else if (room.value?.status === 'waiting') startPolling()
})
</script>

<template>
  <main class="ddz-room app">
    <section class="hero">
      <div class="hero__top">
        <div>
          <p class="hero__tag">宇宙杀 · {{ modeMeta.label }} 联机</p>
          <h1>等待房间</h1>
          <p class="hero__desc">
            选择武将并准备 · 需 {{ requiredPlayers }} 人 · 房间 {{ room?.id?.slice(0, 8) ?? '…' }}
            <span v-if="roomMode === 'identity_5' || roomMode === 'identity_8'">
              · 开局随机分配身份（主公公开）
            </span>
            <span v-if="wsRoomConnected" class="yzs-room__ws"> · 已连接</span>
          </p>
        </div>
        <button type="button" class="hero__logout" @click="handleLeave">← 离开</button>
      </div>
    </section>

    <section v-if="room" class="ddz-room__panel">
      <h2>玩家 ({{ playerCount }}/{{ requiredPlayers }})</h2>
      <ul class="ddz-room__list">
        <li v-for="(p, i) in room.players" :key="p.user_id" class="ddz-room__player">
          <span>
            {{ seatLabel(i) }} · {{ p.username }}{{ p.user_id === room.host_user_id ? '（房主）' : '' }}
          </span>
          <span>
            {{ p.character_id ? '已选将' : '未选将' }}
            · {{ p.ready ? '已准备' : '未准备' }}
          </span>
        </li>
      </ul>
      <div class="ddz-room__actions">
        <button type="button" class="ddz__btn ddz__btn--hint" @click="copyInvite">复制邀请链接</button>
        <button type="button" class="ddz__btn" :disabled="loading || !selectedHeroId" @click="toggleReady">
          {{ selfReady ? '取消准备' : '准备' }}
        </button>
        <button
          v-if="isHost"
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="loading || !canStart"
          @click="handleStart"
        >
          开始对战
        </button>
      </div>
    </section>

    <p v-if="heroesLoading" class="ddz__loading">加载武将…</p>
    <section v-else class="yzs-pick">
      <h2 class="yzs-room__pick-title">选择你的武将</h2>
      <div class="yzs-pick__grid">
        <button
          v-for="hero in heroes"
          :key="hero.id"
          type="button"
          class="yzs-pick__card"
          :class="{ 'yzs-pick__card--selected': selectedHeroId === hero.id }"
          :style="{ '--hero-accent': heroAccentColor(hero) }"
          :disabled="loading"
          @click="pickHero(hero.id)"
        >
          <span class="yzs-pick__kingdom">{{ kingdomLabel(hero.kingdom) }}</span>
          <h2 class="yzs-pick__name">{{ hero.name }}</h2>
          <p class="yzs-pick__hp">体力 {{ hero.max_hp }}</p>
          <ul class="yzs-pick__skills">
            <li
              v-for="skill in hero.skills"
              :key="skill.id"
              :class="{ 'yzs-pick__skill--inactive': skillBlockedInMode(skill, roomMode) }"
            >
              <strong>{{ skill.name }}</strong>
            </li>
          </ul>
        </button>
      </div>
    </section>
  </main>
</template>

<style scoped>
.yzs-room__pick-title {
  margin: 1rem 0 0.75rem;
  font-size: 1.1rem;
}
.yzs-room__ws {
  color: var(--color-success, #4ade80);
}
</style>
