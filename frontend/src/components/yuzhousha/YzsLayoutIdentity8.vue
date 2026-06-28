<script setup lang="ts">
import YzsSeatPanel from './YzsSeatPanel.vue'
import YzsArenaCenter from './YzsArenaCenter.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

const { crossSeats } = useYzsGameInject()
</script>

<template>
  <div class="yzs__arena yzs__arena--identity8">
    <!-- 占位槽 -->
    <div class="yzs__slot-ghost" style="grid-area: s1" />
    <div class="yzs__slot-ghost" style="grid-area: s2" />
    <div class="yzs__slot-ghost" style="grid-area: s8" />
    <div class="yzs__slot-ghost" style="grid-area: s9" />
    <div class="yzs__slot-ghost" style="grid-area: s11" />
    <div class="yzs__slot-ghost" style="grid-area: s19" />

    <YzsSeatPanel
      v-for="cs in crossSeats"
      :key="cs.seat"
      :seat="cs.seat"
      :placement="cs.placement"
      :stack-class="`yzs__oct-seat yzs__oct-seat--${cs.area}`"
    />
    <YzsArenaCenter center-class="yzs__middle--identity8-center" />
  </div>
</template>
