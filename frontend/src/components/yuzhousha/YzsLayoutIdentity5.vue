<script setup lang="ts">
import YzsSeatPanel from './YzsSeatPanel.vue'
import YzsArenaCenter from './YzsArenaCenter.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

const { crossSeats } = useYzsGameInject()
</script>

<template>
  <div class="yzs__arena yzs__arena--identity5">
    <YzsSeatPanel
      v-for="cs in crossSeats"
      :key="cs.seat"
      :seat="cs.seat"
      :placement="cs.placement"
      :stack-class="`yzs__pent-seat yzs__pent-seat--${cs.area}`"
    />
    <YzsArenaCenter center-class="yzs__middle--identity5-center" />
  </div>
</template>
