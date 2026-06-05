<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { fetchYuzhoushaHeroes, fetchYuzhoushaPacks } from '../../api/games'
import { showToast } from '../../composables/useToast'
import { heroAccentColor } from '../../composables/yuzhousha/resolveYzsHeroDisplay'
import {
  YZS_BASIC_CARDS,
  YZS_TRICK_CARDS,
  YZS_WEAPON_CARDS,
  YZS_PACK_LABELS,
  type CodexCardEntry,
} from '../../constants/yzsCardCatalog'
import type { YzsCharacter } from '../../types/yuzhousha'
import { YZS_KINGDOM_LABELS } from '../../types/yuzhousha'

type CodexTab = 'heroes' | 'basic' | 'weapons' | 'tricks'

const router = useRouter()
const activeTab = ref<CodexTab>('heroes')
const searchQuery = ref('')
const kingdomFilter = ref('')

const heroes = ref<YzsCharacter[]>([])
const heroesLoading = ref(false)
const heroCount = ref(0)
const packLabels = ref<Record<string, string>>({ ...YZS_PACK_LABELS })

const tabs = computed(() => [
  { id: 'heroes' as CodexTab, label: '武将库', count: heroCount.value },
  { id: 'basic' as CodexTab, label: '基础牌库', count: YZS_BASIC_CARDS.length },
  { id: 'weapons' as CodexTab, label: '武器库', count: YZS_WEAPON_CARDS.length },
  { id: 'tricks' as CodexTab, label: '锦囊库', count: YZS_TRICK_CARDS.length },
])

const kingdomOptions = [
  { id: '', label: '全部势力' },
  { id: 'shu', label: '蜀' },
  { id: 'wei', label: '魏' },
  { id: 'wu', label: '吴' },
  { id: 'qun', label: '群' },
]

async function loadHeroes() {
  heroesLoading.value = true
  try {
    const res = await fetchYuzhoushaHeroes({ page: 1, page_size: 100 })
    heroes.value = res.heroes
    heroCount.value = res.total
  } catch (err) {
    showToast(err instanceof Error ? err.message : '加载武将失败')
  } finally {
    heroesLoading.value = false
  }
}

async function loadPacks() {
  try {
    const res = await fetchYuzhoushaPacks()
    for (const p of res.packs) {
      packLabels.value[p.hero_pack || p.id] = p.name
    }
  } catch {
    /* optional */
  }
}

onMounted(() => {
  void loadHeroes()
  void loadPacks()
})

watch(kingdomFilter, () => {
  void reloadHeroesFiltered()
})

async function reloadHeroesFiltered() {
  heroesLoading.value = true
  try {
    const res = await fetchYuzhoushaHeroes({
      page: 1,
      page_size: 100,
      kingdom: kingdomFilter.value || undefined,
    })
    heroes.value = res.heroes
    heroCount.value = res.total
  } catch (err) {
    showToast(err instanceof Error ? err.message : '加载武将失败')
  } finally {
    heroesLoading.value = false
  }
}

const normalizedQuery = computed(() => searchQuery.value.trim().toLowerCase())

function matchesQuery(text: string) {
  if (!normalizedQuery.value) return true
  return text.toLowerCase().includes(normalizedQuery.value)
}

const filteredHeroes = computed(() =>
  heroes.value.filter((h) => {
    const blob = [h.name, h.kingdom ?? '', ...(h.skills?.map((s) => `${s.name} ${s.desc}`) ?? [])].join(' ')
    return matchesQuery(blob)
  }),
)

function filterCards(list: CodexCardEntry[]) {
  return list.filter((c) => matchesQuery(`${c.name} ${c.effect} ${c.subtype ?? ''} ${c.tag ?? ''}`))
}

const filteredBasic = computed(() => filterCards(YZS_BASIC_CARDS))
const filteredWeapons = computed(() => filterCards(YZS_WEAPON_CARDS))
const filteredTricks = computed(() => filterCards(YZS_TRICK_CARDS))

function kingdomLabel(k?: string) {
  if (!k) return ''
  return YZS_KINGDOM_LABELS[k] ?? k
}

function packLabel(p?: string) {
  if (!p) return '标准包'
  return packLabels.value[p] ?? p
}

function heroAccent(hero: YzsCharacter) {
  return heroAccentColor(hero)
}

function skillKindLabel(kind: string) {
  switch (kind) {
    case 'lord':
      return '主公技'
    case 'active':
      return '主动技'
    case 'awakening':
      return '觉醒技'
    default:
      return '锁定技'
  }
}
</script>

