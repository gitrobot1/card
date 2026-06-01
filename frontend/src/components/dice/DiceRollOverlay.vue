<script setup lang="ts">
import DiceCube from './DiceCube.vue'
import type { DiceRotation } from '../../composables/useDiceRoll'

defineProps<{
  visible?: boolean
  rolling?: boolean
  playerName?: string
  dice1Value?: number
  dice2Value?: number
  dice1Rolling?: boolean
  dice2Rolling?: boolean
  dice1Rotation?: DiceRotation
  dice2Rotation?: DiceRotation
  sum?: number | null
  size?: number
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="dice-overlay-fade">
      <div v-if="visible" class="dice-roll-overlay">
        <p class="dice-roll-overlay__title">
          <template v-if="rolling">{{ playerName || '玩家' }} 掷骰中…</template>
          <template v-else-if="sum != null">{{ playerName || '玩家' }} · {{ sum }} 点</template>
          <template v-else>掷骰定先手</template>
        </p>
        <div class="dice-roll-overlay__pair">
          <DiceCube
            :value="dice1Value ?? 1"
            :rolling="dice1Rolling ?? false"
            :rotation="dice1Rotation ?? { x: 0, y: 0, z: 0 }"
            :size="size ?? 96"
          />
          <DiceCube
            :value="dice2Value ?? 1"
            :rolling="dice2Rolling ?? false"
            :rotation="dice2Rotation ?? { x: 0, y: 0, z: 0 }"
            :size="size ?? 96"
          />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
