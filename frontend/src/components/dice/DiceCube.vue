<script setup lang="ts">
import { computed } from 'vue'

/** 各点数朝前时立方体的旋转（度） */
const FACE_ROTATIONS: Record<number, { x: number; y: number }> = {
  1: { x: 0, y: 0 },
  2: { x: 0, y: -90 },
  3: { x: -90, y: 0 },
  4: { x: 90, y: 0 },
  5: { x: 0, y: 90 },
  6: { x: 0, y: 180 },
}

const props = withDefaults(
  defineProps<{
    value?: number
    rolling?: boolean
    rotation?: { x: number; y: number; z: number }
    size?: number
  }>(),
  {
    value: 1,
    rolling: false,
    rotation: () => ({ x: 0, y: 0, z: 0 }),
    size: 64,
  },
)

const pipLayouts: Record<number, [number, number][]> = {
  1: [[2, 2]],
  2: [
    [1, 1],
    [3, 3],
  ],
  3: [
    [1, 1],
    [2, 2],
    [3, 3],
  ],
  4: [
    [1, 1],
    [3, 1],
    [1, 3],
    [3, 3],
  ],
  5: [
    [1, 1],
    [3, 1],
    [2, 2],
    [1, 3],
    [3, 3],
  ],
  6: [
    [1, 1],
    [1, 2],
    [1, 3],
    [3, 1],
    [3, 2],
    [3, 3],
  ],
}

const half = computed(() => props.size / 2)

/** 静止且 rotation 仍为初始值时，才按 value 推算朝向；掷完后保持 rotation 避免二次翻转 */
const effectiveRotation = computed(() => {
  if (props.rolling) return props.rotation

  const { x, y, z } = props.rotation
  if (x !== 0 || y !== 0 || z !== 0) return props.rotation

  const face = FACE_ROTATIONS[clampFace(props.value)] ?? FACE_ROTATIONS[1]
  return { x: face.x, y: face.y, z: 0 }
})

const cubeStyle = computed(() => {
  const r = effectiveRotation.value
  return {
    '--dice-size': `${props.size}px`,
    '--dice-half': `${half.value}px`,
    transform: `rotateX(${r.x}deg) rotateY(${r.y}deg) rotateZ(${r.z}deg)`,
  }
})

function clampFace(n: number) {
  return Math.min(6, Math.max(1, Math.round(n)))
}

const faces = [1, 2, 3, 4, 5, 6] as const
</script>

<template>
  <div class="dice-scene" :style="{ width: `${size}px`, height: `${size}px` }">
    <div
      class="dice-cube"
      :class="{ 'dice-cube--rolling': rolling }"
      :style="cubeStyle"
    >
      <div
        v-for="face in faces"
        :key="face"
        class="dice-face"
        :class="`dice-face--${face}`"
      >
        <span
          v-for="(pip, i) in pipLayouts[face]"
          :key="i"
          class="dice-pip"
          :style="{ gridRow: pip[0], gridColumn: pip[1] }"
        />
      </div>
    </div>
  </div>
</template>
