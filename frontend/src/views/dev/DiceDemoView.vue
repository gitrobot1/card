<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import DiceCube from '../../components/dice/DiceCube.vue'
import DiceRoller from '../../components/dice/DiceRoller.vue'
import { useDiceRoll } from '../../composables/useDiceRoll'

const router = useRouter()
const { visible, rolling, value, rotation, roll, hide } = useDiceRoll()
const lastResult = ref<number | null>(null)
const busy = ref(false)

async function rollRandom() {
  if (busy.value) return
  busy.value = true
  lastResult.value = await roll()
  busy.value = false
}

async function rollFixed(n: number) {
  if (busy.value) return
  busy.value = true
  lastResult.value = await roll({ value: n })
  busy.value = false
}
</script>

<template>
  <main class="dice-demo">
    <button class="dice-demo__back" type="button" @click="router.push('/')">← 返回大厅</button>

    <h1 class="dice-demo__title">3D 掷骰子</h1>
    <p class="dice-demo__hint">点击右下角浮层或下方按钮掷骰；可指定点数预览各面朝向。</p>

    <DiceCube
      :value="value"
      :rolling="rolling"
      :rotation="rotation"
      :size="96"
    />

    <p v-if="lastResult != null" class="dice-demo__hint">上次结果：{{ lastResult }}</p>

    <div class="dice-demo__actions">
      <button class="dice-demo__btn" type="button" :disabled="busy" @click="rollRandom">
        随机掷骰
      </button>
      <button
        v-for="n in 6"
        :key="n"
        class="dice-demo__btn dice-demo__btn--ghost"
        type="button"
        :disabled="busy"
        @click="rollFixed(n)"
      >
        掷出 {{ n }}
      </button>
      <button
        class="dice-demo__btn dice-demo__btn--ghost"
        type="button"
        :disabled="!visible"
        @click="hide()"
      >
        隐藏浮层
      </button>
    </div>

    <DiceRoller
      :visible="visible"
      :rolling="rolling"
      :value="value"
      :rotation="rotation"
      placement="bottom-right"
      :size="72"
      @click="rollRandom"
    />
  </main>
</template>
