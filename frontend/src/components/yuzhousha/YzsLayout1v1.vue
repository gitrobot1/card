<script setup lang="ts">
import YzsSeatPanel from './YzsSeatPanel.vue'
import YzsArenaCenter from './YzsArenaCenter.vue'
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

const { opponentSeat } = useYzsGameInject()
</script>

<template>
  <div class="yzs__arena">
    <div class="yzs__opponent-row">
      <YzsSeatPanel :seat="opponentSeat" placement="top" />
    </div>
    <YzsArenaCenter />
  </div>
</template>
