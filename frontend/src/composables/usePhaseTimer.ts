import { onUnmounted, ref, watch, type Ref } from 'vue'

const DEFAULT_SECONDS = 20

export function usePhaseTimer(
  deadlineUnix: Ref<number | undefined>,
  phase: Ref<string | undefined>,
  active: Ref<boolean>,
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
    const deadline = deadlineUnix.value ?? 0
    if (!deadline) {
      secondsLeft.value = DEFAULT_SECONDS
      return
    }
    const left = Math.max(0, Math.ceil(deadline - Date.now() / 1000))
    secondsLeft.value = left

    if (left <= 0 && active.value && !timeoutTriggered && phase.value !== 'finished') {
      timeoutTriggered = true
      void onTimeout()
    }
  }

  function startTimer() {
    clearTimer()
    updateSeconds()
    timerId = setInterval(updateSeconds, 200)
  }

  // 仅在阶段/截止时间变化时重置触发标记，避免 loading 导致 active 闪烁后重复超时
  watch(
    () => [deadlineUnix.value, phase.value] as const,
    () => {
      timeoutTriggered = false
      startTimer()
    },
    { immediate: true },
  )

  watch(active, () => {
    updateSeconds()
  })

  onUnmounted(clearTimer)

  return { secondsLeft, clearTimer }
}