<template>
  <main class="app yzs-codex">
    <section class="hero">
      <div class="hero__top">
        <div>
          <h1>宇宙杀 · 图鉴</h1>
          <p class="hero__desc">武将、基础牌、武器与锦囊效果一览</p>
        </div>
        <button type="button" class="hero__logout" @click="router.push('/games/yuzhousha')">← 返回</button>
      </div>
    </section>

    <div class="yzs-codex__toolbar">
      <nav class="yzs-codex__tabs" aria-label="图鉴分类">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="yzs-codex__tab"
          :class="{ 'yzs-codex__tab--active': activeTab === tab.id }"
          @click="activeTab = tab.id"
        >
          {{ tab.label }}
          <span class="yzs-codex__tab-count">{{ tab.count }}</span>
        </button>
      </nav>

      <input
        v-model="searchQuery"
        type="search"
        class="yzs-codex__search"
        placeholder="搜索名称或效果…"
        aria-label="搜索图鉴"
      />
    </div>

    <div v-if="activeTab === 'heroes'" class="yzs-codex__filters">
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

    <p v-if="activeTab === 'heroes' && heroesLoading" class="ddz__loading">加载武将…</p>

    <!-- 武将库 -->
    <section v-else-if="activeTab === 'heroes'" class="yzs-codex__section">
      <p v-if="filteredHeroes.length === 0" class="yzs-codex__empty">没有匹配的武将</p>
      <div v-else class="yzs-codex__grid yzs-codex__grid--heroes">
        <article
          v-for="hero in filteredHeroes"
          :key="hero.id"
          class="yzs-codex__card yzs-codex__card--hero"
          :style="{ '--hero-accent': heroAccent(hero) }"
        >
          <header class="yzs-codex__card-head">
            <div>
              <h2 class="yzs-codex__name">{{ hero.name }}</h2>
              <p class="yzs-codex__meta">
                <span>{{ kingdomLabel(hero.kingdom) }}</span>
                <span>体力 {{ hero.max_hp }}</span>
                <span>{{ packLabel(hero.pack) }}</span>
              </p>
            </div>
          </header>
          <ul v-if="hero.skills?.length" class="yzs-codex__skills">
            <li v-for="skill in hero.skills" :key="skill.id">
              <div class="yzs-codex__skill-head">
                <strong>{{ skill.name }}</strong>
                <span class="yzs-codex__skill-kind">{{ skillKindLabel(skill.kind) }}</span>
              </div>
              <p class="yzs-codex__effect">{{ skill.desc }}</p>
            </li>
          </ul>
        </article>
      </div>
    </section>

    <!-- 牌库（基础 / 武器 / 锦囊） -->
    <section v-else class="yzs-codex__section">
      <p v-if="activeTab === 'basic' && filteredBasic.length === 0" class="yzs-codex__empty">没有匹配的牌</p>
      <p v-else-if="activeTab === 'weapons' && filteredWeapons.length === 0" class="yzs-codex__empty">没有匹配的牌</p>
      <p v-else-if="activeTab === 'tricks' && filteredTricks.length === 0" class="yzs-codex__empty">没有匹配的牌</p>

      <div v-else class="yzs-codex__grid">
        <article
          v-for="card in activeTab === 'basic'
            ? filteredBasic
            : activeTab === 'weapons'
              ? filteredWeapons
              : filteredTricks"
          :key="card.kind"
          class="yzs-codex__card"
        >
          <header class="yzs-codex__card-head">
            <h2 class="yzs-codex__name">{{ card.name }}</h2>
            <div class="yzs-codex__badges">
              <span v-if="card.subtype" class="yzs-codex__badge">{{ card.subtype }}</span>
              <span v-if="card.range" class="yzs-codex__badge">距离 {{ card.range }}</span>
              <span v-if="card.tag" class="yzs-codex__badge yzs-codex__badge--tag">{{ card.tag }}</span>
            </div>
          </header>
          <p class="yzs-codex__effect">{{ card.effect }}</p>
        </article>
      </div>
    </section>
  </main>
</template>

<style scoped>
.yzs-codex__toolbar {
  max-width: 1080px;
  margin: 0 auto 16px;
  padding: 0 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
}

.yzs-codex__tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.yzs-codex__tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  font-size: 14px;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.yzs-codex__tab--active {
  border-color: var(--color-accent);
  background: var(--color-accent-soft);
  color: var(--color-accent);
  font-weight: 600;
}

.yzs-codex__tab-count {
  font-size: 12px;
  opacity: 0.75;
}

.yzs-codex__search {
  flex: 1 1 200px;
  max-width: 280px;
  padding: 8px 14px;
  border-radius: 10px;
  border: 1px solid var(--color-border);
  font-size: 14px;
  background: var(--color-surface);
}

.yzs-codex__filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-width: 1080px;
  margin: 0 auto 16px;
  padding: 0 16px;
}

.yzs-codex__section {
  max-width: 1080px;
  margin: 0 auto;
  padding: 0 16px 48px;
}

.yzs-codex__empty {
  text-align: center;
  color: var(--color-text-secondary);
  padding: 48px 16px;
}

.yzs-codex__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.yzs-codex__grid--heroes {
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
}

.yzs-codex__card {
  padding: 16px 18px;
  border-radius: 14px;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  box-shadow: var(--shadow-sm);
}

.yzs-codex__card--hero {
  border-left: 4px solid var(--hero-accent, var(--color-accent));
}

.yzs-codex__card-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.yzs-codex__name {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
}

.yzs-codex__meta {
  margin: 4px 0 0;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.yzs-codex__meta span:not(:last-child)::after {
  content: '·';
  margin-left: 8px;
  opacity: 0.5;
}

.yzs-codex__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.yzs-codex__badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--color-surface-muted);
  color: var(--color-text-secondary);
}

.yzs-codex__badge--tag {
  background: #fef3c7;
  color: #92400e;
}

.yzs-codex__effect {
  margin: 0;
  font-size: 14px;
  line-height: 1.65;
  color: var(--color-text-secondary);
}

.yzs-codex__skills {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.yzs-codex__skills li {
  padding-top: 10px;
  border-top: 1px solid var(--color-border-light);
}

.yzs-codex__skills li:first-child {
  padding-top: 0;
  border-top: none;
}

.yzs-codex__skill-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.yzs-codex__skill-head strong {
  font-size: 14px;
}

.yzs-codex__skill-kind {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--color-accent-soft);
  color: var(--color-accent);
}
</style>
