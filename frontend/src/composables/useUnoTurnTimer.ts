import { onUnmounted, ref, watch, type Ref } from 'vue'
import { tickUnoGame } from '../api/games'
import type { UnoState } from '../types/uno'

const DEFAULT_SECONDS = 20

export function useUnoTurnTimer(
  state: Ref<UnoState | null>,
  isMyTurn: Ref<boolean>,
  onTimeout: () => Promise<void>,
) {
  const secondsLeft = ref(DEFAULT_SECONDS)
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
      secondsLeft.value = DEFAULT_SECONDS
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
    () => [state.value?.turn_deadline_unix, state.value?.current_turn, state.value?.phase],
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

export async function defaultUnoTimeoutHandler(
  gameId: string,
  applyState: (s: UnoState) => Promise<void>,
) {
  const next = await tickUnoGame(gameId)
  await applyState(next)
}
