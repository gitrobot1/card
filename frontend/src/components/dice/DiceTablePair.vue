<script setup lang="ts">
import DiceCube from './DiceCube.vue'
import type { DiceRotation } from '../../composables/useDiceRoll'

defineProps<{
  visible?: boolean
  rolling?: boolean
  dice1Value?: number
  dice2Value?: number
  dice1Rolling?: boolean
  dice2Rolling?: boolean
  dice1Rotation?: DiceRotation
  dice2Rotation?: DiceRotation
  size?: number
}>()
</script>

<template>
  <div v-if="visible" class="uno__table-dice">
    <div class="uno__table-dice__pair">
      <DiceCube
        :value="dice1Value ?? 1"
        :rolling="dice1Rolling ?? false"
        :rotation="dice1Rotation ?? { x: 0, y: 0, z: 0 }"
        :size="size ?? 72"
      />
      <DiceCube
        :value="dice2Value ?? 1"
        :rolling="dice2Rolling ?? false"
        :rotation="dice2Rotation ?? { x: 0, y: 0, z: 0 }"
        :size="size ?? 72"
      />
    </div>
  </div>
</template>
