import { onUnmounted, ref, watch, type Ref } from 'vue'
import { tickDouDizhuGame } from '../api/games'
import type { DouDizhuState } from '../types/doudizhu'

export function useTurnTimer(
  state: Ref<DouDizhuState | null>,
  isMyTurn: Ref<boolean>,
  onTimeout: () => Promise<void>,
) {
  const secondsLeft = ref(35)
  let timerId: ReturnType<typeof setInterval> | null = null
  let timeoutTriggered = false

  function clearTimer() {
    if (timerId) {
      clearInterval(timerId)
      timerId = null
    }
  }

  function updateSeconds() {
    const deadline = state.value?.turn_deadline_unix ?? 0
    if (!deadline) {
      secondsLeft.value = 35
      return
    }
    const left = Math.max(0, Math.ceil(deadline - Date.now() / 1000))
    secondsLeft.value = left

    if (left <= 0 && isMyTurn.value && !timeoutTriggered && state.value?.phase !== 'finished') {
      timeoutTriggered = true
      void onTimeout()
    }
  }

  watch(
    () => [state.value?.turn_deadline_unix, state.value?.current_turn, state.value?.calling_index, state.value?.phase],
    () => {
      timeoutTriggered = false
      clearTimer()
      updateSeconds()
      timerId = setInterval(updateSeconds, 200)
    },
    { immediate: true },
  )

  onUnmounted(clearTimer)

  return { secondsLeft, clearTimer }
}

export async function defaultTimeoutHandler(gameId: string, applyState: (s: DouDizhuState) => Promise<void>) {
  const next = await tickDouDizhuGame(gameId)
  await applyState(next)
}
