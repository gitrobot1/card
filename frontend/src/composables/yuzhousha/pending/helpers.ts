import type { YuzhoushaState } from '../../../types/yuzhousha'
import type { PendingContext } from './context'

export function isBusy(ctx: PendingContext) {
  return ctx.loading || ctx.isAnimating
}

export function responseMode(state: YuzhoushaState | null | undefined, mode: string) {
  return state?.phase === 'response' && state.pending?.response_mode === mode
}

export function responseAnyMode(state: YuzhoushaState | null | undefined, modes: string[]) {
  const mode = state?.pending?.response_mode
  return state?.phase === 'response' && !!mode && modes.includes(mode)
}

export function pickFirstTarget(
  ctx: PendingContext,
  options: { zone: string; cardId: string }[],
) {
  const first = options[0]
  if (first) {
    ctx.selectedTargetZone.value = first.zone
    ctx.selectedTargetCardId.value = first.cardId
  }
}
