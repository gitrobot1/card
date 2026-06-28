<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { fetchYuzhoushaHeroes, startYuzhoushaGame } from '../../api/games'
import { showToast } from '../../composables/useToast'
import type { YzsCharacter } from '../../types/yuzhousha'
import { YZS_KINGDOM_LABELS } from '../../types/yuzhousha'
import { heroAccentColor } from '../../composables/yuzhousha/resolveYzsHeroDisplay'
import { skillBlockedInMode } from '../../constants/yzsModes'

const PAGE_SIZE = 12

const router = useRouter()
const route = useRoute()
const gameMode = computed(() => {
  const m = route.query.mode
  if (typeof m === 'string' && m.length > 0) return m
  return '1v1'
})
const heroes = ref<YzsCharacter[]>([])
const total = ref(0)
const page = ref(1)
const totalPages = ref(0)
const selectedId = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const starting = ref(false)
const kingdomFilter = ref('')

const pickSubtitle = computed(() => {
  switch (gameMode.value) {
    case '2v2':
      return '2v2 十字阵 · 选将后队友与敌将随机'
    case '3v3':
      return '3v3 竞技 · 你担任暖色主帅'
    case '3p_chain':
      return '3 人链式 · 杀上家保下家'
    case '3p_ddz':
      return '3 人斗地主 · 你担任地主'
    case 'identity_5':
      return '5 人身份局 · 标准场（1 忠 1 内 2 反）· 你担任主公'
    case 'identity_8':
      return '8 人身份局 · 标准场（2 忠 1 内 4 反）· 你担任主公'
    default:
      return '1v1 单机 · 电脑随机选取剩余武将'
  }
})

const hasMore = computed(() => page.value < totalPages.value)

const kingdomOptions = [
  { id: '', label: '全部' },
  { id: 'shu', label: '蜀' },
  { id: 'wei', label: '魏' },
  { id: 'wu', label: '吴' },
  { id: 'qun', label: '群' },
]

async function fetchPage(pageNum: number, append: boolean) {
  const res = await fetchYuzhoushaHeroes({
    mode: gameMode.value,
    kingdom: kingdomFilter.value || undefined,
    page: pageNum,
    page_size: PAGE_SIZE,
  })
  heroes.value = append ? [...heroes.value, ...res.heroes] : res.heroes
  total.value = res.total
  page.value = res.page
  totalPages.value = res.total_pages
  if (heroes.value.length > 0 && !heroes.value.some((h) => h.id === selectedId.value)) {
    selectedId.value = heroes.value[0].id
  }
}

async function loadHeroes() {
  loading.value = true
  page.value = 1
  try {
    await fetchPage(1, false)
  } catch (err) {
    showToast(err instanceof Error ? err.message : '加载武将失败')
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    await fetchPage(page.value + 1, true)
  } catch (err) {
    showToast(err instanceof Error ? err.message : '加载更多失败')
  } finally {
    loadingMore.value = false
  }
}

onMounted(() => {
  void loadHeroes()
})

watch([gameMode, kingdomFilter], () => {
  selectedId.value = ''
  void loadHeroes()
})

function kingdomLabel(k?: string) {
  if (!k) return ''
  return YZS_KINGDOM_LABELS[k] ?? k
}

function heroAccent(hero: YzsCharacter) {
  return heroAccentColor(hero)
}

async function confirmPick() {
  if (!selectedId.value || starting.value) return
  starting.value = true
  try {
    const state = await startYuzhoushaGame(selectedId.value, gameMode.value)
    await router.push({ name: 'yuzhousha-play', params: { gameId: state.id } })
  } catch (err) {
    showToast(err instanceof Error ? err.message : '开局失败')
  } finally {
    starting.value = false
  }
}
</script>

