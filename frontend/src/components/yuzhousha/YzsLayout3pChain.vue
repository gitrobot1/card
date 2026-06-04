<script setup lang="ts">
import YzsSeatPanel from './YzsSeatPanel.vue'
import YzsArenaCenter from './YzsArenaCenter.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

const { crossSeats } = useYzsGameInject()
</script>

<template>
  <div class="yzs__arena yzs__arena--3p">
    <YzsSeatPanel
      v-for="cs in crossSeats"
      :key="cs.seat"
      :seat="cs.seat"
      :placement="cs.placement"
      :is-teammate="cs.isTeammate"
      :seat-role="cs.seatRole"
      :stack-class="`yzs__chain-seat yzs__chain-seat--${cs.area}`"
    />
    <YzsArenaCenter center-class="yzs__middle--3p-center" />
  </div>
</template>
