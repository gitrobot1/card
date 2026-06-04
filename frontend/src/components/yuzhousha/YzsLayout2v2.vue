<script setup lang="ts">
import YzsSeatPanel from './YzsSeatPanel.vue'
import YzsArenaCenter from './YzsArenaCenter.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

const { crossSeats } = useYzsGameInject()
</script>

<template>
  <div class="yzs__arena yzs__arena--2v2">
    <YzsSeatPanel
      v-for="cs in crossSeats"
      :key="cs.seat"
      :seat="cs.seat"
      :placement="cs.placement"
      :is-teammate="cs.isTeammate"
      :stack-class="`yzs__cross-seat yzs__cross-seat--${cs.area}`"
    />
    <YzsArenaCenter center-class="yzs__middle--2v2-center" />
  </div>
</template>