<template>
  <main class="app">
    <section class="hero">
      <div class="hero__top">
        <div>
          <h1>选择武将</h1>
          <p class="hero__desc">
            {{ pickSubtitle }}
            <span v-if="total > 0" class="yzs-pick__count">
              （已加载 {{ heroes.length }}/{{ total }} 名）
            </span>
          </p>
        </div>
        <button type="button" class="hero__logout" @click="router.push('/games/yuzhousha')">← 返回</button>
      </div>
    </section>

    <div v-if="!loading" class="yzs-pick__filters">
      <button
        v-for="opt in kingdomOptions"
        :key="opt.id || 'all'"
        type="button"
        class="yzs-pick__filter"
        :class="{ 'yzs-pick__filter--active': kingdomFilter === opt.id }"
        @click="kingdomFilter = opt.id"
      >
        {{ opt.label }}
      </button>
    </div>

    <p v-if="loading" class="ddz__loading">加载武将…</p>

    <section v-else class="yzs-pick">
      <div class="yzs-pick__grid">
        <button
          v-for="hero in heroes"
          :key="hero.id"
          type="button"
          class="yzs-pick__card"
          :class="{ 'yzs-pick__card--selected': selectedId === hero.id }"
          :style="{ '--hero-accent': heroAccent(hero) }"
          @click="selectedId = hero.id"
          @dblclick="selectedId = hero.id; confirmPick()"
        >
          <span class="yzs-pick__kingdom">{{ kingdomLabel(hero.kingdom) }}</span>
          <h2 class="yzs-pick__name">{{ hero.name }}</h2>
          <p class="yzs-pick__hp">体力 {{ hero.max_hp }}</p>
          <ul class="yzs-pick__skills">
            <li v-for="skill in hero.skills" :key="skill.id" :class="{ 'yzs-pick__skill--inactive': skillBlockedInMode(skill, gameMode) }">
              <strong>{{ skill.name }}<span v-if="skillBlockedInMode(skill, gameMode)" class="yzs-pick__skill-tag">1v1不可用</span></strong>
              <span>{{ skill.desc }}</span>
            </li>
          </ul>
        </button>
      </div>

      <div v-if="hasMore" class="yzs-pick__more-wrap">
        <button
          type="button"
          class="ddz__btn yzs-pick__more"
          :disabled="loadingMore"
          @click="loadMore"
        >
          {{ loadingMore ? '加载中…' : `加载更多（${heroes.length}/${total}）` }}
        </button>
      </div>

      <button
        type="button"
        class="ddz__btn ddz__btn--primary yzs-pick__start"
        :disabled="!selectedId || starting"
        @click="confirmPick"
      >
        {{ starting ? '开局中…' : '确认选将' }}
      </button>
    </section>
  </main>
</template>

<style scoped>
.yzs-pick__count {
  color: var(--color-text-secondary);
  font-size: 13px;
}

.yzs-pick__filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-width: 960px;
  margin: 0 auto 16px;
  padding: 0 16px;
}

.yzs-pick__filter {
  padding: 6px 14px;
  border-radius: 999px;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  font-size: 13px;
  cursor: pointer;
}

.yzs-pick__filter--active {
  border-color: var(--color-primary, #3b82f6);
  background: color-mix(in srgb, var(--color-primary, #3b82f6) 12%, transparent);
}

.yzs-pick {
  max-width: 960px;
  margin: 0 auto;
  padding: 0 16px 48px;
}

.yzs-pick__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.yzs-pick__more-wrap {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.yzs-pick__more {
  min-width: 200px;
}

.yzs-pick__card {
  text-align: left;
  padding: 20px;
  border-radius: 14px;
  border: 2px solid var(--color-border);
  background: var(--color-surface);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.yzs-pick__card--selected {
  border-color: var(--hero-accent);
  box-shadow: 0 0 0 1px var(--hero-accent);
}

.yzs-pick__kingdom {
  display: inline-block;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--hero-accent) 18%, transparent);
  color: var(--hero-accent);
  margin-bottom: 8px;
}

.yzs-pick__name {
  margin: 0 0 4px;
  font-size: 22px;
}

.yzs-pick__hp {
  margin: 0 0 12px;
  color: var(--color-text-secondary);
  font-size: 13px;
}

.yzs-pick__skills {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.yzs-pick__skills li {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 13px;
  line-height: 1.45;
  color: var(--color-text-secondary);
}

.yzs-pick__skills strong {
  color: var(--color-text);
  font-size: 14px;
}

.yzs-pick__skill--inactive {
  opacity: 0.72;
}

.yzs-pick__skill-tag {
  margin-left: 6px;
  font-size: 11px;
  font-weight: normal;
  color: var(--color-text-secondary);
}

.yzs-pick__start {
  display: block;
  margin: 0 auto;
  min-width: 200px;
}
</style>
