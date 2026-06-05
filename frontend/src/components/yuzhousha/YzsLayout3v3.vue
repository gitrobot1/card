<script setup lang="ts">
import YzsSeatPanel from './YzsSeatPanel.vue'
import YzsArenaCenter from './YzsArenaCenter.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

const { crossSeats } = useYzsGameInject()
</script>

<template>
  <div class="yzs__arena yzs__arena--3v3">
    <YzsSeatPanel
      v-for="cs in crossSeats"
      :key="cs.seat"
      :seat="cs.seat"
      :placement="cs.placement"
      :is-teammate="cs.isTeammate"
      :seat-role="cs.seatRole"
      :stack-class="`yzs__hex-seat yzs__hex-seat--${cs.area}`"
    />
    <YzsArenaCenter center-class="yzs__middle--3v3-center" />
  </div>
</template>
