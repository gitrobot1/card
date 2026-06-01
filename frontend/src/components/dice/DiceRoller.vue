<script setup lang="ts">
import { computed } from 'vue'
import DiceCube from './DiceCube.vue'
import type { DiceRotation } from '../../composables/useDiceRoll'

const props = withDefaults(
  defineProps<{
    visible?: boolean
    rolling?: boolean
    value?: number
    rotation?: DiceRotation
    /** 屏幕上的锚点位置 */
    placement?: 'bottom-right' | 'bottom-left' | 'center' | 'top-right'
    size?: number
    showLabel?: boolean
  }>(),
  {
    visible: false,
    rolling: false,
    value: 1,
    rotation: () => ({ x: 0, y: 0, z: 0 }),
    placement: 'bottom-right',
    size: 64,
    showLabel: true,
  },
)

const emit = defineEmits<{
  click: []
}>()

const labelText = computed(() => {
  if (props.rolling) return '掷骰中…'
  return `点数 ${props.value}`
})
</script>

<template>
  <Teleport to="body">
    <Transition name="dice-roller-fade">
      <div
        v-if="visible"
        class="dice-roller"
        :class="[
          `dice-roller--${placement}`,
          { 'dice-roller--rolling': rolling },
        ]"
        @click="emit('click')"
      >
        <div class="dice-roller__inner">
          <DiceCube
            :value="value"
            :rolling="rolling"
            :rotation="rotation"
            :size="size"
          />
          <p v-if="showLabel" class="dice-roller__label">{{ labelText }}</p>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
